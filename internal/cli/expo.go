package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
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
		host        string
		devicePort  int
		metroPort   int
		timeout     time.Duration
		skipConnect bool
	)

	cmd := &cobra.Command{
		Use:   "expo",
		Short: "Connect a device, configure Metro reverse, and start Expo",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExpo(cmd.Context(), expoOptions{
				ManualHost:  host,
				ManualPort:  devicePort,
				MetroPort:   metroPort,
				Timeout:     timeout,
				SkipConnect: skipConnect,
				ExtraArgs:   args,
			})
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "manual device host (skips discovery, requires --device-port)")
	cmd.Flags().IntVar(&devicePort, "device-port", 0, "manual device port (skips discovery, requires --host)")
	cmd.Flags().IntVar(&metroPort, "port", cfg.DefaultMetroPort, "Metro port to reverse and pass to Expo")
	cmd.Flags().DurationVar(&timeout, "timeout", cfg.DefaultDiscoveryTimeout, "discovery timeout")
	cmd.Flags().BoolVar(&skipConnect, "skip-connect", false, "assume a device is already connected")

	return cmd
}

type expoOptions struct {
	ManualHost  string
	ManualPort  int
	MetroPort   int
	Timeout     time.Duration
	SkipConnect bool
	ExtraArgs   []string
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
