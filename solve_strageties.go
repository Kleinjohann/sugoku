package main

import (
    "fmt"
)

type Context int

const (
    Row Context = iota
    Column
    Box
    Cell
)

func (context Context) String() string {
    switch context {
    case Row:
        return "Row"
    case Column:
        return "Column"
    case Box:
        return "Box"
    case Cell:
        return "Cell"
    }
    return "Unknown"
}

func getCell(context Context, contextIdx int, cellIdx int) (int, int) {
    switch context {
    case Row:
        return contextIdx, cellIdx
    case Column:
        return cellIdx, contextIdx
    case Box:
        boxRowStart, boxColStart := getBoxStartsFromBoxId(contextIdx)
        return boxRowStart + cellIdx/3, boxColStart + cellIdx%3
    case Cell:
        return contextIdx / 9, contextIdx % 9
    }
    return -1, -1
}

type Effect int

const (
    PlaceNumber Effect = iota
    RemoveCandidate
)

type SolutionStep struct {
    strategy      string
    description   string
    sourceContext Context
    sourceIndices []int
    targetCells   [][]int
    targetValues  []uint8
    effectType    Effect
}

func (step SolutionStep) Apply(game *Sudoku) {
    switch step.effectType {
    case PlaceNumber:
        for i, cell := range step.targetCells {
            game.board[cell[0]][cell[1]] = step.targetValues[i]
            updateCandidates(cell[0], cell[1], step.targetValues[i], game)
        }
    case RemoveCandidate:
        for i, cell := range step.targetCells {
            game.candidates[cell[0]][cell[1]][step.targetValues[i]-1] = false
            game.candidatesCount[cell[0]][cell[1]]--
        }
    }
}

type SolveStrategy func(*Sudoku) []SolutionStep

func nakedSingle(game *Sudoku) []SolutionStep {
    var steps []SolutionStep
    var candidates []uint8
    var description string
    for i := 0; i < 9; i++ {
        for j := 0; j < 9; j++ {
            if game.board[i][j] == 0 {
                candidates = getCandidates(game, i, j)
                if len(candidates) == 1 {
                    description = fmt.Sprintf("r%dc%d can only be %d", i+1, j+1, candidates[0])
                    steps = append(steps, SolutionStep{
                        strategy:      "Naked Single",
                        description:   description,
                        sourceContext: Cell,
                        sourceIndices: []int{9*i + j},
                        targetCells:   [][]int{{i, j}},
                        targetValues:  []uint8{candidates[0]},
                        effectType:    PlaceNumber,
                    })
                }
            }
        }
    }
    return steps
}

func hiddenSingle(game *Sudoku) []SolutionStep {
    var steps []SolutionStep
    var description string
    var row, col int
    var count, lastIdx int
    var context Context
    for _, context = range []Context{Row, Column, Box} {
        for contextIdx := range 9 {
            candidateLoop:
            for candidateIdx := range 9 {
                count = 0
                for cell_idx := 0; cell_idx < 9; cell_idx++ {
                    row, col = getCell(context, contextIdx, cell_idx)
                    if game.board[row][col] == 0 && game.candidates[row][col][candidateIdx] {
                        count++
                        lastIdx = cell_idx
                    }
                }
                if count == 1 {
                    contextStr := context.String()
                    row, col = getCell(context, contextIdx, lastIdx)
                    number := uint8(candidateIdx + 1)

                    for _, step := range steps {
                        // avoid steps with duplicate effects
                        if (step.targetCells[0][0] == row &&
                            step.targetCells[0][1] == col &&
                            step.targetValues[0] == number) {
                            continue candidateLoop
                        }
                    }
                    description = fmt.Sprintf("%d can only go in r%dc%d in %s %d",
                                              number,
                                              row+1,
                                              col+1,
                                              contextStr,
                                              contextIdx+1)
                    steps = append(steps, SolutionStep{
                        strategy:      "Hidden Single",
                        description:   description,
                        sourceContext: context,
                        sourceIndices: []int{contextIdx},
                        targetCells:   [][]int{{row, col}},
                        targetValues:  []uint8{number},
                        effectType:    PlaceNumber,
                    })
                }
            }
        }
    }
    return steps
}

var solveStrategies = []SolveStrategy{
    nakedSingle,
    hiddenSingle,
    // nakedPair,
    // hiddenPair,
    // nakedTriple,
    // hiddenTriple,
    // nakedQuad,
    // hiddenQuad,
    // pointingGroup,
    // xWing,
    // swordfish,
    // yWing,
}

func solvableUsingStrategies(game *Sudoku, strategies []SolveStrategy) bool {
    gameCopy := *game
    var steps []SolutionStep
    for !isSolved(gameCopy.board) {
        for _, strategy := range strategies {
            steps = strategy(&gameCopy)
            if len(steps) > 0 {
                break
            }
        }
        if len(steps) == 0 {
            return false
        }
        for _, step := range steps {
            step.Apply(&gameCopy)
        }
    }
    return true
}
