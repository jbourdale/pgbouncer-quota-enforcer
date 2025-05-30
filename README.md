# PgBouncer Quota Enforcer

A high-performance quota enforcement service that processes PostgreSQL query events to track and enforce database usage quotas. This project implements a TCP server that receives and logs PostgreSQL protocol messages as the first step toward building a complete quota enforcement system.

## Features

- **TCP Server**: Accepts connections and logs all received bytes
- **Structured Logging**: Logs with connection IDs, hex previews, and ASCII previews
- **Hexagonal Architecture**: Clean separation of concerns with domain, application, and infrastructure layers
- **Graceful Shutdown**: Proper connection handling and server shutdown
- **Comprehensive Testing**: Unit tests for all components
- **CLI Interface**: Command-line interface with Cobra

## Architecture

The project follows hexagonal architecture principles:

```
cmd/
├── main.go                          # Entry point
internal/
├── app/
│   ├── domain/                      # Domain entities and interfaces
│   │   └── server.go               # TCPServer, ConnectionHandler, ByteLogger interfaces
│   ├── infrastructure/             # External adapters
│   │   ├── tcp_server.go          # TCP server implementation
│   │   ├── connection_handler.go   # Connection handling logic
│   │   ├── byte_logger.go         # Byte logging implementation
│   │   └── *_test.go              # Unit tests
│   ├── interfaces/                 # Primary adapters
│   │   └── cmd.go                 # CLI commands
│   └── service.go                 # Application service layer
pkg/
├── logger/                         # Shared logging utilities
│   ├── logger.go
│   └── logger_test.go
```

## Getting Started

### Prerequisites

- Go 1.21 or later

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd pgbouncer-quota-enforcer
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o bin/pgbouncer-quota-enforcer ./cmd
```

### Usage

#### Start the TCP Server

```bash
# Start server on default port (5432)
./bin/pgbouncer-quota-enforcer server

# Start server on custom port
./bin/pgbouncer-quota-enforcer server --address :8080

# Get help
./bin/pgbouncer-quota-enforcer server --help
```

#### Test the Server

You can test the server by sending data to it:

```bash
# Send test data
printf "Hello PostgreSQL Server" | nc localhost 8080

# Send binary data
echo -e "\x00\x01\x02\xFF" | nc localhost 8080
```

### Example Output

When you connect to the server and send data, you'll see structured logs like:

```
[2025-05-30 16:19:24.098] INFO: Starting server service [address=:8080]
[2025-05-30 16:19:24.100] INFO: TCP server started [address=[::]:8080]
TCP server started on [::]:8080
Press Ctrl+C to stop the server

[2025-05-30 16:19:29.543] INFO: New connection established [connection_id=conn_1, remote_addr=[::1]:55115]
[2025-05-30 16:19:29.544] INFO: Received bytes [length=12, hex_preview=48656c6c6f20536572766572, ascii_preview="Hello Server"] [connection_id=conn_1]
[2025-05-30 16:19:29.544] DEBUG: Full packet hex dump [hex_data=48656c6c6f20536572766572] [connection_id=conn_1]
[2025-05-30 16:19:29.544] INFO: Connection closed by client [connection_id=conn_1, remote_addr=[::1]:55115]
[2025-05-30 16:19:29.544] INFO: Connection closed [connection_id=conn_1, remote_addr=[::1]:55115]
```

## Development

### Running Tests

```bash
# Run unit tests only
make test

# Run integration tests only (requires psql)
make test-integration

# Run all tests
make test-all

# Run tests with verbose output
go test ./... -v
go test -tags=integration ./test/... -v
```

#### Integration Tests

The project includes integration tests that use real PostgreSQL clients (`psql`) to test the complete flow:

- **Purpose**: Test receiving actual PostgreSQL wire protocol data
- **Requirements**: `psql` command must be available on the system
- **What it tests**:
  - Server startup and shutdown
  - Multiple concurrent connections
  - PostgreSQL protocol message reception
  - Byte logging with real protocol data
  - Connection lifecycle management

**Example Integration Test Output**:
```bash
make test-integration

# Shows real PostgreSQL protocol data:
[2025-05-30 16:28:46.807] INFO: Received bytes [length=8, hex_preview=0000000804d2162f, ascii_preview="......./"] [connection_id=conn_1]
[2025-05-30 16:28:46.807] DEBUG: Full packet hex dump [hex_data=0000000804d2162f] [connection_id=conn_1]
```

The captured hex data `0000000804d2162f` represents PostgreSQL's SSLRequest message, demonstrating that our server successfully receives and logs real PostgreSQL wire protocol data.

### Code Quality

The project includes comprehensive unit tests and follows Go best practices:

- **Domain Layer**: Pure business logic with no external dependencies
- **Infrastructure Layer**: Implementations of domain interfaces
- **Application Layer**: Service orchestration and dependency injection
- **Interface Layer**: CLI commands and external API adapters

### Key Components

#### Domain Interfaces

- `TCPServer`: Defines server lifecycle management
- `ConnectionHandler`: Handles individual TCP connections
- `ByteLogger`: Logs received bytes with structured output

#### Infrastructure Implementations

- `StandardTCPServer`: TCP server with graceful shutdown
- `LoggingConnectionHandler`: Reads and processes connection data
- `StandardByteLogger`: Structured logging with hex/ASCII previews

#### Features

- **Connection Management**: Unique connection IDs and proper cleanup
- **Byte Logging**: Hex and ASCII previews with configurable limits
- **Graceful Shutdown**: Context-aware shutdown with timeout handling
- **Error Handling**: Comprehensive error handling and logging
- **Performance**: 4KB read buffers and efficient byte processing

## Future Enhancements

This TCP server is the foundation for building a complete PostgreSQL quota enforcement system. Future enhancements will include:

1. **PostgreSQL Protocol Parsing**: Parse PostgreSQL wire protocol messages
2. **Query Analysis**: Extract and normalize SQL queries
3. **Cost Estimation**: Implement query cost calculation
4. **Quota Tracking**: Track usage per user/connection
5. **Enforcement Actions**: Implement connection limiting via PgBouncer API
6. **Caching Layer**: Redis integration for performance
7. **Metrics & Monitoring**: Prometheus metrics and health checks

## Contributing

1. Follow the hexagonal architecture principles
2. Write comprehensive unit tests for new functionality
3. Use structured logging with appropriate context
4. Handle errors gracefully and provide meaningful messages
5. Follow Go naming conventions and best practices

## License

[Add your license information here] 