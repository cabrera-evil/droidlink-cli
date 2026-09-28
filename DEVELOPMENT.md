# Development Guide

## Setup

Install toolchain with `mise`:

```bash
mise install
```

You'll also need `adb` (Android platform-tools) on `PATH` or under
`$ANDROID_HOME`/`$ANDROID_SDK_ROOT`, and `npx` (Node.js/npm) if you're
working on the `expo` command.

## Build

```bash
make build                     # bin/droidlink
make install                   # ~/.local/bin/droidlink (override INSTALL_DIR)
```

The build stamps `main.Version` from `git describe`; override with `make build VERSION=1.2.3`.

## Run

```bash
make run ARGS="discover"       # go run . discover
make run ARGS="--help"
```

## Test

```bash
make test
```

## Lint

```bash
make lint                      # golangci-lint run ./...
```

## Docker

```bash
make docker-build
make docker-run
```

## Notes

- Runtime defaults (`DROIDLINK_METRO_PORT`, `DROIDLINK_DISCOVERY_TIMEOUT`, `DROIDLINK_LOG_LEVEL`) come from the environment; command flags always override them. There is no file-based config.
- Manually testing `pair`/`connect`/`reverse`/`expo` end-to-end requires a real Android device with Wireless Debugging enabled on the same LAN.
