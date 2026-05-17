package main

import "math/rand"

const (
	GridW = 8
	GridH = 10

	ColorNone   = 0
	ColorRed    = 1
	ColorGreen  = 2
	ColorBlue   = 3
	ColorYellow = 4
	numColors   = 4
)

// Grid holds the board state. Row 0 is the top.
type Grid struct {
	Cells [GridH][GridW]int
}

// GravityOnly compacts each column downward without filling empty cells.
// Returns animation data for every ball that moved.
func (g *Grid) GravityOnly() []FallingBall {
	var falls []FallingBall
	for c := 0; c < GridW; c++ {
		type ballSrc struct{ color, srcRow int }
		var stack []ballSrc
		for r := GridH - 1; r >= 0; r-- {
			if g.Cells[r][c] != ColorNone {
				stack = append(stack, ballSrc{g.Cells[r][c], r})
			}
		}
		for r := GridH - 1; r >= 0; r-- {
			idx := GridH - 1 - r
			if idx < len(stack) {
				ball := stack[idx]
				g.Cells[r][c] = ball.color
				if ball.srcRow != r {
					falls = append(falls, FallingBall{ball.color, c, float64(ball.srcRow), float64(r)})
				}
			} else {
				g.Cells[r][c] = ColorNone
			}
		}
	}
	return falls
}

// NewAttackGrid creates a grid for Ball Attack mode: only the bottom 2 rows
// are filled; rows 0–7 are empty.
func NewAttackGrid() *Grid {
	g := &Grid{}
	for r := GridH - 2; r < GridH; r++ {
		for c := 0; c < GridW; c++ {
			g.Cells[r][c] = rand.Intn(numColors) + 1
		}
	}
	return g
}

// NewGrid creates a randomly filled grid with no pre-existing matches.
func NewGrid() *Grid {
	g := &Grid{}
	for {
		for r := 0; r < GridH; r++ {
			for c := 0; c < GridW; c++ {
				g.Cells[r][c] = rand.Intn(numColors) + 1
			}
		}
		// resolve any initial matches by rerolling those cells
		for {
			matches := g.FindMatches()
			if len(matches) == 0 {
				break
			}
			for _, p := range matches {
				g.Cells[p.R][p.C] = rand.Intn(numColors) + 1
			}
		}
		break
	}
	return g
}

// Snapshot returns a 2D slice copy of the board (safe to read from other goroutines
// only while the caller holds the game mutex).
func (g *Grid) Snapshot() [][]int {
	out := make([][]int, GridH)
	for r := 0; r < GridH; r++ {
		out[r] = make([]int, GridW)
		copy(out[r], g.Cells[r][:])
	}
	return out
}

// Swap exchanges two adjacent cells.
func (g *Grid) Swap(a, b Pos) {
	g.Cells[a.R][a.C], g.Cells[b.R][b.C] = g.Cells[b.R][b.C], g.Cells[a.R][a.C]
}

// FindMatches returns the set of cells that are part of a match:
//   - horizontal run of 3+
//   - vertical run of 3+
//   - 2×2 block of same color
func (g *Grid) FindMatches() []Pos {
	marked := map[Pos]bool{}

	// horizontal runs of 3+
	for r := 0; r < GridH; r++ {
		for c := 0; c < GridW; {
			clr := g.Cells[r][c]
			if clr == ColorNone {
				c++
				continue
			}
			start := c
			for c < GridW && g.Cells[r][c] == clr {
				c++
			}
			if c-start >= 3 {
				for i := start; i < c; i++ {
					marked[Pos{r, i}] = true
				}
			}
		}
	}

	// vertical runs of 3+
	for c := 0; c < GridW; c++ {
		for r := 0; r < GridH; {
			clr := g.Cells[r][c]
			if clr == ColorNone {
				r++
				continue
			}
			start := r
			for r < GridH && g.Cells[r][c] == clr {
				r++
			}
			if r-start >= 3 {
				for i := start; i < r; i++ {
					marked[Pos{i, c}] = true
				}
			}
		}
	}

	// 2×2 blocks
	for r := 0; r < GridH-1; r++ {
		for c := 0; c < GridW-1; c++ {
			clr := g.Cells[r][c]
			if clr == ColorNone {
				continue
			}
			if g.Cells[r][c+1] == clr && g.Cells[r+1][c] == clr && g.Cells[r+1][c+1] == clr {
				marked[Pos{r, c}] = true
				marked[Pos{r, c + 1}] = true
				marked[Pos{r + 1, c}] = true
				marked[Pos{r + 1, c + 1}] = true
			}
		}
	}

	result := make([]Pos, 0, len(marked))
	for p := range marked {
		result = append(result, p)
	}
	return result
}

// ClearMatches removes the given cells from the grid.
func (g *Grid) ClearMatches(matches []Pos) {
	for _, p := range matches {
		g.Cells[p.R][p.C] = ColorNone
	}
}

