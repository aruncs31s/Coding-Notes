---
id: linked_lists_dsa_go
aliases:
  - Linked Lists in Go
  - Singly and Doubly Linked Lists
tags:
  - go
  - dsa
  - linked-list
dg-publish: true
---

# Linked Lists in Go

A **Linked List** is a linear collection of data elements whose order is not given by their physical placement in memory. Instead, each element points to the next (and optionally previous) element via a pointer.

Unlike C, Go features automated Garbage Collection (GC). When you sever a node's reference (and nothing else points to it), Go's runtime will automatically reclaim its memory.

---

## 1. Singly Linked List (Generic Implementation)

Each node points forward to `Next`.

```
[ Head: 10 ] ──> [ 20 ] ──> [ 30 ] ──> nil
```

```go
package main

import "fmt"

// SNode represents a node in a singly linked list.
type SNode[T comparable] struct {
	Val  T
	Next *SNode[T]
}

// SinglyLinkedList manages node pointers and size.
type SinglyLinkedList[T comparable] struct {
	Head *SNode[T]
	Size int
}

func NewSinglyLinkedList[T comparable]() *SinglyLinkedList[T] {
	return &SinglyLinkedList[T]{}
}

// Prepend adds a node to the beginning of the list. Time: O(1)
func (l *SinglyLinkedList[T]) Prepend(val T) {
	newNode := &SNode[T]{Val: val, Next: l.Head}
	l.Head = newNode
	l.Size++
}

// Append adds a node to the end of the list. Time: O(n)
func (l *SinglyLinkedList[T]) Append(val T) {
	newNode := &SNode[T]{Val: val}
	if l.Head == nil {
		l.Head = newNode
		l.Size++
		return
	}

	curr := l.Head
	for curr.Next != nil {
		curr = curr.Next
	}
	curr.Next = newNode
	l.Size++
}

// Delete removes the first occurrence of val. Time: O(n)
func (l *SinglyLinkedList[T]) Delete(val T) bool {
	if l.Head == nil {
		return false
	}

	// Target is at head
	if l.Head.Val == val {
		l.Head = l.Head.Next
		l.Size--
		return true
	}

	curr := l.Head
	for curr.Next != nil && curr.Next.Val != val {
		curr = curr.Next
	}

	if curr.Next != nil {
		curr.Next = curr.Next.Next
		l.Size--
		return true
	}
	return false
}

// Print displays the list contents.
func (l *SinglyLinkedList[T]) Print() {
	curr := l.Head
	for curr != nil {
		fmt.Printf("%v -> ", curr.Val)
		curr = curr.Next
	}
	fmt.Println("nil")
}
```

---

## 2. Doubly Linked List

Each node points both forward (`Next`) and backward (`Prev`), allowing bidirectional traversal and $O(1)$ removal when the node pointer is known.

```
nil <── [ 10 ] <───> [ 20 ] <───> [ 30 ] ──> nil
```

```go
type DNode[T comparable] struct {
	Val  T
	Prev *DNode[T]
	Next *DNode[T]
}

type DoublyLinkedList[T comparable] struct {
	Head *DNode[T]
	Tail *DNode[T]
	Size int
}

func NewDoublyLinkedList[T comparable]() *DoublyLinkedList[T] {
	return &DoublyLinkedList[T]{}
}

// PushFront inserts at the head. Time: O(1)
func (d *DoublyLinkedList[T]) PushFront(val T) {
	newNode := &DNode[T]{Val: val, Next: d.Head}
	if d.Head != nil {
		d.Head.Prev = newNode
	} else {
		d.Tail = newNode // List was empty
	}
	d.Head = newNode
	d.Size++
}

// PushBack inserts at the tail. Time: O(1)
func (d *DoublyLinkedList[T]) PushBack(val T) {
	newNode := &DNode[T]{Val: val, Prev: d.Tail}
	if d.Tail != nil {
		d.Tail.Next = newNode
	} else {
		d.Head = newNode // List was empty
	}
	d.Tail = newNode
	d.Size++
}

// RemoveNode unlinks an arbitrary node in O(1) time.
func (d *DoublyLinkedList[T]) RemoveNode(node *DNode[T]) {
	if node == nil {
		return
	}

	if node.Prev != nil {
		node.Prev.Next = node.Next
	} else {
		d.Head = node.Next // Node was head
	}

	if node.Next != nil {
		node.Next.Prev = node.Prev
	} else {
		d.Tail = node.Prev // Node was tail
	}

	node.Prev = nil
	node.Next = nil
	d.Size--
}
```

---

## 3. Essential Linked List Algorithms

### Algorithm 1: Reverse a Singly Linked List (Iterative)
Three pointers (`prev`, `curr`, `next`) walk the list, re-wiring pointers.

```
Initial:   nil [prev]    10 [curr] ──> 20 [next] ──> 30 ──> nil
Step 1:                  10 [prev] ──> nil,  20 [curr]
Final:     nil <── 10 <── 20 <── 30 [prev]
```

```go
// ReverseList inverts pointers iteratively.
// Time: O(n), Space: O(1)
func ReverseList[T comparable](head *SNode[T]) *SNode[T] {
	var prev *SNode[T] = nil
	curr := head

	for curr != nil {
		nextTemp := curr.Next
		curr.Next = prev
		prev = curr
		curr = nextTemp
	}
	return prev
}
```

### Algorithm 2: Cycle Detection (Floyd's Tortoise and Hare)
Using two pointers moving at different speeds (`slow` moves 1 step, `fast` moves 2 steps):
- If there is a cycle, `fast` will inevitably lap and meet `slow`.
- If `fast` reaches `nil`, no cycle exists.

```go
// HasCycle checks if a linked list contains a cycle.
// Time: O(n), Space: O(1)
func HasCycle[T comparable](head *SNode[T]) bool {
	if head == nil || head.Next == nil {
		return false
	}

	slow := head
	fast := head

	for fast != nil && fast.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next

		if slow == fast {
			return true // Cycle detected!
		}
	}
	return false
}

// DetectCycleStart returns the node where the cycle begins, or nil.
func DetectCycleStart[T comparable](head *SNode[T]) *SNode[T] {
	slow, fast := head, head

	for fast != nil && fast.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next

		if slow == fast {
			// Phase 2: find entrance
			p1 := head
			p2 := slow
			for p1 != p2 {
				p1 = p1.Next
				p2 = p2.Next
			}
			return p1
		}
	}
	return nil
}
```

### Algorithm 3: Merge Two Sorted Lists
Given heads of two sorted lists, splice them together in sorted order.

```go
import "cmp"

// MergeTwoLists merges two sorted linked lists into one sorted list.
// Time: O(n + m), Space: O(1)
func MergeTwoLists[T cmp.Ordered](l1 *SNode[T], l2 *SNode[T]) *SNode[T] {
	dummy := &SNode[T]{}
	tail := dummy

	for l1 != nil && l2 != nil {
		if l1.Val <= l2.Val {
			tail.Next = l1
			l1 = l1.Next
		} else {
			tail.Next = l2
			l2 = l2.Next
		}
		tail = tail.Next
	}

	if l1 != nil {
		tail.Next = l1
	} else if l2 != nil {
		tail.Next = l2
	}

	return dummy.Next
}
```
