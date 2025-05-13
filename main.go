package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"hashi/hashisolver"
)

func main() {
	var inputFile string
	var debug bool

	flag.StringVar(&inputFile, "input", "", "Input puzzle file (use - for stdin)")
	flag.BoolVar(&debug, "debug", false, "Enable debug output")
	flag.Parse()

	var reader io.Reader
	if inputFile == "" || inputFile == "-" {
		reader = os.Stdin
	} else {
		file, err := os.Open(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		reader = file
	}

	puzzle, err := hashisolver.Solve(reader, debug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error solving puzzle: %v\n", err)
		os.Exit(1)
	}

	// Print the solution
	hashisolver.PrintMap(puzzle)
} 