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

	numBricks = 4 // number of brick obstacles placed in Demo mode
)

// Grid holds the board state. Row 0 is the top.
// Bricks are static obstacles that block cells; they are separate from ball colors.
type Grid struct {
	Cells  [GridH][GridW]int
	Bricks [GridH][GridW]bool
}

// columnSegments returns the contiguous row ranges within column c that are not
// occupied by bricks. Each segment is [top, bottom] inclusive.
func (g *Grid) columnSegments(c int) [][2]int {
	var segs [][2]int
	start := 0
	for r := 0; r < GridH; r++ {
		if g.Bricks[r][c] {
			if r > start {
				segs = append(segs, [2]int{start, r - 1})
			}
			start = r + 1
		}
	}
	if start < GridH {
		segs = append(segs, [2]int{start, GridH - 1})
	}
	return segs
}

// GravityOnly compacts each column segment downward without filling empty cells.
// Returns animation data for every ball that moved.
func (g *Grid) GravityOnly() []FallingBall {
	var falls []FallingBall
	for c := 0; c < GridW; c++ {
		for _, seg := range g.columnSegments(c) {
			top, bottom := seg[0], seg[1]
			type ballSrc struct{ color, srcRow int }
			var stack []ballSrc
			for r := bottom; r >= top; r-- {
				if g.Cells[r][c] != ColorNone {
					stack = append(stack, ballSrc{g.Cells[r][c], r})
				}
			}
			writeRow := bottom
			for _, ball := range stack {
				g.Cells[writeRow][c] = ball.color
				if ball.srcRow != writeRow {
					falls = append(falls, FallingBall{ball.color, c, float64(ball.srcRow), float64(writeRow)})
				}
				writeRow--
			}
			for r := writeRow; r >= top; r-- {
				g.Cells[r][c] = ColorNone
			}
		}
	}
	return falls
}

// NewAttackGrid creates a grid for Ball Attack mode: only the bottom 2 rows
// are filled; rows 0–7 are empty. The filled rows are guaranteed match-free.
// No bricks in attack mode.
func NewAttackGrid() *Grid {
	g := &Grid{}
	for r := GridH - 2; r < GridH; r++ {
		for c := 0; c < GridW; c++ {
			g.Cells[r][c] = rand.Intn(numColors) + 1
		}
	}
	for {
		matches := g.FindMatches()
		if len(matches) == 0 {
			break
		}
		for _, p := range matches {
			g.Cells[p.R][p.C] = rand.Intn(numColors) + 1
		}
	}
	return g
}

// NewGrid creates a randomly filled grid with no pre-existing matches and a
// small number of brick obstacles placed at random non-adjacent positions.
func NewGrid() *Grid {
	g := &Grid{}

	// Place bricks first so gravity segments are established before filling.
	for placed := 0; placed < numBricks; {
		r := rand.Intn(GridH)
		c := rand.Intn(GridW)
		if !g.Bricks[r][c] {
			g.Bricks[r][c] = true
			placed++
		}
	}

	// Fill non-brick cells with random colors.
	for r := 0; r < GridH; r++ {
		for c := 0; c < GridW; c++ {
			if !g.Bricks[r][c] {
				g.Cells[r][c] = rand.Intn(numColors) + 1
			}
		}
	}

	// Resolve any initial matches by rerolling matched cells.
	for {
		matches := g.FindMatches()
		if len(matches) == 0 {
			break
		}
		for _, p := range matches {
			g.Cells[p.R][p.C] = rand.Intn(numColors) + 1
		}
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

// BrickSnapshot returns a 2D slice copy of the brick layout.
func (g *Grid) BrickSnapshot() [][]bool {
	out := make([][]bool, GridH)
	for r := 0; r < GridH; r++ {
		out[r] = make([]bool, GridW)
		copy(out[r], g.Bricks[r][:])
	}
	return out
}

// DestroyAdjacentBricks removes any brick that is orthogonally adjacent to a
// matched cell. Returns the positions of the destroyed bricks.
func (g *Grid) DestroyAdjacentBricks(matches []Pos) []Pos {
	dirs := [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	destroyed := map[Pos]bool{}
	for _, p := range matches {
		for _, d := range dirs {
			np := Pos{p.R + d[0], p.C + d[1]}
			if np.R < 0 || np.R >= GridH || np.C < 0 || np.C >= GridW {
				continue
			}
			if g.Bricks[np.R][np.C] && !destroyed[np] {
				destroyed[np] = true
				g.Bricks[np.R][np.C] = false
			}
		}
	}
	result := make([]Pos, 0, len(destroyed))
	for p := range destroyed {
		result = append(result, p)
	}
	return result
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

// ApplyGravity compacts each column segment downward (row GridH-1 = bottom).
func (g *Grid) ApplyGravity() {
	for c := 0; c < GridW; c++ {
		for _, seg := range g.columnSegments(c) {
			top, bottom := seg[0], seg[1]
			stack := make([]int, 0, bottom-top+1)
			for r := bottom; r >= top; r-- {
				if g.Cells[r][c] != ColorNone {
					stack = append(stack, g.Cells[r][c])
				}
			}
			writeRow := bottom
			for _, clr := range stack {
				g.Cells[writeRow][c] = clr
				writeRow--
			}
			for r := writeRow; r >= top; r-- {
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

// GravityAndFill applies gravity per column-segment. Only the topmost segment
// in each column (top == 0) receives new random balls from above the grid;
// segments below a brick compact their existing balls downward but are not
// refilled — spawning a ball from above the grid into a lower segment would
// require it to animate through the brick above it.
func (g *Grid) GravityAndFill() []FallingBall {
	var falls []FallingBall
	for c := 0; c < GridW; c++ {
		for _, seg := range g.columnSegments(c) {
			top, bottom := seg[0], seg[1]

			type ballSrc struct{ color, srcRow int }
			var stack []ballSrc
			for r := bottom; r >= top; r-- {
				if g.Cells[r][c] != ColorNone {
					stack = append(stack, ballSrc{g.Cells[r][c], r})
				}
			}

			// New balls only enter from the very top of the grid.
			if top == 0 {
				gaps := (bottom - top + 1) - len(stack)
				for i := 0; i < gaps; i++ {
					clr := rand.Intn(numColors) + 1
					destRow := float64(gaps - 1 - i) // top==0, so top+gaps-1-i == gaps-1-i
					spawnRow := float64(-(i + 1))
					falls = append(falls, FallingBall{clr, c, spawnRow, destRow})
					stack = append(stack, ballSrc{clr, -1})
				}
			}

			// Write segment bottom-to-top: compact balls, clear remaining rows.
			writeRow := bottom
			for _, ball := range stack {
				g.Cells[writeRow][c] = ball.color
				if ball.srcRow != -1 && ball.srcRow != writeRow {
					falls = append(falls, FallingBall{ball.color, c, float64(ball.srcRow), float64(writeRow)})
				}
				writeRow--
			}
			for r := writeRow; r >= top; r-- {
				g.Cells[r][c] = ColorNone
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
