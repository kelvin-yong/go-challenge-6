// By Kelvin Yong for Go Challenge 6

package mazelib

import (
	"errors"
	"fmt"
)

// MazeReply is a struct to represent the reply form the maze server
type MazeReply struct {
	Survey Survey
	Err    error
}

// directionToMove is a utility method that returns the direction (N,S,E,W)
// to take to move from src to dest coordinates. the coordinates must be adjacent
func directionToMove(src, dest Coordinate) (direction int) {
	dx, dy := dest.X-src.X, dest.Y-src.Y
	switch {
	case dx == 1 && dy == 0:
		direction = E
	case dx == -1 && dy == 0:
		direction = W
	case dx == 0 && dy == 1:
		direction = S
	case dx == 0 && dy == -1:
		direction = N
	}
	return
}

// updatePosition is a utility method that returns the new x, y coordinates
// given a current(x,y) and the direction to move
func updatePosition(cx, cy, direction int) (newx, newy int) {
	newx, newy = cx, cy
	switch direction {
	case N:
		newy--
	case S:
		newy++
	case E:
		newx++
	case W:
		newx--
	}

	// var s string
	// fmt.Scanln(&s)

	return newx, newy
}

//////////// Dijkstra's algorithm ////////////
// see http://www.adjacencyMap-magics.com/articles/shortest_path.php
// we use Dijkstra's algorithm to find out which of the exisiting
// junction is nearest. We should backtrack to the nearest junction.

// adjacencyMap maps a Coordinate(x, y) to a slice of Coordinates which are
// accessible. Essentially it represents graph of nodes and their neighbours
type adjacencyMap map[Coordinate][]Coordinate

const infinity = 1000000

type node struct {
	processed bool
	dist      int
	parent    Coordinate
}

// given a map of nodes, find the node with the unprocessed node
// with the smallest dist and return the map key
func nextNodeID(nodes map[Coordinate]*node) (Coordinate, error) {
	found := false
	shortest := infinity
	var item Coordinate

	for k, node := range nodes {
		if node.processed == false && node.dist < shortest {
			item = k
			shortest = node.dist
			found = true
		}
	}

	if found {
		return item, nil
	}

	return item, errors.New("No more unprocessed nodes with some distance")
}

