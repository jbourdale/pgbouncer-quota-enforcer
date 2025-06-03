package adapters

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgQueryNormalizer_Normalize(t *testing.T) {
	normalizer := NewPgQueryNormalizer()

	tests := []struct {
		name               string
		input              string
		expectedNormalized string
		expectError        bool
	}{
		{
			name:               "Simple SELECT with string literal",
			input:              "SELECT * FROM users WHERE name = 'John'",
			expectedNormalized: "SELECT * FROM users WHERE name = $1",
		},
		{
			name:               "SELECT with numeric literal",
			input:              "SELECT * FROM products WHERE price > 100",
			expectedNormalized: "SELECT * FROM products WHERE price > $1",
		},
		{
			name:               "Multiple parameters",
			input:              "SELECT * FROM orders WHERE user_id = 123 AND status = 'pending'",
			expectedNormalized: "SELECT * FROM orders WHERE user_id = $1 AND status = $2",
		},
		{
			name:               "IN clause with multiple values",
			input:              "SELECT * FROM items WHERE category IN ('electronics', 'books', 'clothing')",
			expectedNormalized: "SELECT * FROM items WHERE category IN ($1, $2, $3)",
		},
		{
			name:               "SELECT with LIMIT and OFFSET",
			input:              "SELECT * FROM posts ORDER BY created_at LIMIT 10 OFFSET 20",
			expectedNormalized: "", // Will verify structure instead of exact match
		},
		{
			name:        "Empty query",
			input:       "",
			expectError: true,
		},
		{
			name:        "Whitespace only query",
			input:       "   \n\t  ",
			expectError: true,
		},
		{
			name:        "Invalid SQL syntax",
			input:       "SELECT * FROM WHERE",
			expectError: true,
		},
		{
			name:               "Complex JOIN query",
			input:              "SELECT u.name, p.title FROM users u JOIN posts p ON u.id = p.user_id WHERE u.active = true",
			expectedNormalized: "SELECT u.name, p.title FROM users u JOIN posts p ON u.id = p.user_id WHERE u.active = $1",
		},
		{
			name:               "Query with functions",
			input:              "SELECT COUNT(*) FROM users WHERE created_at > '2023-01-01'",
			expectedNormalized: "SELECT COUNT(*) FROM users WHERE created_at > $1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizer.Normalize(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.input, result.Original, "Original query should be preserved")

			// Special handling for LIMIT/OFFSET test since parameter ordering may differ
			if tt.name == "SELECT with LIMIT and OFFSET" {
				// Just verify it has LIMIT $x OFFSET $y structure, don't care about exact parameter numbers
				assert.Contains(t, result.Normalized, "LIMIT $", "Should contain LIMIT with parameter")
				assert.Contains(t, result.Normalized, "OFFSET $", "Should contain OFFSET with parameter")
				assert.Contains(t, result.Normalized, "SELECT * FROM posts ORDER BY created_at", "Should preserve query structure")
			} else if tt.expectedNormalized != "" {
				assert.Equal(t, tt.expectedNormalized, result.Normalized, "Normalized query mismatch")
			}

			assert.NotEmpty(t, result.Hash.Value(), "Hash should not be empty")

			// Verify hash consistency
			result2, err := normalizer.Normalize(tt.input)
			require.NoError(t, err)
			assert.Equal(t, result.Hash.Value(), result2.Hash.Value(), "Hash should be consistent for same query")
		})
	}
}

func TestPgQueryNormalizer_HashConsistency(t *testing.T) {
	normalizer := NewPgQueryNormalizer()

	// These queries should produce the same hash after normalization
	equivalentQueries := [][]string{
		{
			"SELECT * FROM users WHERE id = 1",
			"SELECT * FROM users WHERE id = 2",
			"SELECT * FROM users WHERE id = 999",
		},
		{
			"SELECT * FROM products WHERE name = 'Product A'",
			"SELECT * FROM products WHERE name = 'Product B'",
			"SELECT * FROM products WHERE name = 'Something Else'",
		},
		{
			"SELECT * FROM orders WHERE total > 100 AND status = 'pending'",
			"SELECT * FROM orders WHERE total > 50 AND status = 'completed'",
			"SELECT * FROM orders WHERE total > 1000 AND status = 'cancelled'",
		},
	}

	for i, queryGroup := range equivalentQueries {
		t.Run(fmt.Sprintf("EquivalentGroup_%d", i), func(t *testing.T) {
			var hashes []string

			for _, query := range queryGroup {
				result, err := normalizer.Normalize(query)
				require.NoError(t, err)
				hashes = append(hashes, result.Hash.Value())
			}

			// All hashes in the group should be identical
			for j := 1; j < len(hashes); j++ {
				assert.Equal(t, hashes[0], hashes[j],
					"Queries with same structure should have same hash: %s vs %s",
					queryGroup[0], queryGroup[j])
			}
		})
	}
}

func TestPgQueryNormalizer_ComplexQueries(t *testing.T) {
	normalizer := NewPgQueryNormalizer()

	tests := []struct {
		name  string
		input string
	}{
		{
			name: "Complex SELECT with subquery",
			input: `SELECT u.id, u.name, 
					(SELECT COUNT(*) FROM orders o WHERE o.user_id = u.id AND o.status = 'completed') as order_count
					FROM users u 
					WHERE u.created_at > '2023-01-01' AND u.active = true
					ORDER BY u.name LIMIT 10`,
		},
		{
			name: "CTE (Common Table Expression)",
			input: `WITH recent_orders AS (
						SELECT user_id, COUNT(*) as order_count, SUM(total) as total_spent
						FROM orders 
						WHERE created_at > '2023-01-01'
						GROUP BY user_id
					)
					SELECT u.name, ro.order_count, ro.total_spent
					FROM users u
					JOIN recent_orders ro ON u.id = ro.user_id
					WHERE ro.total_spent > 1000`,
		},
		{
			name: "Window function",
			input: `SELECT 
						name, 
						salary,
						ROW_NUMBER() OVER (ORDER BY salary DESC) as rank,
						AVG(salary) OVER (PARTITION BY department) as dept_avg
					FROM employees 
					WHERE department = 'Engineering'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizer.Normalize(tt.input)
			require.NoError(t, err)

			// Verify that complex queries are handled without errors
			assert.NotEmpty(t, result.Normalized, "Normalized query should not be empty")
			assert.NotEmpty(t, result.Hash.Value(), "Hash should not be empty")
			assert.Equal(t, tt.input, result.Original, "Original query should be preserved")

			t.Logf("Original: %s", tt.input)
			t.Logf("Normalized: %s", result.Normalized)
			t.Logf("Hash: %s", result.Hash.Value())
		})
	}
}

func BenchmarkPgQueryNormalizer_Normalize(b *testing.B) {
	normalizer := NewPgQueryNormalizer()
	testQuery := "SELECT u.id, u.name, p.title FROM users u JOIN posts p ON u.id = p.user_id WHERE u.age > 25 AND p.created_at > '2023-01-01' AND p.category IN ('tech', 'science', 'news') ORDER BY p.created_at DESC LIMIT 10 OFFSET 20"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := normalizer.Normalize(testQuery)
		if err != nil {
			b.Fatal(err)
		}
	}
}
