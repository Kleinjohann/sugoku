package main

import (
    "fmt"
    "os"

    "github.com/charmbracelet/bubbles/help"
    "github.com/charmbracelet/bubbles/key"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/charmbracelet/lipgloss/table"
)

type model struct {
    game     Sudoku
    editable [9][9]bool
    cursor   [2]int
    keys     keyMap
    help     help.Model
    tips     string
    width    int
}

type keyMap struct {
    Up                key.Binding
    Down              key.Binding
    Left              key.Binding
    Right             key.Binding
    Up3               key.Binding
    Down3             key.Binding
    Left3             key.Binding
    Right3            key.Binding
    Number            key.Binding
    Candidate         key.Binding
    Delete            key.Binding
    ComputeCandidates key.Binding
    WipeCandidates    key.Binding
    ToggleTips        key.Binding
    NewGame           key.Binding
    Quit              key.Binding
}

var keys = keyMap{
    Up: key.NewBinding(
        key.WithKeys("up", "k"),
        key.WithHelp("↑/k", "move up"),
    ),
    Down: key.NewBinding(
        key.WithKeys("down", "j"),
        key.WithHelp("↓/j", "move down"),
    ),
    Left: key.NewBinding(
        key.WithKeys("left", "h"),
        key.WithHelp("←/h", "move left"),
    ),
    Right: key.NewBinding(
        key.WithKeys("right", "l"),
        key.WithHelp("→/l", "move right"),
    ),
    Up3: key.NewBinding(
        key.WithKeys("shift+up", "K"),
        key.WithHelp("shift+↑/K", "move up 3 cells"),
    ),
    Down3: key.NewBinding(
        key.WithKeys("shift+down", "J"),
        key.WithHelp("shift+↓/J", "move down 3 cells"),
    ),
    Left3: key.NewBinding(
        key.WithKeys("shift+left", "H"),
        key.WithHelp("shift+←/H", "move left 3 cells"),
    ),
    Right3: key.NewBinding(
        key.WithKeys("shift+right", "L"),
        key.WithHelp("shift+→/L", "move right 3 cells"),
    ),
    Number: key.NewBinding(
        key.WithKeys("1", "2", "3", "4", "5", "6", "7", "8", "9"),
        key.WithHelp("1-9", "enter number"),
    ),
    Candidate: key.NewBinding(
        key.WithKeys("!", "@", "#", "$", "%", "^", "&", "*", "("),
        key.WithHelp("shift+1-9", "toggle pencil mark"),
    ),
    Delete: key.NewBinding(
        key.WithKeys("x", "bsp", "del"),
        key.WithHelp("x/bsp/del", "delete number/pencil marks"),
    ),
    ComputeCandidates: key.NewBinding(
        key.WithKeys("c"),
        key.WithHelp("c", "compute all pencil marks"),
    ),
    WipeCandidates: key.NewBinding(
        key.WithKeys("C"),
        key.WithHelp("C", "wipe all pencil marks"),
    ),
    ToggleTips: key.NewBinding(
        key.WithKeys("t"),
        key.WithHelp("t", "toggle tips"),
    ),
    NewGame: key.NewBinding(
        key.WithKeys("n"),
        key.WithHelp("n", "new game"),
    ),
    Quit: key.NewBinding(
        key.WithKeys("q", "esc", "ctrl+c"),
        key.WithHelp("q", "quit"),
    ),
}

