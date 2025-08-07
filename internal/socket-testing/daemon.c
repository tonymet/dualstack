#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>

#define PORT 8080
#define BUFFER_SIZE 1024

int main() {
    int server_fd, new_socket;
    struct sockaddr_in6 address;
    int addrlen = sizeof(address);
    char buffer[BUFFER_SIZE] = {0};
    const char *http_response = "HTTP/1.1 200 OK\nContent-Type: text/plain\nContent-Length: 14\n\nHello, world!\n";
    
    // Create an IPv6 socket
    if ((server_fd = socket(AF_INET6, SOCK_STREAM, 0)) == 0) {
        perror("socket failed");
        exit(EXIT_FAILURE);
    }
    
    // Set SO_REUSEADDR to reuse the address immediately
    int opt = 1;
    if (setsockopt(server_fd, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt))) {
        perror("setsockopt SO_REUSEADDR failed");
        close(server_fd);
        exit(EXIT_FAILURE);
    }

    // Set IPV6_V6ONLY to 0 to allow dual-stack operation.
    // This is the crucial part for the test. If this is 0,
    // the IPv6 socket can also accept IPv4 connections.
    // If it were 1, it would only accept IPv6 connections.
    int v6only = 0;
    if (setsockopt(server_fd, IPPROTO_IPV6, IPV6_V6ONLY, &v6only, sizeof(v6only))) {
        perror("setsockopt IPV6_V6ONLY failed");
        close(server_fd);
        exit(EXIT_FAILURE);
    }
    int returned_v6only;
    socklen_t len = sizeof(returned_v6only);
    if (getsockopt(server_fd, IPPROTO_IPV6, IPV6_V6ONLY, &returned_v6only, &len) == 0) {
        printf("IPV6_V6ONLY is currently set to: %d\n", returned_v6only);
    } else {
        perror("getsockopt IPV6_V6ONLY failed");
    }

    // Configure the server address
    // Use IN6ADDR_LOOPBACK_INIT to bind to [::1]
    memset(&address, 0, sizeof(address));
    address.sin6_family = AF_INET6;
    address.sin6_addr = in6addr_loopback; // This is the [::1] address
    address.sin6_port = htons(PORT);
    
    // Bind the socket to the configured address
    if (bind(server_fd, (struct sockaddr *)&address, addrlen) < 0) {
        perror("bind failed");
        close(server_fd);
        exit(EXIT_FAILURE);
    }
    
    // Start listening for incoming connections
    if (listen(server_fd, 3) < 0) {
        perror("listen failed");
        close(server_fd);
        exit(EXIT_FAILURE);
    }
    
    printf("Server listening on [::1] port %d\n", PORT);
    printf("Test with curl: \n");
    printf("  curl -6 http://[::1]:%d\n", PORT);
    printf("  curl -4 http://127.0.0.1:%d\n\n", PORT);

    while (1) {
        // Accept a new connection
        if ((new_socket = accept(server_fd, (struct sockaddr *)&address, (socklen_t*)&addrlen)) < 0) {
            perror("accept failed");
            continue;
        }

        // Read the incoming request
        read(new_socket, buffer, BUFFER_SIZE);
        printf("Request received:\n---\n%s\n---\n", buffer);
        
        // Send the HTTP response
        write(new_socket, http_response, strlen(http_response));
        
        // Close the client socket
        close(new_socket);
    }
    
    // Close the server socket (unreachable in this infinite loop)
    close(server_fd);
    
    return 0;
}
