package discovery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/libp2p/zeroconf/v2"

	"droidlink/internal/types"
)

// ErrNotFound indicates the awaited service instance was not advertised
// before the timeout elapsed.
var ErrNotFound = errors.New("discovery: service instance not found")

// WaitForPairingInstance browses _adb-tls-pairing._tcp until a device
// advertises the given instance name (as happens after scanning a pairing
// QR code), returning as soon as it appears.
func WaitForPairingInstance(ctx context.Context, instance string, timeout time.Duration) (types.Device, error) {
	browseCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	entries := make(chan *zeroconf.ServiceEntry, entriesBuffer)
	found := make(chan types.Device, 1)

	// Browse closes entries when it returns, so this loop keeps draining
	// until then and never leaves Browse blocked on a full channel.
	go func() {
		defer close(found)

		matched := false

		for entry := range entries {
			if matched || entry.Instance != instance {
				continue
			}

			device := fromServiceEntry(entry, types.ServicePairing)
			if device.Host == "" {
				continue
			}

			matched = true
			found <- device

			cancel()
		}
	}()

	err := zeroconf.Browse(browseCtx, servicePairing, domain, entries)

	device, ok := <-found
	if ok {
		return device, nil
	}

	if parentErr := ctx.Err(); parentErr != nil {
		return types.Device{}, parentErr
	}

	if err != nil {
		return types.Device{}, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}

	return types.Device{}, fmt.Errorf("%w: %q within %s", ErrNotFound, instance, timeout)
}
