package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type State int

const (
	MainMenu State = iota
	PickDifficulty
	PickStrategy
	PickFileToLoad
	WaitScreen
	Playing
	PickFileToSave
	WinScreen
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

var DefaultMenuItems = []list.Item{
	item{title: "New Game", desc: "Generate a new puzzle, you can choose the difficulty"},
	item{title: "New Exercise", desc: "Generate a puzzle which requires a specific strategy for the next deduction"},
	item{title: "Load", desc: "Load a puzzle from a file"},
	item{title: "Quit", desc: "Exit this program"},
}

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

type model struct {
	err                 error
	state               State
	mainMenuModel       list.Model
	difficultyMenuModel list.Model
	strategyMenuModel   list.Model
	filepickerModel     filepicker.Model
	textinputModel      textinput.Model
	game                Sudoku
	tipsGame            Sudoku
	editable            [9][9]bool
	difficulty          int
	cursor              [2]int
	keys                keyMap
	help                help.Model
	strategies          []SolutionStep
	tips                string
	width               int
	boardHeight         int
	boardWidth          int
	cores               int
}

type keyMap struct {
	Enter             key.Binding
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
	ApplyTips         key.Binding
	SaveGame          key.Binding
	QuitToMenu        key.Binding
	Quit              key.Binding
	Abort             key.Binding
}

var keys = keyMap{
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "choose selected option"),
	),
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
	ApplyTips: key.NewBinding(
		key.WithKeys("T"),
		key.WithHelp("T", "apply tips"),
	),
	SaveGame: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "save game"),
	),
	QuitToMenu: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit to menu"),
	),
	Quit: key.NewBinding(
		key.WithKeys("Q", "esc", "ctrl+c"),
		key.WithHelp("Q", "quit"),
	),
	Abort: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "abort"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.QuitToMenu, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Up, k.Down, k.Left, k.Right,
			k.Up3, k.Down3, k.Left3, k.Right3,
			k.Number, k.Candidate, k.Delete,
			k.ComputeCandidates, k.WipeCandidates,
			k.ToggleTips, k.ApplyTips, k.SaveGame,
			k.QuitToMenu, k.Quit,
		},
	}
}

var cursorBackground = lipgloss.Color("3")
var visibleFromCursorBackground = lipgloss.Color("18")
var cursorNumberBackground = lipgloss.Color("6")
var cursorNumberForeground = lipgloss.Color("7")
var cursorCandidatesForeground = lipgloss.Color("0")
var wrongNumberForeground = lipgloss.Color("1")
var completedNumberForeground = lipgloss.Color("2")
var editableForeground = lipgloss.Color("4")
var uneditableForeground = lipgloss.Color("15")

