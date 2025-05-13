package hashisolver

import (
	"bufio"
	"fmt"
	"io"
	"sort"
)

const (
	MaxRow     = 50
	MaxCol     = 50
	MaxIslands = 1600
	MaxBridges = 64000

	// Direction constants
	Up    = 0
	Right = 1
	Down  = 2
	Left  = 3

	DirectionsCount = 4
)

// Island represents a single island in the puzzle
type Island struct {
	X, Y           int
	MaxBridges     int
	CurrentBridges int
	Neighbors      [DirectionsCount]*Island
	NeighborCount  int
}

// Bridge represents a connection between two islands
type Bridge struct {
	Island1   *Island
	Island2   *Island
	Direction int
	Symbol    rune
	Wires     int
	Skip      bool
}

// Puzzle represents the entire hashi puzzle
type Puzzle struct {
	Nodes        [MaxIslands]*Island
	Edges        [MaxBridges]*Bridge
	Rows         int
	Cols         int
	IslandCount  int
	BridgeCount  int
	FullBridges  int // total bridges needed
	BuiltBridges int // bridges built so far
	Solved       [MaxIslands]int
	Attempts     int
}

// IsIsland checks if a character represents an island
func IsIsland(ch rune) bool {
	return (ch >= '1' && ch <= '9') || (ch >= 'a' && ch <= 'c')
}

// IslandToNum converts an island character to the number of bridges
func IslandToNum(ch rune) int {
	if ch >= 'a' && ch <= 'c' {
		return 10 + int(ch) - int('a')
	}
	return int(ch) - int('0')
}

// ScanMap reads the puzzle input from a reader
func ScanMap(reader io.Reader, puzzle *Puzzle) ([][]rune, error) {
	scanner := bufio.NewScanner(reader)
	var rows [][]rune

	for scanner.Scan() {
		line := scanner.Text()
		rows = append(rows, []rune(line))
		if len(rows[len(rows)-1]) > puzzle.Cols {
			puzzle.Cols = len(rows[len(rows)-1])
		}
	}

	puzzle.Rows = len(rows)
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Ensure all rows have the same length
	for i := range rows {
		if len(rows[i]) < puzzle.Cols {
			rows[i] = append(rows[i], make([]rune, puzzle.Cols-len(rows[i]))...)
			for j := len(rows[i]) - (puzzle.Cols - len(rows[i])); j < puzzle.Cols; j++ {
				rows[i][j] = ' '
			}
		}
	}

	return rows, nil
}

// ParseMap processes the map to identify islands and their neighbors
func ParseMap(mapData [][]rune, puzzle *Puzzle) {
	puzzle.Attempts = 0
	puzzle.BridgeCount = 0
	puzzle.IslandCount = 0
	puzzle.FullBridges = 0
	puzzle.BuiltBridges = 0

	// Identify islands
	for i := 0; i < puzzle.Rows; i++ {
		for j := 0; j < puzzle.Cols; j++ {
			if IsIsland(mapData[i][j]) {
				island := &Island{
					X:          j,
					Y:          i,
					MaxBridges: IslandToNum(mapData[i][j]),
				}
				puzzle.FullBridges += island.MaxBridges
				puzzle.Nodes[puzzle.IslandCount] = island
				puzzle.IslandCount++
			}
		}
	}

	// Find neighbors for each island
	for j := 0; j < puzzle.IslandCount; j++ {
		island := puzzle.Nodes[j]
		puzzle.Solved[j] = island.MaxBridges - island.CurrentBridges

		// Check upwards
		for d := 1; d <= island.Y; d++ {
			if island.Y == 0 {
				break
			}
			if IsIsland(mapData[island.Y-d][island.X]) {
				idx := getIsland(puzzle, island.X, island.Y-d)
				if idx != -1 {
					island.Neighbors[Up] = puzzle.Nodes[idx]
					island.NeighborCount++
				}
				break
			}
		}

		// Check downwards
		for d := 1; island.Y+d < puzzle.Rows; d++ {
			if island.Y >= puzzle.Rows-1 {
				break
			}
			if IsIsland(mapData[island.Y+d][island.X]) {
				idx := getIsland(puzzle, island.X, island.Y+d)
				if idx != -1 {
					island.Neighbors[Down] = puzzle.Nodes[idx]
					island.NeighborCount++
				}
				break
			}
		}

		// Check left
		for d := 1; d <= island.X; d++ {
			if island.X == 0 {
				break
			}
			if IsIsland(mapData[island.Y][island.X-d]) {
				idx := getIsland(puzzle, island.X-d, island.Y)
				if idx != -1 {
					island.Neighbors[Left] = puzzle.Nodes[idx]
					island.NeighborCount++
				}
				break
			}
		}

		// Check right
		for d := 1; island.X+d < puzzle.Cols; d++ {
			if island.X >= puzzle.Cols-1 {
				break
			}
			if IsIsland(mapData[island.Y][island.X+d]) {
				idx := getIsland(puzzle, island.X+d, island.Y)
				if idx != -1 {
					island.Neighbors[Right] = puzzle.Nodes[idx]
					island.NeighborCount++
				}
				break
			}
		}
	}

	// Divide by 2 because each bridge is counted twice (once for each island)
	puzzle.FullBridges /= 2
}

