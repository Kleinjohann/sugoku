package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func loadSudoku(filename string) (Sudoku, [9][9]bool) {
	var game Sudoku
	var err error
	var line, cellString, valueString, editableString string
	var splitLine []string
	var cell int
	var value uint64
	var cellEditable bool
	var editable [9][9]bool
	for i := range 9 {
		for j := range 9 {
			editable[i][j] = true
		}
	}
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	game = makeEmptySudoku()
	re := regexp.MustCompile("^[0-9]+;[1-9];[01]$")
	for scanner.Scan() {
		line = scanner.Text()
		if re.MatchString(line) {
			splitLine = strings.Split(line, ";")
			if len(splitLine) != 3 {
				panic(fmt.Sprintf("Not exactly 3 fields: Invalid line %s", line))
			}
			cellString = splitLine[0]
			valueString = splitLine[1]
			editableString = splitLine[2]
			cell, err = strconv.Atoi(cellString)
			if err != nil {
				panic(err)
			}
			value, err = strconv.ParseUint(valueString, 10, 8)
			if err != nil {
				panic(err)
			}
			cellEditable, err = strconv.ParseBool(editableString)
			if err != nil {
				panic(err)
			}
			editable[cell/9][cell%9] = cellEditable
			if cell < 0 || cell > 80 {
				panic(fmt.Sprintf("Invalid cell %d", cell))
			}
			if value < 1 || value > 9 {
				panic(fmt.Sprintf("Invalid value %d", value))
			}
			game.board[cell/9][cell%9] = uint8(value)
		} else {
			panic(fmt.Sprintf("Regex mismatch: Invalid line %s", line))
		}
	}
	if err = scanner.Err(); err != nil {
		panic(err)
	}
	computeCandidates(&game)
	if !isValidBoard(game.board) {
		panic("Invalid Sudoku")
	}
	quit := make(chan bool)
	numSolutions, solution := getNumSolutions(game, quit)
	close(quit)
	if numSolutions == 0 {
		panic("Sudoku has no solution")
	}
	if numSolutions > 1 {
		panic("Sudoku has multiple solutions")
	}
	game.solution = solution
	return game, editable
}

func saveSudoku(filename string, game Sudoku, editable [9][9]bool) {
	var editableCell uint8
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	for i := range 9 {
		for j := range 9 {
			if game.board[i][j] != 0 {
				if editable[i][j] {
					editableCell = 1
				} else {
					editableCell = 0
				}
				fmt.Fprintf(file, "%d;%d;%d\n", i*9+j, game.board[i][j], editableCell)
			}
		}
	}
}
