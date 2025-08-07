package main

import (
    "fmt"
    "log"
    "net"
)

func main() {
    // This listener is bound to the IPv6 loopback address and will accept
    // connections from both [::1] (IPv6) and 127.0.0.1 (IPv4).
    ln, err := net.Listen("tcp", "[::1]:8080")
    if err != nil {
        log.Fatal(err)
    }
    defer ln.Close()

    fmt.Println("Listening on [::1]:8080. Try connecting with 'telnet 127.0.0.1 8080' and 'telnet ::1 8080'")

    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Println(err)
            continue
        }
        fmt.Printf("Accepted connection from %s\n", conn.RemoteAddr())
        conn.Close()
    }
}