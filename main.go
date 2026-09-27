package main

import (
	"fmt"

	"go-template/internal/app"
)

// Version is injected at build time via -ldflags "-X main.Version=...".
var Version = "dev" //nolint:gochecknoglobals // set via -ldflags

func main() {
	fmt.Println(app.Run()) //nolint:forbidigo // template entrypoint writes to stdout
}