func initialModel() model {
	boardHeight := 31
	boardWidth := 0
	game := makeEmptySudoku()
	var editable [9][9]bool
	tipsGame := game
	game.candidates = [9][9][9]bool{}
	strategyDifficulties := getStrategyDifficulties()
	var difficultyMenuItems []list.Item
	difficultyMenuItems = append(difficultyMenuItems, item{title: "0", desc: "Pick a random difficulty"})
	difficultyMenuItems = append(difficultyMenuItems, item{title: "6", desc: "Not solvable using all of the above strategies"})
	for difficulty, strategies := range strategyDifficulties {
		menuItem := item{
			title: strconv.Itoa(difficulty),
			desc:  strings.Join(strategies, ", "),
		}
		difficultyMenuItems = append(difficultyMenuItems, menuItem)
	}
	slices.SortFunc(difficultyMenuItems, func(i, j list.Item) int {
		return int(i.FilterValue()[0]) - int(j.FilterValue()[0])
	})
	var strategyMenuItems []list.Item
	for _, strategy := range SolveStrategies {
		menuItem := item{
			title: strategy.name,
			desc: fmt.Sprintf(
				"Difficulty: %d; Effect Type: %s",
				strategy.difficulty,
				strategy.effectType.String()),
		}
		strategyMenuItems = append(strategyMenuItems, menuItem)
	}
	fp := filepicker.New()
	fp.AllowedTypes = []string{".csv"}
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fp.CurrentDirectory = path.Join(wd, "archive")
	fp.Height = boardHeight
	ti := textinput.New()
	m := model{
		mainMenuModel: list.New(
			DefaultMenuItems,
			list.NewDefaultDelegate(),
			boardWidth,
			boardHeight),
		difficultyMenuModel: list.New(
			difficultyMenuItems,
			list.NewDefaultDelegate(),
			boardWidth,
			boardHeight),
		strategyMenuModel: list.New(
			strategyMenuItems,
			list.NewDefaultDelegate(),
			boardWidth,
			boardHeight),
		filepickerModel: fp,
		textinputModel:  ti,
		game:            game,
		tipsGame:        tipsGame,
		editable:        editable,
		difficulty:      0,
		cursor:          [2]int{4, 4},
		keys:            keys,
		help:            help.New(),
		cores:           -1,
		boardHeight:     boardHeight,
		boardWidth:      boardWidth,
	}
	m.mainMenuModel.Title = "Main Menu"
	m.mainMenuModel.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{keys.Enter}
	}
	m.mainMenuModel.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{keys.Enter}
	}
	m.difficultyMenuModel.Title = "Choose a Difficulty"
	m.difficultyMenuModel.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{keys.Enter}
	}
	m.difficultyMenuModel.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{keys.Enter}
	}
	m.strategyMenuModel.Title = "Choose a Strategy"
	m.strategyMenuModel.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{keys.Enter}
	}
	m.strategyMenuModel.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{keys.Enter}
	}
	m.help.ShowAll = true
	return m
}

func newGame(m *model) {
	game := generateSudokuParallel(m.difficulty, -1, m.cores)
	m.editable = [9][9]bool{}
	for i := range 9 {
		for j := range 9 {
			if game.board[i][j] == 0 {
				m.editable[i][j] = true
			}
		}
	}
	tipsGame := game
	m.game = game
	m.game.candidates = [9][9][9]bool{}
	m.tipsGame = tipsGame
	m.cursor = [2]int{4, 4}
	m.help = help.New()
	m.help.ShowAll = true
	m.tips = ""
	m.state = Playing
}

func newExercise(m *model, strategyName string) {
	exercise := generateStrategyExerciseParallel(strategyName, -1, m.cores)
	game := exercise.game
	firstIdx := exercise.indices[0]
	computeCandidates(game)
	m.editable = [9][9]bool{}
	for i := range 9 {
		for j := range 9 {
			if game.board[i][j] == 0 {
				m.editable[i][j] = true
			}
		}
	}
	for _, step := range exercise.steps[:firstIdx] {
		step.Apply(game)
	}
	tipsGame := game
	m.game = *game
	m.tipsGame = *tipsGame
	m.game.candidates = [9][9][9]bool{}
	m.cursor = [2]int{4, 4}
	m.help = help.New()
	m.help.ShowAll = true
	updateTipsString(m)
	m.state = Playing
}

