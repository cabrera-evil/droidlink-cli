package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"droidlink/internal/types"
)

// deviceStateColumn is the STATE column index in printDeviceStatusTable's
// headers ("NAME", "HOST:PORT", "STATE", "ADVERTISING").
const deviceStateColumn = 2

func printf(format string, args ...any) error {
	if _, err := fmt.Fprintf(os.Stdout, format, args...); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	return nil
}

func printDeviceTable(devices []types.Device) error {
	t := newDeviceTable("NAME", "SERVICE", "HOST:PORT")

	for _, d := range devices {
		t.Row(d.Name, d.Service.String(), d.Endpoint())
	}

	return printf("%s\n", t.Render())
}

// displayEndpoint renders a device's connection identifier for tables. Some
// adb-reported serials aren't "host:port" at all — e.g. modern adb's own
// mDNS auto-tracking lists a device by its raw mDNS instance name
// ("adb-XXXX._adb-tls-connect._tcp") as a second serial alongside the usual
// host:port one — and Device.Endpoint() would otherwise render those as a
// meaningless ":0".
func displayEndpoint(d types.Device) string {
	if d.Host != "" {
		return d.Endpoint()
	}

	return d.Serial
}

func printDeviceStatusTable(devices []types.Device) error {
	t := newDeviceTable("NAME", "HOST:PORT", "STATE", "ADVERTISING")

	t.StyleFunc(func(row, col int) lipgloss.Style {
		if row == table.HeaderRow {
			return styleHeader()
		}

		if col == deviceStateColumn && row >= 0 && row < len(devices) {
			return stateStyle(devices[row].State)
		}

		return styleCell()
	})

	for _, d := range devices {
		advertising := "none"
		if d.Service != types.ServiceUnknown {
			advertising = d.Service.String()
		}

		state := d.State
		if state == "" {
			state = "-" // known only via mDNS; adb has never reported a state for it
		}

		t.Row(d.Name, displayEndpoint(d), state, advertising)
	}

	return printf("%s\n", t.Render())
}
