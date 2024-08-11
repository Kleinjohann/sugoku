package main

import (
    "fmt"
    "os"

    "github.com/charmbracelet/lipgloss"
    "github.com/charmbracelet/lipgloss/table"
    tea "github.com/charmbracelet/bubbletea"
)

type model struct {
    game  Sudoku
    editable [9][9]bool
    cursor   [2]int
}

var cursorBackground = lipgloss.Color("3")
var visibleFromCursorBackground = lipgloss.Color("18")
var cursorNumberBackground = lipgloss.Color("8")
var cursorCandidatesForeground = lipgloss.Color("0")
var wrongNumberForeground = lipgloss.Color("1")
var completedNumberForeground = lipgloss.Color("2")
var editableForeground = lipgloss.Color("4")
var uneditableForeground = lipgloss.Color("15")

func initialModel() model {
    game := generateSudoku()
    editable := [9][9]bool{}
    for i := 0; i < 9; i++ {
        for j := 0; j < 9; j++ {
            if game.board[i][j] == 0 {
                editable[i][j] = true
            }
        }
    }
    game.candidates = [9][9][9]bool{}
    return model{
        game: game,
        editable: editable,
        cursor: [2]int{4, 4},
    }
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case tea.KeyMsg:

        switch msg.String() {

        case "ctrl+c", "q":
            fmt.Print("\n")
            return m, tea.Quit

        case "up", "k":
            m.cursor[0] = (m.cursor[0] - 1 + 9) % 9

        case "down", "j":
            m.cursor[0] = (m.cursor[0] + 1) % 9

        case "left", "h":
            m.cursor[1] = (m.cursor[1] - 1 + 9) % 9

        case "right", "l":
            m.cursor[1] = (m.cursor[1] + 1) % 9

        case "shift+up", "K":
            m.cursor[0] = (m.cursor[0] - 3 + 9) % 9

        case "shift+down", "J":
            m.cursor[0] = (m.cursor[0] + 3) % 9

        case "shift+left", "H":
            m.cursor[1] = (m.cursor[1] - 3 + 9) % 9

        case "shift+right", "L":
            m.cursor[1] = (m.cursor[1] + 3) % 9

        case "1", "2", "3", "4", "5", "6", "7", "8", "9":
            if m.editable[m.cursor[0]][m.cursor[1]] {
                m.game.board[m.cursor[0]][m.cursor[1]] = uint8(msg.String()[0] - '0')
            }

        case "!", "@", "#", "$", "%", "^", "&", "*", "(":
            number := getNumberFromShiftedDigit(msg.String())
            if m.editable[m.cursor[0]][m.cursor[1]] {
                toggleCandidate(m.cursor[0], m.cursor[1], number, &m.game)
            }

        case "x", "backspace", "delete":
            if m.editable[m.cursor[0]][m.cursor[1]] {
                if m.game.board[m.cursor[0]][m.cursor[1]] == 0 {
                    m.game.candidates[m.cursor[0]][m.cursor[1]] = [9]bool{}
                } else {
                    m.game.board[m.cursor[0]][m.cursor[1]] = 0
                }
            }

        case "c":
            computeCandidates(&m.game)

        case "C":
            wipeCandidates(&m.game)
        }
    }

    return m, nil
}

func (m model) View() string {
    rows := [][]string{}
    var boxId int
    for i := 0; i < 3; i++ {
        row := []string{}
        for j := 0; j < 3; j++ {
            boxId = 3 * i + j
            box := getBoxString(boxId, m, pagga, 3, 7)
            row = append(row, box)
        }
        rows = append(rows, row)
    }
    t := table.New().
        Border(lipgloss.NormalBorder()).
        BorderStyle(lipgloss.NewStyle().Foreground(uneditableForeground)).
        BorderRow(true).
        Rows(rows...)

    return t.String()
}

func getCellStyle(m model, row int, col int) lipgloss.Style {
    var foreground lipgloss.Color
    var background lipgloss.Color
    number := m.game.board[row][col]
    cursorRow := m.cursor[0]
    cursorCol := m.cursor[1]
    cursorNumber := m.game.board[cursorRow][cursorCol]
    if cellsSeeEachOther(row, col, cursorRow, cursorCol) {
        background = visibleFromCursorBackground
    }
    if number != 0 && number == cursorNumber {
        background = cursorNumberBackground
    }
    if cursorRow == row && cursorCol == col {
        background = cursorBackground
        foreground = cursorCandidatesForeground
    }
    if !m.editable[row][col] {
        foreground = uneditableForeground
    }
    if number > 0 && m.editable[row][col] {
        foreground = editableForeground
    }
    if numberIsComplete(m.game, number) {
        foreground = completedNumberForeground
    }
    if number != 0 && number != m.game.solution[row][col] {
        foreground = wrongNumberForeground
    }
    return lipgloss.NewStyle().Foreground(foreground).Background(background)
}

func getCellString(game Sudoku, row int, col int, font asciiFont, height int, width int) string {
    var digitString string
    var background string
    digit := game.board[row][col]
    if digit != 0 {
        digitString = font.numbers[int(digit)]
        background = font.background
    } else {
        candidates := getCandidates(&game, row, col)
        digitString = getCandidatesString(candidates)
        background = " "
    }
    cellString := lipgloss.Place(width,
                                 height,
                                 lipgloss.Center,
                                 lipgloss.Center,
                                 digitString,
                                 lipgloss.WithWhitespaceChars(background))
    return cellString
}

func getCandidatesString(candidates []uint8) string {
    var cellString string
    var rowString string
    var rowStrings []string
    var number uint8
    for i := 0; i < 3; i++ {
        rowString = ""
        for j := 0; j < 3; j++ {
            number = uint8(3 * i + j + 1)
            if len(rowString) > 0 {
                rowString += " "
            }
            if contains(candidates, number) {
                rowString += fmt.Sprintf("%d", number)
            } else {
                rowString += " "
            }
        }
        rowStrings = append(rowStrings, rowString)
    }
    cellString = lipgloss.JoinVertical(lipgloss.Left, rowStrings...)
    return cellString
}

func contains(s []uint8, e uint8) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

func getBoxString(boxId int, m model, font asciiFont, height int, width int) string {
    boxRowStart, boxColStart := getBoxStartsFromBoxId(boxId)
    var boxString string
    var cellString string
    var cellStyle lipgloss.Style
    var rowString string
    var rowStrings []string
    for i := boxRowStart; i < boxRowStart + 3; i++ {
        rowString = ""
        for j := boxColStart; j < boxColStart + 3; j++ {
            cellString = getCellString(m.game,
                                       i,
                                       j,
                                       font,
                                       height,
                                       width)
            cellStyle = getCellStyle(m, i, j).SetString(cellString)
            rowString = lipgloss.JoinHorizontal(lipgloss.Top,
                                                rowString,
                                                cellStyle.String())
        }
        rowStrings = append(rowStrings, rowString)
    }
    boxString = lipgloss.JoinVertical(lipgloss.Left, rowStrings...)
    return boxString
}

func getNumberFromShiftedDigit(digit string) int {
    switch digit {
        case "!":
            return 1
        case "@":
            return 2
        case "#":
            return 3
        case "$":
            return 4
        case "%":
            return 5
        case "^":
            return 6
        case "&":
            return 7
        case "*":
            return 8
        case "(":
            return 9
        default:
            return 0
    }
}

func runTui() {
    p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v", err)
        os.Exit(1)
    }
}
