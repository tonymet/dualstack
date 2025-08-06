// ©️ 2025 Anthony Metzidis
/*
multilistener -- listen to ipv4 & ipv6 interfaces or multiple interfaces

go std library can listen to ALL interfaces but cannot listen to all local interfaces
by default.  

use multilistener.ListenLocalLoopback to return a single Listener for all ipv4 & ipv6 interfaces
*/
package multilistener

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

// MultiListener definition (as from previous answer)
type MultiListener struct {
	listeners []net.Listener
	acceptCh  chan net.Conn
	closeCh   chan struct{}
	wg        sync.WaitGroup
	closed    bool
}

type Addresses = []string

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
func NewLocalLoopback(port string) (*MultiListener, error) {
	return NewMultiListener(Addresses{"[::1]:" + port, "127.0.0.1:" + port})
}

// NewMultiListenerRaw returns a MultiListener wrapper of multiple listeners
// 
// useful when raw net.Addr is needed
func NewMultiListenerRaw(listeners []net.Listener) (*MultiListener, error) {
	// todo: see if listeners can be tested
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
func NewMultiListener(addrs Addresses) (*MultiListener, error) {
	// todo: support arbitrary
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

func (dl *MultiListener) Close() error {
	if dl.closed {
		return fmt.Errorf("already closed")
	}
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
	dl.closed = true
	return firstErr
}

func (dl *MultiListener) Addr() net.Addr {
	// Return one of the addresses for display
	return dl.listeners[0].Addr()
}
