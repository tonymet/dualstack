package middleware

import (
	"net"
	"net/http"
)

// LocalOnlyMiddleware checks if a request is coming from a local interface
// by accessing the actual connection's remote address. This version uses a
// type assertion to get the binary IP address directly, avoiding string parsing.
func LocalOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var conn net.Conn
		hijacker, ok := w.(http.Hijacker)
		if ok {
			// Hijack the connection to get the underlying net.Conn
			var err error
			conn, _, err = hijacker.Hijack()
			if err != nil {
				http.Error(w, "Forbidden: Failed to hijack connection", http.StatusForbidden)
				return
			}
		}

		var ip net.IP
		if conn != nil {
			// Get the remote address from the connection.
			remoteAddr := conn.RemoteAddr()
			// Use a type assertion to check if the address is a TCP address.
			if tcpAddr, ok := remoteAddr.(*net.TCPAddr); ok {
				// If it is, we can access its IP field directly.
				ip = tcpAddr.IP
			}
		}

		if ip == nil {
			// Fallback to r.RemoteAddr if we couldn't get the underlying connection
			// or if the type assertion failed. This makes the middleware more robust.
			ipStr, _, err := net.SplitHostPort(r.RemoteAddr)
			if err == nil {
				ip = net.ParseIP(ipStr)
			}
		}

		// Now we have the net.IP object, so we can use its methods directly.
		if ip == nil || !ip.IsLoopback() {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// If the checks pass, proceed to the next handler.
		next.ServeHTTP(w, r)
	})
}
