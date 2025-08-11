package main

import (
	"fmt"
	"net"
)

func main() {
	ipStr := "2001:db8::68" // An example IPv6 address

	ip := net.ParseIP(ipStr)
	if ip == nil {
		fmt.Printf("Invalid IP address: %s\n", ipStr)
		return
	}

	// Check if the parsed IP is an IPv4 address.
	// If it's an IPv6 address, ip.To4() will return nil.
	// if ip.To4() != nil {
	// 	fmt.Printf("Parsed IP is IPv4: %s\n", ip.String())
	// } else if ip.IsLoopback() {
	// 	fmt.Printf("Parsed IP is IPv6 loopback: %s\n", ip.String())
	// } else if ip.IsPrivate() {
	// 	fmt.Printf("Parsed IP is IPv6 private: %s\n", ip.String())
	// } else if ip.IsGlobalUnicast() {
	// 	fmt.Printf("Parsed IP is IPv6 global unicast: %s\n", ip.String())
	// } else {
	// 	fmt.Printf("Parsed IP is IPv6 (non-IPv4, non-loopback, non-private, non-global unicast): %s\n", ip.String())
	// }

	// net.ParseIP handles both IPv4 and IPv6. To check that it's *not* IPv4,
	// you can see if the To4() method returns nil.
	fmt.Println(ip.String())
}