// Helper function to find island by coordinates
func getIsland(puzzle *Puzzle, x, y int) int {
	for i := 0; i < puzzle.IslandCount; i++ {
		if puzzle.Nodes[i].X == x && puzzle.Nodes[i].Y == y {
			return i
		}
	}
	return -1
}

// Helper function to find bridge index by island
func islandIndex(puzzle *Puzzle, is *Island) int {
	for i := 0; i < puzzle.IslandCount; i++ {
		if puzzle.Nodes[i] == is {
			return i
		}
	}
	return -1
}

// ConstructBridge creates a new bridge or returns an existing one
func constructBridge(puzzle *Puzzle, i1, i2 *Island, dir int) *Bridge {
	// Check if a bridge already exists
	for i := 0; i < puzzle.BridgeCount; i++ {
		b := puzzle.Edges[i]
		if (b.Island1 == i1 && b.Island2 == i2) || (b.Island1 == i2 && b.Island2 == i1) {
			return b /* reuse even with 0 wires */
		}
	}

	// Create a new bridge
	bridge := &Bridge{
		Island1:   i1,
		Island2:   i2,
		Direction: dir,
		Wires:     0,
		Skip:      false,
	}
	return bridge
}

// RemoveBridge removes a bridge between islands
func RemoveBridge(puzzle *Puzzle, curr *Island, dir int) {
	i1 := curr
	i2 := curr.Neighbors[dir]

	puzzle.BuiltBridges-- // islands first
	i1.CurrentBridges--
	i2.CurrentBridges--

	idx1 := islandIndex(puzzle, i1)
	idx2 := islandIndex(puzzle, i2)
	puzzle.Solved[idx1] = i1.MaxBridges - i1.CurrentBridges
	puzzle.Solved[idx2] = i2.MaxBridges - i2.CurrentBridges

	// Find the bridge
	var b *Bridge
	for i := 0; i < puzzle.BridgeCount; i++ {
		b = puzzle.Edges[i]
		if b.Skip {
			continue
		}
		if (b.Island1 == i1 && b.Island2 == i2) || (b.Island1 == i2 && b.Island2 == i1) {
			break
		}
	}

	b.Wires--

	if b.Wires == 0 {
		b.Symbol = ' '
		b.Skip = true // Mark the bridge for skipping since it has 0 wires
		return        // We can safely return after marking the bridge
	}

	switch b.Wires {
	case 1:
		if b.Direction == Left || b.Direction == Right {
			b.Symbol = '-'
		} else {
			b.Symbol = '|'
		}
	case 2:
		if b.Direction == Left || b.Direction == Right {
			b.Symbol = '='
		} else {
			b.Symbol = '"'
		}
	}
}

