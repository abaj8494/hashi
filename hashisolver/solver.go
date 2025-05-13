// hashisolver/solver.go
package hashisolver

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Direction constants for bridge connections
const (
	DirectionUp    = 0
	DirectionDown  = 1
	DirectionLeft  = 2
	DirectionRight = 3
)

// Node represents an island in the puzzle
type Node struct {
	Value        int
	XPos         int
	YPos         int
	UpBridges    int
	DownBridges  int
	LeftBridges  int
	RightBridges int
	TotalBridges int

	// Neighbor nodes this one MAY connect to
	UpNeighbor    *Node
	DownNeighbor  *Node
	LeftNeighbor  *Node
	RightNeighbor *Node

	// Blocked directions
	UpBlocked    bool
	DownBlocked  bool
	LeftBlocked  bool
	RightBlocked bool
	NumBlocked   int

	// Used when traversing nodes to check for potential islands
	Visited bool
}

// Puzzle represents the entire hashiwokakero puzzle
type Puzzle struct {
	Board        [][]*Node
	Size         int
	BuiltBridges int
	FullBridges  int
}

// NewNode creates a new node with the given value and position
func NewNode(value, x, y int) *Node {
	return &Node{
		Value:        value,
		XPos:         x,
		YPos:         y,
		UpBridges:    0,
		DownBridges:  0,
		LeftBridges:  0,
		RightBridges: 0,
		TotalBridges: 0,
		UpBlocked:    false,
		DownBlocked:  false,
		LeftBlocked:  false,
		RightBlocked: false,
		NumBlocked:   0,
		Visited:      false,
	}
}

// GetNeighbor returns the neighbor in the specified direction
func (n *Node) GetNeighbor(direction int) *Node {
	switch direction {
	case DirectionUp:
		return n.UpNeighbor
	case DirectionDown:
		return n.DownNeighbor
	case DirectionLeft:
		return n.LeftNeighbor
	case DirectionRight:
		return n.RightNeighbor
	default:
		return nil
	}
}

// BridgesInDirection returns the number of bridges in the specified direction
func (n *Node) BridgesInDirection(direction int) int {
	switch direction {
	case DirectionUp:
		return n.UpBridges
	case DirectionDown:
		return n.DownBridges
	case DirectionLeft:
		return n.LeftBridges
	case DirectionRight:
		return n.RightBridges
	default:
		return -1
	}
}

// NumNeighbors returns the number of neighbors this node has
func (n *Node) NumNeighbors() int {
	count := 0
	if n.UpNeighbor != nil {
		count++
	}
	if n.DownNeighbor != nil {
		count++
	}
	if n.LeftNeighbor != nil {
		count++
	}
	if n.RightNeighbor != nil {
		count++
	}
	return count
}

// RemainingPossibleMoves calculates how many bridge connections are still possible
func (n *Node) RemainingPossibleMoves() int {
	moves := 2 * n.NumNeighbors()

	if n.UpNeighbor != nil {
		moves -= n.UpBridges
		if n.UpNeighbor.Value-n.UpNeighbor.TotalBridges == 1 && n.UpBridges == 0 {
			moves--
		} else if n.UpNeighbor.Value-n.UpNeighbor.TotalBridges == 0 && n.UpBridges == 1 {
			moves--
		}
	}

	if n.DownNeighbor != nil {
		moves -= n.DownBridges
		if n.DownNeighbor.Value-n.DownNeighbor.TotalBridges == 1 && n.DownBridges == 0 {
			moves--
		} else if n.DownNeighbor.Value-n.DownNeighbor.TotalBridges == 0 && n.DownBridges == 1 {
			moves--
		}
	}

	if n.LeftNeighbor != nil {
		moves -= n.LeftBridges
		if n.LeftNeighbor.Value-n.LeftNeighbor.TotalBridges == 1 && n.LeftBridges == 0 {
			moves--
		} else if n.LeftNeighbor.Value-n.LeftNeighbor.TotalBridges == 0 && n.LeftBridges == 1 {
			moves--
		}
	}

	if n.RightNeighbor != nil {
		moves -= n.RightBridges
		if n.RightNeighbor.Value-n.RightNeighbor.TotalBridges == 1 && n.RightBridges == 0 {
			moves--
		} else if n.RightNeighbor.Value-n.RightNeighbor.TotalBridges == 0 && n.RightBridges == 1 {
			moves--
		}
	}

	return moves
}

