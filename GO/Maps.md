---
id: maps_go
aliases:
  - Maps in Go
  - Go Map Internals and Concurrency
tags:
  - go
  - maps
  - concurrency
  - memory
dg-publish: true
---

# Go Maps: Internals, Concurrency & Thread-Safety

A **Map** in Go is an unordered collection of key-value pairs built on a dynamically sized hash table.
While simple to use (`m := make(map[string]int)`), maps have runtime behaviors, memory characteristics, and strict concurrency rules that every Go engineer must master.

```
                    hmap (Map Header)
             +------------------------------+
             | count  | flags | B (2^B bkts)|
             | hash0 (random seed)          |
             | buckets ─────────────────────┼─────────┐
             +------------------------------+         │
                                                      ▼
                                       Array of bmap (Buckets)
                                    +---------------------------+
                                    | tophash: [8]uint8         |
                                    | keys:    [8]KeyType       |
                                    | values:  [8]ValueType     |
                                    | overflow ──> next bmap    |
                                    +---------------------------+
```

---

## 1. Under the Hood: `hmap` & `bmap`

- **Buckets (`bmap`)**: A bucket stores up to **8 key-value pairs**.
- **Top-Hash**: To avoid comparing long keys directly, the runtime hashes the key, takes the top 8 bits, and compares them first.
- **Hash Seed (`hash0`)**: Initialized with a random seed when the map is created. This defends against HashDoS attacks and causes **iteration order to be randomized**.
- **Load Factor**: When the number of elements per bucket exceeds **6.5**, Go initiates rehashing (doubles bucket count and incrementally evacuates old buckets).
- **Non-addressable**: You cannot take the address of a map element (`&m["key"]`) because as the map grows and re-hashes, bucket elements move in memory.

---

## 2. The Fatal Concurrency Trap

In Go, maps are **NOT safe for concurrent use**.
If one goroutine writes to a map while another reads or writes to it, the runtime immediately aborts execution with a fatal panic:

```
fatal error: concurrent map writes
// or
fatal error: concurrent map read and map write
```

> [!CAUTION]
> This panic **cannot** be caught or prevented by `recover()`. The entire Go process crashes immediately. Always synchronize map access across goroutines!

---

## 3. Safe Concurrent Maps (Pattern 1: `sync.RWMutex`)

For general-purpose concurrent applications, wrapping a standard map with `sync.RWMutex` provides type safety, high read concurrency, and clean ergonomics.

```go
package main

import (
	"sync"
)

// SafeMap provides thread-safe access to map[K]V.
type SafeMap[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		data: make(map[K]V),
	}
}

// Get retrieves a value under a read lock (concurrent reads allowed).
func (sm *SafeMap[K, V]) Get(key K) (V, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	val, exists := sm.data[key]
	return val, exists
}

// Set writes a value under an exclusive write lock.
func (sm *SafeMap[K, V]) Set(key K, val V) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.data[key] = val
}

// Delete removes a key under a write lock.
func (sm *SafeMap[K, V]) Delete(key K) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.data, key)
}

// Len returns the current count of elements.
func (sm *SafeMap[K, V]) Len() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.data)
}
```

---

## 4. When to Use `sync.Map` (Pattern 2)

Go provides `sync.Map` in the standard library. However, the Go documentation explicitly notes:
> `sync.Map` is optimized for two specific scenarios:
> 1. When the entry for a given key is **only ever written once but read many times** (e.g., caches that only grow).
> 2. When multiple goroutines read, write, and overwrite entries for **disjoint sets of keys**.

In other scenarios, `sync.RWMutex` with a normal map is significantly faster and type-safe.

```go
package main

import (
	"fmt"
	"sync"
)

func ExampleSyncMapUsage() {
	var m sync.Map

	// 1. Store key-value
	m.Store("session_101", "active")
	m.Store("session_102", "pending")

	// 2. Load
	val, ok := m.Load("session_101")
	if ok {
		fmt.Printf("session_101: %s\n", val.(string))
	}

	// 3. LoadOrStore (Atomic insert-if-not-present)
	actual, loaded := m.LoadOrStore("session_103", "initialized")
	fmt.Printf("Loaded: %v, Value: %s\n", loaded, actual.(string))

	// 4. Iterate using Range
	m.Range(func(key, value any) bool {
		fmt.Printf("Iterate -> Key: %v, Value: %v\n", key, value)
		return true // continue iteration
	})

	// 5. Delete
	m.Delete("session_102")
}
```

---

## 5. Modern `maps` Package Utilities (Go 1.21+)

The standard library `maps` package provides generic helper functions:

```go
package main

import (
	"fmt"
	"maps"
)

func ExampleMapsStandardPackage() {
	m1 := map[string]int{"apple": 5, "banana": 2, "orange": 8}

	// 1. Clone creates a shallow copy
	m2 := maps.Clone(m1)

	// 2. Equal compares two maps
	fmt.Println("m1 equals m2:", maps.Equal(m1, m2)) // true

	// 3. DeleteFunc filters map entries in-place
	maps.DeleteFunc(m1, func(k string, v int) bool {
		return v < 5 // Remove fruits with count < 5
	})
	fmt.Println("After filtering:", m1) // map[apple:5 orange:8]

	// 4. Copy merges src into dst
	maps.Copy(m1, map[string]int{"mango": 12})
	fmt.Println("After merging:", m1)
}
```

---

## ⚠️ Go Map Pitfalls & Best Practices

1. **Reading from a `nil` map is safe**: Returns the zero value (`var m map[string]int; val := m["foo"] // 0`).
2. **Writing to a `nil` map PANICS**:
   ```go
   var m map[string]int
   m["key"] = 1 // panic: assignment to entry in nil map
   ```
   Always initialize with `m := make(map[string]int)` or a literal `{}`.
3. **Maps do not shrink in RAM after deletion**: Deleting millions of keys frees bucket slots, but the underlying allocated buckets remain in memory. To release memory back to the OS, copy remaining entries to a freshly allocated map.
