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

```
sugoku [-difficulty <0-5>] [-print] [-cores <int>] [-seed <int>] [-cpuprofile <file>]
  -difficulty int
        difficulty of the generated sudoku, 0 for random difficulty (default 0)
  -print
        print a generated sudoku and its solution and exit
  -cores int
        number of cores to use, -1 for all cores (default -1)
  -seed int
        seed for random number generator, -1 for random seed (default -1)
  -cpuprofile file
        write cpu profile to file
```

A puzzle's difficulty is given by the difficulty of the hardest strategy required to solve it.
The exact strategy difficulty mapping is as follows (bracketed strategies are not implemented yet):
1. Naked Single, Hidden Single
2. Naked Pair, Naked Triple, Naked Quad, Pointing Group, Box Reduction
3. Hidden Pair, Hidden Triple, Hidden Quad
4. X-Wing, Swordfish, Jellyfish, Skyscraper, (Y-Wing)
5. Not solvable using all of the above

Note that puzzles of difficulty >= 4 are quite rare and may take a while to generate.

When the `-print` flag is set, the program simply prints a generated Sudoku and its solution.
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

- Implement more solving strategies
- Improve the TUI
    - Allow selection of multiple cells to enter multiple candidates at once
    - Improve keymap and hint formatting for narrow terminal windows
- Integrate with a simple web server to play Sudoku in the browser
