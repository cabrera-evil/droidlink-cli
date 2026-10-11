# droidlink

## Overview

A Linux-first CLI that removes the manual IP/port/pairing-code friction from
Android Wireless Debugging and Expo dev-client workflows. It discovers
devices over mDNS, pairs and connects automatically, and wires up `adb
reverse` + `npx expo start` for you.

## Commands

- `droidlink discover` — find Android Wireless Debugging devices/services on the local network.
- `droidlink pair` — discover an unpaired device and pair with it using its pairing code.
  Add `--qr` to show a QR code to scan from *Wireless debugging > Pair device with QR code* instead.
- `droidlink connect` — discover paired devices and connect automatically.
- `droidlink devices` — list discovered/connected devices with status.
- `droidlink disconnect` — disconnect one or all connected devices.
- `droidlink reverse` — configure `adb reverse` for the Metro bundler port (default `8081`).
- `droidlink expo` — connect a device, configure Metro reverse, and start `npx expo start --dev-client --localhost`.

Every command falls back to `--host`/`--port` (and interactive prompts) when
discovery finds nothing or the terminal isn't interactive. Run
`droidlink [command] --help` for the full flag list.

## Requirements

- Go 1.25+ (or use `mise` to install)
- `adb` (Android platform-tools) on `PATH` or under `$ANDROID_HOME`/`$ANDROID_SDK_ROOT`
- `npx` (Node.js/npm) for `droidlink expo`

## Quick Start

```bash
make build
./bin/droidlink discover
```

## Updating

Release builds can update themselves from GitHub Releases:

```bash
droidlink update --check          # report current vs latest, change nothing
droidlink update                  # prompts "Update X -> Y? [y/N]"
droidlink update --yes            # required when stdin is not interactive
droidlink update --version v1.2.3 # install a specific release
```

Dev and dirty builds refuse to update. Set `GITHUB_TOKEN` for rate limits.

## Docker

```bash
docker build -t droidlink .
docker run --rm droidlink
```

Note: droidlink relies on host mDNS multicast and a locally installed `adb`,
so running it inside the container mainly demonstrates the build — real
device discovery needs host networking and `adb` available in the image.

## Tooling

- `make` for build, run, test, tidy, and Docker targets
- `mise` for toolchain management and task aliases

## Development

See `DEVELOPMENT.md` for detailed workflows.