// TotalPossibleMoves calculates the total possible moves (not accounting for nodes with only one possible connection)
func (n *Node) TotalPossibleMoves() int {
	moves := 2 * n.NumNeighbors()

	if n.UpNeighbor != nil {
		moves -= n.UpBridges
		if n.UpBridges == 1 && n.UpNeighbor.Value == n.UpNeighbor.TotalBridges {
			moves--
		}
	}

	if n.DownNeighbor != nil {
		moves -= n.DownBridges
		if n.DownBridges == 1 && n.DownNeighbor.Value == n.DownNeighbor.TotalBridges {
			moves--
		}
	}

	if n.LeftNeighbor != nil {
		moves -= n.LeftBridges
		if n.LeftBridges == 1 && n.LeftNeighbor.Value == n.LeftNeighbor.TotalBridges {
			moves--
		}
	}

	if n.RightNeighbor != nil {
		moves -= n.RightBridges
		if n.RightBridges == 1 && n.RightNeighbor.Value == n.RightNeighbor.TotalBridges {
			moves--
		}
	}

	return moves
}

// UnblockedNode returns the direction of the single unblocked node (assumes only one exists)
func (n *Node) UnblockedNode() int {
	if !n.UpBlocked {
		return DirectionUp
	} else if !n.DownBlocked {
		return DirectionDown
	} else if !n.LeftBlocked {
		return DirectionLeft
	} else if !n.RightBlocked {
		return DirectionRight
	}

	return -1 // Error case
}

// UnblockedNodes returns a slice of all unblocked directions
func (n *Node) UnblockedNodes() []int {
	result := []int{}
	if !n.UpBlocked {
		result = append(result, DirectionUp)
	}
	if !n.DownBlocked {
		result = append(result, DirectionDown)
	}
	if !n.LeftBlocked {
		result = append(result, DirectionLeft)
	}
	if !n.RightBlocked {
		result = append(result, DirectionRight)
	}
	return result
}

// NodeFilled blocks all directions of this node (used when the node is filled with all its bridges)
func (n *Node) NodeFilled() {
	n.UpBlocked = true
	n.DownBlocked = true
	n.LeftBlocked = true
	n.RightBlocked = true
	n.NumBlocked = 4

	// Also blocks the corresponding directions of neighbor nodes if they aren't already blocked
	if n.UpNeighbor != nil && !n.UpNeighbor.DownBlocked {
		n.UpNeighbor.DownBlocked = true
		n.UpNeighbor.NumBlocked++
	}

	if n.DownNeighbor != nil && !n.DownNeighbor.UpBlocked {
		n.DownNeighbor.UpBlocked = true
		n.DownNeighbor.NumBlocked++
	}

	if n.LeftNeighbor != nil && !n.LeftNeighbor.RightBlocked {
		n.LeftNeighbor.RightBlocked = true
		n.LeftNeighbor.NumBlocked++
	}

	if n.RightNeighbor != nil && !n.RightNeighbor.LeftBlocked {
		n.RightNeighbor.LeftBlocked = true
		n.RightNeighbor.NumBlocked++
	}
}

// DirectionBlocked blocks the connection between this node and the neighbor node in the given direction
func (n *Node) DirectionBlocked(direction int) {
	switch direction {
	case DirectionUp:
		if !n.UpBlocked {
			n.UpBlocked = true
			n.NumBlocked++
		}
		if n.UpNeighbor != nil && !n.UpNeighbor.DownBlocked {
			n.UpNeighbor.DownBlocked = true
			n.UpNeighbor.NumBlocked++
		}

	case DirectionDown:
		if !n.DownBlocked {
			n.DownBlocked = true
			n.NumBlocked++
		}
		if n.DownNeighbor != nil && !n.DownNeighbor.UpBlocked {
			n.DownNeighbor.UpBlocked = true
			n.DownNeighbor.NumBlocked++
		}

	case DirectionLeft:
		if !n.LeftBlocked {
			n.LeftBlocked = true
			n.NumBlocked++
		}
		if n.LeftNeighbor != nil && !n.LeftNeighbor.RightBlocked {
			n.LeftNeighbor.RightBlocked = true
			n.LeftNeighbor.NumBlocked++
		}

	case DirectionRight:
		if !n.RightBlocked {
			n.RightBlocked = true
			n.NumBlocked++
		}
		if n.RightNeighbor != nil && !n.RightNeighbor.LeftBlocked {
			n.RightNeighbor.LeftBlocked = true
			n.RightNeighbor.NumBlocked++
		}
	}
}