func (k keyMap) ShortHelp() []key.Binding {
    return []key.Binding{k.NewGame, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
    return [][]key.Binding{
        {k.Up, k.Down, k.Left, k.Right,
         k.Up3, k.Down3, k.Left3, k.Right3,
         k.Number, k.Candidate, k.Delete,
         k.ComputeCandidates, k.WipeCandidates,
         k.ToggleTips, k.NewGame, k.Quit},
    }
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
    m := model{
        game:     game,
        editable: editable,
        cursor:   [2]int{4, 4},
        keys:     keys,
        help:     help.New(),
    }
    m.help.ShowAll = true
    return m
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case tea.WindowSizeMsg:
        m.width = msg.Width

    case tea.KeyMsg:

        switch {

        case key.Matches(msg, keys.Quit):
            fmt.Print("\n")
            return m, tea.Quit

        case key.Matches(msg, keys.NewGame):
            return initialModel(), nil

        case key.Matches(msg, keys.Up):
            m.cursor[0] = (m.cursor[0] - 1 + 9) % 9

        case key.Matches(msg, keys.Down):
            m.cursor[0] = (m.cursor[0] + 1) % 9

        case key.Matches(msg, keys.Left):
            m.cursor[1] = (m.cursor[1] - 1 + 9) % 9

        case key.Matches(msg, keys.Right):
            m.cursor[1] = (m.cursor[1] + 1) % 9

        case key.Matches(msg, keys.Up3):
            m.cursor[0] = (m.cursor[0] - 3 + 9) % 9

        case key.Matches(msg, keys.Down3):
            m.cursor[0] = (m.cursor[0] + 3) % 9

        case key.Matches(msg, keys.Left3):
            m.cursor[1] = (m.cursor[1] - 3 + 9) % 9

        case key.Matches(msg, keys.Right3):
            m.cursor[1] = (m.cursor[1] + 3) % 9

        case key.Matches(msg, keys.Number):
            if m.editable[m.cursor[0]][m.cursor[1]] {
                m.game.board[m.cursor[0]][m.cursor[1]] = uint8(msg.String()[0] - '0')
            }

        case key.Matches(msg, keys.Candidate):
            number := getNumberFromShiftedDigit(msg.String())
            if m.editable[m.cursor[0]][m.cursor[1]] {
                toggleCandidate(m.cursor[0], m.cursor[1], number, &m.game)
            }

        case key.Matches(msg, keys.Delete):
            if m.editable[m.cursor[0]][m.cursor[1]] {
                if m.game.board[m.cursor[0]][m.cursor[1]] == 0 {
                    m.game.candidates[m.cursor[0]][m.cursor[1]] = [9]bool{}
                } else {
                    m.game.board[m.cursor[0]][m.cursor[1]] = 0
                }
            }

        case key.Matches(msg, keys.ComputeCandidates):
            computeCandidates(&m.game)

        case key.Matches(msg, keys.WipeCandidates):
            wipeCandidates(&m.game)

        case key.Matches(msg, keys.ToggleTips):
            toggleTips(&m)
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
            boxId = 3*i + j
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
    renderedTable := t.Render()

    tableWidth := lipgloss.Width(renderedTable)
    tableHeight := lipgloss.Height(renderedTable)

    if isValidSolvedBoard(m.game.board) {
        m.help.ShowAll = false
        m.tips = ""
        winMessage := lipgloss.NewStyle().Foreground(completedNumberForeground).Render("You won!")
        renderedTable = lipgloss.Place(tableWidth, tableHeight, lipgloss.Center, lipgloss.Center, winMessage)
    }

    m.help.Width = m.width - tableWidth - 1
    helpView := m.help.View(m.keys)

    if len(m.tips) > 0 {
        updateTipsString(&m)
    }

    helpView = lipgloss.JoinVertical(lipgloss.Left, helpView, "\n", m.tips)

    return lipgloss.JoinHorizontal(lipgloss.Top,
        renderedTable,
        " ",
        helpView,
    )
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
            number = uint8(3*i + j + 1)
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
    for i := boxRowStart; i < boxRowStart+3; i++ {
        rowString = ""
        for j := boxColStart; j < boxColStart+3; j++ {
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

func updateTipsString(m *model) {
    game := copySudoku(m.game)
    computeCandidates(&game)
    m.tips = "Available hints:\n"
    for strategy := range solveStrategies {
        steps := solveStrategies[strategy](&game)
        for step := range steps {
            m.tips += fmt.Sprintf("%s: %s\n", steps[step].strategy, steps[step].description)
        }
    }
}

func toggleTips(m *model) {
    switch len(m.tips) {
    case 0:
        updateTipsString(m)
    default:
        m.tips = ""
    }
}

func runTui() {
    p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v", err)
        os.Exit(1)
    }
}
