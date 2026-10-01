package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"droidlink/internal/adb"
	"droidlink/internal/config"
	"droidlink/internal/discovery"
	"droidlink/internal/logging"
	"droidlink/internal/ui"
)

// pairMaxAttempts bounds automatic retries of the pairing session-rotation
// glitch (see isTransientPairError). Each attempt re-discovers the device
// and re-prompts for the code, since a failed attempt commonly causes the
// phone to rotate its pairing port and code defensively.
const pairMaxAttempts = 3

// defaultQRTimeout leaves time to open the scanner and scan the code.
const defaultQRTimeout = 2 * time.Minute

const qrPairInstructions = "On your device open Developer options > Wireless debugging > " +
	"Pair device with QR code, then scan:"

func newPairCmd(cfg config.Config) *cobra.Command {
	var (
		host      string
		port      int
		code      string
		timeout   time.Duration
		qr        bool
		qrTimeout time.Duration
	)

	cmd := &cobra.Command{
		Use:   "pair",
		Short: "Discover an unpaired device and pair with it using its pairing code",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if qr {
				return runPairQR(cmd.Context(), qrTimeout)
			}

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
	cmd.Flags().BoolVar(&qr, "qr", false, "show a QR code to scan from Wireless debugging > Pair device with QR code")
	cmd.Flags().DurationVar(&qrTimeout, "qr-timeout", defaultQRTimeout,
		"how long to wait for the device to scan the QR code")
	cmd.MarkFlagsMutuallyExclusive("qr", "host")
	cmd.MarkFlagsMutuallyExclusive("qr", "port")
	cmd.MarkFlagsMutuallyExclusive("qr", "code")

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

	logger := logging.FromContext(ctx)

	var lastErr error

	for attempt := 1; attempt <= pairMaxAttempts; attempt++ {
		if attempt > 1 {
			logger.WarnContext(ctx, "pairing session likely rotated after a transient failure, retrying",
				"attempt", attempt)
		}

		device, resolveErr := resolveDevice(ctx, discovery.BrowsePairing, resolveOptions{
			ManualHost:  opts.ManualHost,
			ManualPort:  opts.ManualPort,
			Timeout:     opts.Timeout,
			SelectLabel: "Select device to pair",
			ManualLabel: "Enter device host:port",
		})
		if resolveErr != nil {
			return fmt.Errorf("find device to pair (open Wireless debugging > Pair with pairing code): %w", resolveErr)
		}

		code := opts.Code
		if code == "" {
			promptedCode, promptErr := ui.PromptPairingCode()
			if promptErr != nil {
				return promptErr
			}

			code = promptedCode
		}

		pairErr := client.Pair(ctx, device.Endpoint(), code)
		if pairErr == nil {
			logger.InfoContext(ctx, "paired", "device", device.String())

			return printf("%s\n", success("Paired. Run `droidlink connect` to connect."))
		}

		if !isTransientPairError(pairErr) || attempt == pairMaxAttempts {
			return fmt.Errorf("pair with %s: %w", device.Endpoint(), pairErr)
		}

		lastErr = pairErr
	}

	return lastErr
}

// runPairQR shows a pairing QR code, waits for the device that scanned it to
// advertise the encoded service name over mDNS, then pairs using the
// encoded password. A failed attempt is not retried: the device discards
// the QR session, so the user must rescan a fresh code.
func runPairQR(ctx context.Context, timeout time.Duration) error {
	client, err := adb.NewClient()
	if err != nil {
		return err
	}

	pairing, err := adb.NewQRPairing()
	if err != nil {
		return err
	}

	err = printf("%s\n\n", qrPairInstructions)
	if err != nil {
		return err
	}

	ui.RenderQR(os.Stdout, pairing.Payload())

	err = printf("%s\n", styleMuted().Render(fmt.Sprintf("Waiting up to %s for the device to scan...", timeout)))
	if err != nil {
		return err
	}

	device, err := discovery.WaitForPairingInstance(ctx, pairing.ServiceName, timeout)
	if err != nil {
		return fmt.Errorf("wait for QR scan (device and computer must share a network with mDNS): %w", err)
	}

	logger := logging.FromContext(ctx)
	logger.DebugContext(ctx, "QR pairing service found", "device", device.String())

	if pairErr := client.Pair(ctx, device.Endpoint(), pairing.Password); pairErr != nil {
		return fmt.Errorf("pair with %s: %w", device.Endpoint(), pairErr)
	}

	logger.InfoContext(ctx, "paired", "device", device.String())

	return printf("%s\n", success("Paired. Run `droidlink connect` to connect."))
}

// isTransientPairError reports whether err looks like adb's well-known
// intermittent pairing wire glitch ("protocol fault (couldn't read status
// message): Success") rather than a genuine rejection (e.g. wrong code).
// Retrying with a freshly discovered port and a freshly prompted code
// resolves it in practice; retrying the exact same code/port does not.
func isTransientPairError(err error) bool {
	return strings.Contains(err.Error(), "protocol fault")
}
