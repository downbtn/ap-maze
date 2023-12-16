package maze

import (
	"container/heap"
	"errors"
	"math"
)

type item struct {
	pos    Coords
	weight int
	index  int
}

type pointQueue []*item

// these methods have to be implemented so we can use container/heap library
// https://pkg.go.dev/container/heap

func (q pointQueue) Len() int {
	return len(q)
}

func (q pointQueue) Less(i, j int) bool {
	return q[i].weight < q[j].weight
}

func (q pointQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}

func (q *pointQueue) Push(x any) {
	n := len(*q)
	item := x.(*item)
	item.index = n
	*q = append(*q, item)
}

func (q *pointQueue) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	item.index = -1
	old[n-1] = nil
	*q = old[:n-1]
	return item
}

// CreateSpt creates a shortest path tree using Dijkstra's algorithm given a
// certain point on a board.
// This is intended to be used with generated mazes, so the coordinates should
// be (2m+1, 2n+1) where m and n are integers (i.e. one of the "cells" used in
// generation and not the tunnels between them).
func (m *Maze) CreateSpt(src Coords) ([][]int, error) {
	if len(m.Board)%2 != 1 || len(m.Board[0])%2 != 1 {
		return nil, errors.New("Invalid board dimensions. Are you sure this is a generated maze?")
	}
	if src.X%2 != 1 || src.Y%2 != 1 {
		return nil, errors.New("Source point is not a \"cell\" (2m+1, 2n+1 form)")
	}

	// "real" refers to the coordinate it would occupy as a cell during
	// generation. I.e., the upper leftmost cell would be an empty space
	// located at (1,1) on the board, but its real coordinate would be
	// (0,0)
	var realHeight = (len(m.Board) - 1) / 2
	var realWidth = (len(m.Board[0]) - 1) / 2
	var realSrc = Coords{X: (src.X - 1) / 2, Y: (src.Y - 1) / 2}

	// https://www.geeksforgeeks.org/dijkstras-shortest-path-algorithm-using-priority_queue-stl

	distances := make([][]int, realHeight)
	for i, _ := range distances {
		distances[i] = make([]int, realWidth)
		for j, _ := range distances[i] {
			distances[i][j] = math.MaxInt
		}
	}

	distances[realSrc.Y][realSrc.X] = 0

	var pq pointQueue = make([]*item, 0, realWidth*realHeight)
	heap.Init(&pq)

	heap.Push(&pq, &item{
		pos:    Coords{X: realSrc.X, Y: realSrc.Y},
		weight: 0,
	})

	for pq.Len() != 0 {
		// get the lowest "weight" square in the queue
		current := pq.Pop().(*item)

		// Check all accessible adjacent squares
		adj := make([]Coords, 0, 4)
		// we *shouldn't* need to check if the coordinate is zero or
		// maximum, because then the board should have a wall there
		if m.Board[current.pos.Y*2][current.pos.X*2+1] == TILE_EMPTY {
			adj = append(adj, Coords{X: current.pos.X, Y: current.pos.Y - 1})
		}
		if m.Board[current.pos.Y*2+2][current.pos.X*2+1] == TILE_EMPTY {
			adj = append(adj, Coords{X: current.pos.X, Y: current.pos.Y + 1})
		}
		if m.Board[current.pos.Y*2+1][current.pos.X*2+2] == TILE_EMPTY {
			adj = append(adj, Coords{X: current.pos.X + 1, Y: current.pos.Y})
		}
		if m.Board[current.pos.Y*2+1][current.pos.X*2] == TILE_EMPTY {
			adj = append(adj, Coords{X: current.pos.X - 1, Y: current.pos.Y})
		}

		for _, point := range adj {
			newDist := distances[current.pos.Y][current.pos.X] + 1
			if newDist < distances[point.Y][point.X] {
				distances[point.Y][point.X] = newDist
				heap.Push(&pq, &item{pos: point, weight: newDist})
			}
		}
	}

	return distances, nil
}