// ApplyGravity compacts each column downward (row GridH-1 = bottom).
func (g *Grid) ApplyGravity() {
	for c := 0; c < GridW; c++ {
		// collect non-empty cells bottom-to-top
		stack := make([]int, 0, GridH)
		for r := GridH - 1; r >= 0; r-- {
			if g.Cells[r][c] != ColorNone {
				stack = append(stack, g.Cells[r][c])
			}
		}
		// rewrite column
		for r := GridH - 1; r >= 0; r-- {
			idx := GridH - 1 - r
			if idx < len(stack) {
				g.Cells[r][c] = stack[idx]
			} else {
				g.Cells[r][c] = ColorNone
			}
		}
	}
}

// FillEmpty replaces all empty cells with random colors.
func (g *Grid) FillEmpty() {
	for r := 0; r < GridH; r++ {
		for c := 0; c < GridW; c++ {
			if g.Cells[r][c] == ColorNone {
				g.Cells[r][c] = rand.Intn(numColors) + 1
			}
		}
	}
}

// FallingBall describes one ball's fall animation after gravity.
type FallingBall struct {
	Color       int
	Col         int
	FromRow     float64 // grid row where the ball starts (may be negative = above the grid)
	ToRow       float64 // grid row where it lands
}

// GravityAndFill applies gravity and fills empty cells, returning animation data
// for every ball that moved or was spawned.
func (g *Grid) GravityAndFill() []FallingBall {
	var falls []FallingBall
	for c := 0; c < GridW; c++ {
		// collect existing balls bottom-to-top, tracking source rows
		type ballSrc struct{ color, srcRow int }
		var stack []ballSrc
		for r := GridH - 1; r >= 0; r-- {
			if g.Cells[r][c] != ColorNone {
				stack = append(stack, ballSrc{g.Cells[r][c], r})
			}
		}
		// how many new balls needed?
		gaps := GridH - len(stack)
		// New balls fill the top `gaps` rows. The rewrite loop (below) assigns
		// stack[GridH-gaps+j] → row j (top-to-bottom within the gap), so we
		// append in the order that places spawnRow=-1 at the bottom of the gap
		// and spawnRow=-gaps at the top. All balls fall the same distance (gaps rows).
		for i := 0; i < gaps; i++ {
			clr := rand.Intn(numColors) + 1
			// stack position GridH-gaps+i → row gaps-1-i
			destRow := float64(gaps - 1 - i)
			spawnRow := float64(-(i + 1)) // -1, -2, ..., -gaps
			falls = append(falls, FallingBall{clr, c, spawnRow, destRow})
			stack = append(stack, ballSrc{clr, -1})
		}
		// rewrite column and record which existing balls moved
		for r := GridH - 1; r >= 0; r-- {
			idx := GridH - 1 - r
			ball := stack[idx]
			g.Cells[r][c] = ball.color
			destRow := float64(r)
			if ball.srcRow == -1 {
				// already recorded above
				continue
			}
			if ball.srcRow != r {
				falls = append(falls, FallingBall{ball.color, c, float64(ball.srcRow), destRow})
			}
		}
	}
	return falls
}

// FindMatchGroups returns match groups in top-to-bottom, left-to-right order.
// Each group is a connected set of cells that all belong to one or more match
// patterns (horizontal run, vertical run, or 2×2 block). When there are multiple
// independent groups, the one with the topmost-leftmost cell comes first.
func (g *Grid) FindMatchGroups() [][]Pos {
	all := g.FindMatches()
	if len(all) == 0 {
		return nil
	}
	matchSet := MatchSet(all)
	visited := map[Pos]bool{}
	var groups [][]Pos

	// Iterate in row-major order so the first unvisited cell of each group
	// determines the sort position naturally.
	for r := 0; r < GridH; r++ {
		for c := 0; c < GridW; c++ {
			p := Pos{r, c}
			if !matchSet[p] || visited[p] {
				continue
			}
			groups = append(groups, bfsGroup(p, matchSet, visited))
		}
	}
	return groups
}

func bfsGroup(start Pos, matchSet map[Pos]bool, visited map[Pos]bool) []Pos {
	dirs := [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	var group []Pos
	queue := []Pos{start}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		if visited[p] {
			continue
		}
		visited[p] = true
		group = append(group, p)
		for _, d := range dirs {
			np := Pos{p.R + d[0], p.C + d[1]}
			if matchSet[np] && !visited[np] {
				queue = append(queue, np)
			}
		}
	}
	return group
}

// MatchSet converts a []Pos slice to a map for O(1) lookup.
func MatchSet(matches []Pos) map[Pos]bool {
	s := make(map[Pos]bool, len(matches))
	for _, p := range matches {
		s[p] = true
	}
	return s
}

// MatchesAsArray converts []Pos to [][2]int for JSON.
func MatchesAsArray(matches []Pos) [][2]int {
	out := make([][2]int, len(matches))
	for i, p := range matches {
		out[i] = [2]int{p.R, p.C}
	}
	return out
}
