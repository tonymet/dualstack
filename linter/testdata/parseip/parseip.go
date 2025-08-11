package main

import (
	"fmt"
	"net"
)

func main() {
	ipStr := "2001:db8::68" // An example IPv6 address

	ip := net.ParseIP(ipStr) // want "call to `net.ParseIP` should be followed by a check for IPv4 or handle IPv6 compatibility"
	if ip == nil {
		fmt.Printf("Invalid IP address: %s\n", ipStr)
		return
	}
	fmt.Println(ip.String())
}

func GoodIpv4() {
	ipStr := "2001:db8::68"

	ip := net.ParseIP(ipStr)
	if ip == nil {
		fmt.Printf("Invalid IP address: %s\n", ipStr)
		return
	}

	if ip.To4() != nil {
		fmt.Printf("Parsed IP is IPv4: %s\n", ip.String())
	}
	fmt.Println(ip.String())
}
