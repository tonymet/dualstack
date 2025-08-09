package middleware

import (
	"log"
	"net"
	"net/http"
	"testing"
	"time"
)

// mockConn is a fake net.Conn for testing purposes.
type mockConn struct {
	net.Conn
	remoteAddr *net.TCPAddr
}

func (m *mockConn) RemoteAddr() net.Addr {
	return m.remoteAddr
}

func (m *mockConn) Close() error {
	log.Printf("Mock connection from %s closed.", m.remoteAddr.IP)
	return nil
}

// mockListener is a fake net.Listener for testing purposes.
type mockListener struct {
	net.Listener
	connChan chan net.Conn
}

func newMockListener() *mockListener {
	return &mockListener{
		connChan: make(chan net.Conn),
	}
}

func (m *mockListener) Accept() (net.Conn, error) {
	conn := <-m.connChan
	return conn, nil
}

func (m *mockListener) Addr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
}

func (m *mockListener) Close() error {
	close(m.connChan)
	return nil
}

// TestFirewallListener verifies that the FirewallListener correctly
// allows connections from loopback addresses and blocks others.
func TestFirewallListener2(t *testing.T) {
	// Create a standard TCP listener on a loopback address.
	// We use port 0 to let the OS choose an available port.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	// Get the address of the listener.
	addr := l.Addr().String()

	// Wrap the standard listener with our firewall middleware.
	firewallListener := NewFirewallListener(l)
	defer firewallListener.Close() // Ensure the listener is closed after the test

	// A simple HTTP handler for our test server.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK")) // nolint:errcheck
	})

	// Start the test server in a goroutine using the firewall listener.
	go func() {
		// http.Serve will block, so we run it in a goroutine.
		err := http.Serve(firewallListener, nil)
		if err != nil && err != http.ErrServerClosed {
			log.Printf("Test server stopped with error: %v", err)
		}
	}()

	// Give the server a moment to start.
	time.Sleep(100 * time.Millisecond)

	// --- Subtest 1: Allowed connection from localhost (127.0.0.1) ---
	t.Run("Allowed connection from localhost", func(t *testing.T) {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatalf("Expected connection from localhost to be successful, but got error: %v", err)
		}
		defer conn.Close()
	})

	// --- Subtest 2: Blocked connection from a non-localhost address ---
	t.Run("Blocked connection from non-localhost", func(t *testing.T) {
		// Create a mock listener and a mock connection with a remote, non-loopback address.
		ml := newMockListener()
		remoteAddr := &net.TCPAddr{IP: net.ParseIP("192.168.1.100"), Port: 54321}
		mockConn := &mockConn{remoteAddr: remoteAddr}

		// Create a FirewallListener that wraps our mock listener
		fml := FirewallListener{ml}

		go func() {
			ml.connChan <- mockConn
		}()

		_, err := fml.Accept()
		if err == nil {
			t.Fatalf("Expected Accept to fail: %v", err)
		}

	})
}
