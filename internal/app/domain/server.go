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

// QueryLogger defines the interface for logging SQL queries
type QueryLogger interface {
	// LogQuery logs a SQL query with connection information
	LogQuery(connectionID string, query string) error

	// LogProtocolMessage logs other protocol messages (startup, auth, etc.)
	LogProtocolMessage(connectionID string, messageType string, details map[string]interface{}) error
}