func (m model) Init() tea.Cmd {
	return m.filepickerModel.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.mainMenuModel.SetSize(msg.Width, m.boardHeight)
		m.difficultyMenuModel.SetSize(msg.Width, m.boardHeight)
		m.strategyMenuModel.SetSize(msg.Width, m.boardHeight)
		m.width = msg.Width
		return m, nil

	case clearErrorMsg:
		m.err = nil

	case tea.KeyMsg:
		// will be handled depending on state below

	default:
		switch m.state {
		case MainMenu:
			m.mainMenuModel, cmd = m.mainMenuModel.Update(msg)
			return m, cmd
		case PickDifficulty:
			m.difficultyMenuModel, cmd = m.difficultyMenuModel.Update(msg)
			return m, cmd
		case PickStrategy:
			m.strategyMenuModel, cmd = m.strategyMenuModel.Update(msg)
			return m, cmd
		case PickFileToLoad:
			m.filepickerModel, cmd = m.filepickerModel.Update(msg)
			return m, cmd
		case PickFileToSave:
			m.textinputModel, cmd = m.textinputModel.Update(msg)
			return m, cmd
		default:
			return m, nil
		}
	}

	keyMsg := msg.(tea.KeyMsg)

	switch m.state {

	case WinScreen:
		switch {

		case key.Matches(keyMsg, keys.Quit):
			fmt.Print("\n")
			return m, tea.Quit

		default:
			m.state = MainMenu
			return m, nil
		}

	case MainMenu:
		switch {

		case key.Matches(keyMsg, keys.Quit):
			if m.mainMenuModel.FilterState() != list.Unfiltered {
				m.mainMenuModel, cmd = m.mainMenuModel.Update(msg)
				return m, cmd
			}
			fmt.Print("\n")
			return m, tea.Quit

		case key.Matches(keyMsg, keys.Enter):
			if m.mainMenuModel.FilterState() == list.Filtering {
				m.mainMenuModel, cmd = m.mainMenuModel.Update(msg)
				return m, cmd
			}

			i, ok := m.mainMenuModel.SelectedItem().(item)
			if !ok {
				return m, nil
			}
			switch i.title {
			case "New Game":
				m.state = PickDifficulty
				return m, nil
			case "New Exercise":
				m.state = PickStrategy
				return m, nil
			case "Load":
				m.state = PickFileToLoad
				return m, nil
			case "Quit":
				return m, tea.Quit
			}

		default:
			m.mainMenuModel, cmd = m.mainMenuModel.Update(msg)
			return m, cmd
		}

	case PickDifficulty:
		switch {

		case key.Matches(keyMsg, keys.Quit):
			if m.difficultyMenuModel.FilterState() != list.Unfiltered {
				m.difficultyMenuModel, cmd = m.difficultyMenuModel.Update(msg)
				return m, cmd
			}
			fmt.Print("\n")
			return m, tea.Quit

		case key.Matches(keyMsg, keys.Enter):
			if m.difficultyMenuModel.FilterState() == list.Filtering {
				m.difficultyMenuModel, cmd = m.difficultyMenuModel.Update(msg)
				return m, cmd
			}

			i, ok := m.difficultyMenuModel.SelectedItem().(item)
			if !ok {
				return m, nil
			}
			difficulty64, err := strconv.ParseInt(i.title, 10, 64)
			if err != nil {
				panic(err)
			}
			m.difficulty = int(difficulty64)
			newGame(&m)
			return m, nil

		default:
			m.difficultyMenuModel, cmd = m.difficultyMenuModel.Update(msg)
			return m, cmd
		}

	case PickStrategy:
		switch {

		case key.Matches(keyMsg, keys.Quit):
			if m.strategyMenuModel.FilterState() != list.Unfiltered {
				m.strategyMenuModel, cmd = m.strategyMenuModel.Update(msg)
				return m, cmd
			}
			fmt.Print("\n")
			return m, tea.Quit

		case key.Matches(keyMsg, keys.Enter):
			if m.strategyMenuModel.FilterState() == list.Filtering {
				m.strategyMenuModel, cmd = m.strategyMenuModel.Update(msg)
				return m, cmd
			}

			i, ok := m.strategyMenuModel.SelectedItem().(item)
			if !ok {
				return m, nil
			}
			newExercise(&m, i.title)
			return m, nil

		default:
			m.strategyMenuModel, cmd = m.strategyMenuModel.Update(msg)
			return m, cmd
		}

	case PickFileToLoad:
		switch {
		case key.Matches(keyMsg, keys.Quit):
			fmt.Print("\n")
			return m, tea.Quit

		case key.Matches(keyMsg, keys.QuitToMenu):
			m.state = MainMenu
			return m, nil

		default:
			m.filepickerModel, cmd = m.filepickerModel.Update(msg)

			// Did the user select a file?
			if didSelect, path := m.filepickerModel.DidSelectFile(msg); didSelect {
				m.game, m.editable = loadSudoku(path)
				tipsGame := m.game
				m.game.candidates = [9][9][9]bool{}
				m.tipsGame = tipsGame
				m.cursor = [2]int{4, 4}
				m.help = help.New()
				m.help.ShowAll = true
				m.tips = ""
				m.state = Playing
				return m, nil
			}

			// Did the user select a disabled file?
			// This is only necessary to display an error to the user.
			if didSelect, path := m.filepickerModel.DidSelectDisabledFile(msg); didSelect {
				// Let's clear the selectedFile and display an error.
				m.err = errors.New(path + " is not valid.")
				return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
			}

			return m, cmd
		}

	case PickFileToSave:
		switch {
		case key.Matches(keyMsg, keys.Enter):
			saveSudoku(
				fmt.Sprintf("archive/%s.csv", m.textinputModel.Value()),
				m.game,
				m.editable)
			m.state = Playing
			return m, nil

		case key.Matches(keyMsg, keys.Abort):
			m.state = Playing
			return m, nil

		default:
			m.textinputModel, cmd = m.textinputModel.Update(msg)
			return m, cmd
		}

	case Playing:
		switch {

		case key.Matches(keyMsg, keys.Quit):
			fmt.Print("\n")
			return m, tea.Quit

		case key.Matches(keyMsg, keys.QuitToMenu):
			m.state = MainMenu

		case key.Matches(keyMsg, keys.SaveGame):
			m.state = PickFileToSave
			currentTime := time.Now()
			m.textinputModel.Placeholder = fmt.Sprintf("%d%02d%02d-%02d%02d-%02d",
				currentTime.Year(),
				currentTime.Month(),
				currentTime.Day(),
				currentTime.Hour(),
				currentTime.Minute(),
				currentTime.Second())
			m.textinputModel.Focus()
			m.textinputModel.CharLimit = 16
			m.textinputModel.Width = 16
			m.textinputModel.Prompt = ""
			return m, nil

		case key.Matches(keyMsg, keys.Up):
			m.cursor[0] = (m.cursor[0] - 1 + 9) % 9

		case key.Matches(keyMsg, keys.Down):
			m.cursor[0] = (m.cursor[0] + 1) % 9

		case key.Matches(keyMsg, keys.Left):
			m.cursor[1] = (m.cursor[1] - 1 + 9) % 9

		case key.Matches(keyMsg, keys.Right):
			m.cursor[1] = (m.cursor[1] + 1) % 9

		case key.Matches(keyMsg, keys.Up3):
			m.cursor[0] = (m.cursor[0] - 3 + 9) % 9

		case key.Matches(keyMsg, keys.Down3):
			m.cursor[0] = (m.cursor[0] + 3) % 9

		case key.Matches(keyMsg, keys.Left3):
			m.cursor[1] = (m.cursor[1] - 3 + 9) % 9

		case key.Matches(keyMsg, keys.Right3):
			m.cursor[1] = (m.cursor[1] + 3) % 9

		case key.Matches(keyMsg, keys.Number):
			if m.editable[m.cursor[0]][m.cursor[1]] {
				number := uint8(keyMsg.String()[0] - '0')
				m.game.board[m.cursor[0]][m.cursor[1]] = number
				if m.game.solution[m.cursor[0]][m.cursor[1]] == number {
					m.tipsGame.board[m.cursor[0]][m.cursor[1]] = number
					computeCandidates(&m.tipsGame)
				}
			}

		case key.Matches(keyMsg, keys.Candidate):
			number := getNumberFromShiftedDigit(keyMsg.String())
			if m.editable[m.cursor[0]][m.cursor[1]] {
				toggleCandidate(m.cursor[0], m.cursor[1], number, &m.game)
			}

		case key.Matches(keyMsg, keys.Delete):
			if m.editable[m.cursor[0]][m.cursor[1]] {
				if m.game.board[m.cursor[0]][m.cursor[1]] == 0 {
					m.game.candidates[m.cursor[0]][m.cursor[1]] = [9]bool{}
				} else {
					m.game.board[m.cursor[0]][m.cursor[1]] = 0
				}
			}

		case key.Matches(keyMsg, keys.ComputeCandidates):
			m.game.candidates = m.tipsGame.candidates

		case key.Matches(keyMsg, keys.WipeCandidates):
			wipeCandidates(&m.game)

		case key.Matches(keyMsg, keys.ToggleTips):
			toggleTips(&m)

		case key.Matches(keyMsg, keys.ApplyTips):
			applyTips(&m)
		}
	}

	if isValidSolvedBoard(m.game.board) {
		m.state = WinScreen
	}

	return m, nil
}