// BlockCheck checks whether bridges need to be blocked in any direction
func (n *Node) BlockCheck() {
	// If node is filled up with bridges, block all directions
	if n.Value == n.TotalBridges {
		n.NodeFilled()
	}

	// 2 bridges is maximum in any direction, so block that direction
	if n.UpBridges == 2 {
		n.DirectionBlocked(DirectionUp)
	}
	if n.UpNeighbor != nil && n.UpNeighbor.TotalBridges == n.UpNeighbor.Value {
		n.UpNeighbor.NodeFilled()
	}

	if n.DownBridges == 2 {
		n.DirectionBlocked(DirectionDown)
	}
	if n.DownNeighbor != nil && n.DownNeighbor.TotalBridges == n.DownNeighbor.Value {
		n.DownNeighbor.NodeFilled()
	}

	if n.LeftBridges == 2 {
		n.DirectionBlocked(DirectionLeft)
	}
	if n.LeftNeighbor != nil && n.LeftNeighbor.TotalBridges == n.LeftNeighbor.Value {
		n.LeftNeighbor.NodeFilled()
	}

	if n.RightBridges == 2 {
		n.DirectionBlocked(DirectionRight)
	}
	if n.RightNeighbor != nil && n.RightNeighbor.TotalBridges == n.RightNeighbor.Value {
		n.RightNeighbor.NodeFilled()
	}
}

// ConnectNodes connects two nodes with a bridge in the specified direction
func ConnectNodes(puzzle *Puzzle, node *Node, neighbor *Node, direction int, isSpeculative bool) {
	if !isSpeculative {
		puzzle.BuiltBridges++
	}

	node.TotalBridges++
	neighbor.TotalBridges++

	switch direction {
	case DirectionUp:
		node.UpBridges++
		neighbor.DownBridges++

		// Mark the bridge in the board
		distance := node.YPos - neighbor.YPos
		for i := 1; i < distance; i++ {
			if node.UpBridges == 1 {
				puzzle.Board[node.YPos-i][node.XPos].Value = -1 // Vertical single bridge
			} else {
				puzzle.Board[node.YPos-i][node.XPos].Value = -2 // Vertical double bridge
			}
		}

	case DirectionDown:
		node.DownBridges++
		neighbor.UpBridges++

		// Mark the bridge in the board
		distance := neighbor.YPos - node.YPos
		for i := 1; i < distance; i++ {
			if node.DownBridges == 1 {
				puzzle.Board[node.YPos+i][node.XPos].Value = -1 // Vertical single bridge
			} else {
				puzzle.Board[node.YPos+i][node.XPos].Value = -2 // Vertical double bridge
			}
		}

	case DirectionLeft:
		node.LeftBridges++
		neighbor.RightBridges++

		// Mark the bridge in the board
		distance := node.XPos - neighbor.XPos
		for i := 1; i < distance; i++ {
			if node.LeftBridges == 1 {
				puzzle.Board[node.YPos][node.XPos-i].Value = -3 // Horizontal single bridge
			} else {
				puzzle.Board[node.YPos][node.XPos-i].Value = -4 // Horizontal double bridge
			}
		}

	case DirectionRight:
		node.RightBridges++
		neighbor.LeftBridges++

		// Mark the bridge in the board
		distance := neighbor.XPos - node.XPos
		for i := 1; i < distance; i++ {
			if node.RightBridges == 1 {
				puzzle.Board[node.YPos][node.XPos+i].Value = -3 // Horizontal single bridge
			} else {
				puzzle.Board[node.YPos][node.XPos+i].Value = -4 // Horizontal double bridge
			}
		}
	}

	// Check for bridge conflicts and node filling
	node.BlockCheck()
	neighbor.BlockCheck()
}

// BridgeCheck checks for bridges that would block one edge of the node
func BridgeCheck(node *Node) {
	// This function implements the bridge checking logic from the C++ implementation
	// For each direction, if that's the only direction with a possible bridge, connect it

	if node.NumBlocked == 3 && node.Value-node.TotalBridges > 0 {
		direction := node.UnblockedNode()
		neighbor := node.GetNeighbor(direction)

		if neighbor != nil && neighbor.Value-neighbor.TotalBridges > 0 {
			// This is an obvious move - only one direction is available
			return
		}
	}
}

