// middleware protects http handlers against remote requests
//
// useful for dual-stack [::] listeners for local services
package middleware

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

// A custom ResponseWriter that implements http.Hijacker to test our middleware.
// This is needed to get to the underlying connection in the middleware code.
type mockHijackResponseWriter struct {
	httptest.ResponseRecorder
	conn net.Conn
}

func (m *mockHijackResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return m.conn, nil, nil
}

// TestLocalOnlyMiddleware is the main test function for our middleware.
func TestLocalOnlyMiddleware(t *testing.T) {
	// A simple handler to confirm the middleware passed the request through.
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Allowed")) //nolint:errcheck
	})

	// Table-driven tests for various IP addresses.
	testCases := []struct {
		name         string
		remoteAddr   string
		expectStatus int
		expectBody   string
		// Use a custom conn for the hijacker test case
		hijackConn net.Conn
	}{
		{
			name:         "IPv4 Loopback",
			remoteAddr:   "127.0.0.1:12345",
			expectStatus: http.StatusOK,
			expectBody:   "Allowed",
		},
		{
			name:         "IPv6 Loopback",
			remoteAddr:   "[::1]:12345",
			expectStatus: http.StatusOK,
			expectBody:   "Allowed",
		},
		{
			name:         "Non-Local IPv4",
			remoteAddr:   "192.168.1.1:12345",
			expectStatus: http.StatusForbidden,
			expectBody:   "Forbidden\n",
		},
		{
			name:         "Non-Local IPv6",
			remoteAddr:   "[2001:db8::1]:12345",
			expectStatus: http.StatusForbidden,
			expectBody:   "Forbidden\n",
		},
		{
			name:         "Malformed RemoteAddr",
			remoteAddr:   "invalid-address",
			expectStatus: http.StatusForbidden,
			expectBody:   "Forbidden\n",
		},
		{
			name:         "Hijack with IPv4 Loopback",
			remoteAddr:   "192.168.1.1:12345", // RemoteAddr is a fake, but the conn is local.
			hijackConn:   &mockConn{remoteAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 54321}},
			expectStatus: http.StatusOK,
			expectBody:   "Allowed",
		},
		{
			name:         "Hijack with Non-Local IPv4",
			remoteAddr:   "127.0.0.1:12345", // RemoteAddr is a fake, but the conn is not local.
			hijackConn:   &mockConn{remoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 54321}},
			expectStatus: http.StatusForbidden,
			expectBody:   "Forbidden\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a request and a recorder for the test.
			req := httptest.NewRequest("GET", "http://example.com/", nil)
			req.RemoteAddr = tc.remoteAddr

			var w http.ResponseWriter
			if tc.hijackConn != nil {
				// Use the mockHijackResponseWriter for hijack tests.
				w = &mockHijackResponseWriter{
					ResponseRecorder: *httptest.NewRecorder(),
					conn:             tc.hijackConn,
				}
			} else {
				// Use the standard ResponseRecorder for other tests.
				w = httptest.NewRecorder()
			}

			// Call the middleware with the test handler.
			LocalOnlyMiddleware(nextHandler).ServeHTTP(w, req)

			// Get the response and check the status code and body.
			var resp *http.Response
			var body string
			switch v := w.(type) {
			case *httptest.ResponseRecorder:
				resp = v.Result()
				body = v.Body.String()
			case *mockHijackResponseWriter:
				resp = v.Result()
				body = v.ResponseRecorder.Body.String()
			default:
				t.Fatalf("unexpected ResponseWriter type: %T", w)
			}

			if resp.StatusCode != tc.expectStatus {
				t.Errorf("Expected status %d, got %d", tc.expectStatus, resp.StatusCode)
			}
			if body != tc.expectBody {
				t.Errorf("Expected body '%s', got '%s'", tc.expectBody, body)
			}
		})
	}
}

// ExampleLocalOnlyMiddleware example of wrapping a common handler
func ExampleLocalOnlyMiddleware() {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Allowed")) // nolint:errcheck
	})
	protectedHandler := LocalOnlyMiddleware(nextHandler)
	// for common apps use http.Handle("/", protectedHandler)
	ts := httptest.NewServer(protectedHandler)
	defer ts.Close()

	// Create a request with a local remote address
	reqLocal := httptest.NewRequest("GET", ts.URL, nil)
	reqLocal.RemoteAddr = "127.0.0.1:12345" // Simulate a local client

	// Create a response recorder
	rrLocal := httptest.NewRecorder()

	// Serve the request through the middleware
	protectedHandler.ServeHTTP(rrLocal, reqLocal)

	fmt.Printf("Local Request Status: %d\n", rrLocal.Result().StatusCode)
	fmt.Printf("Local Request Body: %s\n", rrLocal.Body.String())

	// --- Test Case 2: Request from a non-local IP (should be forbidden) ---
	// Create a request with a non-local remote address
	reqRemote := httptest.NewRequest("GET", ts.URL, nil)
	reqRemote.RemoteAddr = "192.168.1.100:54321" // Simulate a remote client

	// Create a response recorder
	rrRemote := httptest.NewRecorder()

	// Serve the request through the middleware
	protectedHandler.ServeHTTP(rrRemote, reqRemote)

	fmt.Printf("Remote Request Status: %d\n", rrRemote.Result().StatusCode)
	fmt.Printf("Remote Request Body: %s\n", rrRemote.Body.String())
	// Output:
	// Local Request Status: 200
	// Local Request Body: Allowed
	// Remote Request Status: 403
	// Remote Request Body: Forbidden
}
