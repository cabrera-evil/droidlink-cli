package ui

import "github.com/charmbracelet/huh"

// runField runs a single-field huh form under droidlink's shared theme, so
// every interactive prompt looks consistent.
func runField(field huh.Field) error {
	return huh.NewForm(huh.NewGroup(field)).WithTheme(huh.ThemeCharm()).Run()
}