// CanBuildBridge checks if a bridge can be built in the given direction
func CanBuildBridge(puzzle *Puzzle, curr *Island, dir int) bool {
	i1 := curr
	i2 := curr.Neighbors[dir]
	overlap := false

	if i2 == nil {
		return false
	}

	// Forward checking - if islands are full
	if i1.MaxBridges == i1.CurrentBridges || i2.MaxBridges == i2.CurrentBridges {
		return false
	}

	// Check for existing bridges between these islands
	var b *Bridge
	for i := 0; i < puzzle.BridgeCount; i++ {
		b = puzzle.Edges[i]
		if b.Skip {
			continue
		}
		if (b.Island1 == i1 && b.Island2 == i2) || (b.Island1 == i2 && b.Island2 == i1) {
			if b.Wires == 3 { // Max bridges already built
				return false
			}
			overlap = true
			break
		}
	}

	// If it's an existing bridge, we can add more wires without checking for crossings
	if overlap {
		return true
	}

	// Only check for crossings if this is a new bridge
	var bridgePath []struct{ x, y int }

	// Create a path of coordinates that the bridge would occupy
	if i1.X == i2.X { // Vertical bridge
		minY, maxY := i1.Y, i2.Y
		if i1.Y > i2.Y {
			minY, maxY = i2.Y, i1.Y
		}
		for y := minY + 1; y < maxY; y++ {
			bridgePath = append(bridgePath, struct{ x, y int }{i1.X, y})
		}
	} else { // Horizontal bridge
		minX, maxX := i1.X, i2.X
		if i1.X > i2.X {
			minX, maxX = i2.X, i1.X
		}
		for x := minX + 1; x < maxX; x++ {
			bridgePath = append(bridgePath, struct{ x, y int }{x, i1.Y})
		}
	}

	// Check if any existing bridges cross this path
	for i := 0; i < puzzle.BridgeCount; i++ {
		b = puzzle.Edges[i]
		if b.Skip || b.Wires == 0 {
			continue
		}

		j1, j2 := b.Island1, b.Island2

		// Skip bridges between the same islands
		if (j1 == i1 && j2 == i2) || (j1 == i2 && j2 == i1) {
			continue
		}

		// Only perpendicular bridges can cross
		if (i1.X == i2.X && j1.X == j2.X) || (i1.Y == i2.Y && j1.Y == j2.Y) {
			continue
		}

		// Check if this bridge crosses our path
		if i1.X == i2.X { // Our bridge is vertical
			if j1.Y == j2.Y { // Their bridge is horizontal
				minX, maxX := j1.X, j2.X
				if j1.X > j2.X {
					minX, maxX = j2.X, j1.X
				}

				// Check if their horizontal bridge crosses our vertical path
				for _, point := range bridgePath {
					if point.y == j1.Y && point.x >= minX && point.x <= maxX {
						return false
					}
				}
			}
		} else { // Our bridge is horizontal
			if j1.X == j2.X { // Their bridge is vertical
				minY, maxY := j1.Y, j2.Y
				if j1.Y > j2.Y {
					minY, maxY = j2.Y, j1.Y
				}

				// Check if their vertical bridge crosses our horizontal path
				for _, point := range bridgePath {
					if point.x == j1.X && point.y >= minY && point.y <= maxY {
						return false
					}
				}
			}
		}
	}

	return true
}

// AddBridge adds a bridge between islands
func AddBridge(puzzle *Puzzle, curr *Island, dir int) {
	i1 := curr
	i2 := curr.Neighbors[dir]
	b := constructBridge(puzzle, i1, i2, dir)

	if b.Wires == 3 {
		return // Should never happen due to CanBuildBridge check
	}

	b.Wires++

	if b.Skip {
		b.Skip = false
	}

	puzzle.BuiltBridges++
	i1.CurrentBridges++
	i2.CurrentBridges++

	idx1 := islandIndex(puzzle, i1)
	idx2 := islandIndex(puzzle, i2)
	puzzle.Solved[idx1] = i1.MaxBridges - i1.CurrentBridges
	puzzle.Solved[idx2] = i2.MaxBridges - i2.CurrentBridges

	switch b.Wires {
	case 1:
		if b.Direction == Left || b.Direction == Right {
			b.Symbol = '-'
		} else {
			b.Symbol = '|'
		}
		puzzle.Edges[puzzle.BridgeCount] = b
		puzzle.BridgeCount++
	case 2:
		if b.Direction == Left || b.Direction == Right {
			b.Symbol = '='
		} else {
			b.Symbol = '"'
		}
	case 3:
		if b.Direction == Left || b.Direction == Right {
			b.Symbol = 'E'
		} else {
			b.Symbol = '#'
		}
	}
}

