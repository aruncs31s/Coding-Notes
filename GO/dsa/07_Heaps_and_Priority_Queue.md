---
id: heaps_and_priority_queue_dsa_go
aliases:
  - Heaps and Priority Queue in Go
  - container/heap
tags:
  - go
  - dsa
  - heap
  - priority-queue
dg-publish: true
---

# Heaps and Priority Queues in Go

A **Binary Heap** is a complete binary tree implemented sequentially inside an array or slice.
- **Min-Heap**: The value of each node is $\le$ the values of its children (Root is minimum).
- **Max-Heap**: The value of each node is $\ge$ the values of its children (Root is maximum).

```
        Min-Heap Tree:                Array Representation:
              2                       Index: [ 0 | 1 | 2 | 3 | 4 | 5 ]
            /   \                     Value: [ 2 | 4 | 7 | 8 | 9 | 10 ]
           4     7
          / \   /
         8   9 10
```

### Index Math (0-Based)
Given an element at index $i$:
- **Parent**: `(i - 1) / 2`
- **Left Child**: `2*i + 1`
- **Right Child**: `2*i + 2`

---

## 1. Custom Min-Heap from Scratch

Demonstrating `heapifyUp` and `heapifyDown` mechanics.

```go
package main

import (
	"cmp"
	"errors"
	"fmt"
)

var ErrEmptyHeap = errors.New("heap is empty")

type MinHeap[T cmp.Ordered] struct {
	data []T
}

func NewMinHeap[T cmp.Ordered]() *MinHeap[T] {
	return &MinHeap[T]{data: make([]T, 0)}
}

// Push adds an element into the heap. Time: O(log n)
func (h *MinHeap[T]) Push(val T) {
	h.data = append(h.data, val)
	h.heapifyUp(len(h.data) - 1)
}

// Pop removes and returns the minimum element. Time: O(log n)
func (h *MinHeap[T]) Pop() (T, error) {
	var zero T
	if len(h.data) == 0 {
		return zero, ErrEmptyHeap
	}

	minVal := h.data[0]
	lastIdx := len(h.data) - 1

	// Move last element to root and sift down
	h.data[0] = h.data[lastIdx]
	h.data = h.data[:lastIdx]

	if len(h.data) > 0 {
		h.heapifyDown(0)
	}

	return minVal, nil
}

func (h *MinHeap[T]) heapifyUp(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if h.data[i] < h.data[parent] {
			h.data[i], h.data[parent] = h.data[parent], h.data[i]
			i = parent
		} else {
			break
		}
	}
}

func (h *MinHeap[T]) heapifyDown(i int) {
	n := len(h.data)
	for {
		smallest := i
		left := 2*i + 1
		right := 2*i + 2

		if left < n && h.data[left] < h.data[smallest] {
			smallest = left
		}
		if right < n && h.data[right] < h.data[smallest] {
			smallest = right
		}

		if smallest != i {
			h.data[i], h.data[smallest] = h.data[smallest], h.data[i]
			i = smallest
		} else {
			break
		}
	}
}
```

---

## 2. Using Go's Standard `container/heap`

In production Go, use the `container/heap` package. Any collection satisfying `heap.Interface` can be transformed into a heap.

### The `heap.Interface` Contract
```go
type Interface interface {
    sort.Interface // Len(), Less(i, j int) bool, Swap(i, j int)
    Push(x any)    // add x as element Len()
    Pop() any      // remove and return element Len() - 1.
}
```

### Implementing an Item Priority Queue

```go
import (
	"container/heap"
	"fmt"
)

// Item represents an element with a priority in the queue.
type Item struct {
	Value    string
	Priority int // Higher value = higher priority (Max-Heap)
	Index    int // Index in the heap (needed for update operations)
}

// PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int           { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool { return pq[i].Priority > pq[j].Priority } // > for Max-Heap
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*Item)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // GC clean-up
	item.Index = -1
	*pq = old[0 : n-1]
	return item
}

func ExamplePriorityQueueUsage() {
	pq := &PriorityQueue{}
	heap.Init(pq)

	heap.Push(pq, &Item{Value: "low-task", Priority: 1})
	heap.Push(pq, &Item{Value: "critical-bug", Priority: 10})
	heap.Push(pq, &Item{Value: "feature-request", Priority: 5})

	for pq.Len() > 0 {
		item := heap.Pop(pq).(*Item)
		fmt.Printf("Task: %s (Priority: %d)\n", item.Value, item.Priority)
	}
	// Output:
	// Task: critical-bug (Priority: 10)
	// Task: feature-request (Priority: 5)
	// Task: low-task (Priority: 1)
}
```

---

## 3. Top K Elements Pattern

Find the $K$ largest elements in an array using a Min-Heap of fixed size $K$.
Time: $O(n \log k)$, Space: $O(k)$.

```go
type IntMinHeap []int

func (h IntMinHeap) Len() int           { return len(h) }
func (h IntMinHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h IntMinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *IntMinHeap) Push(x any)        { *h = append(*h, x.(int)) }
func (h *IntMinHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func FindKLargest(nums []int, k int) []int {
	h := &IntMinHeap{}
	heap.Init(h)

	for _, num := range nums {
		heap.Push(h, num)
		if h.Len() > k {
			heap.Pop(h) // Drop the smallest among the (k+1) elements
		}
	}
	return *h
}
```
