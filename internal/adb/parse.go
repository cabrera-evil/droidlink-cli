package adb

import (
	"net"
	"strconv"
	"strings"

	"droidlink/internal/types"
)

const (
	devicesHeader = "List of devices attached"
	stateDevice   = "device"

	// minDeviceFields is "<serial> <state>", the minimum adb devices -l
	// prints per device line.
	minDeviceFields = 2
)

// ParseDevicesOutput parses `adb devices -l` stdout into every device adb
// currently knows about, including offline/unauthorized entries (see
// types.Device.State) — callers that only care about usable devices should
// filter on Connected (state == "device").
func ParseDevicesOutput(output string) []types.Device {
	var devices []types.Device

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == devicesHeader {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < minDeviceFields {
			continue
		}

		devices = append(devices, deviceFromFields(fields))
	}

	return devices
}

func deviceFromFields(fields []string) types.Device {
	serial, state := fields[0], fields[1]

	device := types.Device{
		Name:      serial,
		Serial:    serial,
		State:     state,
		Connected: state == stateDevice,
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