// ApplyHeuristics applies initial heuristics to simplify the puzzle
func ApplyHeuristics(puzzle *Puzzle) {
	// Apply simple island heuristics
	for i := 0; i < puzzle.IslandCount; i++ {
		curr := puzzle.Nodes[i]

		// Islands with only one neighbor must connect all bridges to that neighbor
		if curr.NeighborCount == 1 {
			var dir int
			for dir = 0; dir < DirectionsCount && curr.Neighbors[dir] == nil; dir++ {
			}
			for b := 0; b < curr.MaxBridges && CanBuildBridge(puzzle, curr, dir); b++ {
				AddBridge(puzzle, curr, dir)
			}
		}
	}

	// Update solved status
	for i := 0; i < puzzle.IslandCount; i++ {
		curr := puzzle.Nodes[i]
		puzzle.Solved[i] = curr.MaxBridges - curr.CurrentBridges
	}
}

// SolveMap attempts to solve the puzzle using backtracking
func SolveMap(puzzle *Puzzle, debug bool) bool {
	// Early termination condition for excessive attempts
	if puzzle.Attempts >= 10000 {
		return false
	}

	if debug {
		PrintMap(puzzle)
		fmt.Printf("\nI want %d bridges\nI have %d\n", puzzle.FullBridges, puzzle.BuiltBridges)
		fmt.Println()
	}

	puzzle.Attempts++

	// Success condition: all bridges built
	if puzzle.BuiltBridges == puzzle.FullBridges {
		return CheckSolved(puzzle)
	}

	// Sort the islands to process first the ones with fewer options
	type IndexedIsland struct {
		Island *Island
		Index  int
	}
	var islands []IndexedIsland
	for i := 0; i < puzzle.IslandCount; i++ {
		// Only consider islands that still need bridges
		if puzzle.Solved[i] > 0 {
			islands = append(islands, IndexedIsland{
				Island: puzzle.Nodes[i],
				Index:  i,
			})
		}
	}

	// Return false if no islands left to process but not all bridges built
	if len(islands) == 0 && puzzle.BuiltBridges < puzzle.FullBridges {
		return false
	}

	sort.Slice(islands, func(i, j int) bool {
		// Sort by remaining bridges
		a := islands[i].Island
		b := islands[j].Island

		// First prioritize islands with fewer total options
		aOptions := 0
		bOptions := 0
		for dir := 0; dir < DirectionsCount; dir++ {
			if a.Neighbors[dir] != nil && CanBuildBridge(puzzle, a, dir) {
				aOptions++
			}
			if b.Neighbors[dir] != nil && CanBuildBridge(puzzle, b, dir) {
				bOptions++
			}
		}

		// If options differ, prioritize islands with fewer options
		if aOptions != bOptions {
			return aOptions < bOptions
		}

		// Then by remaining bridges
		aRemaining := a.MaxBridges - a.CurrentBridges
		bRemaining := b.MaxBridges - b.CurrentBridges
		if aRemaining != bRemaining {
			return aRemaining < bRemaining
		}

		// Then by neighbor count
		return a.NeighborCount < b.NeighborCount
	})

	// Try to build bridges for each island
	for _, ii := range islands {
		curr := ii.Island
		idx := ii.Index

		for dir := 0; dir < DirectionsCount; dir++ {
			if CanBuildBridge(puzzle, curr, dir) {
				AddBridge(puzzle, curr, dir)
				if SolveMap(puzzle, debug) {
					return true
				}
				RemoveBridge(puzzle, curr, dir)
			}
		}

		// If we have an island with bridges to build but we couldn't build any, this branch failed
		if puzzle.Solved[idx] > 0 {
			break
		}
	}

	return false
}

