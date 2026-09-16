---
id: graphs_dfs_and_bfs_dsa_go
aliases:
  - Graphs DFS and BFS in Go
  - Graph Algorithms
tags:
  - go
  - dsa
  - graphs
  - dfs
  - bfs
  - algorithms
dg-publish: true
---

# Graphs: Representation, DFS, and BFS in Go

A **Graph** $G = (V, E)$ consists of a set of vertices (nodes) and edges connecting pairs of vertices.

---

## 1. Graph Representations in Go

In Go, graphs are most commonly and flexibly represented using an **Adjacency List** via `map[T][]T` with generics.

```go
package main

import (
	"fmt"
)

// Graph represents an unweighted directed/undirected graph.
type Graph[T comparable] struct {
	adjList map[T][]T
}

func NewGraph[T comparable]() *Graph[T] {
	return &Graph[T]{
		adjList: make(map[T][]T),
	}
}

// AddEdge inserts an edge. Set directed=false for undirected graphs.
func (g *Graph[T]) AddEdge(u, v T, directed bool) {
	g.adjList[u] = append(g.adjList[u], v)
	if !directed {
		g.adjList[v] = append(g.adjList[v], u)
	}
}
```

---

## 2. Breadth-First Search (BFS)

### Intuition & Characteristics
- Traverses the graph **level by level** using a **Queue (FIFO)**.
- **Shortest Path Property**: In an unweighted graph, the first time BFS reaches a node, it is guaranteed to be via the shortest path (minimum number of edges).
- **Time Complexity**: $O(V + E)$
- **Space Complexity**: $O(V)$ for the queue and visited set.

```
       (A)  <- Level 0
      /   \
    (B)   (C) <- Level 1
    / \     \
  (D) (E)   (F) <- Level 2
Queue Progression: [A] -> [B, C] -> [C, D, E] -> [D, E, F] -> ...
```

### Standard BFS Traversal

```go
// BFS visits all nodes reachable from start.
func (g *Graph[T]) BFS(start T) []T {
	visited := make(map[T]bool)
	var order []T
	queue := []T{start}

	visited[start] = true

	for len(queue) > 0 {
		// Dequeue
		curr := queue[0]
		queue = queue[1:]
		order = append(order, curr)

		// Explore neighbors
		for _, neighbor := range g.adjList[curr] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	return order
}
```

### Shortest Path in Unweighted Graph (With Path Reconstruction)

```go
// ShortestPath returns the sequence of nodes along the shortest path from start to dest.
func (g *Graph[T]) ShortestPath(start, dest T) ([]T, bool) {
	if start == dest {
		return []T{start}, true
	}

	visited := make(map[T]bool)
	parent := make(map[T]T) // Tracks parent pointers for path backtracking
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

	// Reconstruct path backward from dest to start
	var path []T
	for curr := dest; ; curr = parent[curr] {
		path = append(path, curr)
		if curr == start {
			break
		}
	}

	// Reverse path to [start -> dest]
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path, true
}
```

---

## 3. Depth-First Search (DFS)

### Intuition & Characteristics
- Traverses as deep as possible along each branch before backtracking.
- Implemented either **recursively** (using the Go call stack) or **iteratively** (using an explicit stack).
- **Time Complexity**: $O(V + E)$
- **Space Complexity**: $O(V)$ auxiliary space.

```
       (A)
      /   \
    (B)   (E)
    / \
  (C) (D)
DFS Order: A -> B -> C -> (backtrack) -> D -> (backtrack) -> E
```

### Recursive DFS

```go
// DFSRecursive traverses graph starting at start node.
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
```

### Iterative DFS (Using Explicit Slice Stack)

```go
// DFSIterative traverses graph using an explicit stack.
func (g *Graph[T]) DFSIterative(start T) []T {
	visited := make(map[T]bool)
	var order []T
	stack := []T{start}

	for len(stack) > 0 {
		// Pop from top
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited[curr] {
			continue
		}
		visited[curr] = true
		order = append(order, curr)

		// Push unvisited neighbors (reverse order to match recursive exploration)
		neighbors := g.adjList[curr]
		for i := len(neighbors) - 1; i >= 0; i-- {
			neighbor := neighbors[i]
			if !visited[neighbor] {
				stack = append(stack, neighbor)
			}
		}
	}

	return order
}
```

---

## 4. Cycle Detection

### In Directed Graphs (3-Color / Tri-State DFS)
State tracking:
- `0 (White)`: Unvisited
- `1 (Gray)`: Currently visiting (in recursion stack / ancestor)
- `2 (Black)`: Completely processed

A back-edge pointing to a node currently in state `1 (Gray)` confirms a **cycle**.

```go
// HasCycleDirected checks for cycles in a directed graph.
func (g *Graph[T]) HasCycleDirected() bool {
	state := make(map[T]int) // 0: unvisited, 1: visiting, 2: visited

	var dfs func(node T) bool
	dfs = func(node T) bool {
		state[node] = 1 // Mark visiting

		for _, neighbor := range g.adjList[node] {
			if state[neighbor] == 1 {
				return true // Found back-edge -> Cycle!
			}
			if state[neighbor] == 0 {
				if dfs(neighbor) {
					return true
				}
			}
		}

		state[node] = 2 // Mark visited
		return false
	}

	for node := range g.adjList {
		if state[node] == 0 {
			if dfs(node) {
				return true
			}
		}
	}

	return false
}
```

---

## 5. Topological Sort (Kahn's Algorithm via BFS)

Orders vertices in a Directed Acyclic Graph (DAG) linearly such that for every directed edge $u \to v$, $u$ comes before $v$.

```
Build System / Task Dependency:
  [Compile A] ──> [Link B] ──> [Create Executable C]
```

```go
// TopologicalSort returns a valid topological ordering of vertices,
// or an error if a cycle exists.
// Time: O(V + E), Space: O(V)
func (g *Graph[T]) TopologicalSort() ([]T, error) {
	inDegree := make(map[T]int)

	// Initialize inDegree for all vertices
	for u := range g.adjList {
		if _, exists := inDegree[u]; !exists {
			inDegree[u] = 0
		}
		for _, v := range g.adjList[u] {
			inDegree[v]++
		}
	}

	// Collect all nodes with 0 in-degree into queue
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
		return nil, fmt.Errorf("cycle detected: topological sort impossible")
	}

	return topoOrder, nil
}
```

---

## 📊 BFS vs DFS Comparison

| Metric | Breadth-First Search (BFS) | Depth-First Search (DFS) |
|---|---|---|
| **Data Structure** | Queue (FIFO) | Stack (LIFO or recursion) |
| **Shortest Path** | Guarantees shortest path on unweighted graphs | Does not guarantee shortest path |
| **Memory Consumption** | $O(w)$ (can be large for wide trees/graphs) | $O(d)$ (proportional to max depth) |
| **Primary Use Cases** | Shortest path, level-order traversal, web crawlers | Topological sort, cycle detection, maze solving, connected components |
