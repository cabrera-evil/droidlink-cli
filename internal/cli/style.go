package cli

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// Color palette, matching huh's ThemeCharm accent so prompts and static
// output feel like one system.
const (
	colorAccent  lipgloss.Color = "99"  // violet
	colorSuccess lipgloss.Color = "42"  // green
	colorWarning lipgloss.Color = "214" // amber
	colorError   lipgloss.Color = "204" // pink/red
	colorMuted   lipgloss.Color = "241" // gray
)

func styleHeader() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(colorAccent).Padding(0, 1)
}

func styleCell() lipgloss.Style {
	return lipgloss.NewStyle().Padding(0, 1)
}

func styleSuccess() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(colorSuccess)
}

func styleError() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(colorError)
}

func styleMuted() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(colorMuted)
}

// stateStyle colors an adb device state so device/offline/unauthorized/"-"
// are visually distinct at a glance.
func stateStyle(state string) lipgloss.Style {
	switch state {
	case "device":
		return lipgloss.NewStyle().Padding(0, 1).Foreground(colorSuccess)
	case "":
		return styleMuted().Padding(0, 1)
	default:
		return lipgloss.NewStyle().Padding(0, 1).Foreground(colorWarning)
	}
}

func newDeviceTable(headers ...string) *table.Table {
	return table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(colorMuted)).
		Headers(headers...).
		StyleFunc(func(row, _ int) lipgloss.Style {
			if row == table.HeaderRow {
				return styleHeader()
			}

			return styleCell()
		})
}

func success(msg string) string {
	return styleSuccess().Render("✓ ") + msg
}
