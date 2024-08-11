package main

import (
    "flag"
    "fmt"
    "math/rand/v2"
    "strings"
)

type Sudoku struct {
    board [9][9]uint8
    solution [9][9]uint8
    candidates [9][9][9]bool
}

func makeEmptySudoku() Sudoku {
    var game Sudoku
    for i := 0; i < 9; i++ {
        for j := 0; j < 9; j++ {
            game.board[i][j] = 0
            for k := 0; k < 9; k++ {
                game.candidates[i][j][k] = true
            }
        }
    }
    return game
}

func copySudoku(game Sudoku) Sudoku {
    var newGame Sudoku
    for i := 0; i < 9; i++ {
        for j := 0; j < 9; j++ {
            newGame.board[i][j] = game.board[i][j]
            for k := 0; k < 9; k++ {
                newGame.candidates[i][j][k] = game.candidates[i][j][k]
            }
        }
    }
    return newGame
}

func isValidSet(set []uint8) bool {
    seen := make(map[uint8]bool)
    for _, value := range set {
        if value != 0 {
            if seen[value] {
                return false
            }
            seen[value] = true
        }
    }
    return true
}

func isValidBoard(board [9][9]uint8) bool {
    for i := 0; i < 9; i++ {
        row := board[i][:]

        boxRowStart, boxColumnStart := getBoxStartsFromBoxId(i)

        column := make([]uint8, 9)
        box := make([]uint8, 9)

        for j := 0; j < 9; j++ {
            column[j] = board[j][i]
            box[j] = board[boxRowStart + j / 3][boxColumnStart + j % 3]
        }

        isValidRow := isValidSet(row)
        isValidColumn := isValidSet(column)
        isValidBox := isValidSet(box)

        if (!isValidRow || !isValidColumn || !isValidBox) {
            return false
        }
    }
    return true
}

func isValidSolvedBoard(board [9][9]uint8) bool {
    return (isValidBoard(board) && isSolved(board))
}

func isValidUnsolvedBoard(board [9][9]uint8) bool {
    return (isValidBoard(board) && !isSolved(board))
}

func fillRandomCell(game *Sudoku) {
    row, col := getRandomEmptyCell(game.board)
    candidates := getCandidates(game, row, col)
    insertedValue, err := selectRandomCandidate(candidates)
    if err != nil {
        fillRandomCell(game)
    }
    game.board[row][col] = insertedValue
    updateCandidates(row, col, insertedValue, game, false)
}

func generateSudoku() Sudoku {
    game := makeEmptySudoku()
    var currentSolution [9][9]uint8
    var numSolutions int
    previousNumSolutions := 2
    var previousGame Sudoku
    // fill in 5 random cells according to the sudoku rules without checking for number of solutions
    // I'm pretty sure there cannot be a board with <5 filled cells that has 0 solutions
    for i := 0; i < 5; i++ {
        previousGame = copySudoku(game)
        fillRandomCell(&game)
    }
    // now start checking for number of solutions
    isRetry := false
    for {
        if isRetry {
            numSolutions = previousNumSolutions
        } else {
            numSolutions, currentSolution = getNumSolutions(game)
        }
        if numSolutions == 1 {
            if !isValidUnsolvedBoard(game.board) {
                panic("Invalid Sudoku")
            }
            game.solution = currentSolution
            if !isValidSolvedBoard(game.solution) {
                panic("Invalid Solution")
            }
            return game
        } else if numSolutions == 0 {
            isRetry = true
            game = copySudoku(previousGame)
        } else {
            previousGame = copySudoku(game)
            previousNumSolutions = numSolutions
            fillRandomCell(&game)
            isRetry = false
        }
    }
}

