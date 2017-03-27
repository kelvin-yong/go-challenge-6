package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"bitbucket.org/kelvinyong/gc6/mazelib"
)

//Maze struct
type Maze struct {
	rooms      [][]mazelib.Room
	start      mazelib.Coordinate
	end        mazelib.Coordinate
	icarus     mazelib.Coordinate
	StepsTaken int
}

// GetRoom returns a Room struct
func (m *Maze) GetRoom(x, y int) (*mazelib.Room, error) {
	if x < 0 || y < 0 || x >= m.Width() || y >= m.Height() {
		return &mazelib.Room{}, errors.New("room outside of maze boundaries")
	}

	return &m.rooms[y][x], nil
}

// Width returns width of the maze
func (m *Maze) Width() int { return len(m.rooms[0]) }

// Height returns height of the maze
func (m *Maze) Height() int { return len(m.rooms) }

// Icarus returns the finder's current position
func (m *Maze) Icarus() (x, y int) {
	return m.icarus.X, m.icarus.Y
}

// SetStartPoint sets the location where Icarus will awake
func (m *Maze) SetStartPoint(x, y int) error {
	r, err := m.GetRoom(x, y)

	if err != nil {
		return err
	}

	if r.Treasure {
		return errors.New("can't start in the treasure")
	}

	r.Start = true
	m.icarus = mazelib.Coordinate{x, y}
	return nil
}

// SetTreasure sets the location of the treasure for a given maze
func (m *Maze) SetTreasure(x, y int) error {
	r, err := m.GetRoom(x, y)

	if err != nil {
		return err
	}

	if r.Start {
		return errors.New("can't have the treasure at the start")
	}

	r.Treasure = true
	m.end = mazelib.Coordinate{x, y}
	return nil
}

// LookAround Given Icarus's current location, Discover that room
// Will return ErrVictory if Icarus is at the treasure.
func (m *Maze) LookAround() (mazelib.Survey, error) {
	if m.end.X == m.icarus.X && m.end.Y == m.icarus.Y {
		//fmt.Printf("Victory achieved in %d steps \n", m.StepsTaken)
		if mCount > 0 {
			ms := mazeStats[(mCount-1)/100]
			ms.steps += m.StepsTaken
			ms.times++
		}
		return mazelib.Survey{}, mazelib.ErrVictory
	}

	return m.Discover(m.icarus.X, m.icarus.Y)
}

// Discover Given two points, survey the room.
// Will return error if two points are outside of the maze
func (m *Maze) Discover(x, y int) (mazelib.Survey, error) {
	r, err := m.GetRoom(x, y)
	if err != nil {
		return mazelib.Survey{}, err
	}
	return r.Walls, nil
}

// MoveLeft Moves Icarus's position left one step
// Will not permit moving through walls or out of the maze
func (m *Maze) MoveLeft() error {
	s, e := m.LookAround()
	if e != nil {
		return e
	}
	if s.Left {
		return errors.New("Can't walk through walls")
	}

	x, y := m.Icarus()
	if _, err := m.GetRoom(x-1, y); err != nil {
		return err
	}

	m.icarus = mazelib.Coordinate{x - 1, y}
	m.StepsTaken++
	return nil
}

// MoveRight Moves Icarus's position right one step
// Will not permit moving through walls or out of the maze
func (m *Maze) MoveRight() error {
	s, e := m.LookAround()
	if e != nil {
		return e
	}
	if s.Right {
		return errors.New("Can't walk through walls")
	}

	x, y := m.Icarus()
	if _, err := m.GetRoom(x+1, y); err != nil {
		return err
	}

	m.icarus = mazelib.Coordinate{x + 1, y}
	m.StepsTaken++
	return nil
}

// MoveUp Moves Icarus's position up one step
// Will not permit moving through walls or out of the maze
func (m *Maze) MoveUp() error {
	s, e := m.LookAround()
	if e != nil {
		return e
	}
	if s.Top {
		return errors.New("Can't walk through walls")
	}

	x, y := m.Icarus()
	if _, err := m.GetRoom(x, y-1); err != nil {
		return err
	}

	m.icarus = mazelib.Coordinate{x, y - 1}
	m.StepsTaken++
	return nil
}

