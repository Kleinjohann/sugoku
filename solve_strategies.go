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

func getContextIdx(context Context, row int, col int) int {
    switch context {
    case Row:
        return row
    case Column:
        return col
    case Box:
        return getBoxIdFromCell(row, col)
    default:
        return 9*row + col
    }
}

func inSameBox(indices []int, context Context, contextIdx int) bool {
    var row, col, idx int
    var boxRowStart, boxColStart, boxRowEnd, boxColEnd int
    row, col = getCell(context, contextIdx, indices[0])
    boxRowStart, boxColStart = getBoxStartsFromCell(row, col)
    boxRowEnd, boxColEnd = boxRowStart+2, boxColStart+2
    for _, idx = range indices[1:] {
        row, col = getCell(context, contextIdx, idx)
        if row < boxRowStart || row > boxRowEnd ||
            col < boxColStart || col > boxColEnd {
            return false
        }
    }
    return true
}

func inSameContext(targetContext Context, indices []int, context Context, contextIdx int) bool {
    var idx, otherRow, otherCol int
    row, col := getCell(context, contextIdx, indices[0])
    switch targetContext {
    case Row:
        for _, idx = range indices[1:] {
            otherRow, _ = getCell(context, contextIdx, idx)
            if otherRow != row {
                return false
            }
        }
    case Column:
        for _, idx = range indices[1:] {
            _, otherCol = getCell(context, contextIdx, idx)
            if otherCol != col {
                return false
            }
        }
    case Box:
        boxId := getBoxIdFromCell(row, col)
        for _, idx = range indices[1:] {
            otherRow, otherCol = getCell(context, contextIdx, idx)
            if getBoxIdFromCell(otherRow, otherCol) != boxId {
                return false
            }
        }
    }
    return true
}

