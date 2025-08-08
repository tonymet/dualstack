package main

import (
	"fmt"
	"net"
)

func main() {
	ipStr := "2001:db8::68" // An example IPv6 address

	ip := net.ParseIP(ipStr)

	// net.ParseIP handles both IPv4 and IPv6. To check that it's *not* IPv4,
	// you can see if the To4() method returns nil.
	fmt.Printf(ip.String())
}