// MoveDown Moves Icarus's position down one step
// Will not permit moving through walls or out of the maze
func (m *Maze) MoveDown() error {
	s, e := m.LookAround()
	if e != nil {
		return e
	}
	if s.Bottom {
		return errors.New("Can't walk through walls")
	}

	x, y := m.Icarus()
	if _, err := m.GetRoom(x, y+1); err != nil {
		return err
	}

	m.icarus = mazelib.Coordinate{x, y + 1}
	m.StepsTaken++
	return nil
}

// Utility method to add wall to room
func (m *Maze) addWallAt(cx, cy, direction int) {
	w, h := m.Width(), m.Height()

	if cx >= 0 && cx < w && cy >= 0 && cy < h {
		m.rooms[cy][cx].AddWall(direction)
	}

	nx, ny := cx+mazelib.Delta[direction].X, cy+mazelib.Delta[direction].Y
	if nx >= 0 && nx < w && ny >= 0 && ny < h {
		m.rooms[ny][nx].AddWall(mazelib.Opposite[direction])
	}
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano()) // need to initialize the seed
}

func emptyMaze() *Maze {
	z := Maze{}
	ySize := 30
	xSize := 30

	z.rooms = make([][]mazelib.Room, ySize)
	for y := 0; y < ySize; y++ {
		z.rooms[y] = make([]mazelib.Room, xSize)
		for x := 0; x < xSize; x++ {
			z.rooms[y][x] = mazelib.Room{}
		}
	}

	// Add walls for top and bottom boundaries
	for x := 0; x < xSize; x++ {
		z.rooms[0][x].AddWall(mazelib.N)
		z.rooms[ySize-1][x].AddWall(mazelib.S)
	}

	// Add walls for left and right boundaries
	for y := 0; y < ySize; y++ {
		z.rooms[y][0].AddWall(mazelib.W)
		z.rooms[y][xSize-1].AddWall(mazelib.E)
	}

	return &z
}

func fullMaze() *Maze {
	z := emptyMaze()
	ySize := 30
	xSize := 30

	for y := 0; y < ySize; y++ {
		for x := 0; x < xSize; x++ {
			z.rooms[y][x].Walls = mazelib.Survey{true, true, true, true}
		}
	}

	return z
}

//////////////// A1. Recursive BackTracker Algo ////////////////
func carvePassages(m *Maze, cx int, cy int) {
	directions := []int{mazelib.N, mazelib.S, mazelib.E, mazelib.W}
	mazelib.Shuffle(directions)

	for _, dir := range directions {
		nx, ny := cx+mazelib.Delta[dir].X, cy+mazelib.Delta[dir].Y
		if nextRoom, err := m.GetRoom(nx, ny); err == nil && !nextRoom.Visited {
			curRoom, _ := m.GetRoom(cx, cy)
			curRoom.Visited = true
			curRoom.RmWall(dir)
			nextRoom.Visited = true
			nextRoom.RmWall(mazelib.Opposite[dir])
			carvePassages(m, nx, ny)
		}
	}
}

var mazeName string

