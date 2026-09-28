package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"droidlink/internal/adb"
	"droidlink/internal/logging"
	"droidlink/internal/ui"
)

func newDisconnectCmd() *cobra.Command {
	var (
		host string
		port int
		all  bool
	)

	cmd := &cobra.Command{
		Use:   "disconnect",
		Short: "Disconnect one or all connected devices",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDisconnect(cmd.Context(), disconnectOptions{Host: host, Port: port, All: all})
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "device host to disconnect")
	cmd.Flags().IntVar(&port, "port", 0, "device port to disconnect")
	cmd.Flags().BoolVar(&all, "all", false, "disconnect every connected device")

	return cmd
}

type disconnectOptions struct {
	Host string
	Port int
	All  bool
}

func runDisconnect(ctx context.Context, opts disconnectOptions) error {
	if err := validateManualFlags(opts.Host, opts.Port); err != nil {
		return err
	}

	client, err := adb.NewClient()
	if err != nil {
		return err
	}

	if opts.All {
		if disconnectErr := client.Disconnect(ctx, ""); disconnectErr != nil {
			return fmt.Errorf("disconnect all: %w", disconnectErr)
		}

		logging.FromContext(ctx).InfoContext(ctx, "disconnected all devices")

		return nil
	}

	endpoint := opts.Host
	if endpoint != "" {
		endpoint = fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	} else {
		endpoint, err = selectConnectedEndpoint(ctx, client)
		if err != nil {
			return err
		}
	}

	if disconnectErr := client.Disconnect(ctx, endpoint); disconnectErr != nil {
		return fmt.Errorf("disconnect %s: %w", endpoint, disconnectErr)
	}

	logging.FromContext(ctx).InfoContext(ctx, "disconnected", "device", endpoint)

	return nil
}

func selectConnectedEndpoint(ctx context.Context, client *adb.Client) (string, error) {
	connected, err := client.Devices(ctx)
	if err != nil {
		return "", fmt.Errorf("list adb devices: %w", err)
	}

	if len(connected) == 0 {
		return "", errors.New("no connected devices found; pass --host/--port or --all")
	}

	device, err := ui.SelectDevice(connected, "Select device to disconnect")
	if err != nil {
		return "", err
	}

	return device.Endpoint(), nil
}