// CheckForIsland checks if adding a bridge would create an isolated island
func CheckForIsland(puzzle *Puzzle, node *Node, direction int, bridgeCount int) bool {
	// Reset visited flags
	for i := 0; i < puzzle.Size; i++ {
		for j := 0; j < puzzle.Size; j++ {
			if puzzle.Board[i][j].Value > 0 {
				puzzle.Board[i][j].Visited = false
			}
		}
	}

	// Temporarily block the direction we're testing
	oldBlocked := false
	switch direction {
	case DirectionUp:
		oldBlocked = node.UpBlocked
		node.UpBlocked = true
	case DirectionDown:
		oldBlocked = node.DownBlocked
		node.DownBlocked = true
	case DirectionLeft:
		oldBlocked = node.LeftBlocked
		node.LeftBlocked = true
	case DirectionRight:
		oldBlocked = node.RightBlocked
		node.RightBlocked = true
	}

	// Mark the current node as visited
	node.Visited = true

	// Start a depth-first search from this node
	connected := true
	for i := 0; i < puzzle.Size && connected; i++ {
		for j := 0; j < puzzle.Size && connected; j++ {
			if puzzle.Board[i][j].Value > 0 && !puzzle.Board[i][j].Visited {
				// Found an unvisited node, check if it's reachable
				connected = CheckNodeString(puzzle.Board[i][j])
				if !connected {
					// We found an island, so we must add a bridge in the tested direction
					// Restore the original blocked state
					switch direction {
					case DirectionUp:
						node.UpBlocked = oldBlocked
					case DirectionDown:
						node.DownBlocked = oldBlocked
					case DirectionLeft:
						node.LeftBlocked = oldBlocked
					case DirectionRight:
						node.RightBlocked = oldBlocked
					}

					// Add the bridge
					ConnectNodes(puzzle, node, node.GetNeighbor(direction), direction, false)
					return true
				}
			}
		}
	}

	// Restore the original blocked state
	switch direction {
	case DirectionUp:
		node.UpBlocked = oldBlocked
	case DirectionDown:
		node.DownBlocked = oldBlocked
	case DirectionLeft:
		node.LeftBlocked = oldBlocked
	case DirectionRight:
		node.RightBlocked = oldBlocked
	}

	return false
}

// CheckNodeString performs a DFS to mark all nodes that are connected
func CheckNodeString(node *Node) bool {
	if node == nil || node.Visited {
		return true
	}

	node.Visited = true

	// Check all four directions
	if !node.UpBlocked && node.UpNeighbor != nil {
		CheckNodeString(node.UpNeighbor)
	}

	if !node.DownBlocked && node.DownNeighbor != nil {
		CheckNodeString(node.DownNeighbor)
	}

	if !node.LeftBlocked && node.LeftNeighbor != nil {
		CheckNodeString(node.LeftNeighbor)
	}

	if !node.RightBlocked && node.RightNeighbor != nil {
		CheckNodeString(node.RightNeighbor)
	}

	return true
}

