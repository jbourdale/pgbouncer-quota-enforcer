package adapters

import (
	"fmt"
	"pgbouncer-quota-enforcer/internal/app/domain"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v6"
)

// PgQueryNormalizer implements domain.QueryNormalizer using pg_query library
type PgQueryNormalizer struct{}

// NewPgQueryNormalizer creates a new PgQueryNormalizer
func NewPgQueryNormalizer() domain.QueryNormalizer {
	return &PgQueryNormalizer{}
}

// Normalize normalizes a SQL query using PostgreSQL's actual parser
func (n *PgQueryNormalizer) Normalize(rawQuery string) (domain.NormalizedQuery, error) {
	if rawQuery == "" {
		return domain.NormalizedQuery{}, fmt.Errorf("empty query cannot be normalized")
	}

	// Check for whitespace-only queries
	if strings.TrimSpace(rawQuery) == "" {
		return domain.NormalizedQuery{}, fmt.Errorf("empty query cannot be normalized")
	}

	// Use pg_query to normalize the query
	normalized, err := pg_query.Normalize(rawQuery)
	if err != nil {
		return domain.NormalizedQuery{}, fmt.Errorf("failed to normalize query: %w", err)
	}

	// Use pg_query to generate a fingerprint (hash)
	fingerprint, err := pg_query.Fingerprint(rawQuery)
	if err != nil {
		return domain.NormalizedQuery{}, fmt.Errorf("failed to generate fingerprint: %w", err)
	}

	return domain.NormalizedQuery{
		Original:   rawQuery,
		Normalized: normalized,
		Hash:       domain.NewQueryHash(fingerprint),
	}, nil
}