func resolveRowCol(context1 Context, contextIdx1 int, context2 Context, contextIdx2 int) (int, int) {
    if context1 == Row && context2 == Column {
        return contextIdx1, contextIdx2
    } else if context1 == Column && context2 == Row {
        return contextIdx2, contextIdx1
    }
    panic("Invalid contexts")
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

func getAllUniqueMapValues[keyType constraints.Integer, valueType constraints.Integer](m map[keyType][]valueType, keys []keyType) []valueType {
    var values []valueType
    for _, key := range keys {
        for _, value := range m[key] {
            if !slices.Contains(values, value) {
                values = append(values, value)
            }
        }
    }
    return values
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

func getContextPossibilitiesByCandidate(game *Sudoku, context Context, contextIdx int) map[uint8][]int {
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

func getCandidatePossibilitiesInContext(game *Sudoku, context Context, contextIdx int, candidate uint8) []int {
    var row, col int
    var possibilities []int
    switch context {
    case Cell:
        row, col = getCell(context, contextIdx, 0)
        if game.board[row][col] == 0 && game.candidates[row][col][candidate-1] {
            possibilities = []int{0}
        } else {
            possibilities = []int{}
        }
    default:
        for cellIdx := range 9 {
            row, col = getCell(context, contextIdx, cellIdx)
            if game.board[row][col] == 0 && game.candidates[row][col][candidate-1] {
                possibilities = append(possibilities, cellIdx)
            }
        }
    }
    return possibilities
}

func getCandidatePossibilitiesByContextIdx(game *Sudoku, context Context, candidate uint8) map[int][]int {
    possibilities := make(map[int][]int)
    if context == Cell {
        panic("Cannot get candidate possibilities by context index for cell context")
    }
    for contextIdx := range 9 {
        contextPossibilities := getCandidatePossibilitiesInContext(game, context, contextIdx, candidate)
        possibilities[contextIdx] = contextPossibilities
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
        } else if len(currentValues) == 0 {
            continue
        }
        currentSetIndices = []keyType{currentKey}
        otherKeyLoop:
        for _, otherKey := range keys[currentIdx+1:] {
            otherValues = candidates[otherKey]
            if len(otherValues) > setSize {
                continue
            } else if len(otherValues) == 0 {
                continue
            }
            for _, setIdx := range currentSetIndices {
                if !isSuperset(candidates[setIdx], otherValues) {
                    continue otherKeyLoop
                }
            }
            currentSetIndices = append(currentSetIndices, otherKey)
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

func nakedSet(game *Sudoku, setSize int, strategyName string) []SolutionStep {
    var steps []SolutionStep
    var description string
    var row, col int
    var setIndices, targetCells [][]int
    var targetValues, setCandidates []uint8
    var context Context
    var contextCandidates map[int][]uint8
    for _, context = range []Context{Row, Column, Box} {
        for contextIdx := range 9 {
            contextCandidates = getContextCandidates(game, context, contextIdx)
            setIndices = findSets(contextCandidates, setSize)
            for _, set := range setIndices {
                setCandidates = getAllUniqueMapValues(contextCandidates, set)
                targetCells = [][]int{}
                targetValues = []uint8{}
                for otherIdx, otherCandidates := range contextCandidates {
                    if slices.Contains(set, otherIdx) {
                        continue
                    }
                    row, col = getCell(context, contextIdx, otherIdx)
                    for _, candidate := range otherCandidates {
                        if slices.Contains(setCandidates, candidate) {
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
                contextStr := context.String()
                description = fmt.Sprintf("In %s %d, ", contextStr, contextIdx+1)
                for _, setCandidate := range setCandidates {
                    description += fmt.Sprintf("%d ", setCandidate)
                }
                description += "have to go in"
                for _, cellIdx := range set {
                    row, col = getCell(context, contextIdx, cellIdx)
                    description += fmt.Sprintf(" r%dc%d", row+1, col+1)
                }
                steps = append(steps, SolutionStep{
                    strategy:      strategyName,
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

func nakedPair(game *Sudoku) []SolutionStep {
    return nakedSet(game, 2, "Naked Pair")
}

func nakedTriple(game *Sudoku) []SolutionStep {
    return nakedSet(game, 3, "Naked Triple")
}

func nakedQuad(game *Sudoku) []SolutionStep {
    return nakedSet(game, 4, "Naked Quad")
}

func hiddenSet(game *Sudoku, setSize int, strategyName string) []SolutionStep {
    var steps []SolutionStep
    var description string
    var row, col, cellIdx int
    var candidate uint8
    var targetCells [][]int
    var setIndices []int
    var candidates, setCandidates, targetValues []uint8
    var sets [][]uint8
    var context Context
    var possibilities map[uint8][]int
    for _, context = range []Context{Row, Column, Box} {
        for contextIdx := range 9 {
            possibilities = getContextPossibilitiesByCandidate(game, context, contextIdx)
            sets = findSets(possibilities, setSize)
            for _, setCandidates = range sets {
                setIndices = getAllUniqueMapValues(possibilities, setCandidates)
                targetCells = [][]int{}
                targetValues = []uint8{}
                for _, cellIdx = range setIndices {
                    row, col = getCell(context, contextIdx, cellIdx)
                    candidates = getCandidates(game, row, col)
                    for _, candidate = range candidates {
                        if slices.Contains(setCandidates, candidate) {
                            continue
                        }
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
                contextStr := context.String()
                description = fmt.Sprintf("In %s %d, ", contextStr, contextIdx+1)
                for _, setCandidate := range setCandidates {
                    description += fmt.Sprintf("%d ", setCandidate)
                }
                description += "can only go in"
                for _, cellIdx := range setIndices {
                    row, col = getCell(context, contextIdx, cellIdx)
                    description += fmt.Sprintf(" r%dc%d", row+1, col+1)
                }
                steps = append(steps, SolutionStep{
                    strategy:      strategyName,
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
    return hiddenSet(game, 2, "Hidden Pair")
}

func hiddenTriple(game *Sudoku) []SolutionStep {
    return hiddenSet(game, 3, "Hidden Triple")
}

func hiddenQuad(game *Sudoku) []SolutionStep {
    return hiddenSet(game, 4, "Hidden Quad")
}

func pointingGroup(game *Sudoku) []SolutionStep {
    var steps []SolutionStep
    var description string
    var sourceRow, sourceCol, row, col, boxId, targetBoxId, contextIdx int
    var targetContext Context
    var possibilities map[uint8][]int
    var targetCells   [][]int
    var candidates, targetValues  []uint8
    for boxId = range 9 {
        possibilities = getContextPossibilitiesByCandidate(game, Box, boxId)
        candidates = maps.Keys(possibilities)
        for _, candidate := range candidates {
            numPossibilities := len(possibilities[candidate])
            if numPossibilities < 2 || numPossibilities > 3 {
                continue
            }
            for _, targetContext = range []Context{Row, Column} {
                if !inSameContext(targetContext, possibilities[candidate], Box, boxId) {
                    continue
                }
                sourceRow, sourceCol = getCell(Box, boxId, possibilities[candidate][0])
                contextIdx = getContextIdx(targetContext, sourceRow, sourceCol)
                targetCells = [][]int{}
                targetValues = []uint8{}
                for idx := range 9 {
                    row, col = getCell(targetContext, contextIdx, idx)
                    targetBoxId = getBoxIdFromCell(row, col)
                    if targetBoxId == boxId {
                        continue
                    } else if game.board[row][col] != 0 {
                        continue
                    } else if !game.candidates[row][col][candidate-1] {
                        continue
                    } else if isDuplicateEffect(steps, row, col, candidate) {
                        continue
                    }
                    targetCells = append(targetCells, []int{row, col})
                    targetValues = append(targetValues, candidate)
                }
                if len(targetCells) == 0 {
                    continue
                }
                description = fmt.Sprintf("In %s %d, %d has to be in box %d",
                    targetContext.String(),
                    contextIdx+1,
                    candidate,
                    boxId+1)
                steps = append(steps, SolutionStep{
                    strategy:      "Pointing Group",
                    description:   description,
                    sourceContext: Box,
                    sourceIndices: []int{boxId},
                    targetCells:   targetCells,
                    targetValues:  targetValues,
                    effectType:    RemoveCandidate,
                })
            }
        }
    }
    return steps
}

func boxReduction(game *Sudoku) []SolutionStep {
    var steps []SolutionStep
    var description string
    var row, col, boxId, boxRowStart, boxColStart, boxRowEnd, boxColEnd int
    var context Context
    var possibilities map[uint8][]int
    var targetCells   [][]int
    var candidates, targetValues  []uint8
    for _, context = range []Context{Row, Column} {
        for contextIdx := range 9 {
            possibilities = getContextPossibilitiesByCandidate(game, context, contextIdx)
            candidates = maps.Keys(possibilities)
            for _, candidate := range candidates {
                numPossibilities := len(possibilities[candidate])
                if numPossibilities < 2 || numPossibilities > 3 {
                    continue
                } else if !inSameBox(possibilities[candidate], context, contextIdx) {
                    continue
                }
                row, col = getCell(context, contextIdx, possibilities[candidate][0])
                boxId = getBoxIdFromCell(row, col)
                boxRowStart, boxColStart = getBoxStartsFromBoxId(boxId)
                boxRowEnd, boxColEnd = boxRowStart+2, boxColStart+2
                targetCells = [][]int{}
                targetValues = []uint8{}
                for row = boxRowStart; row <= boxRowEnd; row++ {
                    if context == Row && row == contextIdx {
                        continue
                    }
                    for col = boxColStart; col <= boxColEnd; col++ {
                        if context == Column && col == contextIdx {
                            continue
                        } else if game.board[row][col] != 0 {
                            continue
                        } else if !game.candidates[row][col][candidate-1] {
                            continue
                        } else if isDuplicateEffect(steps, row, col, candidate) {
                            continue
                        }
                        targetCells = append(targetCells, []int{row, col})
                        targetValues = append(targetValues, candidate)
                    }
                }
                if len(targetCells) == 0 {
                    continue
                }
                description = fmt.Sprintf("In box %d, %d has to be in %s %d",
                    boxId+1,
                    candidate,
                    context.String(),
                    contextIdx+1)
                steps = append(steps, SolutionStep{
                    strategy:      "Box Reduction",
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

func basicFish(game *Sudoku, fishSize int, strategyName string) []SolutionStep {
    var steps []SolutionStep
    var candidate uint8
    contexts := []Context{Row, Column}
    var description string
    var otherContext Context
    var otherContextIdx, cellIdx, row, col int
    var otherContextIndices []int
    var possibilities map[int][]int
    var sets, targetCells [][]int
    var targetValues []uint8
    for i, context := range contexts {
        otherContext = contexts[(i+1)%2]
        for candidate = 1; candidate <= 9; candidate++ {
            possibilities = getCandidatePossibilitiesByContextIdx(game, context, candidate)
            sets = findSets(possibilities, fishSize)
            for _, contextIndices := range sets {
                otherContextIndices = getAllUniqueMapValues(possibilities, contextIndices)
                targetCells = [][]int{}
                targetValues = []uint8{}
                for _, otherContextIdx = range otherContextIndices {
                    for cellIdx = range 9 {
                        row, col = resolveRowCol(context, cellIdx, otherContext, otherContextIdx)
                        if slices.Contains(contextIndices, cellIdx) {
                            continue
                        } else if game.board[row][col] != 0 {
                            continue
                        } else if !game.candidates[row][col][candidate-1] {
                            continue
                        } else if isDuplicateEffect(steps, row, col, candidate) {
                            continue
                        }
                        targetCells = append(targetCells, []int{row, col})
                        targetValues = append(targetValues, candidate)
                    }
                }
                if len(targetCells) == 0 {
                    continue
                }
                description = fmt.Sprintf("In %ss", otherContext.String())
                for _, otherContextIdx := range otherContextIndices {
                    description += fmt.Sprintf(" %d", otherContextIdx+1)
                }
                description += fmt.Sprintf(", %d has to be in %ss", candidate, context.String())
                for _, contextIdx := range contextIndices {
                    description += fmt.Sprintf(" %d", contextIdx+1)
                }
                steps = append(steps, SolutionStep{
                    strategy:      strategyName,
                    description:   description,
                    sourceContext: context,
                    sourceIndices: contextIndices,
                    targetCells:   targetCells,
                    targetValues:  targetValues,
                    effectType:    RemoveCandidate,
                })
            }
        }
    }
    return steps
}

func xWing(game *Sudoku) []SolutionStep {
    return basicFish(game, 2, "X-Wing")
}

func swordfish(game *Sudoku) []SolutionStep {
    return basicFish(game, 3, "Swordfish")
}

func jellyfish(game *Sudoku) []SolutionStep {
    return basicFish(game, 4, "Jellyfish")
}

func skyscraper(game *Sudoku) []SolutionStep {
    var steps []SolutionStep
    var candidate uint8
    contexts := []Context{Row, Column}
    var contextIndices, otherContextIndices []int
    var description string
    var otherContext Context
    var contextIdx, otherContextIdx, cellIdx, row, col int
    var otherPossibilities []int
    var possibilities map[int][]int
    var skyscraperTops, sourceCells, targetCells [][]int
    var targetValues []uint8
    for i, context := range contexts {
        otherContext = contexts[(i+1)%2]
        for candidate = 1; candidate <= 9; candidate++ {
            possibilities = getCandidatePossibilitiesByContextIdx(game, context, candidate)
            contextIndices = []int{}
            for contextIdx, otherContextIndices = range possibilities {
                if len(otherContextIndices) == 2 {
                    contextIndices = append(contextIndices, contextIdx)
                }
            }
            contextIdxLoop:
            for _, contextIdx := range contextIndices {
                otherContextIndices = possibilities[contextIdx]
                skyscraperTops = [][]int{}
                for _, otherContextIdx = range otherContextIndices {
                    otherPossibilities = getCandidatePossibilitiesInContext(game, otherContext, otherContextIdx, candidate)
                    if len(otherPossibilities) != 2 {
                        continue contextIdxLoop
                    }
                    for _, cellIdx = range otherPossibilities {
                        if cellIdx != contextIdx {
                            skyscraperTops = append(skyscraperTops, []int{cellIdx, otherContextIdx})
                        }
                    }
                }
                if len(skyscraperTops) < 2 {
                    continue
                } else if skyscraperTops[0][0] == skyscraperTops[1][0] {
                    continue
                }
                targetCells = [][]int{}
                targetValues = []uint8{}
                sourceCells = [][]int{}
                for _, skyscraperTop := range skyscraperTops {
                    row, col = resolveRowCol(context, skyscraperTop[0], otherContext, skyscraperTop[1])
                    sourceCells = append(sourceCells, []int{row, col})
                }
                for row = range 9 {
                    for col = range 9 {
                        if game.board[row][col] != 0 {
                            continue
                        } else if !game.candidates[row][col][candidate-1] {
                            continue
                        } else if !cellsSeeEachOther(row, col, sourceCells[0][0], sourceCells[0][1]) {
                            continue
                        } else if !cellsSeeEachOther(row, col, sourceCells[1][0], sourceCells[1][1]) {
                            continue
                        } else if row == sourceCells[0][0] && col == sourceCells[0][1] {
                            continue
                        } else if row == sourceCells[1][0] && col == sourceCells[1][1] {
                            continue
                        } else if isDuplicateEffect(steps, row, col, candidate) {
                            continue
                        }
                        targetCells = append(targetCells, []int{row, col})
                        targetValues = append(targetValues, candidate)
                    }
                }
                if len(targetCells) == 0 {
                    continue
                }
                description = fmt.Sprintf("Either r%dc%d or r%dc%d has to be %d",
                                          sourceCells[0][0]+1,
                                          sourceCells[0][1]+1,
                                          sourceCells[1][0]+1,
                                          sourceCells[1][1]+1,
                                          candidate)
                steps = append(steps, SolutionStep{
                    strategy:      "Skyscraper",
                    description:   description,
                    sourceContext: otherContext,
                    sourceIndices: otherContextIndices,
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
    nakedTriple,
    nakedQuad,
    pointingGroup,
    boxReduction,
    hiddenPair,
    hiddenTriple,
    hiddenQuad,
    xWing,
    swordfish,
    jellyfish,
    skyscraper,
    // yWing,
}

var strategyDifficulty = map[string]int{
    "Naked Single":   1,
    "Hidden Single":  1,
    "Naked Pair":     2,
    "Naked Triple":   2,
    "Naked Quad":     2,
    "Pointing Group": 2,
    "Box Reduction":  2,
    "Hidden Pair":    3,
    "Hidden Triple":  3,
    "Hidden Quad":    3,
    "X-Wing":         4,
    "Swordfish":      4,
    "Jellyfish":      4,
    "Skyscraper":     4,
    "Y-Wing":         4,
}

var maxDifficulty = 5

var validDifficulties = []int{0, 1, 2, 3, 4, maxDifficulty} // 0 for random difficulty

func rateDifficulty(game *Sudoku) int {
    gameCopy := *game
    var steps []SolutionStep
    var difficulty int
    for !isSolved(gameCopy.board) {
        for _, strategy := range solveStrategies {
            steps = strategy(&gameCopy)
            if len(steps) > 0 {
                break
            }
        }
        if len(steps) == 0 {
            return maxDifficulty
        }
        for _, step := range steps {
            step.Apply(&gameCopy)
            difficulty = max(difficulty, strategyDifficulty[step.strategy])
        }
    }
    return difficulty
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
