# Sudoku Generator and Solver in Go

Toy project to learn Go. Generates and solves Sudoku puzzles and provides a simple TUI to play them.

Unlike most Sudoku generators I found, this one does not start with a solved puzzle and removes numbers.
Instead, it starts with an empty grid and fills it with random numbers. I did not like the idea of having
to generate a valid solved puzzle first, even though that should be quite straightforward by shuffling a
hard-coded solution, so I went for this approach instead, because I like the challenge of approaching it
differently than the algorithms I found when doing some research before starting.

## Installation

Requires Go, see [here](https://golang.org/doc/install) for installation instructions.

```bash
go install github.com/kleinjohann/sugoku@latest
```

Note that this will install the program in your `$GOPATH/bin` directory. Make sure that this directory is in your `$PATH` to be able to run the program from anywhere.

## Usage

```bash
sugoku [--print]
```

When the `--print` flag is set, the program simply prints a generated Sudoku and its solution.
Otherwise, you are presented with a TUI to solve a randomly generated Sudoku puzzle.

Example screenshot of the TUI:

![](/images/tui.png)

Example output of `sugoku --print`:

```
Generated Sudoku:
   |-----------|-----------|-----------|
   |     3   1 | 9   2   7 |     5     |
   |           |           |           |
   | 9   5     |     3   6 |           |
   |           |           |           |
   |         2 |           | 7   3     |
   |-----------|-----------|-----------|
   | 2         | 6   5     |         3 |
   |           |           |           |
   |     7     |     8   9 | 2         |
   |           |           |           |
   |           | 2   7     | 6         |
   |-----------|-----------|-----------|
   |     4     |           |           |
   |           |           |           |
   |         5 |     1     | 3         |
   |           |           |           |
   | 7   8   9 | 5       3 | 4       1 |
   |-----------|-----------|-----------|

Solution:
   |-----------|-----------|-----------|
   | 4   3   1 | 9   2   7 | 8   5   6 |
   |           |           |           |
   | 9   5   7 | 8   3   6 | 1   4   2 |
   |           |           |           |
   | 8   6   2 | 1   4   5 | 7   3   9 |
   |-----------|-----------|-----------|
   | 2   1   8 | 6   5   4 | 9   7   3 |
   |           |           |           |
   | 5   7   6 | 3   8   9 | 2   1   4 |
   |           |           |           |
   | 3   9   4 | 2   7   1 | 6   8   5 |
   |-----------|-----------|-----------|
   | 1   4   3 | 7   9   2 | 5   6   8 |
   |           |           |           |
   | 6   2   5 | 4   1   8 | 3   9   7 |
   |           |           |           |
   | 7   8   9 | 5   6   3 | 4   2   1 |
   |-----------|-----------|-----------|

```

## Planned Improvements

- Add a difficulty metric and let the user choose the difficulty of the generated puzzle
    - Implement more solving strategies
    - Rank the strategies by difficulty and always apply them in order of increasing difficulty
    - Puzzle difficulty is then determined by number and difficulty of required strategies
- Improve the TUI
    - Allow selection of multiple cells to enter multiple candidates at once
    - Improve keymap and hint formatting for narrow terminal windows
- Integrate with a simple web server to play Sudoku in the browser
