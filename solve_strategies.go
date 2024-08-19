package main

import (
    "fmt"
    "golang.org/x/exp/constraints"
    "golang.org/x/exp/maps"
    "slices"
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

func getContextCandidates(game *Sudoku, context Context, contextIdx int) map[int][]uint8 {
    var row, col int
    candidates := make(map[int][]uint8)
    switch context {
    case Cell:
        row, col = getCell(context, contextIdx, 0)
        if game.board[row][col] == 0 {
            candidates[0] = getCandidates(game, row, col)
        }
    default:
        for cellIdx := range 9 {
            row, col = getCell(context, contextIdx, cellIdx)
            if game.board[row][col] == 0 {
                candidates[cellIdx] = getCandidates(game, row, col)
            }
        }
    }
    return candidates
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

func isDuplicateEffect(steps []SolutionStep, row int, col int, value uint8) bool {
    for _, step := range steps {
        for i, targetCell := range step.targetCells {
            if targetCell[0] == row &&
                    targetCell[1] == col &&
                    step.targetValues[i] == value {
                return true
            }
        }
    }
    return false
}

func isSuperset[Int constraints.Integer](setA []Int, setB []Int) bool {
    if len(setA) < len(setB) {
        return isSuperset(setB, setA)
    }
    for _, item := range setB {
        if !slices.Contains(setA, item) {
            return false
        }
    }
    return true
}

func getContextCandidatePossibilities(game *Sudoku, context Context, contextIdx int) map[uint8][]int {
    var row, col int
    var candidates []uint8
    possibilities := make(map[uint8][]int)
    switch context {
    case Cell:
        row, col = getCell(context, contextIdx, 0)
        candidates = getCandidates(game, row, col)
        for _, candidate := range candidates {
            possibilities[candidate] = []int{0}
        }
    default:
        contextCandidates := getContextCandidates(game, context, contextIdx)
        for i, candidates := range contextCandidates {
            for _, candidate := range candidates {
                possibilities[candidate] = append(possibilities[candidate], i)
            }
        }
    }
    return possibilities
}

func findSets[keyType constraints.Integer, valueType constraints.Integer](candidates map[keyType][]valueType, setSize int) [][]keyType {
    var setIndices [][]keyType
    var currentSetIndices []keyType
    var currentValues, otherValues []valueType
    numCells := len(candidates)
    if numCells < setSize {
        return setIndices
    }
    keys := maps.Keys(candidates)
    for currentIdx, currentKey := range keys[:numCells-1] {
        currentValues = candidates[currentKey]
        if len(currentValues) > setSize {
            continue
        }
        currentSetIndices = []keyType{currentKey}
        for _, otherKey := range keys[currentIdx+1:] {
            otherValues = candidates[otherKey]
            if len(otherValues) > setSize {
                continue
            }
            if isSuperset(currentValues, otherValues) {
                currentSetIndices = append(currentSetIndices, otherKey)
            }
        }
        if len(currentSetIndices) == setSize {
            setIndices = append(setIndices, currentSetIndices)
        }
    }
    return setIndices
}

type SolveStrategy func(*Sudoku) []SolutionStep

func nakedSingle(game *Sudoku) []SolutionStep {
    var steps []SolutionStep
    var candidates []uint8
    var description string
    for i := 0; i < 9; i++ {
        for j := 0; j < 9; j++ {
            if game.board[i][j] == 0 && game.candidatesCount[i][j] == 1 {
                candidates = getCandidates(game, i, j)
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
                        if step.targetCells[0][0] == row &&
                            step.targetCells[0][1] == col &&
                            step.targetValues[0] == number {
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

func nakedPair(game *Sudoku) []SolutionStep {
    var steps []SolutionStep
    var description string
    var row, col int
    var pairIndices, targetCells [][]int
    var targetValues, pairCandidates []uint8
    var context Context
    var contextCandidates map[int][]uint8
    for _, context = range []Context{Row, Column, Box} {
        for contextIdx := range 9 {
            contextCandidates = getContextCandidates(game, context, contextIdx)
            pairIndices = findSets(contextCandidates, 2)
            for _, pair := range pairIndices {
                pairCandidates = contextCandidates[pair[0]]
                targetCells = [][]int{}
                targetValues = []uint8{}
                for otherIdx, otherCandidates := range contextCandidates {
                    if slices.Contains(pair, otherIdx) {
                        continue
                    }
                    // printBoard(game.board)
                    // fmt.Println(context.String(), contextIdx, pair, pairCandidates)
                    row, col = getCell(context, contextIdx, otherIdx)
                    for _, candidate := range otherCandidates {
                        if slices.Contains(pairCandidates, candidate) {
                            if isDuplicateEffect(steps, row, col, candidate) {
                                continue
                            }
                            targetCells = append(targetCells, []int{row, col})
                            targetValues = append(targetValues, candidate)
                        }
                    }
                }
                if len(targetCells) == 0 {
                    continue
                }
                row1, col1 := getCell(context, contextIdx, pair[0])
                row2, col2 := getCell(context, contextIdx, pair[1])
                contextStr := context.String()
                description = fmt.Sprintf("In %s %d, %d and %d have to go in r%dc%d and r%dc%d",
                    contextStr,
                    contextIdx+1,
                    pairCandidates[0],
                    pairCandidates[1],
                    row1+1,
                    col1+1,
                    row2+1,
                    col2+1)
                // fmt.Println(description)
                steps = append(steps, SolutionStep{
                    strategy:      "Naked Pair",
                    description:   description,
                    sourceContext: context,
                    sourceIndices: []int{contextIdx},
                    targetCells:   targetCells,
                    targetValues:  targetValues,
                    effectType:    RemoveCandidate,
                })
            }
        }
    }
    return steps
}

func hiddenPair(game *Sudoku) []SolutionStep {
    var steps []SolutionStep
    var description string
    var row, col, cellIdx int
    var candidate uint8
    var targetCells [][]int
    var pairIndices []int
    var candidates, pairCandidates, targetValues []uint8
    var pairs [][]uint8
    var context Context
    var possibilities map[uint8][]int
    for _, context = range []Context{Row, Column, Box} {
        for contextIdx := range 9 {
            possibilities = getContextCandidatePossibilities(game, context, contextIdx)
            pairs = findSets(possibilities, 2)
            for _, pairCandidates = range pairs {
                pairIndices = possibilities[pairCandidates[0]]
                targetCells = [][]int{}
                targetValues = []uint8{}
                for _, cellIdx = range pairIndices {
                    row, col = getCell(context, contextIdx, cellIdx)
                    candidates = getCandidates(game, row, col)
                    for _, candidate = range candidates {
                        if slices.Contains(pairCandidates, candidate) {
                            continue
                        }
                        // printBoard(game.board)
                        // fmt.Println(context.String(), contextIdx, pairIndices, pairCandidates)
                        if isDuplicateEffect(steps, row, col, candidate) {
                            continue
                        }
                        targetCells = append(targetCells, []int{row, col})
                        targetValues = append(targetValues, candidate)
                    }
                }
                if len(targetCells) == 0 {
                    continue
                }
                row1, col1 := getCell(context, contextIdx, pairIndices[0])
                row2, col2 := getCell(context, contextIdx, pairIndices[1])
                contextStr := context.String()
                description = fmt.Sprintf("In %s %d, %d and %d can only go in r%dc%d and r%dc%d",
                    contextStr,
                    contextIdx+1,
                    pairCandidates[0],
                    pairCandidates[1],
                    row1+1,
                    col1+1,
                    row2+1,
                    col2+1)
                // fmt.Println(description)
                steps = append(steps, SolutionStep{
                    strategy:      "Hidden Pair",
                    description:   description,
                    sourceContext: context,
                    sourceIndices: []int{contextIdx},
                    targetCells:   targetCells,
                    targetValues:  targetValues,
                    effectType:    RemoveCandidate,
                })
            }
        }
    }
    return steps
}

var solveStrategies = []SolveStrategy{
    nakedSingle,
    hiddenSingle,
    nakedPair,
    hiddenPair,
    // nakedTriple,
    // hiddenTriple,
    // nakedQuad,
    // hiddenQuad,
    // pointingGroup,
    // boxReduction,
    // xWing,
    // skyscraper,
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
