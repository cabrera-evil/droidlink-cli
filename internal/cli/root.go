// Package cli wires droidlink's Cobra commands. Each command's RunE is a
// thin shim into a package-private runX function; business logic lives in
// internal/adb, internal/discovery, and internal/ui.
package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"droidlink/internal/config"
	"droidlink/internal/logging"
)

// Execute builds and runs the droidlink root command with a signal-aware
// context, returning any error from execution.
func Execute(version string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	root := newRootCmd()
	root.Version = version

	if err := root.ExecuteContext(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
			fmt.Fprintln(os.Stderr, "error:", err)
		}

		return err
	}

	return nil
}

func newRootCmd() *cobra.Command {
	cfg := config.Load()

	logLevel := cfg.DefaultLogLevel
	noColor := false

	root := &cobra.Command{
		Use:           "droidlink",
		Short:         "Simplify Android Wireless Debugging and Expo dev-client workflows",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			level, err := logging.ParseLevel(logLevel)
			if err != nil {
				return fmt.Errorf("--log-level: %w", err)
			}

			logger := logging.New(logging.Options{Level: level, NoColor: noColor})
			cmd.SetContext(logging.WithLogger(cmd.Context(), logger))

			return nil
		},
	}

	root.PersistentFlags().StringVar(&logLevel, "log-level", cfg.DefaultLogLevel, "log level: debug|info|warn|error")
	root.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored log output")

	root.AddCommand(
		newDiscoverCmd(cfg),
		newPairCmd(cfg),
		newConnectCmd(cfg),
		newDevicesCmd(cfg),
		newDisconnectCmd(),
		newReverseCmd(cfg),
		newExpoCmd(cfg),
	)

	return root
}
