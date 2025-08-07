/*
Package multilistener -- listen to all loopback interfaces, or a mixed slice of
ipv4 & ipv6 interfaces

go std library (net.Listen("tcp", ":8080")) can listen to ALL interfaces 
but cannot listen to all local interfaces by default.  

use multilistener.ListenLocalLoopback to return a single Listener for all ipv4 &
ipv6 loopback interfaces

©️ 2025 Anthony Metzidis
*/
package multilistener

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

// MultiListener implements net.Listener interface 
//
// multiplexes multiple net.Listeners concurrently looping over Accept()
type MultiListener struct {
	listeners []net.Listener
	acceptCh  chan net.Conn
	closeCh   chan struct{}
	wg        sync.WaitGroup
}

type Addresses = []string

// net.Addr.Network() implementation
func (dl *MultiListener) Network() string {
	return "tcp+multi"
}
func (dl *MultiListener) String() string {
	if len(dl.listeners) < 1 {
		return ""
	}
	var (
		r strings.Builder
		i int
	)
	// pre-alloc by common IPv4 size
	r.Grow(len(dl.listeners) * 15)
	for i = 0; i < len(dl.listeners)-1; i++ {
		ln := dl.listeners[i]
		r.WriteString(ln.Addr().String())
		r.WriteString(",")
	}
	r.WriteString(dl.listeners[i].Addr().String())
	return r.String()
}

// NewLocalLoopback returns Multilistener on ipv6 & ipv4 loopback addresses
//
// ipv6 is the preferred address when Addr() is called
func NewLocalLoopback(port string) (*MultiListener, error) {
	return NewMultiListener(Addresses{"[::1]:" + port, "127.0.0.1:" + port})
}

// NewMultiListenerRaw returns a MultiListener wrapper of multiple listeners
// 
// useful when raw net.Addr or net.Listener is needed
func NewMultiListenerRaw(listeners []net.Listener) (*MultiListener, error) {
	dl := &MultiListener{
		listeners: listeners,
		acceptCh:  make(chan net.Conn),
		closeCh:   make(chan struct{}),
	}
	for _, l := range dl.listeners {
		dl.wg.Add(1)
		go dl.acceptLoop(l)
	}
	return dl, nil
}

// NewMultiListener returns multilistener over slice of []string 
//
// see net.Dial and net.Listen for the string format of the address
// e.g. "[::1]:8080" for ipv6 and "127.0.0.1:8080" for ipv4
func NewMultiListener(addrs Addresses) (*MultiListener, error) {
	var listeners = make([]net.Listener, 0, len(addrs))
	for _, addr := range addrs {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, fmt.Errorf("listen error: %v", err)
		}
		listeners = append(listeners, ln)
	}
	dl := &MultiListener{
		listeners: listeners,
		acceptCh:  make(chan net.Conn),
		closeCh:   make(chan struct{}),
	}
	for _, l := range dl.listeners {
		dl.wg.Add(1)
		go dl.acceptLoop(l)
	}
	return dl, nil
}

// AllAddr returns all the addresses, comma-separated
//
// NOTE: NOT A VALID IP ADDRESS . Use Addr() for a valid address
func (dl *MultiListener) AllAddr() net.Addr {
	return dl
}

func (dl *MultiListener) acceptLoop(l net.Listener) {
	defer dl.wg.Done()
	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-dl.closeCh:
				return
			default:
				continue
			}
		}
		dl.acceptCh <- conn
	}
}

func (dl *MultiListener) Accept() (net.Conn, error) {
	select {
	case conn := <-dl.acceptCh:
		return conn, nil
	case <-dl.closeCh:
		return nil, fmt.Errorf("listener closed")
	}
}

// Close closes all internal channels
//
// do not defer Close() if passing to http.Server 
func (dl *MultiListener) Close() error {
	close(dl.closeCh)
	// todo clean up ugly
	var firstErr error
	for _, l := range dl.listeners {
		if err := l.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	dl.wg.Wait()
	close(dl.acceptCh)
	return firstErr
}

// Addr returns the preferred (first) interface Addr
func (dl *MultiListener) Addr() net.Addr {
	// Return one of the addresses for display
	return dl.listeners[0].Addr()
}