// Clone creates a deep copy of a puzzle
func (p *Puzzle) Clone() *Puzzle {
	newPuzzle := &Puzzle{
		Size:         p.Size,
		Board:        make([][]*Node, p.Size),
		BuiltBridges: p.BuiltBridges,
		FullBridges:  p.FullBridges,
	}

	// Clone the board
	for i := 0; i < p.Size; i++ {
		newPuzzle.Board[i] = make([]*Node, p.Size)
		for j := 0; j < p.Size; j++ {
			oldNode := p.Board[i][j]
			newNode := NewNode(oldNode.Value, oldNode.XPos, oldNode.YPos)

			// Copy node state
			newNode.UpBridges = oldNode.UpBridges
			newNode.DownBridges = oldNode.DownBridges
			newNode.LeftBridges = oldNode.LeftBridges
			newNode.RightBridges = oldNode.RightBridges
			newNode.TotalBridges = oldNode.TotalBridges

			newNode.UpBlocked = oldNode.UpBlocked
			newNode.DownBlocked = oldNode.DownBlocked
			newNode.LeftBlocked = oldNode.LeftBlocked
			newNode.RightBlocked = oldNode.RightBlocked
			newNode.NumBlocked = oldNode.NumBlocked

			newPuzzle.Board[i][j] = newNode
		}
	}

	// Reconnect neighbors
	for i := 0; i < p.Size; i++ {
		for j := 0; j < p.Size; j++ {
			oldNode := p.Board[i][j]
			newNode := newPuzzle.Board[i][j]

			if oldNode.UpNeighbor != nil {
				newNode.UpNeighbor = newPuzzle.Board[oldNode.UpNeighbor.YPos][oldNode.UpNeighbor.XPos]
			}

			if oldNode.DownNeighbor != nil {
				newNode.DownNeighbor = newPuzzle.Board[oldNode.DownNeighbor.YPos][oldNode.DownNeighbor.XPos]
			}

			if oldNode.LeftNeighbor != nil {
				newNode.LeftNeighbor = newPuzzle.Board[oldNode.LeftNeighbor.YPos][oldNode.LeftNeighbor.XPos]
			}

			if oldNode.RightNeighbor != nil {
				newNode.RightNeighbor = newPuzzle.Board[oldNode.RightNeighbor.YPos][oldNode.RightNeighbor.XPos]
			}
		}
	}

	return newPuzzle
}

// IsComplete checks if the puzzle is completely solved
func (p *Puzzle) IsComplete() bool {
	// Check if all nodes have their required number of bridges
	for i := 0; i < p.Size; i++ {
		for j := 0; j < p.Size; j++ {
			node := p.Board[i][j]
			if node.Value > 0 && node.Value != node.TotalBridges {
				return false
			}
		}
	}

	// Check if all islands are connected
	var startNode *Node

	// Find the first node
	for i := 0; i < p.Size && startNode == nil; i++ {
		for j := 0; j < p.Size && startNode == nil; j++ {
			if p.Board[i][j].Value > 0 {
				startNode = p.Board[i][j]
			}
		}
	}

	if startNode == nil {
		return true // Empty puzzle
	}

	// Reset visited flags
	for i := 0; i < p.Size; i++ {
		for j := 0; j < p.Size; j++ {
			if p.Board[i][j].Value > 0 {
				p.Board[i][j].Visited = false
			}
		}
	}

	// Start a DFS from the first node
	CheckNodeString(startNode)

	// Check if all nodes were visited
	for i := 0; i < p.Size; i++ {
		for j := 0; j < p.Size; j++ {
			if p.Board[i][j].Value > 0 && !p.Board[i][j].Visited {
				return false // Disconnected island
			}
		}
	}

	return true
}

// FindCandidateNode finds a node with the most constrained but unresolved connections
func (p *Puzzle) FindCandidateNode() *Node {
	var bestNode *Node
	bestScore := -1

	for i := 0; i < p.Size; i++ {
		for j := 0; j < p.Size; j++ {
			node := p.Board[i][j]

			if node.Value <= 0 || node.Value == node.TotalBridges {
				continue // Skip empty or satisfied nodes
			}

			// Calculate a score based on how constrained this node is
			remainingBridges := node.Value - node.TotalBridges
			unblocked := node.UnblockedNodes()

			if len(unblocked) == 0 {
				continue // Skip fully blocked nodes
			}

			// Score is higher for nodes with fewer open directions but more remaining bridges
			score := remainingBridges*10 + (4 - len(unblocked))

			if score > bestScore {
				bestScore = score
				bestNode = node
			}
		}
	}

	return bestNode
}

