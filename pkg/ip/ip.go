package ip

import (
	"net"
)

// CIDR2IPRange takes an IP network represented in CIDR notation and
// returns the start and end IP addresses of the range it represents.
func CIDR2IPRange(ipNet *net.IPNet) (startIP net.IP, endIP net.IP) {
	endId := make(net.IP, len(ipNet.IP))
	for i := range ipNet.Mask {
		endId[i] = ipNet.IP[i] | ^ipNet.Mask[i]
	}

	return ipNet.IP, endId
}

// IsSubnet checks if ipNet1 is a subnet of ipNet2.
// It returns true if ipNet1 is a subnet of ipNet2 and false otherwise.
func IsSubnet(ipNet1, ipNet2 *net.IPNet) bool {
	isSubnet := ipNet2.Contains(ipNet1.IP)

	if isSubnet {
		for i := range ipNet2.Mask {
			if ipNet2.Mask[i] > ipNet1.Mask[i] {
				return false
			}
		}
	}
	return isSubnet
}
