package main

import (
	"cmp"
	"container/heap"
	"errors"
	"fmt"
	"testing"
)

// --- Arrays & Strings ---
func Reverse[T any](s []T) {
	left, right := 0, len(s)-1
	for left < right {
		s[left], s[right] = s[right], s[left]
		left++
		right--
	}
}

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
		windowSum += nums[i] - nums[i-k]
		if windowSum > maxSum {
			maxSum = windowSum
		}
	}
	return maxSum
}

// --- Linked List ---
type SNode[T comparable] struct {
	Val  T
	Next *SNode[T]
}

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

// --- Stack & Queue ---
type Stack[T any] struct {
	items []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{items: make([]T, 0)}
}

func (s *Stack[T]) Push(val T) {
	s.items = append(s.items, val)
}

func (s *Stack[T]) Pop() (T, error) {
	var zero T
	if len(s.items) == 0 {
		return zero, errors.New("empty")
	}
	lastIdx := len(s.items) - 1
	val := s.items[lastIdx]
	s.items = s.items[:lastIdx]
	return val, nil
}

type CircularQueue[T any] struct {
	data  []T
	head  int
	tail  int
	count int
}

func NewCircularQueue[T any](initialCap int) *CircularQueue[T] {
	if initialCap < 2 {
		initialCap = 4
	}
	return &CircularQueue[T]{data: make([]T, initialCap)}
}

func (q *CircularQueue[T]) Enqueue(val T) {
	if q.count == len(q.data) {
		newData := make([]T, len(q.data)*2)
		for i := 0; i < q.count; i++ {
			newData[i] = q.data[(q.head+i)%len(q.data)]
		}
		q.data = newData
		q.head = 0
		q.tail = q.count
	}
	q.data[q.tail] = val
	q.tail = (q.tail + 1) % len(q.data)
	q.count++
}

func (q *CircularQueue[T]) Dequeue() (T, error) {
	var zero T
	if q.count == 0 {
		return zero, errors.New("empty")
	}
	val := q.data[q.head]
	q.head = (q.head + 1) % len(q.data)
	q.count--
	return val, nil
}

// --- Binary Search ---
func BinarySearch[T cmp.Ordered](arr []T, target T) int {
	low, high := 0, len(arr)-1
	for low <= high {
		mid := low + (high-low)/2
		if arr[mid] == target {
			return mid
		} else if arr[mid] < target {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return -1
}

func FirstOccurrence(nums []int, target int) int {
	low, high := 0, len(nums)-1
	res := -1
	for low <= high {
		mid := low + (high-low)/2
		if nums[mid] == target {
			res = mid
			high = mid - 1
		} else if nums[mid] < target {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return res
}

func SearchRotated(nums []int, target int) int {
	low, high := 0, len(nums)-1
	for low <= high {
		mid := low + (high-low)/2
		if nums[mid] == target {
			return mid
		}
		if nums[low] <= nums[mid] {
			if target >= nums[low] && target < nums[mid] {
				high = mid - 1
			} else {
				low = mid + 1
			}
		} else {
			if target > nums[mid] && target <= nums[high] {
				low = mid + 1
			} else {
				high = mid - 1
			}
		}
	}
	return -1
}

// --- Trees & BST ---
type TreeNode[T cmp.Ordered] struct {
	Val   T
	Left  *TreeNode[T]
	Right *TreeNode[T]
}

func InsertBST[T cmp.Ordered](node *TreeNode[T], val T) *TreeNode[T] {
	if node == nil {
		return &TreeNode[T]{Val: val}
	}
	if val < node.Val {
		node.Left = InsertBST(node.Left, val)
	} else if val > node.Val {
		node.Right = InsertBST(node.Right, val)
	}
	return node
}

func InorderTraversal[T cmp.Ordered](root *TreeNode[T]) []T {
	var res []T
	var dfs func(n *TreeNode[T])
	dfs = func(n *TreeNode[T]) {
		if n == nil {
			return
		}
		dfs(n.Left)
		res = append(res, n.Val)
		dfs(n.Right)
	}
	dfs(root)
	return res
}

// --- Heap ---
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

// --- Graphs ---
type Graph[T comparable] struct {
	adjList map[T][]T
}

func NewGraph[T comparable]() *Graph[T] {
	return &Graph[T]{adjList: make(map[T][]T)}
}

func (g *Graph[T]) AddEdge(u, v T, directed bool) {
	g.adjList[u] = append(g.adjList[u], v)
	if !directed {
		g.adjList[v] = append(g.adjList[v], u)
	}
}

func (g *Graph[T]) BFS(start T) []T {
	visited := make(map[T]bool)
	var order []T
	queue := []T{start}
	visited[start] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		order = append(order, curr)

		for _, neighbor := range g.adjList[curr] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}
	return order
}

func (g *Graph[T]) DFSRecursive(start T) []T {
	visited := make(map[T]bool)
	var order []T

	var dfs func(node T)
	dfs = func(node T) {
		visited[node] = true
		order = append(order, node)
		for _, neighbor := range g.adjList[node] {
			if !visited[neighbor] {
				dfs(neighbor)
			}
		}
	}
	dfs(start)
	return order
}

func (g *Graph[T]) ShortestPath(start, dest T) ([]T, bool) {
	if start == dest {
		return []T{start}, true
	}
	visited := make(map[T]bool)
	parent := make(map[T]T)
	queue := []T{start}
	visited[start] = true
	found := false

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		if curr == dest {
			found = true
			break
		}
		for _, neighbor := range g.adjList[curr] {
			if !visited[neighbor] {
				visited[neighbor] = true
				parent[neighbor] = curr
				queue = append(queue, neighbor)
			}
		}
	}

	if !found {
		return nil, false
	}

	var path []T
	for curr := dest; ; curr = parent[curr] {
		path = append(path, curr)
		if curr == start {
			break
		}
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path, true
}

func (g *Graph[T]) TopologicalSort() ([]T, error) {
	inDegree := make(map[T]int)
	for u := range g.adjList {
		if _, exists := inDegree[u]; !exists {
			inDegree[u] = 0
		}
		for _, v := range g.adjList[u] {
			inDegree[v]++
		}
	}

	var queue []T
	for node, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, node)
		}
	}

	var topoOrder []T
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		topoOrder = append(topoOrder, curr)
		for _, neighbor := range g.adjList[curr] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(topoOrder) != len(inDegree) {
		return nil, fmt.Errorf("cycle detected")
	}
	return topoOrder, nil
}