func createMaze(mazeType int) *Maze {
	var m *Maze
	var xSize, ySize int

	oldName := mazeName

	switch mazeType {
	case 1:
		mazeName = "Empty maze"
		m = emptyMaze()
		xSize = m.Width()
		ySize = m.Height()

	case 2:
		mazeName = "Linear maze"
		m = emptyMaze()
		xSize = m.Width()
		ySize = m.Height()

		for y := 0; y < ySize-1; y++ {
			for x := 0; x < xSize; x++ {
				if (y%2 == 0 && x != xSize-1) || (y%2 == 1 && x != 0) {
					m.rooms[y][x].AddWall(mazelib.S)
					m.rooms[y+1][x].AddWall(mazelib.N)
				}
			}
		}

	case 3:
		mazeName = "Perfect recursive maze"
		m = fullMaze()
		xSize = m.Width()
		ySize = m.Height()
		carvePassages(m, rand.Intn(xSize), rand.Intn(ySize))

	case 4:
		mazeName = "Horizontal pocket maze"
		m = emptyMaze()
		xSize = m.Width()
		ySize = m.Height()

		for y := 0; y < ySize-1; y++ {
			for x := 1; x < xSize-1; x++ {
				m.rooms[y][x].AddWall(mazelib.S)
				m.rooms[y+1][x].AddWall(mazelib.N)
			}
			m.rooms[y][0].AddWall(mazelib.E)
			m.rooms[y][1].AddWall(mazelib.W)
		}

	case 5:
		mazeName = "Vertical pocket maze"
		m = emptyMaze()
		xSize = m.Width()
		ySize = m.Height()

		for y := 1; y < ySize-1; y++ {
			for x := 0; x < xSize-1; x++ {
				m.rooms[y][x].AddWall(mazelib.E)
				m.rooms[y][x+1].AddWall(mazelib.W)
			}
		}

		y := 0
		if rand.Intn(2) == 0 {
			y = ySize - 1
		}
		for x := 0; x < xSize-1; x++ {
			m.rooms[y][x].AddWall(mazelib.E)
			m.rooms[y][x+1].AddWall(mazelib.W)
		}

	case 6:
		type edge struct {
			x, y, dir int
		}

		mazeName = "Kruskal Algorithm Maze"
		m = fullMaze()
		xSize = m.Width()
		ySize = m.Height()

		// create edges
		edges := make([]edge, 0, 2*xSize*ySize-ySize-xSize)
		for x := 0; x < xSize; x++ {
			for y := 0; y < ySize; y++ {
				if y > 0 {
					edges = append(edges, edge{x, y, mazelib.N})
				}
				if x > 0 {
					edges = append(edges, edge{x, y, mazelib.W})
				}
			}
		}

		// shuffle the edges
		for i := range edges {
			j := rand.Intn(i + 1)
			edges[i], edges[j] = edges[j], edges[i]
		}

		sets := make(map[int]map[mazelib.Coordinate]int, xSize*ySize/2)

		for i, edge := range edges {
			x, y, dir := edge.x, edge.y, edge.dir
			thisCoor := mazelib.Coordinate{x, y}
			nextCoor := mazelib.Coordinate{x + mazelib.Delta[dir].X, y + mazelib.Delta[dir].Y}

			thisSetID, nextSetID := -1, -1
			for id, m := range sets {
				if _, ok := m[thisCoor]; ok {
					thisSetID = id
				}
				if _, ok := m[nextCoor]; ok {
					nextSetID = id
				}
				if thisSetID >= 0 && nextSetID >= 0 {
					// found both id
					break
				}
			}

			if thisSetID == nextSetID && thisSetID != -1 {
				// the 2 Coordinate are in the same set, do nothing
				continue
			}

			if thisSetID == nextSetID && thisSetID == -1 {
				// they are not connected anyway, form a new set
				newSet := make(map[mazelib.Coordinate]int)
				newSet[thisCoor], newSet[nextCoor] = 0, 0
				sets[i] = newSet
			} else if thisSetID == -1 {
				// thisCoor will be absorbed into existing set
				sets[nextSetID][thisCoor] = 0
			} else if nextSetID == -1 {
				// nextCoor will be absorbed into existing set
				sets[thisSetID][nextCoor] = 0
			} else {
				// absorb one existing set till another existing set
				thisSet := sets[thisSetID]
				nextSet := sets[nextSetID]

				for k, v := range nextSet {
					thisSet[k] = v
				}
				delete(sets, nextSetID)
			}
			// remove the walls linking to them
			m.rooms[thisCoor.Y][thisCoor.X].RmWall(dir)
			m.rooms[nextCoor.Y][nextCoor.X].RmWall(mazelib.Opposite[dir])
		}

	default:
		panic("Invalid maze type")
	}

	if oldName != mazeName {
		fmt.Println(mazeName)
	}

	// set a startingPoint for Icarus
	sx, sy := rand.Intn(xSize), rand.Intn(ySize)
	m.SetStartPoint(sx, sy)

	// set endingPoint (treasure) for Icarus
	for {
		tx, ty := rand.Intn(xSize), rand.Intn(ySize)
		if err := m.SetTreasure(tx, ty); err == nil {
			break
		}
	}

	return m
}

var trace = false