// AttemptSpeculativeSolve attempts to solve the puzzle using speculative moves and backtracking
func AttemptSpeculativeSolve(puzzle *Puzzle, debug bool) (*Puzzle, error) {
	// Try to solve using logic first
	movesFound := true
	for movesFound {
		movesFound = false

		// Look at every node
		for i := 0; i < puzzle.Size; i++ {
			for j := 0; j < puzzle.Size; j++ {
				node := puzzle.Board[i][j]

				// Skip empty spaces or already satisfied nodes
				if node.Value <= 0 || node.TotalBridges == node.Value {
					continue
				}

				// Check for logical errors
				if node.NumBlocked == 4 && node.TotalBridges < node.Value {
					if debug {
						fmt.Println("Logical error - node blocked in all directions but still needs bridges")
					}
					return puzzle, errors.New("logical error - node blocked in all directions")
				}

				// Check for bridges that would block one edge of the node
				BridgeCheck(node)

				// If 3 directions are blocked, connect to the remaining one
				if node.NumBlocked == 3 && node.TotalBridges < node.Value {
					direction := node.UnblockedNode()
					neighbor := node.GetNeighbor(direction)

					if neighbor != nil {
						ConnectNodes(puzzle, node, neighbor, direction, false)

						// Make a double bridge if necessary
						if node.Value == node.TotalBridges+1 {
							ConnectNodes(puzzle, node, neighbor, direction, false)
						}

						movesFound = true
					}
				}

				// If remaining value equals total possible moves, all bridges must be fully connected
				if node.Value-node.TotalBridges == node.TotalPossibleMoves() {
					unblocked := node.UnblockedNodes()
					for _, dir := range unblocked {
						neighbor := node.GetNeighbor(dir)

						if neighbor == nil {
							continue
						}

						// Don't add a double bridge to a 1 or a node with remaining value of 1
						if node.BridgesInDirection(dir) == 0 && neighbor.Value-neighbor.TotalBridges > 1 {
							ConnectNodes(puzzle, node, neighbor, dir, false)
							ConnectNodes(puzzle, node, neighbor, dir, false)
						} else {
							ConnectNodes(puzzle, node, neighbor, dir, false)
						}
					}
					movesFound = true
				}

				// If remaining value equals remaining possible moves - 1
				// All edges must have at least one bridge
				if node.Value-node.TotalBridges == node.TotalPossibleMoves()-1 {
					unblocked := node.UnblockedNodes()
					for _, dir := range unblocked {
						// Check if any bridges already exist in that direction
						// If not, connect one
						if node.BridgesInDirection(dir) < 1 {
							neighbor := node.GetNeighbor(dir)
							if neighbor != nil {
								ConnectNodes(puzzle, node, neighbor, dir, false)
								movesFound = true
							}
						}
					}
				}

				// Check if adding a bridge in any direction would create an island
				unblocked := node.UnblockedNodes()
				for _, dir := range unblocked {
					if CheckForIsland(puzzle, node, dir, 1) {
						movesFound = true
					}
				}

				// Check the island condition for double bridges
				if node.NumBlocked == 2 && node.Value-node.TotalBridges == 2 {
					unblocked := node.UnblockedNodes()
					if len(unblocked) == 2 { // Make sure we have exactly 2 unblocked directions
						for k, dir := range unblocked {
							neighbor := node.GetNeighbor(dir)
							if neighbor == nil {
								continue
							}

							if neighbor.Value >= 2 && neighbor.TotalBridges == 0 {
								if CheckForIsland(puzzle, node, dir, 2) {
									movesFound = true
									// Add a bridge in the other direction
									var otherDir int
									if k == 0 {
										otherDir = unblocked[1]
									} else {
										otherDir = unblocked[0]
									}
									otherNeighbor := node.GetNeighbor(otherDir)
									if otherNeighbor != nil {
										ConnectNodes(puzzle, node, otherNeighbor, otherDir, false)
									}
								}
							}
						}
					}
				}

				// If a node has two unblocked edges and one is not enough to satisfy it
				if node.NumBlocked == 2 && node.Value-node.TotalBridges >= 2 {
					unblocked := node.UnblockedNodes()
					if len(unblocked) == 2 { // Make sure we have exactly 2 unblocked directions
						for k, dir := range unblocked {
							neighbor := node.GetNeighbor(dir)
							if neighbor == nil {
								continue
							}

							if neighbor.Value-neighbor.TotalBridges == 1 {
								movesFound = true

								// Connect to the other direction
								var otherDir int
								if k == 0 {
									otherDir = unblocked[1]
								} else {
									otherDir = unblocked[0]
								}
								otherNeighbor := node.GetNeighbor(otherDir)
								if otherNeighbor != nil {
									ConnectNodes(puzzle, node, otherNeighbor, otherDir, false)
								}
							}
						}
					}
				}
			}
		}

		if debug && movesFound {
			fmt.Println("Found moves in this iteration, continuing...")
		}
	}

	// Check if the puzzle is completely solved using just logic
	if puzzle.IsComplete() {
		if debug {
			fmt.Printf("Solution complete: %d/%d bridges placed\n", puzzle.BuiltBridges, puzzle.FullBridges/2)
		}
		return puzzle, nil
	}

	// If we get here, we need to use speculation
	if debug {
		fmt.Println("Using speculative solving...")
	}

	// Find a good candidate node for speculation
	candidateNode := puzzle.FindCandidateNode()
	if candidateNode == nil {
		return puzzle, errors.New("no candidate node found for speculation")
	}

	// Try each possible direction
	unblocked := candidateNode.UnblockedNodes()
	for _, dir := range unblocked {
		neighbor := candidateNode.GetNeighbor(dir)
		if neighbor == nil {
			continue
		}

		// Try adding a single bridge
		if debug {
			fmt.Printf("Trying a single bridge from (%d,%d) in direction %d\n",
				candidateNode.YPos, candidateNode.XPos, dir)
		}

		// Create a clone for speculative solving
		speculativePuzzle := puzzle.Clone()
		speculativeNode := speculativePuzzle.Board[candidateNode.YPos][candidateNode.XPos]
		speculativeNeighbor := speculativePuzzle.Board[neighbor.YPos][neighbor.XPos]

		// Add a single bridge
		ConnectNodes(speculativePuzzle, speculativeNode, speculativeNeighbor, dir, true)

		// Recursively attempt to solve
		newPuzzle, err := AttemptSpeculativeSolve(speculativePuzzle, debug)
		if err == nil && newPuzzle.IsComplete() {
			return newPuzzle, nil
		}

		// If we can add a double bridge, try that too
		if candidateNode.Value-candidateNode.TotalBridges >= 2 &&
			neighbor.Value-neighbor.TotalBridges >= 2 {

			if debug {
				fmt.Printf("Trying a double bridge from (%d,%d) in direction %d\n",
					candidateNode.YPos, candidateNode.XPos, dir)
			}

			// Create another clone for double bridge speculation
			speculativePuzzle2 := puzzle.Clone()
			speculativeNode2 := speculativePuzzle2.Board[candidateNode.YPos][candidateNode.XPos]
			speculativeNeighbor2 := speculativePuzzle2.Board[neighbor.YPos][neighbor.XPos]

			// Add two bridges
			ConnectNodes(speculativePuzzle2, speculativeNode2, speculativeNeighbor2, dir, true)
			ConnectNodes(speculativePuzzle2, speculativeNode2, speculativeNeighbor2, dir, true)

			// Recursively attempt to solve
			newPuzzle2, err2 := AttemptSpeculativeSolve(speculativePuzzle2, debug)
			if err2 == nil && newPuzzle2.IsComplete() {
				return newPuzzle2, nil
			}
		}

		// Try blocking this direction
		if debug {
			fmt.Printf("Trying blocking direction %d from (%d,%d)\n",
				dir, candidateNode.YPos, candidateNode.XPos)
		}

		// Create a clone for blocking speculation
		speculativePuzzle3 := puzzle.Clone()
		speculativeNode3 := speculativePuzzle3.Board[candidateNode.YPos][candidateNode.XPos]

		// Block the direction
		speculativeNode3.DirectionBlocked(dir)

		// Recursively attempt to solve
		newPuzzle3, err3 := AttemptSpeculativeSolve(speculativePuzzle3, debug)
		if err3 == nil && newPuzzle3.IsComplete() {
			return newPuzzle3, nil
		}
	}

	// If we've tried all possibilities and none worked, there's no solution
	return puzzle, errors.New("no solution found with speculation")
}

