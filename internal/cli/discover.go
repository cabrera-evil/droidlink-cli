package cli

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"droidlink/internal/config"
	"droidlink/internal/discovery"
	"droidlink/internal/logging"
	"droidlink/internal/types"
)

func newDiscoverCmd(cfg config.Config) *cobra.Command {
	var (
		timeout time.Duration
		service string
	)

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover Android Wireless Debugging devices on the local network",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDiscover(cmd.Context(), discoverOptions{Timeout: timeout, Service: service})
		},
	}

	cmd.Flags().DurationVar(&timeout, "timeout", cfg.DefaultDiscoveryTimeout, "discovery timeout")
	cmd.Flags().StringVar(&service, "service", "both", "service to browse: pairing|connect|both")

	return cmd
}

type discoverOptions struct {
	Timeout time.Duration
	Service string
}

func runDiscover(ctx context.Context, opts discoverOptions) error {
	logger := logging.FromContext(ctx)

	devices, err := browseByService(ctx, opts.Service, discovery.Options{Timeout: opts.Timeout})
	if err != nil {
		if errors.Is(err, discovery.ErrUnavailable) {
			return fmt.Errorf("mDNS discovery unavailable: %w (try --host/--port on pair/connect instead)", err)
		}

		return err
	}

	if len(devices) == 0 {
		logger.InfoContext(ctx, "no devices found", "timeout", opts.Timeout)

		return printf(
			"No devices found within %s. Ensure Wireless Debugging is enabled and the device is on the same network.\n",
			opts.Timeout,
		)
	}

	return printDeviceTable(devices)
}

func browseByService(ctx context.Context, service string, opts discovery.Options) ([]types.Device, error) {
	switch service {
	case "pairing":
		return discovery.BrowsePairing(ctx, opts)
	case "connect":
		return discovery.BrowseConnect(ctx, opts)
	case "both":
		return browseBoth(ctx, opts)
	default:
		return nil, fmt.Errorf("unknown --service %q: must be pairing, connect, or both", service)
	}
}

func browseBoth(ctx context.Context, opts discovery.Options) ([]types.Device, error) {
	logger := logging.FromContext(ctx)

	var (
		wg                     sync.WaitGroup
		pairing, connect       []types.Device
		pairingErr, connectErr error
	)

	wg.Add(2) //nolint:mnd // exactly two concurrent browses (pairing + connect)

	go func() {
		defer wg.Done()

		pairing, pairingErr = discovery.BrowsePairing(ctx, opts)
	}()

	go func() {
		defer wg.Done()

		connect, connectErr = discovery.BrowseConnect(ctx, opts)
	}()

	wg.Wait()

	if pairingErr != nil && !errors.Is(pairingErr, context.Canceled) {
		logger.WarnContext(ctx, "pairing discovery failed", "error", pairingErr)
	}

	if connectErr != nil && !errors.Is(connectErr, context.Canceled) {
		logger.WarnContext(ctx, "connect discovery failed", "error", connectErr)
	}

	if pairingErr != nil && connectErr != nil {
		return nil, pairingErr
	}

	return append(pairing, connect...), nil
}