func trialMaze(mazeType, tries int) int {
	if tries < 0 {
		return 0
	}

	totalSteps := 0

	for i := 0; i < tries; i++ {
		m := createMaze(mazeType)
		s, _ := m.LookAround()
		if trace {
			mazelib.PrintMaze(m)
		}
		replies := make(chan mazelib.MazeReply)
		steps := mazelib.FindTreasure(replies)
		replies <- mazelib.MazeReply{s, nil}

		currentCount := 0
		for step := range steps {
			currentCount++
			switch step {
			case mazelib.N:
				m.MoveUp()
			case mazelib.S:
				m.MoveDown()
			case mazelib.E:
				m.MoveRight()
			case mazelib.W:
				m.MoveLeft()
			}
			s, e := m.LookAround()

			if trace {
				mazelib.PrintMaze(m)
				var input string
				fmt.Scanln(&input)
				if input != "" {
					trace = false
				}
			}
			replies <- mazelib.MazeReply{s, e}
		}
		if trace {
			fmt.Print("#################### Solved ####################\n\n")
		}
		totalSteps += currentCount
	}

	return totalSteps
}

var mCount int
var nowKruskal bool

type mStat struct {
	steps   int
	times   int
	kruskal bool
}

var mazeStats = make([]*mStat, 0)

// changes the maze type for every 100 mazes
// base on statistics
func makeMaze() *Maze {
	if mCount%100 == 0 {

		switch mCount / 100 {
		// for the first hundred maze, use kruskal
		case 0:
			nowKruskal = true
		// for the second hundred maze, use pocket
		case 1:
			nowKruskal = false
		// otherwise, decide base on past stats
		default:
			var ksteps, ktimes, psteps, ptimes int
			for _, ms := range mazeStats {
				if ms.kruskal {
					ksteps += ms.steps
					ktimes += ms.times
				} else {
					psteps += ms.steps
					ptimes += ms.times
				}
			}
			if ktimes == 0 {
				nowKruskal = true // solver cannot solve kruskal at all, give kruskal
			} else if ptimes == 0 {
				nowKruskal = true // solver cannot solve pocket, give pocket
			} else {
				nowKruskal = (ksteps / ktimes) > (psteps / ptimes)
			}
			fmt.Printf("Kruskal: %d steps %d times. Average %d\n", ksteps, ktimes, ksteps/ktimes)
			fmt.Printf("Pocket: %d steps %d times. Average %d\n", psteps, ptimes, psteps/ptimes)
		}
		mazeStats = append(mazeStats, &mStat{kruskal: nowKruskal})
		if nowKruskal {
			fmt.Printf("\nUsing Kruskal for %d to %d\n", mCount, mCount+99)
		} else {
			fmt.Printf("\nUsing Pocket for %d to %d\n", mCount, mCount+99)
		}
	}

	mCount++
	if nowKruskal {
		return createMaze(6)
	}
	return createMaze(5)
}

func main() {
	trace = false
	if trace {
		fmt.Print("**Tracing has been turned on. Enter any input and press return to turn off trace.**\n\n\n")
	}

	// generating statistics for different maze types
	/*
		fmt.Println("##Statistics For Different Maze##")
		const tries = 1000
		for mt := 1; mt <= 6; mt++ {
			fmt.Println(tries, "times. Average:", trialMaze(mt, tries)/tries)
		}
	*/
	// Pitting generator against solver
	fmt.Println("\n\n##Results between  generator and solver##")
	for i := 0; i < 500; i++ {
		m := makeMaze()
		s, _ := m.LookAround()
		if trace {
			mazelib.PrintMaze(m)
		}
		replies := make(chan mazelib.MazeReply)
		steps := mazelib.FindTreasure(replies)
		replies <- mazelib.MazeReply{s, nil}

		for step := range steps {
			switch step {
			case mazelib.N:
				m.MoveUp()
			case mazelib.S:
				m.MoveDown()
			case mazelib.E:
				m.MoveRight()
			case mazelib.W:
				m.MoveLeft()
			}
			s, e := m.LookAround()

			if trace {
				mazelib.PrintMaze(m)
				var input string
				fmt.Scanln(&input)
				if input != "" {
					trace = false
				}
			}
			replies <- mazelib.MazeReply{s, e}
		}
		if trace {
			fmt.Print("#################### Solved ####################\n\n")
		}
	}

	// print out the stats
	var steps, times int
	for _, ms := range mazeStats {
		steps += ms.steps
		times += ms.times
	}
	fmt.Println("\nTotal steps:", steps, "\nTimes:", times, "\nAverage:", steps/times)

}
