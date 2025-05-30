package domain

import (
	"context"
	"net"
)

// TCPServer defines the interface for a TCP server
type TCPServer interface {
	// Start begins listening for TCP connections on the specified address
	Start(ctx context.Context, address string) error

	// Stop gracefully shuts down the server
	Stop(ctx context.Context) error

	// Address returns the address the server is listening on
	Address() string
}

// ConnectionHandler defines the interface for handling TCP connections
type ConnectionHandler interface {
	// HandleConnection processes an incoming TCP connection
	HandleConnection(ctx context.Context, conn net.Conn) error
}

// ByteLogger defines the interface for logging received bytes
type ByteLogger interface {
	// LogBytes logs the received bytes with connection information
	LogBytes(connectionID string, data []byte) error
}
