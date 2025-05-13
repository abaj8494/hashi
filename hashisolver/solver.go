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
	X, Y          int
	MaxBridges    int
	CurrentBridges int
	Neighbors     [DirectionsCount]*Island
	NeighborCount int
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
	Nodes    [MaxIslands]*Island
	Edges    [MaxBridges]*Bridge
	Rows     int
	Cols     int
	IslandCount  int
	BridgeCount  int
	FullBridges  int // total bridges needed
	BuiltBridges int // bridges built so far
	Solved   [MaxIslands]int
	Attempts int
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
func findBridge(puzzle *Puzzle, is *Island) int {
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
			return b
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

// CanBuildBridge checks if a bridge can be built in the given direction
func CanBuildBridge(puzzle *Puzzle, curr *Island, dir int) bool {
	i1 := curr
	i2 := curr.Neighbors[dir]
	overlap := false
	
	if i2 == nil {
		return false
	}
	
	// Forward checking
	if i1.MaxBridges == i1.CurrentBridges || i2.MaxBridges == i2.CurrentBridges {
		return false
	}
	
	// Check existing bridges
	var b *Bridge
	for i := 0; i < puzzle.BridgeCount; i++ {
		b = puzzle.Edges[i]
		if (b.Island1 == i1 && b.Island2 == i2) || (b.Island1 == i2 && b.Island2 == i1) {
			if b.Wires == 3 {
				return false
			}
			overlap = true
		}
	}
	
	// Initialize crossover check array
	soln := make([][]int, puzzle.Rows)
	for i := range soln {
		soln[i] = make([]int, puzzle.Cols)
	}
	
	// Cells with islands cannot be crossed
	for i := 0; i < puzzle.IslandCount; i++ {
		is := puzzle.Nodes[i]
		soln[is.Y][is.X] = is.MaxBridges
	}
	
	// Mark existing bridges
	for i := 0; i < puzzle.BridgeCount; i++ {
		b = puzzle.Edges[i]
		if b.Skip {
			continue
		}
		i1 := b.Island1
		i2 := b.Island2
		dist := abs(i1.X-i2.X) + abs(i1.Y-i2.Y)
		
		switch b.Direction {
		case Up:
			for i := 1; i < dist; i++ {
				soln[i1.Y-i][i1.X] = b.Wires
			}
		case Down:
			for i := 1; i < dist; i++ {
				soln[i1.Y+i][i1.X] = b.Wires
			}
		case Left:
			for i := 1; i < dist; i++ {
				soln[i1.Y][i1.X-i] = b.Wires
			}
		case Right:
			for i := 1; i < dist; i++ {
				soln[i1.Y][i1.X+i] = b.Wires
			}
		}
	}
	
	// Check for collisions if this isn't an overlap of an existing bridge
	if !overlap {
		dist := abs(i1.X-i2.X) + abs(i1.Y-i2.Y)
		
		switch dir {
		case Up:
			for i := 1; i < dist; i++ {
				if soln[i1.Y-i][i1.X] > 0 {
					return false
				}
			}
		case Down:
			for i := 1; i < dist; i++ {
				if soln[i1.Y+i][i1.X] > 0 {
					return false
				}
			}
		case Left:
			for i := 1; i < dist; i++ {
				if soln[i1.Y][i1.X-i] > 0 {
					return false
				}
			}
		case Right:
			for i := 1; i < dist; i++ {
				if soln[i1.Y][i1.X+i] > 0 {
					return false
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
	puzzle.BuiltBridges++
	i1.CurrentBridges++
	i2.CurrentBridges++
	
	idx1 := findBridge(puzzle, i1)
	idx2 := findBridge(puzzle, i2)
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

// RemoveBridge removes a bridge between islands
func RemoveBridge(puzzle *Puzzle, curr *Island, dir int) {
	i1 := curr
	i2 := curr.Neighbors[dir]
	
	puzzle.BuiltBridges--
	i1.CurrentBridges--
	i2.CurrentBridges--
	
	idx1 := findBridge(puzzle, i1)
	idx2 := findBridge(puzzle, i2)
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
	
	if b.Wires == 0 {
		b.Skip = true
		return
	}
  if (b.Wires == 1) {
    b->Wires = 0;
    b->Symbol = ' ';
    return;
  }
	
	b.Wires--
	
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
	default:
		b.Symbol = ' '
	}
}

// ApplyHeuristics applies initial heuristics to simplify the puzzle
func ApplyHeuristics(puzzle *Puzzle) {
	for i := 0; i < puzzle.IslandCount; i++ {
		curr := puzzle.Nodes[i]
		
		// Build bridges for high-value islands
		if curr.MaxBridges >= 10 {
			for dir := 0; dir < DirectionsCount; dir++ {
				if CanBuildBridge(puzzle, curr, dir) {
					AddBridge(puzzle, curr, dir)
				}
			}
		}
		if curr.MaxBridges >= 11 {
			for dir := 0; dir < DirectionsCount; dir++ {
				if CanBuildBridge(puzzle, curr, dir) {
					AddBridge(puzzle, curr, dir)
				}
			}
		}
		if curr.MaxBridges == 12 {
			for dir := 0; dir < DirectionsCount; dir++ {
				if CanBuildBridge(puzzle, curr, dir) {
					AddBridge(puzzle, curr, dir)
				}
			}
		}
		
		// Islands with only one neighbor must connect all bridges to that neighbor
		if curr.NeighborCount == 1 {
			var dir int
			for dir = 0; curr.Neighbors[dir] == nil; dir++ {
			}
			for b := 0; b < curr.MaxBridges; b++ {
				if CanBuildBridge(puzzle, curr, dir) {
					AddBridge(puzzle, curr, dir)
				}
			}
		}
		
		// Handle islands with 2 neighbors
		if curr.NeighborCount == 2 {
			if curr.MaxBridges >= 4 {
				var dir int
				for dir = 0; curr.Neighbors[dir] == nil; dir++ {
				}
				if CanBuildBridge(puzzle, curr, dir) {
					AddBridge(puzzle, curr, dir)
				}
				dir++
				for ; dir < DirectionsCount && curr.Neighbors[dir] == nil; dir++ {
				}
				if dir < DirectionsCount && CanBuildBridge(puzzle, curr, dir) {
					AddBridge(puzzle, curr, dir)
				}
			}
			if curr.MaxBridges >= 5 {
				var dir int
				for dir = 0; curr.Neighbors[dir] == nil; dir++ {
				}
				if CanBuildBridge(puzzle, curr, dir) {
					AddBridge(puzzle, curr, dir)
				}
				dir++
				for ; dir < DirectionsCount && curr.Neighbors[dir] == nil; dir++ {
				}
				if dir < DirectionsCount && CanBuildBridge(puzzle, curr, dir) {
					AddBridge(puzzle, curr, dir)
				}
			}
			if curr.MaxBridges >= 6 {
				var dir int
				for dir = 0; curr.Neighbors[dir] == nil; dir++ {
				}
				if CanBuildBridge(puzzle, curr, dir) {
					AddBridge(puzzle, curr, dir)
				}
				dir++
				for ; dir < DirectionsCount && curr.Neighbors[dir] == nil; dir++ {
				}
				if dir < DirectionsCount && CanBuildBridge(puzzle, curr, dir) {
					AddBridge(puzzle, curr, dir)
				}
			}
		}
		
		// Handle islands with 3 neighbors
		if curr.NeighborCount == 3 {
			for maxVal := 7; maxVal <= 9; maxVal++ {
				if curr.MaxBridges >= maxVal {
					var dir int
					for dir = 0; dir < DirectionsCount && curr.Neighbors[dir] == nil; dir++ {
					}
					if dir < DirectionsCount && CanBuildBridge(puzzle, curr, dir) {
						AddBridge(puzzle, curr, dir)
					}
					dir++
					for ; dir < DirectionsCount && curr.Neighbors[dir] == nil; dir++ {
					}
					if dir < DirectionsCount && CanBuildBridge(puzzle, curr, dir) {
						AddBridge(puzzle, curr, dir)
					}
					dir++
					for ; dir < DirectionsCount && curr.Neighbors[dir] == nil; dir++ {
					}
					if dir < DirectionsCount && CanBuildBridge(puzzle, curr, dir) {
						AddBridge(puzzle, curr, dir)
					}
				}
			}
		}
	}
	
	// Update solved status
	for i := 0; i < puzzle.IslandCount; i++ {
		curr := puzzle.Nodes[i]
		puzzle.Solved[i] = curr.MaxBridges - curr.CurrentBridges
	}
}

// ShouldBuildBridge evaluates if building a bridge is a good decision
func ShouldBuildBridge(puzzle *Puzzle, curr *Island, dir int) bool {
	AddBridge(puzzle, curr, dir)
	ret := true
	b1, b2 := 0, 0
	i1 := findBridge(puzzle, curr)
	i2 := findBridge(puzzle, curr.Neighbors[dir])
	
	for i := 0; i < DirectionsCount; i++ {
		if CanBuildBridge(puzzle, curr, i) {
			b1++
		}
		if CanBuildBridge(puzzle, curr.Neighbors[dir], i) {
			b2++
		}
	}
	if (b1 == 0 && puzzle.Solved[i1] > 0) || (b2 == 0 && puzzle.Solved[i2] > 0) {
		ret = false
	}
	RemoveBridge(puzzle, curr, dir)
	return ret
}

// CleanPuzzle resets the state of the puzzle
func CleanPuzzle(puzzle *Puzzle) {
	puzzle.Attempts = 0
	for i := 0; i < puzzle.IslandCount; i++ {
		if puzzle.Solved[i] != 0 {
			curr := puzzle.Nodes[i]
			for dir := 0; dir < DirectionsCount; dir++ {
				if curr.Neighbors[dir] != nil {
					RemoveBridge(puzzle, curr, dir)
				}
			}
		}
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

// SolveMap attempts to solve the puzzle using backtracking
func SolveMap(puzzle *Puzzle, debug bool) bool {
	if puzzle.Attempts == 2000 {
		CleanPuzzle(puzzle)
	}
	
	if debug {
		PrintMap(puzzle)
		fmt.Printf("\nI want %d bridges\nI have %d\n", puzzle.FullBridges, puzzle.BuiltBridges)
		fmt.Println()
		for i := 0; i < puzzle.IslandCount; i++ {
			fmt.Printf("%.2d ", puzzle.Solved[i])
		}
		fmt.Println()
	}
	
	puzzle.Attempts++
	
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
		islands = append(islands, IndexedIsland{
			Island: puzzle.Nodes[i],
			Index:  i,
		})
	}
	
	sort.Slice(islands, func(i, j int) bool {
		// Sort by remaining bridges
		a := islands[i].Island
		b := islands[j].Island
		
		// First sort by remaining bridges
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
		
		if puzzle.Solved[idx] == 0 {
			continue // Island already satisfied
		}
		
		for dir := 0; dir < DirectionsCount; dir++ {
			if CanBuildBridge(puzzle, curr, dir) && ShouldBuildBridge(puzzle, curr, dir) {
				AddBridge(puzzle, curr, dir)
				if SolveMap(puzzle, debug) {
					return true
				}
				RemoveBridge(puzzle, curr, dir)
			}
		}
	}
	
	return false
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
		
		dist := abs(b.Island1.X-b.Island2.X) + abs(b.Island1.Y-b.Island2.Y) - 1
		
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
	
	// Apply initial heuristics
	ApplyHeuristics(puzzle)
	
	// Solve the puzzle
	success := SolveMap(puzzle, debug)
	
	if !success {
		return puzzle, fmt.Errorf("could not solve the puzzle")
	}
	
	return puzzle, nil
} 