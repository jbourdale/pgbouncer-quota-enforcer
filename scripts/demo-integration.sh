#!/bin/bash

# Demo script for PgBouncer Quota Enforcer Integration Test
# This script demonstrates the TCP server receiving PostgreSQL protocol data

set -e

echo "🚀 PgBouncer Quota Enforcer - Integration Test Demo"
echo "=================================================="
echo

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo "❌ psql command not found. Please install PostgreSQL client tools."
    echo "   On macOS: brew install libpq"
    echo "   On Ubuntu: sudo apt-get install postgresql-client"
    exit 1
fi

echo "✅ psql found: $(which psql)"
echo

# Build the application
echo "🔨 Building application..."
make build
echo

# Run the integration test
echo "🧪 Running Integration Test..."
echo "This will start the TCP server and send PostgreSQL queries to it."
echo "Watch the logs to see the PostgreSQL wire protocol data being received."
echo
echo "Press Ctrl+C at any time to stop..."
echo

# Run integration tests with verbose output
make test-integration

echo
echo "🎉 Integration test completed!"
echo "You can see how the server receives and logs real PostgreSQL protocol data."
echo
echo "Next steps:"
echo "- Implement PostgreSQL wire protocol parsing"
echo "- Extract SQL queries from the protocol messages"
echo "- Add query cost estimation and quota tracking" 