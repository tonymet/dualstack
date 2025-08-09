package middleware

import (
	"errors"
	"net"
)

var ErrFirewall = errors.New("blocked remote addr")
var ErrIPError = errors.New("error reading remote IP")

// FirewallListener wraps a net.Listener to block non-localhost connections.
type FirewallListener struct {
	net.Listener
}

// Accept is the middleware for our firewall. It wraps the underlying Accept call,
// inspects the connection's IP address, and blocks it if it's not a localhost address.
func (fl *FirewallListener) Accept() (net.Conn, error) {
	conn, err := fl.Listener.Accept()
	if err != nil {
		return nil, err
	}
	tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr)
	// if we can't read the IP, block with IPError
	if !ok {
		conn.Close()
		return conn, &net.OpError{Err: ErrIPError}
	}
	if !tcpAddr.IP.IsLoopback() {
		conn.Close()
		return conn, &net.OpError{Err: ErrFirewall}
	}
	return conn, nil
}

// NewFirewallListener creates and returns a new FirewallListener that wraps an existing listener.
func NewFirewallListener(l net.Listener) *FirewallListener {
	return &FirewallListener{l}
}
