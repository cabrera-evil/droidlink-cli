package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"droidlink/internal/adb"
	"droidlink/internal/config"
	"droidlink/internal/logging"
)

var errNpxNotFound = errors.New("expo: npx binary not found on PATH")

func newExpoCmd(cfg config.Config) *cobra.Command {
	var (
		host             string
		devicePort       int
		metroPort        int
		timeout          time.Duration
		skipConnect      bool
		skipInstallCheck bool
	)

	cmd := &cobra.Command{
		Use:   "expo",
		Short: "Connect a device, configure Metro reverse, and start Expo",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExpo(cmd.Context(), expoOptions{
				ManualHost:       host,
				ManualPort:       devicePort,
				MetroPort:        metroPort,
				Timeout:          timeout,
				SkipConnect:      skipConnect,
				SkipInstallCheck: skipInstallCheck,
				ExtraArgs:        args,
			})
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "manual device host (skips discovery, requires --device-port)")
	cmd.Flags().IntVar(&devicePort, "device-port", 0, "manual device port (skips discovery, requires --host)")
	cmd.Flags().IntVar(&metroPort, "port", cfg.DefaultMetroPort, "Metro port to reverse and pass to Expo")
	cmd.Flags().DurationVar(&timeout, "timeout", cfg.DefaultDiscoveryTimeout, "discovery timeout")
	cmd.Flags().BoolVar(&skipConnect, "skip-connect", false, "assume a device is already connected")
	cmd.Flags().BoolVar(&skipInstallCheck, "skip-install-check", false, "assume the dev client is already installed")

	return cmd
}

type expoOptions struct {
	ManualHost       string
	ManualPort       int
	MetroPort        int
	Timeout          time.Duration
	SkipConnect      bool
	SkipInstallCheck bool
	ExtraArgs        []string
}

func runExpo(ctx context.Context, opts expoOptions) error {
	if err := validateManualFlags(opts.ManualHost, opts.ManualPort); err != nil {
		return err
	}

	if _, err := exec.LookPath("npx"); err != nil {
		return fmt.Errorf("%w: install Node.js/npm and ensure npx is on PATH", errNpxNotFound)
	}

	client, err := adb.NewClient()
	if err != nil {
		return err
	}

	serial, err := resolveExpoSerial(ctx, client, opts)
	if err != nil {
		return err
	}

	if !opts.SkipInstallCheck {
		if installErr := ensureDevClientInstalled(ctx, client, serial); installErr != nil {
			return installErr
		}
	}

	if reverseErr := applyReverse(ctx, client, serial, opts.MetroPort); reverseErr != nil {
		return reverseErr
	}

	return runExpoStart(ctx, opts)
}

func resolveExpoSerial(ctx context.Context, client *adb.Client, opts expoOptions) (string, error) {
	if opts.SkipConnect {
		return resolveSerial(ctx, client, "")
	}

	device, err := resolveAndConnect(ctx, client, connectOptions{
		ManualHost: opts.ManualHost,
		ManualPort: opts.ManualPort,
		Timeout:    opts.Timeout,
	})
	if err != nil {
		return "", err
	}

	return device.Serial, nil
}

// ensureDevClientInstalled checks whether the Expo dev client is already
// installed on serial and, if not, builds and installs it via
// `npx expo run:android`.
func ensureDevClientInstalled(ctx context.Context, client *adb.Client, serial string) error {
	pkg, err := androidPackageName(ctx)
	if err != nil {
		return err
	}

	logger := logging.FromContext(ctx)

	installed, err := client.IsPackageInstalled(ctx, serial, pkg)
	if err != nil {
		return err
	}

	if installed {
		logger.InfoContext(ctx, "dev client already installed", "package", pkg, "serial", serial)

		return nil
	}

	logger.InfoContext(ctx, "dev client not installed, building and installing", "package", pkg, "serial", serial)

	// `expo run:android --device` matches by Expo's own device name (the adb
	// `model:` field for network-connected devices, per @expo/cli's
	// AndroidDeviceManager.resolveFromNameAsync), never by adb serial — so
	// the wireless host:port serial must be resolved to that name first.
	deviceName, err := adbDeviceName(ctx, client, serial)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "npx", "expo", "run:android", "--device", deviceName, "--no-bundler")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if runErr := cmd.Run(); runErr != nil {
		return fmt.Errorf("npx expo run:android: %w", runErr)
	}

	return nil
}

