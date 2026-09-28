package discovery

import (
	"time"

	"github.com/libp2p/zeroconf/v2"

	"droidlink/internal/types"
)

// fromServiceEntry converts a zeroconf.ServiceEntry into a types.Device,
// preferring an IPv4 address and falling back to IPv6.
func fromServiceEntry(entry *zeroconf.ServiceEntry, svcType types.ServiceType) types.Device {
	host := ""

	switch {
	case len(entry.AddrIPv4) > 0:
		host = entry.AddrIPv4[0].String()
	case len(entry.AddrIPv6) > 0:
		host = entry.AddrIPv6[0].String()
	}

	return types.Device{
		Name:         entry.Instance,
		Host:         host,
		Port:         entry.Port,
		Service:      svcType,
		Paired:       svcType == types.ServiceConnect,
		DiscoveredAt: time.Now(),
	}
}