// --- UNIT TESTS ---

func TestDSA(t *testing.T) {
	// 1. Array & Slices
	arr := []int{1, 2, 3, 4, 5}
	Reverse(arr)
	if arr[0] != 5 || arr[4] != 1 {
		t.Fatalf("Reverse failed, got: %v", arr)
	}

	maxSum := MaxSubarraySumFixed([]int{2, 1, 5, 1, 3, 2}, 3)
	if maxSum != 9 {
		t.Fatalf("MaxSubarraySumFixed failed: expected 9, got %d", maxSum)
	}

	// 2. Linked List
	head := &SNode[int]{Val: 1, Next: &SNode[int]{Val: 2, Next: &SNode[int]{Val: 3}}}
	rev := ReverseList(head)
	if rev.Val != 3 || rev.Next.Val != 2 || rev.Next.Next.Val != 1 {
		t.Fatalf("ReverseList failed")
	}

	// 3. Stack & Queue
	st := NewStack[string]()
	st.Push("hello")
	st.Push("world")
	val, _ := st.Pop()
	if val != "world" {
		t.Fatalf("Stack pop failed: expected 'world', got %s", val)
	}

	cq := NewCircularQueue[int](2)
	cq.Enqueue(10)
	cq.Enqueue(20)
	cq.Enqueue(30) // triggers dynamic resize
	q1, _ := cq.Dequeue()
	q2, _ := cq.Dequeue()
	if q1 != 10 || q2 != 20 {
		t.Fatalf("CircularQueue failed: expected 10, 20, got %d, %d", q1, q2)
	}

	// 4. Binary Search
	sortedArr := []int{2, 5, 8, 12, 16, 23, 38, 56, 72, 91}
	idx := BinarySearch(sortedArr, 23)
	if idx != 5 {
		t.Fatalf("BinarySearch failed: expected 5, got %d", idx)
	}
	if BinarySearch(sortedArr, 999) != -1 {
		t.Fatalf("BinarySearch failed for missing element")
	}

	dupArr := []int{2, 4, 5, 5, 5, 5, 7, 9}
	firstIdx := FirstOccurrence(dupArr, 5)
	if firstIdx != 2 {
		t.Fatalf("FirstOccurrence failed: expected 2, got %d", firstIdx)
	}

	rotated := []int{4, 5, 6, 7, 0, 1, 2}
	if SearchRotated(rotated, 0) != 4 {
		t.Fatalf("SearchRotated failed: expected 4, got %d", SearchRotated(rotated, 0))
	}

	// 5. BST
	var bst *TreeNode[int]
	vals := []int{50, 30, 70, 20, 40, 60, 80}
	for _, v := range vals {
		bst = InsertBST(bst, v)
	}
	inorder := InorderTraversal(bst)
	expectedInorder := []int{20, 30, 40, 50, 60, 70, 80}
	for i, v := range expectedInorder {
		if inorder[i] != v {
			t.Fatalf("BST Inorder mismatch at index %d", i)
		}
	}

	// 6. Heap
	h := &IntMinHeap{20, 5, 15, 3}
	heap.Init(h)
	min := heap.Pop(h).(int)
	if min != 3 {
		t.Fatalf("MinHeap Pop expected 3, got %d", min)
	}

	// 7. Graph BFS, DFS, Shortest Path, Topological Sort
	g := NewGraph[string]()
	g.AddEdge("A", "B", true)
	g.AddEdge("A", "C", true)
	g.AddEdge("B", "D", true)
	g.AddEdge("C", "D", true)
	g.AddEdge("D", "E", true)

	bfsOrder := g.BFS("A")
	if len(bfsOrder) != 5 || bfsOrder[0] != "A" {
		t.Fatalf("BFS traversal unexpected: %v", bfsOrder)
	}

	dfsOrder := g.DFSRecursive("A")
	if len(dfsOrder) != 5 || dfsOrder[0] != "A" {
		t.Fatalf("DFS traversal unexpected: %v", dfsOrder)
	}

	path, ok := g.ShortestPath("A", "E")
	if !ok || len(path) != 4 { // A -> B -> D -> E (or A -> C -> D -> E)
		t.Fatalf("ShortestPath failed: path=%v, ok=%v", path, ok)
	}

	dag := NewGraph[string]()
	dag.AddEdge("cook", "eat", true)
	dag.AddEdge("buy", "cook", true)
	dag.AddEdge("wash", "cook", true)
	topo, err := dag.TopologicalSort()
	if err != nil || len(topo) != 4 {
		t.Fatalf("TopologicalSort failed: %v, order=%v", err, topo)
	}
}
