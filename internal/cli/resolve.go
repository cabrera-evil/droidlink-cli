package cli

import (
	"context"
	"errors"
	"fmt"
	"time"

	"droidlink/internal/discovery"
	"droidlink/internal/logging"
	"droidlink/internal/types"
	"droidlink/internal/ui"
)

// browseFunc matches discovery.BrowsePairing/BrowseConnect's signature.
type browseFunc func(context.Context, discovery.Options) ([]types.Device, error)

// resolveOptions configures resolveDevice.
type resolveOptions struct {
	ManualHost  string
	ManualPort  int
	Timeout     time.Duration
	SelectLabel string
	ManualLabel string
}

// resolveDevice finds a single target device, shared by pair, connect, and
// expo: manual --host/--port flags take precedence; otherwise it browses via
// browse, falling back to an interactive manual prompt if nothing is found.
func resolveDevice(ctx context.Context, browse browseFunc, opts resolveOptions) (types.Device, error) {
	if opts.ManualHost != "" {
		return types.Device{Host: opts.ManualHost, Port: opts.ManualPort}, nil
	}

	logger := logging.FromContext(ctx)

	devices, err := browse(ctx, discovery.Options{Timeout: opts.Timeout})
	if err != nil {
		if !errors.Is(err, discovery.ErrUnavailable) {
			return types.Device{}, err
		}

		logger.WarnContext(ctx, "mDNS discovery unavailable, falling back to manual entry", "error", err)

		devices = nil
	}

	if len(devices) == 0 {
		host, port, promptErr := ui.PromptHostPort(opts.ManualLabel)
		if promptErr != nil {
			return types.Device{}, fmt.Errorf("no devices found: %w", promptErr)
		}

		return types.Device{Host: host, Port: port}, nil
	}

	device, err := ui.SelectDevice(devices, opts.SelectLabel)
	if err != nil {
		return types.Device{}, err
	}

	return device, nil
}

func validateManualFlags(host string, port int) error {
	if (host == "") != (port == 0) {
		return errors.New("--host and --port must be set together")
	}

	return nil
}

// usableDevices filters an `adb devices -l` result down to entries in the
// "device" state, excluding offline/unauthorized ones that aren't usable
// targets for reverse/connect checks.
func usableDevices(devices []types.Device) []types.Device {
	usable := make([]types.Device, 0, len(devices))

	for _, d := range devices {
		if d.Connected {
			usable = append(usable, d)
		}
	}

	return usable
}
