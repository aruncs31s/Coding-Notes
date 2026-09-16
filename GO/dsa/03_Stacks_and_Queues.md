---
id: stacks_and_queues_dsa_go
aliases:
  - Stacks and Queues in Go
  - LIFO and FIFO
tags:
  - go
  - dsa
  - stack
  - queue
dg-publish: true
---

# Stacks and Queues in Go

**Stacks (LIFO)** and **Queues (FIFO)** are fundamental linear data structures. Go does not provide built-in `Stack` or `Queue` types in the standard library; instead, they are built using slices, linked lists, or ring buffers.

---

## 1. Stack (LIFO: Last In, First Out)

Elements are pushed and popped from the same end (the "top").

```
Push(30) ──> | 30 | (Top)
             | 20 |
             | 10 | (Bottom)
             +----+
Pop()    <── returns 30
```

### Generic Slice-Based Stack

```go
package main

import (
	"errors"
	"fmt"
)

var ErrStackEmpty = errors.New("stack is empty")

// Stack is a generic LIFO data structure.
type Stack[T any] struct {
	items []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{items: make([]T, 0)}
}

// Push adds an element to the top of the stack. Time: Amortized O(1)
func (s *Stack[T]) Push(val T) {
	s.items = append(s.items, val)
}

// Pop removes and returns the top element. Time: O(1)
func (s *Stack[T]) Pop() (T, error) {
	var zero T
	if s.IsEmpty() {
		return zero, ErrStackEmpty
	}

	lastIndex := len(s.items) - 1
	val := s.items[lastIndex]
	// Optional: Avoid memory retention when T holds pointer types
	s.items[lastIndex] = zero
	s.items = s.items[:lastIndex]
	return val, nil
}

// Peek returns top item without removing it. Time: O(1)
func (s *Stack[T]) Peek() (T, error) {
	var zero T
	if s.IsEmpty() {
		return zero, ErrStackEmpty
	}
	return s.items[len(s.items)-1], nil
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

func (s *Stack[T]) Len() int {
	return len(s.items)
}
```

### Advanced Pattern: Min Stack
Design a stack that supports `Push`, `Pop`, `Top`, and retrieving the minimum element in **$O(1)$** time.

```go
type MinStack struct {
	stack    []int
	minStack []int
}

func NewMinStack() *MinStack {
	return &MinStack{}
}

func (ms *MinStack) Push(val int) {
	ms.stack = append(ms.stack, val)
	if len(ms.minStack) == 0 || val <= ms.minStack[len(ms.minStack)-1] {
		ms.minStack = append(ms.minStack, val)
	}
}

func (ms *MinStack) Pop() {
	if len(ms.stack) == 0 {
		return
	}
	top := ms.stack[len(ms.stack)-1]
	ms.stack = ms.stack[:len(ms.stack)-1]

	if top == ms.minStack[len(ms.minStack)-1] {
		ms.minStack = ms.minStack[:len(ms.minStack)-1]
	}
}

func (ms *MinStack) Top() int {
	return ms.stack[len(ms.stack)-1]
}

func (ms *MinStack) GetMin() int {
	return ms.minStack[len(ms.minStack)-1]
}
```

---

## 2. Queue (FIFO: First In, First Out)

Elements are enqueued at the back (tail) and dequeued from the front (head).

```
Enqueue(40) ──> [ 10 | 20 | 30 | 40 ] ──> Dequeue() returns 10
                  ▲                ▲
                Front             Rear
```

> [!WARNING]
> **Slice Queue Pitfall**: Simple sub-slicing `q = q[1:]` does **not** free the underlying memory of the dropped elements and results in continuous memory leaks. Use either a pointer-based queue, or a dynamic circular buffer.

### Generic Circular Queue (Ring Buffer)

```go
var ErrQueueEmpty = errors.New("queue is empty")

// CircularQueue dynamically resizes when full and prevents memory leaks on dequeue.
type CircularQueue[T any] struct {
	data  []T
	head  int
	tail  int
	count int
}

func NewCircularQueue[T any](initialCap int) *CircularQueue[T] {
	if initialCap < 2 {
		initialCap = 8
	}
	return &CircularQueue[T]{
		data: make([]T, initialCap),
	}
}

// Enqueue adds an element to the rear of the queue. Time: Amortized O(1)
func (q *CircularQueue[T]) Enqueue(val T) {
	if q.count == len(q.data) {
		q.resize(len(q.data) * 2)
	}
	q.data[q.tail] = val
	q.tail = (q.tail + 1) % len(q.data)
	q.count++
}

// Dequeue removes and returns the front element. Time: O(1)
func (q *CircularQueue[T]) Dequeue() (T, error) {
	var zero T
	if q.count == 0 {
		return zero, ErrQueueEmpty
	}

	val := q.data[q.head]
	q.data[q.head] = zero // GC clean-up
	q.head = (q.head + 1) % len(q.data)
	q.count--

	// Optional: Shrink buffer when heavily underutilized
	if q.count > 0 && q.count <= len(q.data)/4 && len(q.data) > 8 {
		q.resize(len(q.data) / 2)
	}

	return val, nil
}

func (q *CircularQueue[T]) Peek() (T, error) {
	if q.count == 0 {
		var zero T
		return zero, ErrQueueEmpty
	}
	return q.data[q.head], nil
}

func (q *CircularQueue[T]) IsEmpty() bool {
	return q.count == 0
}

func (q *CircularQueue[T]) Len() int {
	return q.count
}

func (q *CircularQueue[T]) resize(newCap int) {
	newData := make([]T, newCap)
	for i := 0; i < q.count; i++ {
		newData[i] = q.data[(q.head+i)%len(q.data)]
	}
	q.data = newData
	q.head = 0
	q.tail = q.count
}
```

---

## 3. Monotonic Stack Pattern

Used to find the **Next Greater Element** (or Next Smaller Element) in $O(n)$ time.

```go
// NextGreaterElements returns an array where result[i] is the next greater element
// to the right of nums[i], or -1 if none exists.
// Time: O(n), Space: O(n)
func NextGreaterElements(nums []int) []int {
	n := len(nums)
	result := make([]int, n)
	for i := range result {
		result[i] = -1
	}

	// Stack stores indices of elements
	stack := make([]int, 0)

	for i := 0; i < n; i++ {
		// While current element is greater than stack's top element
		for len(stack) > 0 && nums[i] > nums[stack[len(stack)-1]] {
			topIdx := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			result[topIdx] = nums[i]
		}
		stack = append(stack, i)
	}

	return result
}
```
