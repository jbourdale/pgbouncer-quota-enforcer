package adapters

import (
	"fmt"
	"io"

	"github.com/jackc/pgx/v5/pgproto3"
)

// PostgreSQLParser handles parsing of PostgreSQL wire protocol messages
type PostgreSQLParser struct {
	backend *pgproto3.Backend
}

// NewPostgreSQLParser creates a new PostgreSQL protocol parser
func NewPostgreSQLParser(reader io.Reader, writer io.Writer) *PostgreSQLParser {
	backend := pgproto3.NewBackend(reader, writer)
	return &PostgreSQLParser{
		backend: backend,
	}
}

// ParsedMessage represents a parsed PostgreSQL protocol message
type ParsedMessage struct {
	Type    string
	Query   string
	Details map[string]interface{}
}

// ReadMessage reads and parses the next PostgreSQL protocol message
func (p *PostgreSQLParser) ReadMessage() (*ParsedMessage, error) {
	msg, err := p.backend.Receive()
	if err != nil {
		return nil, fmt.Errorf("failed to receive message: %w", err)
	}

	return p.parseMessage(msg)
}

// parseMessage converts a pgproto3 message to our ParsedMessage format
func (p *PostgreSQLParser) parseMessage(msg pgproto3.Message) (*ParsedMessage, error) {
	switch m := msg.(type) {
	case *pgproto3.Query:
		return &ParsedMessage{
			Type:  "Query",
			Query: m.String,
			Details: map[string]interface{}{
				"sql": m.String,
			},
		}, nil

	case *pgproto3.Parse:
		return &ParsedMessage{
			Type:  "Parse",
			Query: m.Query,
			Details: map[string]interface{}{
				"name":           m.Name,
				"query":          m.Query,
				"parameter_oids": m.ParameterOIDs,
			},
		}, nil

	case *pgproto3.StartupMessage:
		details := make(map[string]interface{})
		for k, v := range m.Parameters {
			details[k] = v
		}
		details["protocol_version"] = m.ProtocolVersion

		return &ParsedMessage{
			Type:    "StartupMessage",
			Details: details,
		}, nil

	case *pgproto3.PasswordMessage:
		return &ParsedMessage{
			Type: "PasswordMessage",
			Details: map[string]interface{}{
				"password_length": len(m.Password),
			},
		}, nil

	case *pgproto3.Bind:
		return &ParsedMessage{
			Type: "Bind",
			Details: map[string]interface{}{
				"destination_portal": m.DestinationPortal,
				"prepared_statement": m.PreparedStatement,
				"parameter_count":    len(m.Parameters),
			},
		}, nil

	case *pgproto3.Execute:
		return &ParsedMessage{
			Type: "Execute",
			Details: map[string]interface{}{
				"portal":   m.Portal,
				"max_rows": m.MaxRows,
			},
		}, nil

	case *pgproto3.Describe:
		return &ParsedMessage{
			Type: "Describe",
			Details: map[string]interface{}{
				"object_type": string(m.ObjectType),
				"name":        m.Name,
			},
		}, nil

	case *pgproto3.Sync:
		return &ParsedMessage{
			Type:    "Sync",
			Details: map[string]interface{}{},
		}, nil

	case *pgproto3.Terminate:
		return &ParsedMessage{
			Type:    "Terminate",
			Details: map[string]interface{}{},
		}, nil

	case *pgproto3.Flush:
		return &ParsedMessage{
			Type:    "Flush",
			Details: map[string]interface{}{},
		}, nil

	case *pgproto3.Close:
		return &ParsedMessage{
			Type: "Close",
			Details: map[string]interface{}{
				"object_type": string(m.ObjectType),
				"name":        m.Name,
			},
		}, nil

	default:
		return &ParsedMessage{
			Type: fmt.Sprintf("Unknown_%T", msg),
			Details: map[string]interface{}{
				"message": fmt.Sprintf("%+v", msg),
			},
		}, nil
	}
}
