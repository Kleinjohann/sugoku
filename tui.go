package main

import (
    "fmt"
    "os"
    "strings"

    "github.com/charmbracelet/lipgloss"
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
            return m, tea.Quit

        case "`":
            m.cursor[1] = 0

        case "$":
            m.cursor[1] = 8

        case "g":
            m.cursor[0] = 0

        case "G":
            m.cursor[0] = 8

        case "up", "k":
            if m.cursor[0] > 0 {
                m.cursor[0]--
            }

        case "down", "j":
            if m.cursor[0] < 8 {
                m.cursor[0]++
            }

        case "left", "h":
            if m.cursor[1] > 0 {
                m.cursor[1]--
            }

        case "right", "l":
            if m.cursor[1] < 8 {
                m.cursor[1]++
            }

        case "1", "2", "3", "4", "5", "6", "7", "8", "9":
            if m.editable[m.cursor[0]][m.cursor[1]] {
                m.game.board[m.cursor[0]][m.cursor[1]] = uint8(msg.String()[0] - '0')
            }

        case "x":
            if m.editable[m.cursor[0]][m.cursor[1]] {
                m.game.board[m.cursor[0]][m.cursor[1]] = 0
            }
        }
    }

    return m, nil
}

func (m model) View() string {
    var currentCell string
    var currentStyle lipgloss.Style
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
            currentStyle = lipgloss.NewStyle()
            if j == 0 {
                builder.WriteString(leftPad + "|" + hPad)
            } else if j % 3 == 0 && j != 0 {
                builder.WriteString(hPad + "|" + hPad)
            } else {
                builder.WriteString(strings.Repeat(hPad, 2) + " ")
            }
            if m.editable[i][j] {
                currentStyle = currentStyle.Foreground(editableForeground)
            } else {
                currentStyle = currentStyle.Foreground(uneditableForeground)
            }
            if m.game.board[i][j] == 0 {
                currentCell = " "
            } else {
                currentCell = fmt.Sprintf("%d", m.game.board[i][j])
                if m.game.board[i][j] == m.game.board[m.cursor[0]][m.cursor[1]] {
                    currentStyle = currentStyle.Background(cursorNumberBackground)
                }
            }
            if m.cursor[0] == i && m.cursor[1] == j {
                currentStyle = currentStyle.Background(cursorBackground)
            } else if cellsSeeEachOther(m.cursor[0], m.cursor[1], i, j) {
                currentStyle = currentStyle.Background(visibleFromCursorBackground)
            }
            if m.game.board[i][j] != 0 && m.game.board[i][j] != m.game.solution[i][j] {
                currentStyle = currentStyle.Foreground(wrongNumberForeground)
            } else if numberIsComplete(m.game, m.game.board[i][j]) {
                currentStyle = currentStyle.Foreground(completedNumberForeground)
            }
            builder.WriteString(currentStyle.Render(currentCell))
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
    return builder.String()
}

func runTui() {
    p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v", err)
        os.Exit(1)
    }
}
