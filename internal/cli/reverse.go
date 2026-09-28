package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"droidlink/internal/adb"
	"droidlink/internal/config"
	"droidlink/internal/logging"
	"droidlink/internal/ui"
)

func newReverseCmd(cfg config.Config) *cobra.Command {
	var (
		port   int
		serial string
	)

	cmd := &cobra.Command{
		Use:   "reverse",
		Short: "Configure adb reverse for the Metro bundler port",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runReverse(cmd.Context(), reverseOptions{Port: port, Serial: serial})
		},
	}

	cmd.Flags().IntVar(&port, "port", cfg.DefaultMetroPort, "Metro port to reverse")
	cmd.Flags().StringVar(&serial, "serial", "", "adb serial to target (required if multiple devices are connected)")

	return cmd
}

type reverseOptions struct {
	Port   int
	Serial string
}

func runReverse(ctx context.Context, opts reverseOptions) error {
	client, err := adb.NewClient()
	if err != nil {
		return err
	}

	serial, err := resolveSerial(ctx, client, opts.Serial)
	if err != nil {
		return err
	}

	return applyReverse(ctx, client, serial, opts.Port)
}

// applyReverse runs `adb reverse` for the given serial/port, logging
// success. Shared by reverse and expo.
func applyReverse(ctx context.Context, client *adb.Client, serial string, port int) error {
	if err := client.Reverse(ctx, serial, port); err != nil {
		return fmt.Errorf("reverse tcp:%d: %w", port, err)
	}

	logging.FromContext(ctx).InfoContext(ctx, "reverse configured", "port", port, "serial", serial)

	return nil
}

// resolveSerial picks the adb serial to target: the explicit --serial flag,
// the sole connected device, or an interactive choice among several.
func resolveSerial(ctx context.Context, client *adb.Client, serial string) (string, error) {
	if serial != "" {
		return serial, nil
	}

	connected, err := client.Devices(ctx)
	if err != nil {
		return "", fmt.Errorf("list adb devices: %w", err)
	}

	if len(connected) == 0 {
		return "", errors.New("no connected devices found; run `droidlink connect` first")
	}

	device, err := ui.SelectDevice(connected, "Select device to reverse")
	if err != nil {
		return "", err
	}

	return device.Serial, nil
}
