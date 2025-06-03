package domain

import (
	"time"
)

// QueryHash represents a normalized query hash for tracking purposes
type QueryHash struct {
	value string
}

// NewQueryHash creates a new QueryHash
func NewQueryHash(hash string) QueryHash {
	return QueryHash{value: hash}
}

// String returns the hash value
func (q QueryHash) String() string {
	return q.value
}

// Value returns the hash value
func (q QueryHash) Value() string {
	return q.value
}

// Query represents a SQL query with metadata
type Query struct {
	Raw          string
	Normalized   string
	Hash         QueryHash
	ConnectionID string
	UserID       string
	Database     string
	Timestamp    time.Time
	Parameters   []interface{}
}

// NewQuery creates a new Query
func NewQuery(raw, connectionID string) *Query {
	return &Query{
		Raw:          raw,
		ConnectionID: connectionID,
		Timestamp:    time.Now(),
	}
}

// QueryNormalizer defines the interface for normalizing SQL queries
type QueryNormalizer interface {
	// Normalize takes a raw SQL query and returns the normalized version
	Normalize(rawQuery string) (NormalizedQuery, error)
}

// NormalizedQuery represents a normalized query result for quota tracking
type NormalizedQuery struct {
	Original   string
	Normalized string
	Hash       QueryHash
}

// QueryParameter represents a parameter extracted from a query
type QueryParameter struct {
	Position int
	Value    interface{}
	Type     string
}

// QueryAnalyzer defines the interface for analyzing queries
type QueryAnalyzer interface {
	// AnalyzeQuery processes a query and returns analysis results
	AnalyzeQuery(query *Query) (*QueryAnalysis, error)
}

// QueryAnalysis represents the analysis result of a query
type QueryAnalysis struct {
	Query         *Query
	EstimatedCost int64
	QueryType     QueryType
	Tables        []string
	Operations    []QueryOperation
}

// QueryType represents the type of SQL operation
type QueryType string

const (
	QueryTypeSelect QueryType = "SELECT"
	QueryTypeInsert QueryType = "INSERT"
	QueryTypeUpdate QueryType = "UPDATE"
	QueryTypeDelete QueryType = "DELETE"
	QueryTypeCreate QueryType = "CREATE"
	QueryTypeDrop   QueryType = "DROP"
	QueryTypeAlter  QueryType = "ALTER"
	QueryTypeOther  QueryType = "OTHER"
)

// QueryOperation represents a specific operation in a query
type QueryOperation struct {
	Type          string
	Table         string
	Complexity    int
	EstimatedRows int64
}