// given a current point (x, y), figure out the length of the shortest path
// for each of the junctions we want to reach.
// Return a list of steps for the nearest junction
func shortestPath(cx, cy int, graph adjacencyMap, junctions adjacencyMap) []int {
	destinations := make(map[Coordinate]bool, len(junctions))
	for k := range junctions {
		destinations[k] = true
	}

	// initialise
	source := Coordinate{cx, cy}

	n := len(graph)
	nodes := make(map[Coordinate]*node, n)
	for k := range graph {
		nodes[k] = &node{processed: false, dist: infinity}
	}
	nodes[source].dist = 0

	for {
		cur, err := nextNodeID(nodes)
		if err != nil {
			break
		}

		if destinations[cur] {
			delete(destinations, cur)
			if len(destinations) == 0 {
				// found shortest path for all destinations
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

	// find which junction is nearest
	numSteps := infinity
	var junction Coordinate

	for dest := range junctions {
		if nodes[dest].dist < numSteps {
			numSteps = nodes[dest].dist
			junction = dest
		}
	}

	// get the route of the target junction to go to
	route := make([]Coordinate, numSteps)
	target := junction
	for target != source {
		numSteps--
		route[numSteps] = target
		target = nodes[target].parent
	}

	directions := make([]int, 0, numSteps)
	prev := source
	for _, cur := range route {
		direction := directionToMove(prev, cur)
		directions = append(directions, direction)
		prev = cur
	}
	return directions
}

//////////// Main Solver ////////////
type bounds struct {
	xmin, ymin, xmax, ymax int
}

var boundary bounds
var lastDirection int

// A junction is a node that has at least one unvisited neighbour.
// Junction A and B may both point to another node (cx, cy) as
// an unvisited place.  If cx, cy is later newly visited,
// it must be removed from both A and B.
func cleanUpJunctions(cx, cy int, junctions adjacencyMap) {
	for k, paths := range junctions {
		newPaths := make([]Coordinate, 0, 4)

		for _, path := range paths {
			if path.X != cx || path.Y != cy {
				newPaths = append(newPaths, path)
			}
		}

		if len(newPaths) == 0 {
			delete(junctions, k)
		} else if len(newPaths) < len(paths) {
			junctions[k] = newPaths
		}
	}
}

// Given the current location (x, y) and the possible
// coordinates to move, pick the best way to go.
// Requires knowledge of estimated bounds and where has been visited
// Only useful if the maze has few walls
func priortisePaths(cx, cy int, paths []Coordinate, visited map[Coordinate]bool) {
	if len(paths) < 2 {
		// there's nothing to prioritise if you have only 1 or 0 paths.
		return
	}
	unexplored := make([]int, len(paths))
	for i, path := range paths {
		dx, dy := path.X-cx, path.Y-cy
		var startx, endx, starty, endy int
		switch {
		case dy == -1: //move north
			startx = boundary.xmin
			endx = boundary.xmax
			starty = boundary.ymin
			endy = path.Y
		case dy == 1: //move south
			startx = boundary.xmin
			endx = boundary.xmax
			starty = path.Y
			endy = boundary.ymax
		case dx == -1: //move west
			startx = boundary.ymin
			endx = path.X
			starty = boundary.ymin
			endy = boundary.ymax
		case dx == 1: // move east
			startx = path.X
			endx = boundary.ymin
			starty = boundary.ymin
			endy = boundary.ymax
		}

		u := 0
		for x := startx; x <= endx; x++ {
			for y := starty; y <= endy; y++ {
				if _, found := visited[Coordinate{x, y}]; !found {
					u++
				}
			}
		}
		unexplored[i] = u
	}

	// now we have the number of unexplored rooms for each direction
	// find what is the max and the respective index
	highestCount := unexplored[0]
	highestIndex := 0

	for i := 1; i < len(unexplored); i++ {
		if unexplored[i] > highestCount {
			highestCount = unexplored[i]
			highestIndex = i
		}
	}

	paths[0], paths[highestIndex] = paths[highestIndex], paths[0]
	unexplored[0], unexplored[highestIndex] = unexplored[highestIndex], unexplored[0]
}

// FindTreasure receives the surround surveys on replies channel
// and recommends the steps on the output channel.
// The algorithm behind FindTreasure is essentially Tremaux, but instead
// of backtracking from a deadend to get to the next junction, it uses
// Dijkstra's algoritm to find the nearest junction to go do.
// For a 15x10 empty maze, Tremaux takes an average of 95 steps
// whereas FindTreasure takes 85 steps alone.
// FindTreasure combined with priortisePaths reduces the average to 77
// Note: FindTreasure (with or without path prioritising) DOES NOT perform
// better than Tremaux for mazes with no loops
func FindTreasure(replies <-chan MazeReply) <-chan int {
	steps := make(chan int)

	// clear boundary at the start
	boundary = bounds{}
	estWidth, estHeight := 1, 1

	// relative x and y to starting position
	cx, cy := 0, 0

	// graph maps a (x, y) to connected (x, y)s
	graph := make(adjacencyMap)

	// keep track of all junctions that have at least one unvisited (x, y)
	junctions := make(adjacencyMap)

	// visited tracks the rooms that have been visited
	visited := make(map[Coordinate]bool)

	go func() {
		for {
			visited[Coordinate{cx, cy}] = true
			reply := <-replies

			survey, err := reply.Survey, reply.Err
			if err == ErrVictory {
				// solved, we are done
				break
			}

			//possible directions it can go from current x, y
			dirs := make([]int, 0, 4)
			if !survey.Left {
				dirs = append(dirs, W)
			}
			if !survey.Right {
				dirs = append(dirs, E)
			}
			if !survey.Top {
				dirs = append(dirs, N)
			}
			if !survey.Bottom {
				dirs = append(dirs, S)
			}
			Shuffle(dirs)

			// get all possible paths/coordinates you can go
			paths := make([]Coordinate, 0, 4)
			for _, dir := range dirs {
				newX, newY := cx+Delta[dir].X, cy+Delta[dir].Y
				paths = append(paths, Coordinate{newX, newY})

				// explore the new boundary
				if newX > boundary.xmax {
					boundary.xmax = newX
				} else if newX < boundary.xmin {
					boundary.xmin = newX
				}
				if newY > boundary.ymax {
					boundary.ymax = newY
				} else if newY < boundary.ymin {
					boundary.ymin = newY
				}
			}

			// add possible paths to graph
			graph[Coordinate{cx, cy}] = paths

			//add reverse directions to graph
			for _, path := range paths {
				newPaths, found := graph[path]
				if !found {
					newPaths = make([]Coordinate, 0, 4)
				}
				newPaths = append(newPaths, Coordinate{cx, cy})
				graph[path] = newPaths
			}

			if (estWidth != boundary.xmax-boundary.xmin+1) ||
				(estHeight != boundary.ymax-boundary.ymin+1) {
				estWidth = boundary.xmax - boundary.xmin + 1
				estHeight = boundary.ymax - boundary.ymin + 1
			}

			// of all the possible paths, how many unvisited previously?
			uvPaths := make([]Coordinate, 0, 4)
			for _, path := range paths {
				if !visited[path] {
					uvPaths = append(uvPaths, path)
				}
			}

			unvisited := len(uvPaths)
			var nextCoor Coordinate
			var nextDir int

			if unvisited == 0 {
				// deadend, need to backtrack
				stepsBack := shortestPath(cx, cy, graph, junctions)
				if len(stepsBack) == 0 {
					fmt.Println("Visited all places, but can't find treasured!")
					break
				}

				// backtrack as prescribed to a junction with a unvisted neighbour
				for _, dir := range stepsBack {
					cx, cy = updatePosition(cx, cy, dir)
					steps <- dir
					// read the reply from server and discard, not important
					// since they are all visited steps
					<-replies
				}

				if len(junctions[Coordinate{cx, cy}]) == 0 {
					panic("Invalid map. There is a one way wall")
				}

				// pick a new path to go
				priortisePaths(cx, cy, uvPaths, visited)
				nextCoor = junctions[Coordinate{cx, cy}][0]
				nextDir = directionToMove(Coordinate{cx, cy}, nextCoor)
			} else {
				priortisePaths(cx, cy, uvPaths, visited)
				if unvisited > 1 {
					// more than 1 path, remember this junction so we can come back
					junctions[Coordinate{cx, cy}] = uvPaths[1:]
				}

				nextCoor = uvPaths[0]
				nextDir = directionToMove(Coordinate{cx, cy}, nextCoor)
			}

			cx, cy = updatePosition(cx, cy, nextDir)
			cleanUpJunctions(cx, cy, junctions)
			steps <- nextDir
		}

		close(steps)
	}()

	return steps
}

// Tremaux receives the surround surveys on replies channel
// and recommends the steps on the output channel.
// This is not used for the Go-Challenge. Instead it is provided
// as a baseline to improve on maze solving algorithm
func Tremaux(replies <-chan MazeReply) <-chan int {
	steps := make(chan int)

	// relative x and y to starting position
	cx, cy := 0, 0

	// keep track of all junctions that have at least one unvisited (x, y)
	junctions := make(adjacencyMap)

	// visited tracks the rooms that have been visited
	visited := make(map[Coordinate]bool)

	// sequence of steps took. useful for backtracking
	sequence := []Coordinate{}
	go func() {
		for {
			visited[Coordinate{cx, cy}] = true
			sequence = append(sequence, Coordinate{cx, cy})
			reply := <-replies

			survey, err := reply.Survey, reply.Err
			if err == ErrVictory {
				// solved, we are done
				break
			}

			//possible directions it can go from current x, y
			dirs := make([]int, 0, 4)
			if !survey.Left {
				dirs = append(dirs, W)
			}
			if !survey.Right {
				dirs = append(dirs, E)
			}
			if !survey.Top {
				dirs = append(dirs, N)
			}
			if !survey.Bottom {
				dirs = append(dirs, S)
			}
			Shuffle(dirs)

			// get all possible paths/coordinates you can go
			paths := make([]Coordinate, 0, 4)
			for _, dir := range dirs {
				newX, newY := cx+Delta[dir].X, cy+Delta[dir].Y
				paths = append(paths, Coordinate{newX, newY})
			}

			// of all the possible paths, how many unvisited previously?
			uvPaths := make([]Coordinate, 0, 4)
			for _, path := range paths {
				if !visited[path] {
					uvPaths = append(uvPaths, path)
				}
			}

			unvisited := len(uvPaths)
			var nextCoor Coordinate
			var nextDir int

			if unvisited == 0 {
				// deadend, need to backtrack
				if len(sequence) < 2 {
					fmt.Println("Visited all places, but can't find treasured!")
					break
				}

				// backtrack a step
				nextCoor = sequence[len(sequence)-2]
				sequence = sequence[:len(sequence)-2]
			} else {
				if unvisited > 1 {
					// more than 1 path, remember this junction so we can come back
					junctions[Coordinate{cx, cy}] = uvPaths[1:]
				}

				nextCoor = uvPaths[0]
			}

			nextDir = directionToMove(Coordinate{cx, cy}, nextCoor)
			cx, cy = updatePosition(cx, cy, nextDir)
			cleanUpJunctions(cx, cy, junctions)
			steps <- nextDir
		}

		close(steps)
	}()

	return steps
}
