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

func runPrint(difficulty int) {
	if !slices.Contains(ValidDifficulties, difficulty) {
		log.Fatalf("difficulty must be between 0 and %d", len(ValidDifficulties)-1)
	} else if difficulty == 0 {
		difficulty = ValidDifficulties[rand.IntN(len(ValidDifficulties)-1)+1]
	}
	game := generateSudokuParallel(difficulty, -1, -1)
	println("Generated Sudoku:")
	printBoard(game.board)
	println("Solution:")
	printBoard(game.solution)
}

func main() {
	var (
		cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
		print      = flag.Int("print", -1, "print a generated sudoku of `difficulty` and its solution and exit")
	)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: sugoku [-print <difficulty>] [-cpuprofile <file>]\n")
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

	if *print > -1 {
		runPrint(*print)
	} else {
		runTui()
	}
}
