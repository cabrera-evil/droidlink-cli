// Package ui is the only droidlink package that imports promptui. It holds
// pure input/output helpers with no direct adb or discovery calls.
package ui

import (
	"errors"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"

	"droidlink/internal/types"
	"droidlink/internal/utils"
)

// ErrNoSelection indicates no device was available to select from.
var ErrNoSelection = errors.New("ui: no device selected")

// ErrNotInteractive indicates a prompt was needed but stdin isn't a
// terminal, so the caller should ask for explicit flags instead.
var ErrNotInteractive = errors.New("ui: not running in an interactive terminal")

// SelectDevice shows an interactive arrow-key list of devices and returns
// the chosen one. If exactly one device is given, it is returned immediately
// without prompting. If devices is empty, ErrNoSelection is returned.
func SelectDevice(devices []types.Device, label string) (types.Device, error) {
	if len(devices) == 0 {
		return types.Device{}, ErrNoSelection
	}

	if len(devices) == 1 {
		return devices[0], nil
	}

	if !utils.IsInteractive(os.Stdin) {
		return types.Device{}, fmt.Errorf(
			"%w: multiple devices found, pass --host/--port or --serial to disambiguate", ErrNotInteractive,
		)
	}

	items := make([]string, len(devices))
	for i, d := range devices {
		items[i] = d.String()
	}

	prompt := promptui.Select{
		Label: label,
		Items: items,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return types.Device{}, fmt.Errorf("select device: %w", err)
	}

	return devices[idx], nil
}
