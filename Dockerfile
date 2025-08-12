# Start with a base image that has the Go runtime
FROM golang:alpine AS builder

# Set the working directory
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o /ip6check ./cmd/ip6check

FROM golang:alpine
COPY --from=builder /ip6check /ip6check
VOLUME ["/workspace"]
WORKDIR /workspace
# Set the entrypoint for the action
ENTRYPOINT ["/ip6check"]