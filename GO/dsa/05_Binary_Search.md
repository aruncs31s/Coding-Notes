---
id: binary_search_dsa_go
aliases:
  - Binary Search in Go
  - Binary Search Variants
tags:
  - go
  - dsa
  - binary-search
  - algorithms
dg-publish: true
---

# Binary Search in Go

**Binary Search** is an efficient divide-and-conquer search algorithm that works on **sorted** collections. It runs in $O(\log n)$ time by repeatedly cutting the search space in half.

```
Target = 23
Array: [ 2, 5, 8, 12, 16, 23, 38, 56, 72, 91 ]
         ▲             ▲                    ▲
        low           mid                 high
                      (16 < 23 -> search right half)
```

---

## 1. Avoiding Integer Overflow

A common bug in languages like C/C++/Java/Go is:
```go
mid := (low + high) / 2 // Potential overflow if low + high exceeds max int!
```
Always use:
```go
mid := low + (high - low)/2
// or using bit shift:
mid := int(uint(low+high) >> 1)
```

---

## 2. Standard Binary Search (Iterative & Recursive)

### Iterative Implementation (Recommended)
Time: $O(\log n)$, Space: $O(1)$

```go
package main

import (
	"cmp"
	"fmt"
)

// BinarySearch returns the index of target in sorted slice, or -1 if not found.
func BinarySearch[T cmp.Ordered](arr []T, target T) int {
	low := 0
	high := len(arr) - 1

	for low <= high {
		mid := low + (high-low)/2

		switch {
		case arr[mid] == target:
			return mid
		case arr[mid] < target:
			low = mid + 1 // Target is in right half
		case arr[mid] > target:
			high = mid - 1 // Target is in left half
		}
	}

	return -1 // Not found
}
```

### Recursive Implementation
Time: $O(\log n)$, Space: $O(\log n)$ due to call stack.

```go
func BinarySearchRecursive[T cmp.Ordered](arr []T, target T, low, high int) int {
	if low > high {
		return -1
	}

	mid := low + (high-low)/2

	if arr[mid] == target {
		return mid
	}
	if arr[mid] > target {
		return BinarySearchRecursive(arr, target, low, mid-1)
	}
	return BinarySearchRecursive(arr, target, mid+1, high)
}
```

---

## 3. Boundary Search: First and Last Occurrence

When an array contains duplicate target elements, standard binary search can return any matching index. We often need the **exact bounds**.

```
Array:  [ 2,  4,  5,  5,  5,  5,  7,  9 ]
                      ▲   ▲
                  First   Last
```

### First Occurrence (Lower Bound for equality)

```go
// FirstOccurrence finds the first index of target, or -1 if not present.
// Time: O(log n), Space: O(1)
func FirstOccurrence(nums []int, target int) int {
	low, high := 0, len(nums)-1
	result := -1

	for low <= high {
		mid := low + (high-low)/2

		if nums[mid] == target {
			result = mid
			high = mid - 1 // Continue searching to the left
		} else if nums[mid] < target {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return result
}
```

### Last Occurrence (Upper Bound for equality)

```go
// LastOccurrence finds the last index of target, or -1 if not present.
// Time: O(log n), Space: O(1)
func LastOccurrence(nums []int, target int) int {
	low, high := 0, len(nums)-1
	result := -1

	for low <= high {
		mid := low + (high-low)/2

		if nums[mid] == target {
			result = mid
			low = mid + 1 // Continue searching to the right
		} else if nums[mid] < target {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return result
}
```

---

## 4. Search in Rotated Sorted Array

An array sorted in ascending order is rotated at some unknown pivot (e.g., `[4, 5, 6, 7, 0, 1, 2]`).

**Key Insight**: At least one half of the array is always strictly sorted.

```go
// SearchRotated searches for target in a rotated sorted array without duplicates.
// Time: O(log n), Space: O(1)
func SearchRotated(nums []int, target int) int {
	low, high := 0, len(nums)-1

	for low <= high {
		mid := low + (high-low)/2

		if nums[mid] == target {
			return mid
		}

		// Check if left half is sorted
		if nums[low] <= nums[mid] {
			if target >= nums[low] && target < nums[mid] {
				high = mid - 1 // Target lies within sorted left half
			} else {
				low = mid + 1  // Target is in right half
			}
		} else {
			// Right half must be sorted
			if target > nums[mid] && target <= nums[high] {
				low = mid + 1  // Target lies within sorted right half
			} else {
				high = mid - 1 // Target is in left half
			}
		}
	}
	return -1
}
```

---

## 5. Go Standard Library Solutions

Go includes built-in binary search utilities in both `sort` and `slices` (since Go 1.21).

### `slices.BinarySearch` (Go 1.21+)
Returns index and boolean indicating whether the target was found:

```go
import "slices"

func ExampleSlicesBinarySearch() {
	numbers := []int{10, 20, 30, 40, 50}

	idx, found := slices.BinarySearch(numbers, 30)
	fmt.Printf("found: %v, index: %d\n", found, idx) // true, 2

	// If not found, idx is insertion position to maintain sorted order
	idx, found = slices.BinarySearch(numbers, 35)
	fmt.Printf("found: %v, insert at: %d\n", found, idx) // false, 3
}
```

### `sort.Search` (Predicate-Based Binary Search)
Finds smallest index `i` in `[0, n)` where `f(i)` is true.

```go
import "sort"

func ExampleSortSearch() {
	data := []int{1, 3, 5, 7, 9, 11}
	target := 7

	// Condition: data[i] >= target
	idx := sort.Search(len(data), func(i int) bool {
		return data[i] >= target
	})

	if idx < len(data) && data[idx] == target {
		fmt.Printf("Target %d found at index %d\n", target, idx)
	}
}
```
