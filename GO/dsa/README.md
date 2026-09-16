---
id: dsa_in_go_readme
aliases:
  - DSA in Go
  - Data Structures and Algorithms in Go
tags:
  - go
  - dsa
  - computer-science
  - algorithms
dg-publish: true
---

# Data Structures and Algorithms in Go

A comprehensive, production-grade guide to implementing core data structures and algorithmic patterns in Go. Every topic contains explanations, complexity analysis, visual diagrams, and modern Go implementations (using Go 1.18+ Generics where applicable).

---

## 📚 Table of Contents

| # | Topic | Key Concepts Covered | Link |
|---|---|---|---|
| 01 | **Arrays & Slices** | Memory layout, dynamic resizing, two-pointer, sliding window | [[01_Arrays_and_Strings\|01 Arrays and Strings]] |
| 02 | **Linked Lists** | Singly & Doubly linked list, reversal, cycle detection (Floyd's) | [[02_Linked_Lists\|02 Linked Lists]] |
| 03 | **Stacks & Queues** | Slice vs pointer backing, Circular Queue, Min Stack, Monotonic Stack | [[03_Stacks_and_Queues\|03 Stacks and Queues]] |
| 04 | **Hash Tables & Sets** | Go `map` internals, buckets, custom hash table, `Set[T]` pattern | [[04_Hash_Tables_and_Sets\|04 Hash Tables and Sets]] |
| 05 | **Binary Search** | Mid overflow safety, lower/upper bound, rotated array, `slices.BinarySearch` | [[05_Binary_Search\|05 Binary Search]] |
| 06 | **Trees & BST** | Binary tree node, BST CRUD, In/Pre/Post traversals, Tree BFS | [[06_Trees_and_BST\|06 Trees and BST]] |
| 07 | **Heaps & Priority Queue** | Complete binary tree math, Min/Max heap, `container/heap` interface | [[07_Heaps_and_Priority_Queue\|07 Heaps and Priority Queue]] |
| 08 | **Graphs: DFS & BFS** | Adjacency lists, Graph DFS, Graph BFS, Cycle detection, Topological Sort | [[08_Graphs_DFS_and_BFS\|08 Graphs: DFS & BFS]] |

---

## 🐹 Go-Specific DSA Essentials

### 1. Value vs. Pointer Receivers
In Go, structs are passed by value (copied) unless a pointer is used.
For data structures that mutate their internal state (e.g., modifying head of a linked list, appending to an encapsulated slice, adjusting tree pointers), always use **pointer receivers** (`(s *Stack[T])`).

```go
// Modifies the actual struct instance
func (s *Stack[T]) Push(val T) {
    s.items = append(s.items, val)
}
```

### 2. Slices: Under the Hood
A slice in Go is a 3-word header containing:
1. `Data`: Pointer to the underlying array (`unsafe.Pointer`)
2. `Len`: Number of elements currently present (`int`)
3. `Cap`: Capacity before reallocating memory (`int`)

```
Slice Header:  [ ptr | len: 3 | cap: 5 ]
                 │
                 ▼
Underlying Array: [ 10 | 20 | 30 | _ | _ ]
```

- When `append()` exceeds `cap`, Go doubles the capacity (for small slices) or grows by ~1.25x (for larger slices).
- Pre-allocating slice capacity (`make([]T, 0, expectedCap)`) avoids costly memory allocations and GC thrashing.

### 3. Generics and Type Constraints
DSA in Go leverages modern type parameters:
- `any`: Alias for `interface{}` — allows any type (stacks, queues, trees).
- `comparable`: Restricts types to those supporting `==` and `!=` (hash maps, sets, graph node IDs).
- `cmp.Ordered` (from package `cmp` in Go 1.21+): Restricts to types supporting `<`, `<=`, `>`, `>=` (ints, floats, strings for BSTs, Heaps, and Binary Search).

```go
import "cmp"

// Node definition works with any ordered type
type Node[T cmp.Ordered] struct {
    Val   T
    Left  *Node[T]
    Right *Node[T]
}
```

---

## ⚡ Complexity Cheat Sheet

| Data Structure / Algorithm | Access | Search | Insertion | Deletion | Space |
|---|---|---|---|---|---|
| **Array / Slice** | $O(1)$ | $O(n)$ | $O(n)$ (amortized $O(1)$ at end) | $O(n)$ | $O(n)$ |
| **Singly Linked List** | $O(n)$ | $O(n)$ | $O(1)$ (at head) | $O(1)$ (with pointer) | $O(n)$ |
| **Doubly Linked List** | $O(n)$ | $O(n)$ | $O(1)$ (head/tail) | $O(1)$ (with node ptr) | $O(n)$ |
| **Stack / Queue** | $O(n)$ | $O(n)$ | $O(1)$ | $O(1)$ | $O(n)$ |
| **Hash Table** | - | $O(1)$ avg / $O(n)$ worst | $O(1)$ avg / $O(n)$ worst | $O(1)$ avg / $O(n)$ worst | $O(n)$ |
| **Binary Search Tree** | $O(\log n)$ avg / $O(n)$ | $O(\log n)$ avg / $O(n)$ | $O(\log n)$ avg / $O(n)$ | $O(\log n)$ avg / $O(n)$ | $O(n)$ |
| **Binary Heap** | - | $O(n)$ | $O(\log n)$ | $O(\log n)$ (extract-min/max) | $O(n)$ |
| **Binary Search** | - | $O(\log n)$ | - | - | $O(1)$ iterative |
| **BFS / DFS (Graph)** | - | $O(V + E)$ | - | - | $O(V)$ |
