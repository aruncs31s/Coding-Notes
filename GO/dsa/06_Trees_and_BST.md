---
id: trees_and_bst_dsa_go
aliases:
  - Trees and BST in Go
  - Binary Search Tree
tags:
  - go
  - dsa
  - trees
  - bst
  - algorithms
dg-publish: true
---

# Trees and Binary Search Trees (BST) in Go

A **Binary Tree** is a hierarchical data structure in which each node has at most two children: `Left` and `Right`.
A **Binary Search Tree (BST)** enforces the invariant:
$$\text{Left.Val} < \text{Root.Val} < \text{Right.Val}$$

```
        50
       /  \
     30    70
    /  \   / \
   20  40 60  80
```

---

## 1. Node Definition & Tree Structure

```go
package main

import (
	"cmp"
	"fmt"
)

// TreeNode represents a node in a generic binary search tree.
type TreeNode[T cmp.Ordered] struct {
	Val   T
	Left  *TreeNode[T]
	Right *TreeNode[T]
}

// BST encapsulates root pointer.
type BST[T cmp.Ordered] struct {
	Root *TreeNode[T]
}

func NewBST[T cmp.Ordered]() *BST[T] {
	return &BST[T]{}
}
```

---

## 2. Core BST Operations

### Insertion
Time: Average $O(\log n)$, Worst $O(n)$

```go
func (t *BST[T]) Insert(val T) {
	t.Root = insertNode(t.Root, val)
}

func insertNode[T cmp.Ordered](node *TreeNode[T], val T) *TreeNode[T] {
	if node == nil {
		return &TreeNode[T]{Val: val}
	}

	if val < node.Val {
		node.Left = insertNode(node.Left, val)
	} else if val > node.Val {
		node.Right = insertNode(node.Right, val)
	}
	// Note: duplicate values ignored in standard BST
	return node
}
```

### Search
Time: Average $O(\log n)$, Worst $O(n)$

```go
func (t *BST[T]) Search(val T) bool {
	curr := t.Root
	for curr != nil {
		switch {
		case val == curr.Val:
			return true
		case val < curr.Val:
			curr = curr.Left
		case val > curr.Val:
			curr = curr.Right
		}
	}
	return false
}
```

### Deletion (Handling 0, 1, or 2 Children)
When deleting a node with 2 children, replace its value with its **inorder successor** (the minimum node in the right subtree) and delete the successor.

```go
func (t *BST[T]) Delete(val T) {
	t.Root = deleteNode(t.Root, val)
}

func deleteNode[T cmp.Ordered](node *TreeNode[T], val T) *TreeNode[T] {
	if node == nil {
		return nil
	}

	if val < node.Val {
		node.Left = deleteNode(node.Left, val)
	} else if val > node.Val {
		node.Right = deleteNode(node.Right, val)
	} else {
		// Found node to delete!

		// Case 1 & 2: 0 or 1 child
		if node.Left == nil {
			return node.Right
		} else if node.Right == nil {
			return node.Left
		}

		// Case 3: 2 children -> find inorder successor (min in right subtree)
		minRight := findMin(node.Right)
		node.Val = minRight.Val
		node.Right = deleteNode(node.Right, minRight.Val)
	}
	return node
}

func findMin[T cmp.Ordered](node *TreeNode[T]) *TreeNode[T] {
	curr := node
	for curr.Left != nil {
		curr = curr.Left
	}
	return curr
}
```

---

## 3. Depth-First Tree Traversals (DFS)

### Inorder Traversal (Left -> Root -> Right)
> [!NOTE]
> Inorder traversal of a BST visits nodes in strictly **sorted order**.

```go
// Inorder recursive. Time: O(n), Space: O(h)
func InorderTraversal[T cmp.Ordered](root *TreeNode[T]) []T {
	var result []T
	var dfs func(n *TreeNode[T])
	dfs = func(n *TreeNode[T]) {
		if n == nil {
			return
		}
		dfs(n.Left)
		result = append(result, n.Val)
		dfs(n.Right)
	}
	dfs(root)
	return result
}
```

### Preorder (Root -> Left -> Right) & Postorder (Left -> Right -> Root)

```go
func PreorderTraversal[T cmp.Ordered](root *TreeNode[T]) []T {
	var result []T
	var dfs func(n *TreeNode[T])
	dfs = func(n *TreeNode[T]) {
		if n == nil {
			return
		}
		result = append(result, n.Val)
		dfs(n.Left)
		dfs(n.Right)
	}
	dfs(root)
	return result
}

func PostorderTraversal[T cmp.Ordered](root *TreeNode[T]) []T {
	var result []T
	var dfs func(n *TreeNode[T])
	dfs = func(n *TreeNode[T]) {
		if n == nil {
			return
		}
		dfs(n.Left)
		dfs(n.Right)
		result = append(result, n.Val)
	}
	dfs(root)
	return result
}
```

---

## 4. Breadth-First Tree Traversal (Level-Order BFS)

Visits tree nodes level by level using a queue.
Time: $O(n)$, Space: $O(w)$ where $w$ is maximum tree width.

```go
func LevelOrder[T cmp.Ordered](root *TreeNode[T]) [][]T {
	var result [][]T
	if root == nil {
		return result
	}

	queue := []*TreeNode[T]{root}

	for len(queue) > 0 {
		levelSize := len(queue)
		currentLevel := make([]T, 0, levelSize)

		for i := 0; i < levelSize; i++ {
			node := queue[0]
			queue = queue[1:]

			currentLevel = append(currentLevel, node.Val)

			if node.Left != nil {
				queue = append(queue, node.Left)
			}
			if node.Right != nil {
				queue = append(queue, node.Right)
			}
		}

		result = append(result, currentLevel)
	}

	return result
}
```

---

## 5. Validate Binary Search Tree (BST Invariant Check)

A common mistake is checking only `node.Left.Val < node.Val && node.Right.Val > node.Val`. Every node in the left subtree must be smaller than the root, and every node in the right subtree must be larger.

```go
func IsValidBST(root *TreeNode[int]) bool {
	return validate(root, nil, nil)
}

func validate(node *TreeNode[int], minVal, maxVal *int) bool {
	if node == nil {
		return true
	}

	if (minVal != nil && node.Val <= *minVal) || (maxVal != nil && node.Val >= *maxVal) {
		return false
	}

	return validate(node.Left, minVal, &node.Val) && validate(node.Right, &node.Val, maxVal)
}
```
