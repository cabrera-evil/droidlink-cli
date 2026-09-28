package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"droidlink/internal/adb"
	"droidlink/internal/config"
	"droidlink/internal/discovery"
	"droidlink/internal/logging"
	"droidlink/internal/types"
)

func newConnectCmd(cfg config.Config) *cobra.Command {
	var (
		host    string
		port    int
		timeout time.Duration
		all     bool
	)

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Discover paired devices and connect automatically",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runConnect(cmd.Context(), connectOptions{
				ManualHost: host,
				ManualPort: port,
				Timeout:    timeout,
				All:        all,
			})
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "manual device host (skips discovery, requires --port)")
	cmd.Flags().IntVar(&port, "port", 0, "manual device port (skips discovery, requires --host)")
	cmd.Flags().DurationVar(&timeout, "timeout", cfg.DefaultDiscoveryTimeout, "discovery timeout")
	cmd.Flags().BoolVar(&all, "all", false, "connect to every discovered device")

	return cmd
}

type connectOptions struct {
	ManualHost string
	ManualPort int
	Timeout    time.Duration
	All        bool
}

func runConnect(ctx context.Context, opts connectOptions) error {
	if err := validateManualFlags(opts.ManualHost, opts.ManualPort); err != nil {
		return err
	}

	client, err := adb.NewClient()
	if err != nil {
		return err
	}

	if opts.All {
		return connectAll(ctx, client, opts)
	}

	device, err := resolveAndConnect(ctx, client, opts)
	if err != nil {
		return err
	}

	logging.FromContext(ctx).InfoContext(ctx, "connected", "device", device.String())

	return nil
}

// resolveAndConnect finds a target device (manual flags, single discovered
// match, or interactive selection among _adb-tls-connect._tcp results),
// skips it if adb already reports it connected, and connects otherwise.
// Shared by connect and expo. The returned device's Serial is always set to
// its host:port endpoint, which is how adb addresses wireless devices.
func resolveAndConnect(ctx context.Context, client *adb.Client, opts connectOptions) (types.Device, error) {
	connected, err := client.Devices(ctx)
	if err != nil {
		return types.Device{}, fmt.Errorf("list adb devices: %w", err)
	}

	device, err := resolveDevice(ctx, discovery.BrowseConnect, resolveOptions{
		ManualHost:  opts.ManualHost,
		ManualPort:  opts.ManualPort,
		Timeout:     opts.Timeout,
		SelectLabel: "Select device to connect",
		ManualLabel: "Enter device host:port",
	})
	if err != nil {
		return types.Device{}, err
	}

	device.Serial = device.Endpoint()

	if isAlreadyConnected(connected, device) {
		logging.FromContext(ctx).InfoContext(ctx, "already connected", "device", device.String())

		return device, nil
	}

	if connectErr := client.Connect(ctx, device.Endpoint()); connectErr != nil {
		return types.Device{}, fmt.Errorf("connect %s: %w", device.Endpoint(), connectErr)
	}

	return device, nil
}

func isAlreadyConnected(connected []types.Device, target types.Device) bool {
	for _, d := range connected {
		if d.Endpoint() == target.Endpoint() {
			return true
		}
	}

	return false
}

func connectAll(ctx context.Context, client *adb.Client, opts connectOptions) error {
	logger := logging.FromContext(ctx)

	connected, err := client.Devices(ctx)
	if err != nil {
		return fmt.Errorf("list adb devices: %w", err)
	}

	discovered, err := discovery.BrowseConnect(ctx, discovery.Options{Timeout: opts.Timeout})
	if err != nil {
		return fmt.Errorf("browse for devices: %w", err)
	}

	if len(discovered) == 0 {
		logger.InfoContext(ctx, "no devices found")

		return nil
	}

	var failures int

	for _, device := range discovered {
		if isAlreadyConnected(connected, device) {
			logger.InfoContext(ctx, "already connected", "device", device.String())
			continue
		}

		if connectErr := client.Connect(ctx, device.Endpoint()); connectErr != nil {
			logger.ErrorContext(ctx, "connect failed", "device", device.String(), "error", connectErr)
			failures++

			continue
		}

		logger.InfoContext(ctx, "connected", "device", device.String())
	}

	if failures > 0 {
		return fmt.Errorf("%d of %d devices failed to connect", failures, len(discovered))
	}

	return nil
}