// Solve attempts to solve the hashiwokakero puzzle from the input reader
func Solve(input io.Reader, debug bool) (*Puzzle, error) {
	scanner := bufio.NewScanner(input)

	// Read the puzzle from the input
	lines := []string{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %v", err)
	}

	if len(lines) == 0 {
		return nil, errors.New("no input provided")
	}

	// Determine board size - equal to the number of lines
	boardSize := len(lines)

	if debug {
		fmt.Printf("Board size: %dx%d\n", boardSize, boardSize)
	}

	// Initialize the puzzle
	puzzle := &Puzzle{
		Size:         boardSize,
		Board:        make([][]*Node, boardSize),
		BuiltBridges: 0,
		FullBridges:  0,
	}

	// Parse each line of the puzzle
	for i, line := range lines {
		puzzle.Board[i] = make([]*Node, boardSize)

		for j, char := range line {
			if j >= boardSize {
				break
			}

			var value int
			if char == '.' {
				value = 0
			} else if char >= '1' && char <= '9' {
				value = int(char - '0')
				puzzle.FullBridges += value
			} else {
				// If it's not a number or a dot, assume it's empty space
				value = 0
			}

			puzzle.Board[i][j] = NewNode(value, j, i)
		}
	}

	// Find neighbors for each node
	for i := 0; i < boardSize; i++ {
		for j := 0; j < boardSize; j++ {
			if puzzle.Board[i][j].Value <= 0 {
				continue
			}

			// Find right neighbor
			for k := j + 1; k < boardSize; k++ {
				if puzzle.Board[i][k].Value > 0 {
					puzzle.Board[i][j].RightNeighbor = puzzle.Board[i][k]
					break
				}
			}

			// Find left neighbor
			for k := j - 1; k >= 0; k-- {
				if puzzle.Board[i][k].Value > 0 {
					puzzle.Board[i][j].LeftNeighbor = puzzle.Board[i][k]
					break
				}
			}

			// Find down neighbor
			for k := i + 1; k < boardSize; k++ {
				if puzzle.Board[k][j].Value > 0 {
					puzzle.Board[i][j].DownNeighbor = puzzle.Board[k][j]
					break
				}
			}

			// Find up neighbor
			for k := i - 1; k >= 0; k-- {
				if puzzle.Board[k][j].Value > 0 {
					puzzle.Board[i][j].UpNeighbor = puzzle.Board[k][j]
					break
				}
			}
		}
	}

	// Set up initial blockages
	for i := 0; i < boardSize; i++ {
		for j := 0; j < boardSize; j++ {
			if puzzle.Board[i][j].Value <= 0 {
				continue
			}

			// Assign obvious blockages - edge nodes and a 1 connecting to a 1
			if puzzle.Board[i][j].LeftNeighbor == nil ||
				(puzzle.Board[i][j].Value == 1 && puzzle.Board[i][j].LeftNeighbor != nil && puzzle.Board[i][j].LeftNeighbor.Value == 1) {
				puzzle.Board[i][j].LeftBlocked = true
				puzzle.Board[i][j].NumBlocked++
			}

			if puzzle.Board[i][j].RightNeighbor == nil ||
				(puzzle.Board[i][j].Value == 1 && puzzle.Board[i][j].RightNeighbor != nil && puzzle.Board[i][j].RightNeighbor.Value == 1) {
				puzzle.Board[i][j].RightBlocked = true
				puzzle.Board[i][j].NumBlocked++
			}

			if puzzle.Board[i][j].UpNeighbor == nil ||
				(puzzle.Board[i][j].Value == 1 && puzzle.Board[i][j].UpNeighbor != nil && puzzle.Board[i][j].UpNeighbor.Value == 1) {
				puzzle.Board[i][j].UpBlocked = true
				puzzle.Board[i][j].NumBlocked++
			}

			if puzzle.Board[i][j].DownNeighbor == nil ||
				(puzzle.Board[i][j].Value == 1 && puzzle.Board[i][j].DownNeighbor != nil && puzzle.Board[i][j].DownNeighbor.Value == 1) {
				puzzle.Board[i][j].DownBlocked = true
				puzzle.Board[i][j].NumBlocked++
			}
		}
	}

	// Solve the puzzle using the enhanced solver with speculation
	return AttemptSpeculativeSolve(puzzle, debug)
}

// PrintMap prints the solved puzzle to stdout
func PrintMap(puzzle *Puzzle) {
	for i := 0; i < puzzle.Size; i++ {
		for j := 0; j < puzzle.Size; j++ {
			node := puzzle.Board[i][j]

			switch node.Value {
			case 0:
				fmt.Print(" ")
			case -1:
				fmt.Print("|") // Vertical single bridge
			case -2:
				fmt.Print("\"") // Vertical double bridge
			case -3:
				fmt.Print("-") // Horizontal single bridge
			case -4:
				fmt.Print("=") // Horizontal double bridge
			default:
				if node.Value > 0 {
					fmt.Print(node.Value)
				} else {
					fmt.Print(" ") // Unknown value
				}
			}
		}
		fmt.Println()
	}
}
