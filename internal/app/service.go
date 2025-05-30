package app

import (
	"context"
	"pgbouncer-quota-enforcer/internal/app/domain"
	"pgbouncer-quota-enforcer/internal/app/infrastructure"
	"pgbouncer-quota-enforcer/pkg/logger"
)

// ServerService provides the high-level application service for the TCP server
type ServerService struct {
	tcpServer domain.TCPServer
	logger    logger.Logger
}

// ServerConfig holds configuration for the server service
type ServerConfig struct {
	Address string
}

// NewServerService creates a new ServerService with all dependencies wired up
func NewServerService(config ServerConfig) *ServerService {
	// Create logger
	log := logger.NewSimpleLogger()

	// Create byte logger
	byteLogger := infrastructure.NewStandardByteLogger(log)

	// Create connection handler
	connHandler := infrastructure.NewLoggingConnectionHandler(byteLogger, log)

	// Create TCP server
	tcpServer := infrastructure.NewStandardTCPServer(connHandler, log)

	return &ServerService{
		tcpServer: tcpServer,
		logger:    log,
	}
}

// Start starts the TCP server
func (s *ServerService) Start(ctx context.Context, address string) error {
	s.logger.Info("Starting server service", "address", address)
	return s.tcpServer.Start(ctx, address)
}

// Stop stops the TCP server
func (s *ServerService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping server service")
	return s.tcpServer.Stop(ctx)
}

// Address returns the address the server is listening on
func (s *ServerService) Address() string {
	return s.tcpServer.Address()
}
