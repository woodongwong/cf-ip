package ip

import (
	"net"
	"testing"
)

func parseCIDR(cidr string) *net.IPNet {
	_, ipNet, _ := net.ParseCIDR(cidr)
	return ipNet
}

// Test IsSubnet
func TestIsSubnet(t *testing.T) {
	testCases := []struct {
		ipNet1   *net.IPNet
		ipNet2   *net.IPNet
		expected bool
	}{
		{
			parseCIDR("192.168.1.0/24"),
			parseCIDR("192.168.2.0/24"),
			false,
		},
		{
			parseCIDR("192.168.1.0/31"),
			parseCIDR("192.168.1.0/32"),
			false,
		},
		{
			parseCIDR("192.168.1.0/24"),
			parseCIDR("192.168.2.0/23"),
			false,
		},
		{
			parseCIDR("10.0.0.0/16"),
			parseCIDR("10.0.0.0/8"),
			true,
		},
		{
			parseCIDR("2001:db8::/64"),
			parseCIDR("2001:db8:1234::/64"),
			false,
		},
	}

	for _, tc := range testCases {
		actual := IsSubnet(tc.ipNet1, tc.ipNet2)
		if actual != tc.expected {
			t.Errorf("IsSubnet(%q, %q) = %v; expected %v", tc.ipNet1, tc.ipNet2, actual, tc.expected)
		}
	}
}

// Test CIDR2IPRange
func TestCIDR2IPRange(t *testing.T) {
	testCases := []struct {
		ipNet   *net.IPNet
		startIP net.IP
		endIP   net.IP
	}{
		{
			parseCIDR("192.168.0.0/24"),
			net.ParseIP("192.168.0.0"),
			net.ParseIP("192.168.0.255"),
		},
		{
			parseCIDR("192.168.2.0/23"),
			net.ParseIP("192.168.2.0"),
			net.ParseIP("192.168.3.255"),
		},
		{
			parseCIDR("fe80:7a34:7a34:dfe2::/64"),
			net.ParseIP("fe80:7a34:7a34:dfe2::"),
			net.ParseIP("fe80:7a34:7a34:dfe2:ffff:ffff:ffff:ffff"),
		},
	}

	for _, tc := range testCases {
		startIP, endIP := CIDR2IPRange(tc.ipNet)
		if !net.IP.Equal(tc.startIP, startIP) || !net.IP.Equal(tc.endIP, endIP) {
			t.Errorf("IsSubnet(%v) = %v, %v; expected %v, %v", tc.ipNet, startIP, endIP, tc.startIP, tc.endIP)
		}
	}
}
