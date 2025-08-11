package a

import (
	"fmt"
	"net"
)

func validUsage() {
	ip := net.ParseIP("::1")
	if ip == nil {
		return
	}
	// Correctly check for an IPv4 address.
	if ip4 := ip.To4(); ip4 != nil {
		fmt.Println("This is a valid IPv4 address.")
	}

	// Correctly use a loop that works for both IPv4 and IPv6.
	for i := range ip {
		_ = ip[i]
	}
}

func invalidSlice() {
	ip := net.ParseIP("192.168.1.1")
	if ip == nil {
		return
	}
	// This slice assumes a fixed length of 4, which is an error.
	_ = ip[0:4] // want "fixed-length slice of 4 on a net.IP variable may fail with IPv6"
}

func invalidIndex() {
	ip := net.ParseIP("::ffff:192.0.2.1")
	if ip == nil {
		return
	}
	// This index assumes the IP is 4 bytes long and will panic on IPv6.
	_ = ip[3] // want "fixed index on a net.IP variable may be an IPv4 assumption"
	// This index is also a common IPv4 assumption.
	_ = ip[4] // want "fixed index on a net.IP variable may be an IPv4 assumption"
}
