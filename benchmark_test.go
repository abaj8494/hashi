package main

import (
	"fmt"
	"strings"
	"testing"

	"hashi/hashisolver"
)

// BenchmarkSolver benchmarks the solver with different board sizes
func BenchmarkSolver(b *testing.B) {
	// Test with different board sizes
	sizes := []struct {
		rows int
		cols int
	}{
		{3, 3},   // Small
		{5, 5},   // Small-medium
		{8, 8},   // Medium
		{10, 10}, // Medium-large
		{15, 15}, // Large
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%dx%d", size.rows, size.cols), func(b *testing.B) {
			// Generate a puzzle using bridgen before starting the benchmark
			puzzle, err := runBridgenCommand(size.rows, size.cols)
			if err != nil {
				b.Fatalf("Failed to generate puzzle: %v", err)
			}

			// Reset the timer for the actual benchmark
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				// Create a new reader for each iteration to avoid reading an empty stream
				reader := strings.NewReader(puzzle)
				
				// Solve the puzzle (without debug output)
				_, err := hashisolver.Solve(reader, false)
				if err != nil {
					b.Fatalf("Failed to solve puzzle: %v", err)
				}
			}
		})
	}
}

// benchmarkHeuristicsVsNoHeuristics compares solving with and without heuristics
func BenchmarkHeuristicsVsNoHeuristics(b *testing.B) {
	// Use a medium-sized puzzle for the comparison
	puzzle, err := runBridgenCommand(8, 8)
	if err != nil {
		b.Fatalf("Failed to generate puzzle: %v", err)
	}

	b.Run("WithHeuristics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			reader := strings.NewReader(puzzle)
			p := hashisolver.NewPuzzle()
			mapData, _ := hashisolver.ScanMap(reader, p)
			hashisolver.ParseMap(mapData, p)
			
			// Apply heuristics
			hashisolver.ApplyHeuristics(p)
			
			// Solve the puzzle
			hashisolver.SolveMap(p, false)
		}
	})

	b.Run("WithoutHeuristics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			reader := strings.NewReader(puzzle)
			p := hashisolver.NewPuzzle()
			mapData, _ := hashisolver.ScanMap(reader, p)
			hashisolver.ParseMap(mapData, p)
			
			// Skip heuristics
			
			// Solve the puzzle
			hashisolver.SolveMap(p, false)
		}
	})
}

// Create a benchmark that measures memory allocation
func BenchmarkMemoryUsage(b *testing.B) {
	sizes := []struct {
		rows int
		cols int
	}{
		{5, 5},   // Small
		{10, 10}, // Medium
		{15, 15}, // Large
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%dx%d", size.rows, size.cols), func(b *testing.B) {
			puzzle, err := runBridgenCommand(size.rows, size.cols)
			if err != nil {
				b.Fatalf("Failed to generate puzzle: %v", err)
			}

			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				reader := strings.NewReader(puzzle)
				_, err := hashisolver.Solve(reader, false)
				if err != nil {
					b.Fatalf("Failed to solve puzzle: %v", err)
				}
			}
		})
	}
} 