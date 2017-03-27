// Copyright © 2015 Steve Francia <spf@spf13.com>.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//

package commands

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"bitbucket.org/kelvinyong/gc6/mazelib"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//Maze struct
type Maze struct {
	rooms      [][]mazelib.Room
	start      mazelib.Coordinate
	end        mazelib.Coordinate
	icarus     mazelib.Coordinate
	StepsTaken int
}

// Tracking the current maze being solved

// WARNING: This approach is not safe for concurrent use
// This server is only intended to have a single client at a time
// We would need a different and more complex approach if we wanted
// concurrent connections than these simple package variables
var currentMaze *Maze
var scores []int

// Defining the daedalus command.
// This will be called as 'laybrinth daedalus'
var daedalusCmd = &cobra.Command{
	Use:     "daedalus",
	Aliases: []string{"deadalus", "server"},
	Short:   "Start the laybrinth creator",
	Long: `Daedalus's job is to create a challenging Labyrinth for his opponent
  Icarus to solve.

  Daedalus runs a server which Icarus clients can connect to to solve laybrinths.`,
	Run: func(cmd *cobra.Command, args []string) {
		RunServer()
	},
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano()) // need to initialize the seed
	gin.SetMode(gin.ReleaseMode)

	RootCmd.AddCommand(daedalusCmd)
}

// RunServer runs the web server
func RunServer() {
	// Adding handling so that even when ctrl+c is pressed we still print
	// out the results prior to exiting.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		printResults()
		os.Exit(1)
	}()

	// Using gin-gonic/gin to handle our routing
	r := gin.Default()
	v1 := r.Group("/")
	{
		v1.GET("/awake", GetStartingPoint)
		v1.GET("/move/:direction", MoveDirection)
		v1.GET("/done", End)
	}

	r.Run(":" + viper.GetString("port"))
}

// End ends a session and prints the results.
// Called by Icarus when he has reached
//   the number of times he wants to solve the laybrinth.
func End(c *gin.Context) {
	printResults()
	os.Exit(1)
}

// GetStartingPoint initializes a new maze and places Icarus in his awakening location
func GetStartingPoint(c *gin.Context) {
	initializeMaze()
	startRoom, err := currentMaze.Discover(currentMaze.Icarus())
	if err != nil {
		fmt.Println("Icarus is outside of the maze. This shouldn't ever happen")
		fmt.Println(err)
		os.Exit(-1)
	}
	mazelib.PrintMaze(currentMaze)

	c.JSON(http.StatusOK, mazelib.Reply{Survey: startRoom})
}

// MoveDirection is API response to the /move/:direction address
func MoveDirection(c *gin.Context) {
	var err error

	switch c.Param("direction") {
	case "left":
		err = currentMaze.MoveLeft()
	case "right":
		err = currentMaze.MoveRight()
	case "down":
		err = currentMaze.MoveDown()
	case "up":
		err = currentMaze.MoveUp()
	}

	var r mazelib.Reply

	if err != nil {
		r.Error = true
		r.Message = err.Error()
		c.JSON(409, r)
		return
	}

	s, e := currentMaze.LookAround()

	if e != nil {
		if e == mazelib.ErrVictory {
			scores = append(scores, currentMaze.StepsTaken)
			r.Victory = true
			r.Message = fmt.Sprintf("Victory achieved in %d steps \n", currentMaze.StepsTaken)
		} else {
			r.Error = true
			r.Message = err.Error()
		}
	}

	r.Survey = s

	c.JSON(http.StatusOK, r)
}

func initializeMaze() {
	currentMaze = createMaze()
}

// Print to the terminal the average steps to solution for the current session
func printResults() {
	fmt.Printf("Labyrinth solved %d times with an avg of %d steps\n", len(scores), mazelib.AvgScores(scores))
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
		// update statistics
		ms := mazeStats[(mCount-1)/100]
		ms.steps += m.StepsTaken
		ms.times++

		fmt.Printf("Victory achieved in %d steps \n", m.StepsTaken)
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

// Creates a maze without any walls
// Good starting point for additive algorithms
func emptyMaze() *Maze {
	z := Maze{}
	ySize := viper.GetInt("height")
	xSize := viper.GetInt("width")

	z.rooms = make([][]mazelib.Room, ySize)
	for y := 0; y < ySize; y++ {
		z.rooms[y] = make([]mazelib.Room, xSize)
		for x := 0; x < xSize; x++ {
			z.rooms[y][x] = mazelib.Room{}
		}
	}

	// Add perimeter walls for top and bottom
	for x := 0; x < xSize; x++ {
		z.rooms[0][x].AddWall(mazelib.N)
		z.rooms[ySize-1][x].AddWall(mazelib.S)
	}

	// Add perimeter walls for left and right
	for y := 0; y < ySize; y++ {
		z.rooms[y][0].AddWall(mazelib.W)
		z.rooms[y][xSize-1].AddWall(mazelib.E)
	}

	return &z
}

// Creates a maze with all walls
// Good starting point for subtractive algorithms
func fullMaze() *Maze {
	z := emptyMaze()
	ySize := viper.GetInt("height")
	xSize := viper.GetInt("width")

	for y := 0; y < ySize; y++ {
		for x := 0; x < xSize; x++ {
			z.rooms[y][x].Walls = mazelib.Survey{true, true, true, true}
		}
	}

	return z
}

// MAZE CREATION CODES STARTS HERE

// creates a maze based on Kruskal's algorithm
// http://weblog.jamisbuck.org/2011/1/3/maze-generation-kruskal-s-algorithm
func createMazeKruskal() *Maze {
	type edge struct {
		x, y, dir int
	}

	m := fullMaze()
	xSize := m.Width()
	ySize := m.Height()

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

	return m
}

// create a maze full of vertical pockets (tunnels) which
// are either facing up or down
func createMazePocket() *Maze {
	m := emptyMaze()
	xSize := m.Width()
	ySize := m.Height()

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

	return m
}

// some variables to keep track of statistics
var mCount int
var nowKruskal bool

type mStat struct {
	steps   int
	times   int
	kruskal bool
}

var mazeStats = make([]*mStat, 0)

// getMaze changes the maze type for every 100 mazes
// for the first 100 mazes, use Kruskal
// for the next 100 mazes, use Pocket
// subsequent mazes depends on past performance of solver
func getMaze() *Maze {
	if mCount%100 == 0 {
		switch mCount / 100 {
		case 0:
			nowKruskal = true
		case 1:
			nowKruskal = false
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
		}
		mazeStats = append(mazeStats, &mStat{kruskal: nowKruskal})
	}

	mCount++
	if nowKruskal {
		return createMazeKruskal()
	}
	return createMazePocket()
}

func createMaze() *Maze {
	m := getMaze()
	ySize := m.Height()
	xSize := m.Width()

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