// adbDeviceName looks up the adb-reported name (the `model:` field from
// `adb devices -l`, adb.go's Device.Name) for serial, which is what Expo's
// own device resolution matches against for --device, as opposed to the
// adb serial droidlink otherwise uses everywhere else.
func adbDeviceName(ctx context.Context, client *adb.Client, serial string) (string, error) {
	devices, err := client.Devices(ctx)
	if err != nil {
		return "", fmt.Errorf("list adb devices: %w", err)
	}

	for _, d := range devices {
		if d.Serial == serial {
			return d.Name, nil
		}
	}

	return "", fmt.Errorf("no adb device found with serial %q", serial)
}

// androidPackageName resolves the Android application ID that `expo
// run:android` will actually build and install.
//
// If prebuild has already generated a native android/ project,
// `expo run:android` reuses it as-is and does not re-run prebuild, so that
// project's build.gradle applicationId is the ground truth; app.json's
// android.package can drift from it until prebuild is explicitly re-run, so
// it isn't a safe source once a native project exists. Otherwise (no native
// project yet) the app config is evaluated via `npx expo config --json`,
// since expo run:android will prebuild fresh from it.
func androidPackageName(ctx context.Context) (string, error) {
	if pkg, ok, err := androidPackageFromNativeProject(); err != nil {
		return "", err
	} else if ok {
		logging.FromContext(ctx).InfoContext(ctx, "using package from existing native android project", "package", pkg)

		return pkg, nil
	}

	output, err := exec.CommandContext(ctx, "npx", "expo", "config", "--json").Output()
	if err != nil {
		return "", fmt.Errorf("npx expo config --json: %w", err)
	}

	var parsed struct {
		Android struct {
			Package string `json:"package"`
		} `json:"android"`
	}

	if unmarshalErr := json.Unmarshal(output, &parsed); unmarshalErr != nil {
		return "", fmt.Errorf("parse expo config output: %w", unmarshalErr)
	}

	if parsed.Android.Package == "" {
		return "", errors.New("expo config: android.package is not set in app.json/app.config")
	}

	return parsed.Android.Package, nil
}

// applicationIDPattern matches a Gradle applicationId declaration, e.g.
// `applicationId "com.example.app"` or `applicationId 'com.example.app'`
// (Groovy allows either quote style) or `applicationId =
// "com.example.app"` (Kotlin DSL). The mandatory trailing quote keeps it
// from matching `applicationIdSuffix`.
var applicationIDPattern = regexp.MustCompile(`applicationId\s*=?\s*["']([^"']+)["']`)

// androidPackageFromNativeProject reads applicationId from an
// already-prebuilt native android project's build.gradle(.kts), if one
// exists, reporting ok=false (not an error) when prebuild hasn't run yet.
func androidPackageFromNativeProject() (string, bool, error) {
	paths := []string{
		filepath.Join("android", "app", "build.gradle"),
		filepath.Join("android", "app", "build.gradle.kts"),
	}

	for _, path := range paths {
		data, readErr := os.ReadFile(path)
		if errors.Is(readErr, os.ErrNotExist) {
			continue
		}

		if readErr != nil {
			return "", false, fmt.Errorf("read %s: %w", path, readErr)
		}

		match := applicationIDPattern.FindSubmatch(data)
		if match == nil {
			return "", false, fmt.Errorf("%s: applicationId not found", path)
		}

		return string(match[1]), true, nil
	}

	return "", false, nil
}

func runExpoStart(ctx context.Context, opts expoOptions) error {
	args := append(
		[]string{"expo", "start", "--dev-client", "--localhost", "--port", strconv.Itoa(opts.MetroPort)},
		opts.ExtraArgs...,
	)

	cmd := exec.CommandContext(ctx, "npx", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logging.FromContext(ctx).InfoContext(ctx, "starting expo", "port", opts.MetroPort)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npx expo start: %w", err)
	}

	return nil
}
