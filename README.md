# life lesson
> whenever you port code, make sure it is correct before translating otherwise you will introduce a suicidal amount of bugs.

# constraint satisfaction problem

this is what _real_ AI looks like. 

i've rewritten one of my uni assessements in Go because I think that it is a more suitable language for the task.

there was a time where I exclusively wrote code in C --- a braver, but more arrogant and stupid time. now I just learn all of the languages as they are required.

# structure

anyways, the old c code is here, along with this fellas cpp implementation that I had to use for the port because my credit-level code was riddled with bugs

# usage

`cat puzzle.txt | go run main.go`

or
`go run main.go -input puzzle.txt`

`go run main.go -input puzzle.txt -debug`

## regression tests

`go test -v` verbose, duh

`go test -run TestSolverWithKnownPuzzles` specific test, duh
