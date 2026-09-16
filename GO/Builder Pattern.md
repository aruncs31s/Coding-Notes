---
id: builder_pattern_go
aliases:
  - Builder Pattern in Go
  - Functional Options Pattern
tags:
  - go
  - design-patterns
  - creational-patterns
dg-publish: true
---

# Builder Pattern in Go

The **Builder Pattern** is a creational design pattern that lets you construct complex objects step by step. In Go, two primary approaches exist:
1. **The Classic Fluent Builder**: Uses method chaining returning `*Builder`.
2. **The Functional Options Pattern**: The idiomatic Go standard, using closures and variadic functions.

---

## 1. Classic Fluent Builder Pattern

Best suited when an object requires multi-step construction where order or conditional staging matters.

```
+----------------+          +------------------+          +-----------------+
|  QueryBuilder  | -------> | .Select("users") | -------> | .Where("age>18")|
+----------------+          +------------------+          +-----------------+
                                                                   │
                                                            .Build()
                                                                   ▼
                                                          +-----------------+
                                                          |      Query      |
                                                          +-----------------+
```

### Complete Implementation

```go
package main

import (
	"errors"
	"fmt"
	"strings"
)

// SQLQuery is the final immutable product.
type SQLQuery struct {
	table      string
	columns    []string
	whereConds []string
	limit      int
	offset     int
}

// String generates the raw SQL query.
func (q *SQLQuery) String() string {
	cols := "*"
	if len(q.columns) > 0 {
		cols = strings.Join(q.columns, ", ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s", cols, q.table)

	if len(q.whereConds) > 0 {
		query += " WHERE " + strings.Join(q.whereConds, " AND ")
	}
	if q.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", q.limit)
	}
	if q.offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", q.offset)
	}

	return query
}

// SQLQueryBuilder builds SQLQuery step by step.
type SQLQueryBuilder struct {
	query *SQLQuery
	err   error
}

func NewQueryBuilder() *SQLQueryBuilder {
	return &SQLQueryBuilder{
		query: &SQLQuery{
			columns:    make([]string, 0),
			whereConds: make([]string, 0),
		},
	}
}

func (b *SQLQueryBuilder) From(table string) *SQLQueryBuilder {
	if b.err != nil {
		return b
	}
	if strings.TrimSpace(table) == "" {
		b.err = errors.New("table name cannot be empty")
		return b
	}
	b.query.table = table
	return b
}

func (b *SQLQueryBuilder) Select(columns ...string) *SQLQueryBuilder {
	if b.err != nil {
		return b
	}
	b.query.columns = append(b.query.columns, columns...)
	return b
}

func (b *SQLQueryBuilder) Where(condition string) *SQLQueryBuilder {
	if b.err != nil {
		return b
	}
	b.query.whereConds = append(b.query.whereConds, condition)
	return b
}

func (b *SQLQueryBuilder) Limit(limit int) *SQLQueryBuilder {
	if b.err != nil {
		return b
	}
	if limit < 0 {
		b.err = errors.New("limit cannot be negative")
		return b
	}
	b.query.limit = limit
	return b
}

func (b *SQLQueryBuilder) Offset(offset int) *SQLQueryBuilder {
	if b.err != nil {
		return b
	}
	b.query.offset = offset
	return b
}

// Build finalizes validation and produces the product.
func (b *SQLQueryBuilder) Build() (*SQLQuery, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.query.table == "" {
		return nil, errors.New("cannot build query without a target table")
	}
	return b.query, nil
}

func main() {
	query, err := NewQueryBuilder().
		From("users").
		Select("id", "name", "email").
		Where("status = 'active'").
		Where("age >= 18").
		Limit(10).
		Offset(20).
		Build()

	if err != nil {
		fmt.Printf("Builder error: %v\n", err)
		return
	}

	fmt.Println(query.String())
	// Output: SELECT id, name, email FROM users WHERE status = 'active' AND age >= 18 LIMIT 10 OFFSET 20
}
```

---

## 2. Functional Options Pattern (Idiomatic Go)

In Go, constructors often take optional configuration parameters. Rather than having multiple constructors (`NewServer`, `NewServerWithTimeout`, etc.) or passing large config structs with default zeroes, the **Functional Options Pattern** is widely used (e.g., in gRPC, Uber, Google libraries).

### Architecture

```
NewServer(addr, opts...)
   │
   ├─ Apply defaults
   ├─ Loop through opts: opt(server)
   └─ Return configured & validated server
```

### Complete Implementation

```go
package main

import (
	"fmt"
	"time"
)

type Server struct {
	host       string
	port       int
	timeout    time.Duration
	maxConns   int
	tlsEnabled bool
}

// ServerOption is a function signature for modifying Server.
type ServerOption func(*Server)

// WithPort sets server port.
func WithPort(port int) ServerOption {
	return func(s *Server) {
		s.port = port
	}
}

// WithTimeout sets connection timeout.
func WithTimeout(d time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = d
	}
}

// WithMaxConnections sets connection limit.
func WithMaxConnections(n int) ServerOption {
	return func(s *Server) {
		s.maxConns = n
	}
}

// WithTLS enables TLS.
func WithTLS(enabled bool) ServerOption {
	return func(s *Server) {
		s.tlsEnabled = enabled
	}
}

// NewServer builds and validates the server with defaults and options.
func NewServer(host string, options ...ServerOption) (*Server, error) {
	// 1. Default configuration
	srv := &Server{
		host:       host,
		port:       8080,
		timeout:    30 * time.Second,
		maxConns:   100,
		tlsEnabled: false,
	}

	// 2. Apply user options
	for _, opt := range options {
		opt(srv)
	}

	// 3. Validation
	if srv.port <= 0 || srv.port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", srv.port)
	}

	return srv, nil
}

func main() {
	// Server with defaults
	s1, _ := NewServer("localhost")
	fmt.Printf("Default Server: %+v\n", s1)

	// Server customized with options
	s2, _ := NewServer("0.0.0.0",
		WithPort(9090),
		WithTimeout(10*time.Second),
		WithTLS(true),
		WithMaxConnections(500),
	)
	fmt.Printf("Custom Server: %+v\n", s2)
}
```

---

## 3. Comparison: Fluent Builder vs Functional Options

| Feature | Fluent Builder | Functional Options |
|---|---|---|
| **Go Idiomatic Rating** | Moderate (popular in SQL/HTTP builders) | **High** (de-facto standard in Go libraries) |
| **Call Syntax** | `b.SetA().SetB().Build()` | `New(required, WithA(), WithB())` |
| **Immutability** | Created at `Build()` | Produced directly by constructor |
| **Error Handling** | Accumulated or checked at `.Build()` | Checked at constructor invocation |
| **Best Used For** | Multi-phase query/command construction | Struct configuration & dependency injection |
