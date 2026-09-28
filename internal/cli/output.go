package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"droidlink/internal/types"
)

// tabwriter column-spacing constants (minwidth, tabwidth, padding, padchar).
const (
	tableMinWidth = 0
	tableTabWidth = 4
	tablePadding  = 2
	tablePadChar  = ' '
)

func newTable() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, tableMinWidth, tableTabWidth, tablePadding, tablePadChar, 0)
}

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
	w := newTable()

	if _, err := fmt.Fprintln(w, "NAME\tSERVICE\tHOST:PORT"); err != nil {
		return fmt.Errorf("write device table: %w", err)
	}

	for _, d := range devices {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n", d.Name, d.Service, d.Endpoint()); err != nil {
			return fmt.Errorf("write device table: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush device table: %w", err)
	}

	return nil
}

func printDeviceStatusTable(devices []types.Device) error {
	w := newTable()

	if _, err := fmt.Fprintln(w, "NAME\tHOST:PORT\tCONNECTED\tADVERTISING"); err != nil {
		return fmt.Errorf("write device table: %w", err)
	}

	for _, d := range devices {
		advertising := "none"
		if d.Service != types.ServiceUnknown {
			advertising = d.Service.String()
		}

		if _, err := fmt.Fprintf(w, "%s\t%s\t%t\t%s\n", d.Name, d.Endpoint(), d.Connected, advertising); err != nil {
			return fmt.Errorf("write device table: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush device table: %w", err)
	}

	return nil
}
