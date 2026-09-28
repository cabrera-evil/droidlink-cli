package ui

import (
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"

	"github.com/charmbracelet/huh"

	"droidlink/internal/utils"
)

const maxPort = 65535

var pairingCodePattern = regexp.MustCompile(`^\d{6}$`)

// PromptPairingCode asks for and validates a 6-digit Android pairing code.
func PromptPairingCode() (string, error) {
	if !utils.IsInteractive(os.Stdin) {
		return "", fmt.Errorf("%w: pass --code explicitly", ErrNotInteractive)
	}

	var code string

	field := huh.NewInput().
		Title("Pairing code").
		Placeholder("123456").
		Validate(func(s string) error {
			if !pairingCodePattern.MatchString(s) {
				return errors.New("pairing code must be exactly 6 digits")
			}

			return nil
		}).
		Value(&code)

	if err := runField(field); err != nil {
		return "", fmt.Errorf("prompt pairing code: %w", err)
	}

	return code, nil
}

// PromptHostPort asks for manual "host:port" entry, used as a fallback when
// discovery finds nothing or the environment can't use mDNS.
func PromptHostPort(label string) (string, int, error) {
	if !utils.IsInteractive(os.Stdin) {
		return "", 0, fmt.Errorf("%w: pass --host/--port explicitly", ErrNotInteractive)
	}

	var input string

	field := huh.NewInput().
		Title(label).
		Placeholder("192.168.1.5:5555").
		Validate(func(s string) error {
			_, _, err := splitHostPort(s)
			return err
		}).
		Value(&input)

	if err := runField(field); err != nil {
		return "", 0, fmt.Errorf("prompt host:port: %w", err)
	}

	return splitHostPort(input)
}

func runField(field huh.Field) error {
	return huh.NewForm(huh.NewGroup(field)).WithTheme(huh.ThemeCharm()).Run()
}

func splitHostPort(input string) (string, int, error) {
	host, portStr, err := net.SplitHostPort(input)
	if err != nil {
		return "", 0, fmt.Errorf("invalid host:port %q: %w", input, err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > maxPort {
		return "", 0, fmt.Errorf("invalid port in %q", input)
	}

	return host, port, nil
}
