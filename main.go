package main

import (
	"context"
	"errors"
	"os"

	"droidlink/internal/cli"
)

// Version is injected at build time via -ldflags "-X main.Version=...".
var Version = "dev" //nolint:gochecknoglobals // set via -ldflags

const sigintExitCode = 130

func main() {
	if err := cli.Execute(Version); err != nil {
		if errors.Is(err, context.Canceled) {
			os.Exit(sigintExitCode)
		}

		os.Exit(1)
	}
}
