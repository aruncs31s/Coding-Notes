---
id: hash_tables_and_sets_dsa_go
aliases:
  - Hash Tables and Sets in Go
  - Go Maps Internals
tags:
  - go
  - dsa
  - hash-table
  - set
dg-publish: true
---

# Hash Tables and Sets in Go

A **Hash Table** associates keys with values using a hash function to compute an index into an array of buckets.

---

## 1. Go's Built-in `map` Under the Hood

In Go, a `map` is a pointer to an `hmap` struct:
- **Buckets**: An array of `bmap` structs. Each bucket holds up to **8 key-value pairs**.
- **Top-Hash (tophash)**: An 8-byte array storing the high 8 bits of the hash for each key to quickly check keys without full equality comparisons.
- **Load Factor**: When the load factor reaches **6.5** pairs per bucket, Go triggers evacuation/rehashing to double the buckets.
- **Key Restrictions**: The key type must be `comparable` (supports `==` and `!=`). Slices, maps, and functions cannot be used as map keys.

---

## 2. Implementing a Generic `Set[T]`

Go does not have a built-in `Set` type. The standard, idiomatic Go way is to use `map[T]struct{}`.
`struct{}` is an empty struct that occupies **0 bytes of memory**, making it strictly more efficient than `map[T]bool`.

```go
package main

import "fmt"

// Set implements an unordered collection of unique elements.
type Set[T comparable] struct {
	elements map[T]struct{}
}

func NewSet[T comparable](items ...T) *Set[T] {
	s := &Set[T]{elements: make(map[T]struct{})}
	for _, item := range items {
		s.Add(item)
	}
	return s
}

func (s *Set[T]) Add(val T) {
	s.elements[val] = struct{}{}
}

func (s *Set[T]) Remove(val T) {
	delete(s.elements, val)
}

func (s *Set[T]) Contains(val T) bool {
	_, exists := s.elements[val]
	return exists
}

func (s *Set[T]) Size() int {
	return len(s.elements)
}

func (s *Set[T]) Elements() []T {
	result := make([]T, 0, len(s.elements))
	for k := range s.elements {
		result = append(result, k)
	}
	return result
}

// Intersection returns a new set with elements present in both sets.
func (s *Set[T]) Intersection(other *Set[T]) *Set[T] {
	res := NewSet[T]()
	// Optimize: iterate through smaller set
	first, second := s, other
	if first.Size() > second.Size() {
		first, second = second, first
	}

	for k := range first.elements {
		if second.Contains(k) {
			res.Add(k)
		}
	}
	return res
}
```

---

## 3. Custom Hash Table Implementation (Separate Chaining)

Below is an explicit implementation demonstrating how collision resolution via chaining works.

```
Bucket 0: nil
Bucket 1: [ "cat": 5 ] ──> [ "tac": 8 ] ──> nil   (collision handled by linked list)
Bucket 2: [ "dog": 3 ] ──> nil
```

```go
import "hash/fnv"

type entry[K comparable, V any] struct {
	key   K
	value V
	next  *entry[K, V]
}

type HashTable[K comparable, V any] struct {
	buckets []*entry[K, V]
	size    int
	cap     int
}

func NewHashTable[K comparable, V any](initialCap int) *HashTable[K, V] {
	if initialCap < 16 {
		initialCap = 16
	}
	return &HashTable[K, V]{
		buckets: make([]*entry[K, V], initialCap),
		cap:     initialCap,
	}
}

// hashFn computes bucket index using FNV-1a hash
func (ht *HashTable[K, V]) hashFn(key K) int {
	h := fnv.New32a()
	h.Write([]byte(fmt.Sprintf("%v", key)))
	return int(h.Sum32()) % ht.cap
}

// Put inserts or updates a key-value pair. Time: Average O(1)
func (ht *HashTable[K, V]) Put(key K, value V) {
	idx := ht.hashFn(key)
	curr := ht.buckets[idx]

	// Update existing key
	for curr != nil {
		if curr.key == key {
			curr.value = value
			return
		}
		curr = curr.next
	}

	// Insert new entry at head of chain
	newEntry := &entry[K, V]{
		key:   key,
		value: value,
		next:  ht.buckets[idx],
	}
	ht.buckets[idx] = newEntry
	ht.size++
}

// Get retrieves the value for a key. Time: Average O(1)
func (ht *HashTable[K, V]) Get(key K) (V, bool) {
	idx := ht.hashFn(key)
	curr := ht.buckets[idx]

	for curr != nil {
		if curr.key == key {
			return curr.value, true
		}
		curr = curr.next
	}

	var zero V
	return zero, false
}

// Delete removes a key from the table. Time: Average O(1)
func (ht *HashTable[K, V]) Delete(key K) bool {
	idx := ht.hashFn(key)
	curr := ht.buckets[idx]
	var prev *entry[K, V]

	for curr != nil {
		if curr.key == key {
			if prev == nil {
				ht.buckets[idx] = curr.next
			} else {
				prev.next = curr.next
			}
			ht.size--
			return true
		}
		prev = curr
		curr = curr.next
	}
	return false
}
```