func getNumSolutions(game Sudoku) (int, [9][9]uint8) {
    currentGame := copySudoku(game)
    var candidates []uint8
    var err error
    var previousGame Sudoku
    numSolutions := 0
    currentNumSolutions := -1
    lastSolution := [9][9]uint8{}
    currentSolution := [9][9]uint8{}

    for row := 0; row < 9; row++ {
        for col := 0; col < 9; col++ {
            if currentGame.board[row][col] == 0 {
                candidates = getCandidates(&currentGame, row, col)
                currentNumSolutions = 0
                for _, candidate := range candidates {
                    previousGame = copySudoku(currentGame)
                    currentGame.board[row][col] = candidate
                    updateCandidates(row, col, candidate, &currentGame, false)
                    currentSolution, err = solveSudoku(currentGame)
                    if err == nil {
                        currentNumSolutions++
                        if currentSolution != lastSolution {
                            numSolutions++
                            if numSolutions > 1 {
                                return numSolutions, currentSolution
                            }
                            lastSolution = currentSolution
                        }
                    }
                    currentGame = copySudoku(previousGame)
                }
                if currentNumSolutions == 0 {
                    return 0, currentSolution
                }
            }
        }
    }
    if numSolutions != 1 {
        panic("Not exactly one solution after looping through all cells")
    }
    // if the last call to solveSudoku was unsuccessful,
    // currentSolution does not contain a valid solution
    if err != nil {
        return numSolutions, lastSolution
    }
    return numSolutions, currentSolution
}

func solveSudoku(game Sudoku) ([9][9]uint8, error) {
    currentGame := copySudoku(game)
    var row, col int
    var candidates []uint8
    var previousGame Sudoku

    if isSolved(currentGame.board) {
        if !isValidSolvedBoard(game.board) {
            panic("Solution found, but invalid")
        }
        return currentGame.board, nil
    }

    row, col = getMostConstrainedCell(&currentGame)
    candidates = getCandidates(&currentGame, row, col)
    if len(candidates) == 0 {
        return currentGame.board, fmt.Errorf("No candidates available")
    }
    for _, candidate := range candidates {
        previousGame = copySudoku(currentGame)
        currentGame.board[row][col] = candidate
        updateCandidates(row, col, candidate, &currentGame, false)
        solution, err := solveSudoku(currentGame)
        if err == nil{
            return solution, nil
        }
        game = copySudoku(previousGame)
    }

    return currentGame.board, fmt.Errorf("No solution found")
}

func isSolved(board [9][9]uint8) bool {
    for i := 0; i < 9; i++ {
        for j := 0; j < 9; j++ {
            if board[i][j] == 0 {
                return false
            }
        }
    }
    return true
}

func getBoxStartsFromBoxId(boxId int) (int, int) {
    boxRowStart := boxId / 3 * 3
    boxColumnStart := boxId % 3 * 3
    return boxRowStart, boxColumnStart
}

func getBoxStartsFromCell(row int, col int) (int, int) {
    boxId := (row / 3) * 3 + (col / 3)
    return getBoxStartsFromBoxId(boxId)
}

func getCandidates(game *Sudoku, row int, col int) []uint8 {
    candidates := make([]uint8, 0)
    for i := 1; i < 10; i++ {
        if game.candidates[row][col][i - 1] {
            candidates = append(candidates, uint8(i))
        }
    }
    return candidates
}

func getMostConstrainedCell(game *Sudoku) (int, int) {
    minCandidates := 10
    var row, col int
    for i := 0; i < 9; i++ {
        for j := 0; j < 9; j++ {
            if game.board[i][j] == 0 {
                candidates := getCandidates(game, i, j)
                if len(candidates) < minCandidates {
                    minCandidates = len(candidates)
                    row = i
                    col = j
                }
            }
        }
    }
    return row, col
}

func getRandomEmptyCell(board [9][9]uint8) (int, int) {
    for {
        row := rand.IntN(9)
        col := rand.IntN(9)
        if board[row][col] == 0 {
            return row, col
        }
    }
}

