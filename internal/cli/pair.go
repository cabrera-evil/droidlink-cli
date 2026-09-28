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
	"droidlink/internal/ui"
)

func newPairCmd(cfg config.Config) *cobra.Command {
	var (
		host    string
		port    int
		code    string
		timeout time.Duration
	)

	cmd := &cobra.Command{
		Use:   "pair",
		Short: "Discover an unpaired device and pair with it using its pairing code",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runPair(cmd.Context(), pairOptions{
				ManualHost: host,
				ManualPort: port,
				Code:       code,
				Timeout:    timeout,
			})
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "manual device host (skips discovery, requires --port)")
	cmd.Flags().IntVar(&port, "port", 0, "manual device port (skips discovery, requires --host)")
	cmd.Flags().StringVar(&code, "code", "", "6-digit pairing code shown on the device (prompted if omitted)")
	cmd.Flags().DurationVar(&timeout, "timeout", cfg.DefaultDiscoveryTimeout, "discovery timeout")

	return cmd
}

type pairOptions struct {
	ManualHost string
	ManualPort int
	Code       string
	Timeout    time.Duration
}

func runPair(ctx context.Context, opts pairOptions) error {
	if err := validateManualFlags(opts.ManualHost, opts.ManualPort); err != nil {
		return err
	}

	client, err := adb.NewClient()
	if err != nil {
		return err
	}

	device, err := resolveDevice(ctx, discovery.BrowsePairing, resolveOptions{
		ManualHost:  opts.ManualHost,
		ManualPort:  opts.ManualPort,
		Timeout:     opts.Timeout,
		SelectLabel: "Select device to pair",
		ManualLabel: "Enter device host:port",
	})
	if err != nil {
		return fmt.Errorf("find device to pair (open Wireless debugging > Pair with pairing code): %w", err)
	}

	code := opts.Code
	if code == "" {
		code, err = ui.PromptPairingCode()
		if err != nil {
			return err
		}
	}

	if pairErr := client.Pair(ctx, device.Endpoint(), code); pairErr != nil {
		return fmt.Errorf("pair with %s: %w", device.Endpoint(), pairErr)
	}

	logging.FromContext(ctx).InfoContext(ctx, "paired", "device", device.String())

	return printf("Paired successfully. Run `droidlink connect` to connect.\n")
}
