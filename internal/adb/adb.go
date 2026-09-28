// Package adb is the only droidlink package that shells out to the system
// adb binary. It knows nothing about mDNS discovery or interactive UI.
package adb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"droidlink/internal/types"
)

// ErrNotFound indicates the adb binary could not be located.
var ErrNotFound = errors.New("adb: binary not found on PATH")

// CommandError wraps a non-zero adb exit, preserving its combined output for
// display to the user.
type CommandError struct {
	Args     []string
	ExitCode int
	Output   string
}

// Error implements the error interface.
func (e *CommandError) Error() string {
	return fmt.Sprintf("adb %s: exit %d: %s", strings.Join(e.Args, " "), e.ExitCode, strings.TrimSpace(e.Output))
}

// Client wraps a located adb binary for command execution.
type Client struct {
	binPath string
}

// NewClient locates adb on the system and returns a ready-to-use Client.
func NewClient() (*Client, error) {
	path, err := LocateBinary()
	if err != nil {
		return nil, err
	}

	return &Client{binPath: path}, nil
}

// LocateBinary finds adb on PATH, falling back to common Android SDK
// install locations (Android Studio doesn't always add platform-tools to
// PATH), returning ErrNotFound with the checked locations if nothing works.
func LocateBinary() (string, error) {
	if path, err := exec.LookPath("adb"); err == nil {
		return path, nil
	}

	candidates := candidatePaths()

	for _, path := range candidates {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}

	return "", fmt.Errorf(
		"%w (checked PATH and: %s; install Android platform-tools and ensure adb is reachable)",
		ErrNotFound, strings.Join(candidates, ", "),
	)
}

func candidatePaths() []string {
	var paths []string

	for _, envVar := range []string{"ANDROID_HOME", "ANDROID_SDK_ROOT"} {
		if v := os.Getenv(envVar); v != "" {
			paths = append(paths, filepath.Join(v, "platform-tools", "adb"))
		}
	}

	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, "Android", "Sdk", "platform-tools", "adb"))
	}

	return paths
}

// Devices runs `adb devices -l` and returns devices currently known to adb
// in the "device" (usable) state.
func (c *Client) Devices(ctx context.Context) ([]types.Device, error) {
	output, err := c.run(ctx, "devices", "-l")
	if err != nil {
		return nil, fmt.Errorf("devices: %w", err)
	}

	return ParseDevicesOutput(output), nil
}

// Pair runs `adb pair host:port code`. Platform-tools 30.0.0+ accept the
// pairing code as a non-interactive second positional argument; older
// versions require an interactive TTY prompt and are not supported here —
// their failure surfaces via the returned CommandError.
func (c *Client) Pair(ctx context.Context, hostPort, code string) error {
	if _, err := c.run(ctx, "pair", hostPort, code); err != nil {
		return fmt.Errorf("pair %s: %w", hostPort, err)
	}

	return nil
}

// Connect runs `adb connect host:port`.
func (c *Client) Connect(ctx context.Context, hostPort string) error {
	if _, err := c.run(ctx, "connect", hostPort); err != nil {
		return fmt.Errorf("connect %s: %w", hostPort, err)
	}

	return nil
}

// Disconnect runs `adb disconnect [host:port]`; an empty hostPort
// disconnects every device, per adb's own semantics.
func (c *Client) Disconnect(ctx context.Context, hostPort string) error {
	args := []string{"disconnect"}
	if hostPort != "" {
		args = append(args, hostPort)
	}

	if _, err := c.run(ctx, args...); err != nil {
		return fmt.Errorf("disconnect: %w", err)
	}

	return nil
}

// Reverse runs `adb reverse tcp:<port> tcp:<port>`, scoped to serial via
// `-s` when non-empty (required once more than one device is connected).
func (c *Client) Reverse(ctx context.Context, serial string, port int) error {
	spec := fmt.Sprintf("tcp:%d", port)

	args := []string{}
	if serial != "" {
		args = append(args, "-s", serial)
	}

	args = append(args, "reverse", spec, spec)

	if _, err := c.run(ctx, args...); err != nil {
		return fmt.Errorf("reverse %s: %w", spec, err)
	}

	return nil
}

func (c *Client) run(ctx context.Context, args ...string) (string, error) {
	// binPath resolved via LookPath/known SDK paths, args are internal, not shell input.
	cmd := exec.CommandContext(ctx, c.binPath, args...) //nolint:gosec // see comment above

	output, err := cmd.CombinedOutput()
	if err == nil {
		return string(output), nil
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return "", fmt.Errorf("adb %s: %w", strings.Join(args, " "), err)
	}

	return "", &CommandError{Args: args, ExitCode: exitErr.ExitCode(), Output: string(output)}
}
