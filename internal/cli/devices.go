package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"droidlink/internal/adb"
	"droidlink/internal/config"
	"droidlink/internal/discovery"
	"droidlink/internal/logging"
	"droidlink/internal/types"
)

func newDevicesCmd(cfg config.Config) *cobra.Command {
	var (
		timeout time.Duration
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "devices",
		Short: "List discovered and connected Android devices",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDevices(cmd.Context(), devicesOptions{Timeout: timeout, JSON: jsonOut})
		},
	}

	cmd.Flags().DurationVar(&timeout, "timeout", cfg.DefaultDiscoveryTimeout, "mDNS timeout (0 disables discovery)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output as JSON")

	return cmd
}

type devicesOptions struct {
	Timeout time.Duration
	JSON    bool
}

func runDevices(ctx context.Context, opts devicesOptions) error {
	logger := logging.FromContext(ctx)

	client, err := adb.NewClient()
	if err != nil {
		return err
	}

	connected, err := client.Devices(ctx)
	if err != nil {
		return fmt.Errorf("list adb devices: %w", err)
	}

	merged := connected

	if opts.Timeout > 0 {
		discovered, browseErr := browseByService(ctx, "both", discovery.Options{Timeout: opts.Timeout})
		if browseErr != nil {
			logger.WarnContext(ctx, "mDNS discovery failed, showing adb-known devices only", "error", browseErr)
		} else {
			merged = mergeDevices(connected, discovered)
		}
	}

	sort.Slice(merged, func(i, j int) bool { return merged[i].Endpoint() < merged[j].Endpoint() })

	if opts.JSON {
		return writeJSON(os.Stdout, merged)
	}

	return printDeviceStatusTable(merged)
}

func mergeDevices(connected, discovered []types.Device) []types.Device {
	byEndpoint := make(map[string]types.Device, len(connected)+len(discovered))

	for _, d := range connected {
		byEndpoint[d.Endpoint()] = d
	}

	for _, d := range discovered {
		existing, ok := byEndpoint[d.Endpoint()]
		if !ok {
			byEndpoint[d.Endpoint()] = d
			continue
		}

		existing.Service = d.Service
		existing.Paired = existing.Paired || d.Paired
		byEndpoint[d.Endpoint()] = existing
	}

	merged := make([]types.Device, 0, len(byEndpoint))
	for _, d := range byEndpoint {
		merged = append(merged, d)
	}

	return merged
}