func cellsSeeEachOther(row1 int, col1 int, row2 int, col2 int) bool {
    return (row1 == row2 || col1 == col2 || (row1 / 3 == row2 / 3 && col1 / 3 == col2 / 3))
}

func numberIsComplete(game Sudoku, number uint8) bool {
    if number == 0 {
        return false
    }
    for i := 0; i < 9; i++ {
        for j := 0; j < 9; j++ {
            if game.solution[i][j] == number && game.board[i][j] != number {
                return false
            }
        }
    }
    return true
}

func printBoard(board [9][9]uint8) {
    leftPad := "   "
    hPad := " "
    vPad := "\n"
    cellWidth := 3 * (2 + 2 * len(hPad)) - 1
    builder := new(strings.Builder)
    builder.WriteString(leftPad + "|")
    builder.WriteString(strings.Repeat("-", cellWidth))
    builder.WriteString("|")
    builder.WriteString(strings.Repeat("-", cellWidth))
    builder.WriteString("|")
    builder.WriteString(strings.Repeat("-", cellWidth))
    builder.WriteString("|" + vPad)
    for i := 0; i < 9; i++ {
        for j := 0; j < 9; j++ {
            if j == 0 {
                builder.WriteString(leftPad + "|" + hPad)
            } else if j % 3 == 0 && j != 0 {
                builder.WriteString(hPad + "|" + hPad)
            } else {
                builder.WriteString(strings.Repeat(hPad, 2) + " ")
            }
            if board[i][j] == 0 {
                builder.WriteString(" ")
            } else {
                builder.WriteString(fmt.Sprintf("%d", board[i][j]))
            }
        }
        builder.WriteString(hPad + "|" + vPad + leftPad + "|")
        if i % 3 == 2 {
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

func selectRandomCandidate(candidates []uint8) (uint8, error) {
    if len(candidates) == 0 {
        return 0, fmt.Errorf("No candidates available")
    }
    return candidates[rand.IntN(len(candidates))], nil
}

func computeCandidates(game *Sudoku) {
    game.candidates = [9][9][9]bool{}
    for i := 0; i < 9; i++ {
        for j := 0; j < 9; j++ {
            for k := 0; k < 9; k++ {
                game.candidates[i][j][k] = true
            }
        }
    }
    for i := 0; i < 9; i++ {
        for j := 0; j < 9; j++ {
            if game.board[i][j] != 0 {
                updateCandidates(i, j, game.board[i][j], game, false)
            }
        }
    }
}

func toggleCandidate(row int, col int, candidate int, game *Sudoku) {
    game.candidates[row][col][candidate - 1] = !game.candidates[row][col][candidate - 1]
}

func updateCandidates(changedRow int, changedColumn int, insertedValue uint8, game *Sudoku, setTo bool) {
    for i := 0; i < 9; i++ {
        game.candidates[changedRow][i][insertedValue - 1] = setTo
        game.candidates[i][changedColumn][insertedValue - 1] = setTo
    }
    boxRowStart, boxColumnStart := getBoxStartsFromCell(changedRow, changedColumn)
    for i := boxRowStart; i < boxRowStart + 3; i++ {
        for j := boxColumnStart; j < boxColumnStart + 3; j++ {
            game.candidates[i][j][insertedValue - 1] = setTo
        }
    }
}

func wipeCandidates(game *Sudoku) {
    game.candidates = [9][9][9]bool{}
}

func runPrint() {
    sudoku := generateSudoku()
    println("Generated Sudoku:")
    printBoard(sudoku.board)
    println("Solution:")
    printBoard(sudoku.solution)
}

func main() {
    var (
        print = flag.Bool("print", false, "print a generated sudoku and its solution and exit")
    )
    flag.Usage = func() {
        fmt.Fprintf(flag.CommandLine.Output(),
            "Usage: sugoku [-print]\n")
        flag.PrintDefaults()
    }
    flag.Parse()

    if *print {
        runPrint()
    } else {
        runTui()
    }
}
