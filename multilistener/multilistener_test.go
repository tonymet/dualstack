package multilistener

import (
	"io"
	"net/http"
	"sync"
	"fmt"
	"testing"
	"time"
)

func TestMultiListener(t *testing.T) {
	ml, err := NewLocalLoopback("8080") // Use port 0 to get a random available port
	if err != nil {
		t.Fatalf("Failed to create MultiListener: %v", err)
	}
	defer ml.Close()

	// Start an HTTP server on the MultiListener
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		http.Serve(ml, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "Hello from MultiListener!")
		}))
	}()

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
	dual, err := NewLocalLoopback("8080")
	if err != nil {
		panic(err)
	}
	defer dual.Close()
	fmt.Printf("Serving HTTP %+v\n", dual.AllAddr())
	fmt.Printf("Preferred Addr: %+v\n", dual.Addr())
	go http.Serve(dual, nil)
	// Output:
	// Serving HTTP [::1]:8080,127.0.0.1:8080
	// Preferred Addr: [::1]:8080
}