package adb

import (
	"net"
	"strconv"
	"strings"

	"droidlink/internal/types"
)

const devicesHeader = "List of devices attached"

// ParseDevicesOutput parses `adb devices -l` stdout into devices currently
// known to adb. Only entries in the "device" state are included; offline or
// unauthorized entries aren't usable for connect/reverse/disconnect and are
// skipped.
func ParseDevicesOutput(output string) []types.Device {
	var devices []types.Device

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == devicesHeader {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 || fields[1] != "device" {
			continue
		}

		devices = append(devices, deviceFromFields(fields))
	}

	return devices
}

func deviceFromFields(fields []string) types.Device {
	serial := fields[0]

	device := types.Device{
		Name:      serial,
		Serial:    serial,
		Connected: true,
	}

	if host, portStr, splitErr := net.SplitHostPort(serial); splitErr == nil {
		if port, atoiErr := strconv.Atoi(portStr); atoiErr == nil {
			device.Host = host
			device.Port = port
		}
	}

	for _, field := range fields[2:] {
		if model, ok := strings.CutPrefix(field, "model:"); ok {
			device.Name = model
		}
	}

	return device
}
