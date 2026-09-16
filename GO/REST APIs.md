---
id: rest_apis_go
aliases:
  - REST APIs in Go
  - Go Web APIs
tags:
  - go
  - api
  - rest
  - http
  - web
dg-publish: true
---

# Building Production REST APIs in Go

Go has first-class support for HTTP networking. Since **Go 1.22**, the standard library's `http.ServeMux` natively supports HTTP methods, wildcards, and path parameters (`GET /users/{id}`), eliminating the necessity of external routers for many microservices.

```
Incoming Request
      │
      ▼
[ Middleware Chain ] ──> (Logging -> Recovery -> Auth -> CORS)
      │
      ▼
[ Handler Layer ]    ──> Parse JSON, validate input, call Service
      │
      ▼
[ Service Layer ]    ──> Core business rules & transactions
      │
      ▼
[ Repository Layer ] ──> SQL/GORM, Redis, or external API calls
```

---

## 1. Clean API Response Envelope

Standardizing your JSON responses ensures frontend clients and API consumers receive predictable structures.

```go
package main

import (
	"encoding/json"
	"net/http"
)

type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func RespondJSON[T any](w http.ResponseWriter, status int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(APIResponse[T]{
		Success: true,
		Data:    data,
	})
}

func RespondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(APIResponse[any]{
		Success: false,
		Error:   message,
	})
}
```

---

## 2. Idiomatic Go 1.22+ Native Routing & Handlers

Go 1.22 introduced:
- Method matching in routes (`"GET /items"`, `"POST /items"`)
- Path variables via `r.PathValue("id")`
- Automatic `405 Method Not Allowed` when routes match a different method

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type CreateProductRequest struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type ProductHandler struct {
	mu       sync.RWMutex
	products map[int]Product
	nextID   int
}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{
		products: make(map[int]Product),
		nextID:   1,
	}
}

// GET /api/v1/products
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	items := make([]Product, 0, len(h.products))
	for _, p := range h.products {
		items = append(items, p)
	}

	RespondJSON(w, http.StatusOK, items)
}

// GET /api/v1/products/{id}
func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	h.mu.RLock()
	product, exists := h.products[id]
	h.mu.RUnlock()

	if !exists {
		RespondError(w, http.StatusNotFound, "Product not found")
		return
	}

	RespondJSON(w, http.StatusOK, product)
}

// POST /api/v1/products
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Malformed request payload")
		return
	}

	// Validation
	if strings.TrimSpace(req.Name) == "" {
		RespondError(w, http.StatusUnprocessableEntity, "Name is required")
		return
	}
	if req.Price <= 0 {
		RespondError(w, http.StatusUnprocessableEntity, "Price must be greater than zero")
		return
	}

	h.mu.Lock()
	newProduct := Product{
		ID:    h.nextID,
		Name:  req.Name,
		Price: req.Price,
	}
	h.products[h.nextID] = newProduct
	h.nextID++
	h.mu.Unlock()

	RespondJSON(w, http.StatusCreated, newProduct)
}
```

---

## 3. Composable Middleware Pipeline

```go
import (
	"log"
	"time"
)

type Middleware func(http.Handler) http.Handler

// CreateStack chains multiple middlewares
func CreateStack(xs ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(xs) - 1; i >= 0; i-- {
			next = xs[i](next)
		}
		return next
	}
}

// LoggingMiddleware logs request path, method, and latency
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// RecoveryMiddleware prevents server crashes on panic
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Unhandled panic: %v", err)
				RespondError(w, http.StatusInternalServerError, "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
```

---

## 4. Server Assembly & Execution

```go
func main() {
	handler := NewProductHandler()

	mux := http.NewServeMux()

	// Register Go 1.22 route patterns
	mux.HandleFunc("GET /api/v1/products", handler.List)
	mux.HandleFunc("GET /api/v1/products/{id}", handler.GetByID)
	mux.HandleFunc("POST /api/v1/products", handler.Create)

	// Wrap mux with middleware stack
	stack := CreateStack(
		RecoveryMiddleware,
		LoggingMiddleware,
	)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      stack(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("REST API server running at http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
```