func (m model) View() string {
	switch m.state {

	case Playing:
		rows := [][]string{}
		var boxId int
		for i := range 3 {
			row := []string{}
			for j := range 3 {
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

		m.boardWidth = lipgloss.Width(renderedTable)
		m.boardHeight = lipgloss.Height(renderedTable)

		m.help.Width = m.width - m.boardWidth - 1
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

	case WinScreen:
		m.help.ShowAll = false
		m.help.Width = m.width
		helpView := "    " + m.help.View(m.keys)
		m.tips = ""
		winMessage := lipgloss.NewStyle().Foreground(completedNumberForeground).Render("You won!")
		renderedTable := lipgloss.Place(m.width, m.boardHeight-1, lipgloss.Center, lipgloss.Center, winMessage)
		return lipgloss.JoinVertical(lipgloss.Left,
			renderedTable,
			helpView,
		)

	case MainMenu:
		return lipgloss.NewStyle().Margin(1, 2).Render(m.mainMenuModel.View())

	case PickDifficulty:
		return lipgloss.NewStyle().Margin(1, 2).Render(m.difficultyMenuModel.View())

	case PickStrategy:
		return lipgloss.NewStyle().Margin(1, 2).Render(m.strategyMenuModel.View())

	case PickFileToLoad:
		var s strings.Builder
		s.WriteString("\n  ")
		if m.err != nil {
			s.WriteString(m.filepickerModel.Styles.DisabledFile.Render(m.err.Error()))
		} else {
			s.WriteString("Pick a file:")
		}
		s.WriteString("\n\n" + m.filepickerModel.View() + "\n")
		return s.String()

	case PickFileToSave:
		return lipgloss.Place(
			m.width,
			m.boardHeight-1,
			lipgloss.Center,
			lipgloss.Center,
			fmt.Sprintf(
				"Saving current puzzle as ./archive/%s.csv\n\n%s",
				m.textinputModel.View(),
				"(esc to abort)")+"\n")
	}

	return ""
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
	if number == 0 && cursorNumber != 0 && m.game.candidates[row][col][cursorNumber-1] {
		background = cursorNumberBackground
		foreground = cursorCandidatesForeground
	}
	if cursorRow == row && cursorCol == col {
		background = cursorBackground
		foreground = cursorCandidatesForeground
	}
	if number > 0 && m.editable[row][col] {
		foreground = editableForeground
	}
	if number != 0 && number == cursorNumber && (row != cursorRow || col != cursorCol) {
		background = cursorNumberBackground
		foreground = cursorNumberForeground
	}
	if !m.editable[row][col] {
		foreground = uneditableForeground
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
	for i := range 3 {
		rowString = ""
		for j := range 3 {
			number = uint8(3*i + j + 1)
			if len(rowString) > 0 {
				rowString += " "
			}
			if slices.Contains(candidates, number) {
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
	if !isValidBoard(m.game.board) {
		m.tips = "You made a mistake!"
		return
	}
	for _, strategy := range SolveStrategies {
		steps := strategy.Apply(&m.tipsGame)
		if len(steps) > 0 {
			m.strategies = steps
			m.tips = steps[0].strategyName + ":\n"
			for step := range steps {
				m.tips += fmt.Sprintf("%s\n", steps[step].description)
			}
			m.tips += "\nPress 'T' to apply all tips"
			return
		}
	}
	m.tips = "No hints available"
}

func applyTips(m *model) {
	for _, step := range m.strategies {
		step.Apply(&m.tipsGame)
		step.Apply(&m.game)
	}
	m.strategies = []SolutionStep{}
	updateTipsString(m)
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
