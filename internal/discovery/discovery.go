// Package discovery is the only droidlink package that imports zeroconf. It
// knows nothing about adb or interactive UI.
package discovery

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/zeroconf/v2"

	"droidlink/internal/types"
)

// ErrUnavailable indicates mDNS discovery could not be started, typically
// because a multicast UDP socket could not be opened (missing network
// interface, firewall, or sandboxing).
var ErrUnavailable = errors.New("discovery: mDNS unavailable")

const (
	servicePairing = "_adb-tls-pairing._tcp"
	serviceConnect = "_adb-tls-connect._tcp"
	domain         = "local."

	entriesBuffer = 8
)

// Options configures a browse operation.
type Options struct {
	Timeout time.Duration
}

// BrowsePairing browses _adb-tls-pairing._tcp, the service Android
// advertises only while the device's "pair with code" screen is open.
func BrowsePairing(ctx context.Context, opts Options) ([]types.Device, error) {
	return browse(ctx, servicePairing, types.ServicePairing, opts)
}

// BrowseConnect browses _adb-tls-connect._tcp, the service Android
// advertises whenever Wireless Debugging is toggled on, regardless of
// pairing state.
func BrowseConnect(ctx context.Context, opts Options) ([]types.Device, error) {
	return browse(ctx, serviceConnect, types.ServiceConnect, opts)
}

// browse runs zeroconf.Browse, which blocks until browseCtx is done (or a
// hard failure, e.g. no usable multicast interface), closing the entries
// channel itself right before returning. A collector goroutine drains
// entries into devices and exits once that close happens.
func browse(ctx context.Context, service string, svcType types.ServiceType, opts Options) ([]types.Device, error) {
	browseCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	entries := make(chan *zeroconf.ServiceEntry, entriesBuffer)

	var (
		mu      sync.Mutex
		devices []types.Device
		seen    = make(map[string]bool)
	)

	done := make(chan struct{})

	go func() {
		defer close(done)

		for entry := range entries {
			device := fromServiceEntry(entry, svcType)
			if device.Host == "" {
				continue
			}

			mu.Lock()

			if !seen[device.Endpoint()] {
				seen[device.Endpoint()] = true

				devices = append(devices, device)
			}

			mu.Unlock()
		}
	}()

	err := zeroconf.Browse(browseCtx, service, domain, entries)

	<-done

	// Distinguish a real SIGINT/parent cancellation from opts.Timeout simply
	// elapsing (which only cancels browseCtx, not ctx) so callers can react
	// to the former without treating "found nothing before timeout" as an
	// error.
	if parentErr := ctx.Err(); parentErr != nil {
		return nil, parentErr
	}

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}

	mu.Lock()
	defer mu.Unlock()

	return devices, nil
}