// CleanPuzzle resets the state of the puzzle
func CleanPuzzle(puzzle *Puzzle) {
	puzzle.BuiltBridges = 0

	// Reset all islands
	for i := 0; i < puzzle.IslandCount; i++ {
		puzzle.Nodes[i].CurrentBridges = 0
		puzzle.Solved[i] = puzzle.Nodes[i].MaxBridges
	}

	// Reset all bridges
	for i := 0; i < puzzle.BridgeCount; i++ {
		puzzle.Edges[i].Wires = 0
		puzzle.Edges[i].Symbol = ' '
		puzzle.Edges[i].Skip = false
	}
}

// CheckSolved checks if the puzzle is completely solved
func CheckSolved(puzzle *Puzzle) bool {
	for i := 0; i < puzzle.IslandCount; i++ {
		if puzzle.Solved[i] != 0 {
			return false
		}
	}
	return true
}

// PrintMap outputs the current state of the puzzle
func PrintMap(puzzle *Puzzle) {
	soln := make([][]rune, puzzle.Rows)
	for i := range soln {
		soln[i] = make([]rune, puzzle.Cols)
		for j := range soln[i] {
			soln[i][j] = ' '
		}
	}

	// Add islands
	for i := 0; i < puzzle.IslandCount; i++ {
		max := puzzle.Nodes[i].MaxBridges
		if max == 10 {
			soln[puzzle.Nodes[i].Y][puzzle.Nodes[i].X] = 'a'
		} else if max == 11 {
			soln[puzzle.Nodes[i].Y][puzzle.Nodes[i].X] = 'b'
		} else if max == 12 {
			soln[puzzle.Nodes[i].Y][puzzle.Nodes[i].X] = 'c'
		} else {
			soln[puzzle.Nodes[i].Y][puzzle.Nodes[i].X] = rune('0' + max)
		}
	}

	// Add bridges
	for i := 0; i < puzzle.BridgeCount; i++ {
		b := puzzle.Edges[i]
		if b.Skip {
			continue
		}

		dist := 0
		if b.Island1.X == b.Island2.X {
			dist = abs(b.Island1.Y - b.Island2.Y)
		} else {
			dist = abs(b.Island1.X - b.Island2.X)
		}
		dist -= 1 // Adjust for islands themselves

		for j := 0; j < dist; j++ {
			switch b.Direction {
			case Up:
				soln[b.Island1.Y-j-1][b.Island1.X] = b.Symbol
			case Down:
				soln[b.Island1.Y+j+1][b.Island1.X] = b.Symbol
			case Left:
				soln[b.Island1.Y][b.Island1.X-j-1] = b.Symbol
			case Right:
				soln[b.Island1.Y][b.Island1.X+j+1] = b.Symbol
			}
		}
	}

	// Print the solution
	for i := 0; i < puzzle.Rows; i++ {
		for j := 0; j < puzzle.Cols; j++ {
			fmt.Printf("%c", soln[i][j])
		}
		fmt.Println()
	}
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// NewPuzzle creates a new puzzle instance
func NewPuzzle() *Puzzle {
	return &Puzzle{}
}

// Solve provides the main functionality for solving a hashi puzzle
func Solve(reader io.Reader, debug bool) (*Puzzle, error) {
	puzzle := NewPuzzle()

	// Read the map
	mapData, err := ScanMap(reader, puzzle)
	if err != nil {
		return nil, err
	}

	// Parse the map
	ParseMap(mapData, puzzle)

	// Apply initial heuristics to simplify the puzzle
	ApplyHeuristics(puzzle)

	// Solve the puzzle
	success := SolveMap(puzzle, debug)

	if !success {
		return puzzle, fmt.Errorf("could not solve the puzzle")
	}

	return puzzle, nil
}
