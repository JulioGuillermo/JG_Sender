package connection

import (
	"net"
	"net/netip"
)

func GetIPS() []*netip.Prefix {
	ips := []*netip.Prefix{}

	inters, err := net.Interfaces()
	if err != nil {
		return ips
	}

	for _, i := range inters {
		addr, err := i.Addrs()
		if err == nil {
			for _, a := range addr {
				pre, err := netip.ParsePrefix(a.String())
				if err == nil {
					ip := pre.Addr()
					if ip.Is4() && !ip.IsLoopback() && !ip.IsMulticast() && !ip.IsUnspecified() {
						ips = append(ips, &pre)
					}
				}
			}
		}
	}

	return ips
}
