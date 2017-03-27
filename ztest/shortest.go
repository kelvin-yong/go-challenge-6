package main

import (
	"errors"
	"fmt"
)

type node struct {
	processed bool
	dist      int
	parent    int
}

var graph = map[int][]int{
	1:  []int{2, 5},
	2:  []int{1},
	3:  []int{4, 7},
	4:  []int{3, 8},
	5:  []int{1, 6, 7},
	6:  []int{5, 7, 10},
	7:  []int{3, 6},
	8:  []int{4, 12},
	9:  []int{5, 10, 13},
	10: []int{6, 9, 11, 14},
	11: []int{10, 12, 15},
	12: []int{8, 11},
	13: []int{9, 14},
	14: []int{10, 13},
	15: []int{11, 16},
	16: []int{15},
}

const infinity = 100000

func nextNodeID(nodes map[int]*node) (int, error) {
	shortest := infinity
	item := -1

	for k, node := range nodes {
		if node.processed == false && node.dist < shortest {
			item = k
			shortest = node.dist
		}
	}

	if item == -1 {
		return item, errors.New("No more unprocessed nodes with some distance")
	}

	return item, nil
}

func mainX() {
	const source = 6
	interested := map[int]bool{
		1:  false,
		2:  false,
		3:  false,
		4:  false,
		5:  false,
		10: false,
		11: false,
		12: false,
		13: false,
		14: false,
		16: false,
	}

	// initialise
	n := len(graph)
	nodes := make(map[int]*node, n)
	for k := range graph {
		nodes[k] = &node{processed: false, dist: infinity, parent: -1}
	}
	nodes[source].dist = 0

	for {
		cur, err := nextNodeID(nodes)
		if err != nil {
			break
		}

		if _, found := interested[cur]; found {
			delete(interested, cur)
			if len(interested) == 0 {
				break
			}
		}

		nodes[cur].processed = true

		// for each of item's neighbour
		for _, neighbour := range graph[cur] {
			if nodes[neighbour].processed {
				continue
			}
			if nodes[cur].dist+1 < nodes[neighbour].dist {
				nodes[neighbour].dist = nodes[cur].dist + 1
				nodes[neighbour].parent = cur
			}
		}
	}

	for _, dest := range []int{1, 2, 3, 4, 5, 10, 11, 12, 13, 14, 16} {
		fmt.Print(dest, ":", nodes[dest].dist, "->")
		numSteps := nodes[dest].dist
		route := make([]int, numSteps)

		target := dest
		for target != source {
			numSteps--
			route[numSteps] = target
			target = nodes[target].parent
		}
		fmt.Println(route)
	}

}
