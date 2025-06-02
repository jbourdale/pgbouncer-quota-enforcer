package adapters

import (
	"context"
	"fmt"
	"net"
	"pgbouncer-quota-enforcer/internal/app/domain"
	"pgbouncer-quota-enforcer/pkg/logger"
	"sync"
)

// StandardTCPServer implements domain.TCPServer
type StandardTCPServer struct {
	handler   domain.ConnectionHandler
	logger    logger.Logger
	listener  net.Listener
	wg        sync.WaitGroup
	mu        sync.RWMutex
	address   string
	isRunning bool
}

// NewStandardTCPServer creates a new StandardTCPServer
func NewStandardTCPServer(handler domain.ConnectionHandler, log logger.Logger) domain.TCPServer {
	return &StandardTCPServer{
		handler: handler,
		logger:  log,
	}
}

// Start begins listening for TCP connections on the specified address
func (s *StandardTCPServer) Start(ctx context.Context, address string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("server is already running")
	}

	// Create TCP listener
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", address, err)
	}

	s.listener = listener
	s.address = listener.Addr().String()
	s.isRunning = true

	s.logger.Info("TCP server started", "address", s.address)

	// Start accepting connections in a goroutine
	s.wg.Add(1)
	go s.acceptConnections(ctx)

	return nil
}

// Stop gracefully shuts down the server
func (s *StandardTCPServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	s.logger.Info("Stopping TCP server")

	// Close listener to stop accepting new connections
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			s.logger.Error("Error closing listener: %v", err)
		}
	}

	s.isRunning = false

	// Wait for all connection handlers to finish with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("TCP server stopped gracefully")
		return nil
	case <-ctx.Done():
		s.logger.Error("Timeout waiting for connections to close")
		return ctx.Err()
	}
}

// Address returns the address the server is listening on
func (s *StandardTCPServer) Address() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.address
}

// acceptConnections accepts incoming connections and spawns handlers
func (s *StandardTCPServer) acceptConnections(ctx context.Context) {
	defer s.wg.Done()

	for {
		// Accept connection with context awareness
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				s.logger.Info("Stopped accepting connections due to context cancellation")
				return
			default:
				// Check if server is still running
				s.mu.RLock()
				isRunning := s.isRunning
				s.mu.RUnlock()

				if !isRunning {
					s.logger.Info("Stopped accepting connections (server stopped)")
					return
				}

				s.logger.Error("Error accepting connection: %v", err)
				continue
			}
		}

		// Handle connection in a separate goroutine
		s.wg.Add(1)
		go func(c net.Conn) {
			defer s.wg.Done()

			if err := s.handler.HandleConnection(ctx, c); err != nil {
				s.logger.Error("Error handling connection: %v", err)
			}
		}(conn)
	}
}
