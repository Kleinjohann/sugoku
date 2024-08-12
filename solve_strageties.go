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

type Effect int
const (
    PlaceNumber Effect = iota
    RemoveCandidate
)

type SolutionStep struct {
    strategy string
    description string
    sourceContext Context
    sourceIndices []int
    targetCells [][]int
    targetValues []uint8
    effectType Effect
}

func (step SolutionStep) Apply(game *Sudoku) {
    switch step.effectType {
    case PlaceNumber:
        for i, cell := range step.targetCells {
            game.board[cell[0]][cell[1]] = step.targetValues[i]
        }
    case RemoveCandidate:
        for i, cell := range step.targetCells {
            game.candidates[cell[0]][cell[1]][step.targetValues[i] - 1] = false
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
                        strategy: "Naked Single",
                        description: description,
                        sourceContext: Cell,
                        sourceIndices: []int{9*i + j},
                        targetCells: [][]int{{i, j}},
                        targetValues: []uint8{candidates[0]},
                        effectType: PlaceNumber,
                    })
                }
            }
        }
    }
    return steps
}

var solveStrategies = []SolveStrategy{
    nakedSingle,
    // hiddenSingle,
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
