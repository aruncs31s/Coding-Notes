---
id: arrays_and_strings_dsa_go
aliases:
  - Arrays and Strings in Go
  - Slice Algorithms
tags:
  - go
  - dsa
  - arrays
  - algorithms
dg-publish: true
---

# Arrays, Slices & Common Patterns in Go

In Go, **arrays** have a fixed length specified at compile time (`[5]int`), whereas **slices** are dynamically sized views over underlying arrays (`[]int`). Almost all algorithmic problem solving in Go uses slices.

---

## 1. Two-Pointer Technique

The two-pointer technique uses two indices to traverse a slice simultaneously, typically reducing $O(n^2)$ time complexity to $O(n)$ with $O(1)$ extra space.

### Pattern A: Opposite Ends (Converging)
Used for palindrome verification, reversing elements, or finding pairs in sorted arrays.

```go
package main

import "fmt"

// Reverse reverses a slice in-place using two converging pointers.
// Time: O(n), Space: O(1)
func Reverse[T any](s []T) {
	left, right := 0, len(s)-1
	for left < right {
		s[left], s[right] = s[right], s[left]
		left++
		right--
	}
}

// TwoSumSorted finds two indices (1-based) whose values sum up to target in a sorted slice.
// Time: O(n), Space: O(1)
func TwoSumSorted(numbers []int, target int) (int, int, bool) {
	left, right := 0, len(numbers)-1

	for left < right {
		sum := numbers[left] + numbers[right]
		switch {
		case sum == target:
			return left + 1, right + 1, true
		case sum < target:
			left++ // Need a larger sum
		case sum > target:
			right-- // Need a smaller sum
		}
	}
	return -1, -1, false
}
```

### Pattern B: Fast and Slow Pointers (Read/Write)
Used for in-place removal or deduplication without allocating extra arrays.

```go
// RemoveDuplicates removes duplicates from a sorted slice in-place.
// Returns the length of the deduplicated prefix.
// Time: O(n), Space: O(1)
func RemoveDuplicates(nums []int) int {
	if len(nums) == 0 {
		return 0
	}

	slow := 0
	for fast := 1; fast < len(nums); fast++ {
		if nums[fast] != nums[slow] {
			slow++
			nums[slow] = nums[fast]
		}
	}
	return slow + 1
}
```

---

## 2. Sliding Window Technique

Used to process contiguous subarrays/substrings without recomputing overlapping ranges.

### Pattern A: Fixed-Size Sliding Window
Find the maximum sum of any contiguous subarray of size `k`.

```
Array: [2,  1,  5,  1,  3,  2],  k = 3
Window 1: [2, 1, 5] -> sum = 8
Window 2:    [1, 5, 1] -> sum = 8 - 2 + 1 = 7
Window 3:       [5, 1, 3] -> sum = 7 - 1 + 3 = 9 (Max)
```

```go
// MaxSubarraySumFixed finds the maximum sum of a subarray of fixed length k.
// Time: O(n), Space: O(1)
func MaxSubarraySumFixed(nums []int, k int) int {
	if len(nums) < k || k <= 0 {
		return 0
	}

	windowSum := 0
	for i := 0; i < k; i++ {
		windowSum += nums[i]
	}

	maxSum := windowSum
	for i := k; i < len(nums); i++ {
		// Slide window: add incoming element, remove outgoing element
		windowSum += nums[i] - nums[i-k]
		if windowSum > maxSum {
			maxSum = windowSum
		}
	}
	return maxSum
}
```

### Pattern B: Dynamic-Size Sliding Window
Find the minimal length of a contiguous subarray whose sum is $\ge$ `target`.

```go
import "math"

// MinSubArrayLen finds the shortest subarray length with sum >= target.
// Time: O(n), Space: O(1)
func MinSubArrayLen(target int, nums []int) int {
	minLen := math.MaxInt32
	left := 0
	currentSum := 0

	for right := 0; right < len(nums); right++ {
		currentSum += nums[right]

		// Contract the window from the left while condition holds
		for currentSum >= target {
			windowLen := right - left + 1
			if windowLen < minLen {
				minLen = windowLen
			}
			currentSum -= nums[left]
			left++
		}
	}

	if minLen == math.MaxInt32 {
		return 0
	}
	return minLen
}
```

---

## 3. Prefix Sum Technique

Precomputing prefix sums allows range queries ($L$ to $R$) in $O(1)$ time after $O(n)$ preprocessing.

$$\text{Prefix}[i] = \sum_{j=0}^{i-1} \text{nums}[j]$$
$$\text{Sum}(L, R) = \text{Prefix}[R+1] - \text{Prefix}[L]$$

```go
type PrefixSum struct {
	prefix []int
}

func NewPrefixSum(nums []int) *PrefixSum {
	prefix := make([]int, len(nums)+1)
	for i, v := range nums {
		prefix[i+1] = prefix[i] + v
	}
	return &PrefixSum{prefix: prefix}
}

// RangeSum returns the sum of elements from index left to right (inclusive).
// Time: O(1)
func (ps *PrefixSum) RangeSum(left, right int) int {
	return ps.prefix[right+1] - ps.prefix[left]
}
```

---

## 💡 Go Best Practices for Slice Manipulation
1. **Avoid slice memory leaks**: When sub-slicing a large slice `large[0:2]`, the garbage collector cannot free the underlying array of `large`. Use `copy()` to isolate small subsets:
   ```go
   sub := make([]int, 2)
   copy(sub, large[0:2])
   ```
2. **Preallocate slices with known capacity**:
   ```go
   // Good: 1 allocation
   res := make([]int, 0, len(input))
   // Bad: repeated reallocation as slice doubles
   var res []int
   ```
