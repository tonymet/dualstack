// middleware/testing package with mock interfaces for testing net.Listener
package testing

import (
	"log"
	"net"
)

// MockConn is a fake net.Conn for testing purposes.
type MockConn struct {
	net.Conn
	remoteAddr *net.TCPAddr
}

func (m *MockConn) RemoteAddr() net.Addr {
	return m.remoteAddr
}

func (m *MockConn) Close() error {
	log.Printf("Mock connection from %s closed.", m.remoteAddr.IP)
	return nil
}

// MockListener is a fake net.Listener for testing purposes.
type MockListener struct {
	net.Listener
	connChan chan net.Conn
}

func NewMockListener() *MockListener {
	return &MockListener{
		connChan: make(chan net.Conn),
	}
}

func (m *MockListener) Accept() (net.Conn, error) {
	conn := <-m.connChan
	return conn, nil
}

func (m *MockListener) Addr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
}

func (m *MockListener) Close() error {
	close(m.connChan)
	return nil
}
