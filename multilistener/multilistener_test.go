package multilistener

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

// TestListenOnlyIPV6 sanity check that IPv6 Socket cannot receive ipv4 requests
//
// AI claimed IPV6 ::1 C sockets also accepted ipv4.  This test confirms go sockets do not
// see also cmd/simple-listener and test with telnet
// I feel like I'm taking crazy pills.
func TestListenOnlyIPV6(t *testing.T) {
	ready := make(chan struct{})
	port := "8084"
	ml, err := net.Listen("tcp", "[::1]:"+port) // Listen only on IPv6 loopback
	if err != nil {
		t.Fatalf("Failed to create MultiListener for IPv6 only: %v", err)
	}

	go func() {
		close(ready)
		if err := http.Serve(ml, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := io.WriteString(w, "Hello from IPv6 Listener!"); err != nil {
				t.Errorf("error writing string: %v", err)
			}
		})); err != nil {
			t.Errorf("http serve error: %v", err)
		}
	}()
	// server is ready
	<-ready
	// Attempt to connect to the IPv4 loopback address
	testAddr := "127.0.0.1:" + port
	t.Logf("Attempting to connect to %s (IPv4) to an IPv6-only listener", testAddr)
	resp, err := http.Get("http://" + testAddr)
	if err != nil {
		t.Logf("Expected error when connecting to IPv4 address on IPv6-only listener: %v", err)
		// This is the expected behavior, as the listener is only on IPv6
		return
	}
	defer resp.Body.Close()
	// If we reach here, it means the IPv4 connection succeeded unexpectedly
	t.Fatalf("Unexpected success connecting to IPv4 address %s on an IPv6-only listener. Status: %d", testAddr, resp.StatusCode)
}

func TestMultiListener(t *testing.T) {
	ready := make(chan struct{})
	ml, err := NewLocalLoopback("0") // Use port 0 to get a random available port
	if err != nil {
		t.Fatalf("Failed to create MultiListener: %v", err)
	}
	// Start an HTTP server on the MultiListener
	go func() {
		close(ready)
		if err := http.Serve(ml, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := io.WriteString(w, "Hello from MultiListener!"); err != nil {
				t.Errorf("error writeString")
			}
		})); err != nil {
			t.Errorf("http serve error")
		}
	}()

	<-ready
	// Get the addresses of the listeners
	addrs := ml.AllAddr().String()
	t.Logf("MultiListener serving on: %s", addrs)

	// Test connecting to each address
	for _, addr := range ml.listeners {
		testAddr := addr.Addr().String()
		t.Logf("Attempting to connect to %s", testAddr)

		// Give the server a moment to start
		time.Sleep(100 * time.Millisecond)

		resp, err := http.Get("http://" + testAddr)
		if err != nil {
			t.Errorf("Failed to connect to %s: %v", testAddr, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status OK for %s, got %d", testAddr, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read response body from %s: %v", testAddr, err)
		}
		expectedBody := "Hello from MultiListener!"
		if string(body) != expectedBody {
			t.Errorf("Expected body 'package multilistener")
		}
	}
}

// NewLocalLoopback  when you want to listen to ipv6 & ipv4 loopback with one listener
func ExampleNewLocalLoopback() {
	http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World!\n")
	})

	dual, err := NewLocalLoopback("8129")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Serving HTTP %+v\n", dual.AllAddr())
	fmt.Printf("Preferred Addr: %+v\n", dual.Addr())
	go http.Serve(dual, nil) //nolint:errcheck
	// Output:
	// Serving HTTP [::1]:8129,127.0.0.1:8129
	// Preferred Addr: [::1]:8129
}

func TestMultiListenerDoubleClose(t *testing.T) {
	ml, err := NewLocalLoopback("0")
	if err != nil {
		t.Fatalf("Failed to create MultiListener: %v", err)
	}

	// First close should be successful
	err = ml.Close()
	if err != nil {
		t.Errorf("First close failed unexpectedly: %v", err)
	}

	// Second close should ideally not panic and return an error or be a no-op
	// We use a defer func to recover from a potential panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MultiListener.Close() panicked on second call: %v", r)
		}
	}()

	err = ml.Close()
	if err == nil {
		t.Errorf("Second close succeeded unexpectedly, should return an error or be a no-op")
	}
	t.Logf("Second close returned: %v (expected error or no-op)", err)
}

func TestMultiListenerCloseBeforeServe(t *testing.T) {
	//t.Skip()
	ml, err := NewLocalLoopback("0")
	if err != nil {
		t.Fatalf("Failed to create MultiListener: %v", err)
	}

	// Close the listener immediately after creation, before serving
	err = ml.Close()
	if err != nil {
		t.Fatalf("Failed to close MultiListener before serving: %v", err)
	}

	// Attempt to serve on a closed listener, should return an error
	serveErr := http.Serve(ml, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("Handler should not be called on a closed listener")
	}))

	if serveErr == nil {
		t.Errorf("http.Serve on a closed listener did not return an error")
	} else {
		t.Logf("http.Serve on a closed listener returned expected error: %v", serveErr)
	}
}
