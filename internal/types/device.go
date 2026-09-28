// Package types holds data structures shared across droidlink's internal
// packages (adb, discovery, ui, cli).
package types

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

// ServiceType identifies which mDNS service advertised a device.
type ServiceType int

// Recognized mDNS service types for Android Wireless Debugging.
const (
	ServiceUnknown ServiceType = iota
	ServicePairing
	ServiceConnect
)

// String implements fmt.Stringer.
func (s ServiceType) String() string {
	switch s {
	case ServicePairing:
		return "pairing"
	case ServiceConnect:
		return "connect"
	case ServiceUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// Device represents one discovered or adb-known Android device.
//
// Paired is a best-effort inference, not a real ADB concept: Android only
// advertises _adb-tls-pairing._tcp while the device's "pair with code" screen
// is open, and _adb-tls-connect._tcp whenever wireless debugging is toggled
// on regardless of pairing state. Paired really means "seen via the connect
// service", implying adb already trusts it enough to offer a connect port.
type Device struct {
	Name      string      `json:"name"`
	Host      string      `json:"host"`
	Port      int         `json:"port"`
	Service   ServiceType `json:"service"`
	Serial    string      `json:"serial,omitempty"`
	Paired    bool        `json:"paired"`
	Connected bool        `json:"connected"`
	// State is adb's own reported state (e.g. "device", "offline",
	// "unauthorized") when this device was seen via `adb devices -l`, and
	// empty when it's only known from mDNS discovery. Kept distinct from
	// Connected so a device that adb sees but hasn't fully authorized isn't
	// indistinguishable from one adb has never seen at all.
	State        string    `json:"state,omitempty"`
	DiscoveredAt time.Time `json:"discoveredAt,omitzero"`
}

// String returns a human-readable "name (host:port)" label for UI lists.
func (d Device) String() string {
	return fmt.Sprintf("%s (%s)", d.Name, d.Endpoint())
}

// Endpoint returns "host:port" suitable for adb commands.
func (d Device) Endpoint() string {
	return net.JoinHostPort(d.Host, strconv.Itoa(d.Port))
}
