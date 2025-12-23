package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"runtime/pprof"
	"slices"
	"strings"
)

func printBoard(board [9][9]uint8) {
	leftPad := "   "
	hPad := " "
	vPad := "\n"
	cellWidth := 3*(2+2*len(hPad)) - 1
	builder := new(strings.Builder)
	builder.WriteString(leftPad + "|")
	builder.WriteString(strings.Repeat("-", cellWidth))
	builder.WriteString("|")
	builder.WriteString(strings.Repeat("-", cellWidth))
	builder.WriteString("|")
	builder.WriteString(strings.Repeat("-", cellWidth))
	builder.WriteString("|" + vPad)
	for i := range 9 {
		for j := range 9 {
			if j == 0 {
				builder.WriteString(leftPad + "|" + hPad)
			} else if j%3 == 0 && j != 0 {
				builder.WriteString(hPad + "|" + hPad)
			} else {
				builder.WriteString(strings.Repeat(hPad, 2) + " ")
			}
			if board[i][j] == 0 {
				builder.WriteString(" ")
			} else {
				fmt.Fprintf(builder, "%d", board[i][j])
			}
		}
		builder.WriteString(hPad + "|" + vPad + leftPad + "|")
		if i%3 == 2 {
			builder.WriteString(strings.Repeat("-", cellWidth))
			builder.WriteString("|")
			builder.WriteString(strings.Repeat("-", cellWidth))
			builder.WriteString("|")
			builder.WriteString(strings.Repeat("-", cellWidth))
		} else {
			builder.WriteString(strings.Repeat(" ", cellWidth))
			builder.WriteString("|")
			builder.WriteString(strings.Repeat(" ", cellWidth))
			builder.WriteString("|")
			builder.WriteString(strings.Repeat(" ", cellWidth))
		}
		builder.WriteString("|" + vPad)
	}
	fmt.Println(builder.String())
}

func runPrint(game Sudoku) {
	println("Generated Sudoku:")
	printBoard(game.board)
	println("Solution:")
	printBoard(game.solution)
}

func main() {
	var (
		cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
		print      = flag.Bool("print", false, "print a generated sudoku and its solution and exit")
		seed       = flag.Int("seed", -1, "seed for random number generator, -1 for random seed")
		cores      = flag.Int("cores", -1, "number of cores to use, -1 for all cores")
		difficulty = flag.Int("difficulty", 0, "difficulty of the generated sudoku, 0 for random difficulty (default 0)")
		load       = flag.String("load", "", "load a sudoku from a file")
	)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: sugoku [-difficulty <0-5>] [-print] [-cores <int>] [-seed <int>] [-cpuprofile <file>] [-load <file>]\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	if !slices.Contains(validDifficulties, *difficulty) {
		log.Fatalf("difficulty must be between 0 and %d", len(validDifficulties)-1)
	} else if *difficulty == 0 {
		*difficulty = validDifficulties[rand.IntN(len(validDifficulties)-1)+1]
	}

	var game Sudoku
	var editable [9][9]bool
	if *load != "" {
		game, editable = loadSudoku(*load)
	} else {
		game = generateSudokuParallel(*difficulty, *seed, *cores)
		for i := range 9 {
			for j := range 9 {
				if game.board[i][j] == 0 {
					editable[i][j] = true
				}
			}
		}
	}

	if *print {
		runPrint(game)
	} else {
		runTui(game, editable, *difficulty, *cores)
	}
}
