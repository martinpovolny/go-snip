package main

import (
	"strings"
	"testing"
)

// parseBoard builds a Grid from a compact text representation.
// Each character represents a cell: R=Red G=Green B=Blue Y=Yellow .=Empty
// Lines are trimmed; blank lines are skipped.
func parseBoard(s string) *Grid {
	g := &Grid{}
	r := 0
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		c := 0
		for _, ch := range strings.Fields(line) {
			if c >= GridW {
				break
			}
			switch ch {
			case "R":
				g.Cells[r][c] = ColorRed
			case "G":
				g.Cells[r][c] = ColorGreen
			case "B":
				g.Cells[r][c] = ColorBlue
			case "Y":
				g.Cells[r][c] = ColorYellow
			default:
				g.Cells[r][c] = ColorNone
			}
			c++
		}
		r++
		if r >= GridH {
			break
		}
	}
	return g
}

func posSet(matches []Pos) map[Pos]bool {
	s := map[Pos]bool{}
	for _, p := range matches {
		s[p] = true
	}
	return s
}

func TestHorizontalRun(t *testing.T) {
	g := parseBoard(`
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. R R R . . . .
	`)
	matches := g.FindMatches()
	got := posSet(matches)
	want := map[Pos]bool{{9, 1}: true, {9, 2}: true, {9, 3}: true}
	if len(got) != len(want) {
		t.Fatalf("horizontal: got %v, want %v", got, want)
	}
	for p := range want {
		if !got[p] {
			t.Errorf("missing %v", p)
		}
	}
}

func TestVerticalRun(t *testing.T) {
	g := parseBoard(`
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . G . . . . .
		. . G . . . . .
		. . G . . . . .
	`)
	matches := g.FindMatches()
	got := posSet(matches)
	want := map[Pos]bool{{7, 2}: true, {8, 2}: true, {9, 2}: true}
	if len(got) != len(want) {
		t.Fatalf("vertical: got %v, want %v", got, want)
	}
	for p := range want {
		if !got[p] {
			t.Errorf("missing %v", p)
		}
	}
}

func TestTwoByTwoBlock(t *testing.T) {
	g := parseBoard(`
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . B B . . .
		. . . B B . . .
	`)
	matches := g.FindMatches()
	got := posSet(matches)
	want := map[Pos]bool{{8, 3}: true, {8, 4}: true, {9, 3}: true, {9, 4}: true}
	if len(got) != len(want) {
		t.Fatalf("2x2: got %v, want %v", got, want)
	}
	for p := range want {
		if !got[p] {
			t.Errorf("missing %v", p)
		}
	}
}

func TestNoMatch(t *testing.T) {
	g := parseBoard(`
		R G B Y R G B Y
		G B Y R G B Y R
		B Y R G B Y R G
		Y R G B Y R G B
		R G B Y R G B Y
		G B Y R G B Y R
		B Y R G B Y R G
		Y R G B Y R G B
		R G B Y R G B Y
		G B Y R G B Y R
	`)
	matches := g.FindMatches()
	if len(matches) != 0 {
		t.Errorf("expected no matches, got %v", matches)
	}
}

func TestGravity(t *testing.T) {
	g := parseBoard(`
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. R . . . . . .
		. . . . . . . .
	`)
	g.ApplyGravity()
	if g.Cells[9][1] != ColorRed {
		t.Errorf("expected Red at (9,1) after gravity, got %d", g.Cells[9][1])
	}
	if g.Cells[8][1] != ColorNone {
		t.Errorf("expected empty at (8,1) after gravity, got %d", g.Cells[8][1])
	}
}

func TestSwapCreatesMatch(t *testing.T) {
	g := parseBoard(`
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		R R G R . . . .
	`)
	// swap col 2 (G) with col 3 (R) → bottom row becomes R R R G
	g.Swap(Pos{9, 2}, Pos{9, 3})
	matches := g.FindMatches()
	got := posSet(matches)
	for _, p := range []Pos{{9, 0}, {9, 1}, {9, 2}} {
		if !got[p] {
			t.Errorf("expected %v in match set, got %v", p, got)
		}
	}
}

func TestCascade(t *testing.T) {
	// Setup: horizontal Y Y Y Y match at row 7.
	// Above it are 3 Bs in col 0 (rows 5,6,8) separated by the Y row.
	// Clearing the Ys causes gravity to stack all 3 Bs into consecutive rows,
	// creating a vertical match in the cascade check.
	g := parseBoard(`
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		B . . . . . . .
		B . . . . . . .
		Y Y Y Y . . . .
		B B . . . . . .
		. . . . . . . .
	`)
	m1 := g.FindMatches()
	if len(m1) == 0 {
		t.Fatal("expected initial match (Y Y Y Y)")
	}
	g.ClearMatches(m1)
	g.ApplyGravity()
	// After gravity, col 0 has B at rows 7,8,9 → vertical match.
	m2 := g.FindMatches()
	got2 := posSet(m2)
	for _, p := range []Pos{{7, 0}, {8, 0}, {9, 0}} {
		if !got2[p] {
			t.Errorf("cascade: expected %v in second match, got %v", p, got2)
		}
	}
}

func TestMatchGroupOrder(t *testing.T) {
	// Two separate horizontal runs: one near top-left, one near bottom-right.
	// FindMatchGroups must return the top-left group first.
	g := parseBoard(`
		. . . . . . . .
		. . . . . . . .
		R R R . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . G G G
	`)
	groups := g.FindMatchGroups()
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	// First group must contain row 2 cells (the R R R).
	firstRowMin := GridH
	for _, p := range groups[0] {
		if p.R < firstRowMin {
			firstRowMin = p.R
		}
	}
	if firstRowMin != 2 {
		t.Errorf("first group should be at row 2 (R R R), got min row %d", firstRowMin)
	}
}

func TestMatchGroupOrderLeftRight(t *testing.T) {
	// Two separate runs on the same row: left group must come first.
	g := parseBoard(`
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		R R R . B B B .
	`)
	groups := g.FindMatchGroups()
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	// First group must be the leftmost (R R R at cols 0-2).
	firstMinCol := GridW
	for _, p := range groups[0] {
		if p.C < firstMinCol {
			firstMinCol = p.C
		}
	}
	if firstMinCol != 0 {
		t.Errorf("first group should start at col 0 (R R R), got min col %d", firstMinCol)
	}
}

func TestOverlappingPatterns(t *testing.T) {
	// A horizontal run overlapping with a 2×2 block should mark all cells.
	g := parseBoard(`
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. Y Y Y . . . .
		. Y Y . . . . .
	`)
	// Horizontal run at row 8 cols 1-3, plus 2×2 at rows 8-9 cols 1-2.
	matches := g.FindMatches()
	got := posSet(matches)
	// All five Y cells should be marked.
	for _, p := range []Pos{{8, 1}, {8, 2}, {8, 3}, {9, 1}, {9, 2}} {
		if !got[p] {
			t.Errorf("overlapping: expected %v, got %v", p, got)
		}
	}
}

// ── GravityOnly ───────────────────────────────────────────────────────────────

func TestGravityOnly_AlreadySettled(t *testing.T) {
	g := &Grid{}
	g.Cells[GridH-1][0] = ColorRed
	g.Cells[GridH-1][1] = ColorBlue

	falls := g.GravityOnly()
	if len(falls) != 0 {
		t.Fatalf("settled balls should produce no falls, got %d", len(falls))
	}
	if g.Cells[GridH-1][0] != ColorRed || g.Cells[GridH-1][1] != ColorBlue {
		t.Fatal("settled balls should not move")
	}
}

func TestGravityOnly_SingleBallDropsToBottom(t *testing.T) {
	g := &Grid{}
	g.Cells[0][0] = ColorGreen // lone ball at top of empty column

	falls := g.GravityOnly()

	if len(falls) != 1 {
		t.Fatalf("expected 1 fall record, got %d", len(falls))
	}
	f := falls[0]
	if f.Color != ColorGreen {
		t.Errorf("color: got %d, want %d (Green)", f.Color, ColorGreen)
	}
	if f.FromRow != 0 {
		t.Errorf("FromRow: got %v, want 0", f.FromRow)
	}
	if f.ToRow != float64(GridH-1) {
		t.Errorf("ToRow: got %v, want %d", f.ToRow, GridH-1)
	}
	if g.Cells[GridH-1][0] != ColorGreen {
		t.Error("ball should land at the bottom row")
	}
}

func TestGravityOnly_DoesNotFillEmpty(t *testing.T) {
	// After dropping one ball there should be exactly GridH-1 empty cells in that column.
	g := &Grid{}
	g.Cells[0][0] = ColorRed

	g.GravityOnly()

	empty := 0
	for r := 0; r < GridH; r++ {
		if g.Cells[r][0] == ColorNone {
			empty++
		}
	}
	if empty != GridH-1 {
		t.Errorf("GravityOnly should leave %d empty cells, found %d non-empty", GridH-1, GridH-1-empty)
	}
}

func TestGravityOnly_LandsOnExistingBall(t *testing.T) {
	// Ball at row GridH-3 should land on top of ball at GridH-1.
	g := &Grid{}
	g.Cells[GridH-1][0] = ColorYellow
	g.Cells[GridH-3][0] = ColorRed

	falls := g.GravityOnly()

	if len(falls) != 1 {
		t.Fatalf("expected 1 fall (Red), got %d", len(falls))
	}
	if falls[0].Color != ColorRed {
		t.Errorf("expected Red to fall, got color %d", falls[0].Color)
	}
	if falls[0].ToRow != float64(GridH-2) {
		t.Errorf("ToRow: got %v, want %d", falls[0].ToRow, GridH-2)
	}
}

func TestGravityOnly_StackOrder(t *testing.T) {
	// Two balls at rows 0 and 2; both compact to the bottom, order preserved.
	g := &Grid{}
	g.Cells[0][0] = ColorRed
	g.Cells[2][0] = ColorBlue

	g.GravityOnly()

	// GravityOnly collects bottom-to-top: Blue(row2) first, Red(row0) second.
	// Writes: Cells[GridH-1]=Blue, Cells[GridH-2]=Red.
	if g.Cells[GridH-1][0] != ColorBlue {
		t.Errorf("bottom cell: got %d, want Blue(%d)", g.Cells[GridH-1][0], ColorBlue)
	}
	if g.Cells[GridH-2][0] != ColorRed {
		t.Errorf("second cell: got %d, want Red(%d)", g.Cells[GridH-2][0], ColorRed)
	}
	for r := 0; r < GridH-2; r++ {
		if g.Cells[r][0] != ColorNone {
			t.Errorf("row %d should be empty, got %d", r, g.Cells[r][0])
		}
	}
}

// ── GravityAndFill ────────────────────────────────────────────────────────────

func TestGravityAndFill_EmptyColumnFilled(t *testing.T) {
	g := &Grid{}
	g.GravityAndFill()

	for r := 0; r < GridH; r++ {
		if g.Cells[r][0] == ColorNone {
			t.Errorf("row %d col 0 should be filled, got empty", r)
		}
	}
}

func TestGravityAndFill_NewBallsHaveNegativeFromRow(t *testing.T) {
	g := &Grid{} // completely empty
	for _, f := range g.GravityAndFill() {
		if f.FromRow >= 0 {
			t.Errorf("spawned ball should have FromRow<0, got %v", f.FromRow)
		}
	}
}

func TestGravityAndFill_FullColumnUnchanged(t *testing.T) {
	// Fill every column so no spawning happens anywhere.
	g := &Grid{}
	for r := 0; r < GridH; r++ {
		for c := 0; c < GridW; c++ {
			g.Cells[r][c] = ColorRed
		}
	}
	before := g.Cells

	falls := g.GravityAndFill()

	if g.Cells != before {
		t.Error("full grid must not change")
	}
	if len(falls) != 0 {
		t.Errorf("full grid should produce no fall records, got %d", len(falls))
	}
}

func TestGravityAndFill_ExistingBallMovesAndGapFilled(t *testing.T) {
	// One ball at row 0; gravity moves it to the bottom, new balls fill above it.
	g := &Grid{}
	g.Cells[0][0] = ColorRed

	falls := g.GravityAndFill()

	if g.Cells[GridH-1][0] == ColorNone {
		t.Error("bottom cell should be non-empty after fill")
	}

	var existing, spawned int
	for _, f := range falls {
		if f.Col != 0 {
			continue
		}
		if f.FromRow < 0 {
			spawned++
		} else {
			existing++
		}
	}
	if existing != 1 {
		t.Errorf("expected 1 existing-ball fall record for col 0, got %d", existing)
	}
	if spawned != GridH-1 {
		t.Errorf("expected %d spawned-ball records, got %d", GridH-1, spawned)
	}
}

// ── ClearMatches + GravityOnly round-trip (Ball Attack) ───────────────────────

func TestClearAndGravityOnly_AttackCycle(t *testing.T) {
	// Vertical match of 3 at the bottom; one orphan ball above.
	// After clear + GravityOnly the orphan falls to bottom; no fill.
	g := parseBoard(`
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . . . . . . .
		. . B . . . . .
		. . R . . . . .
		. . R . . . . .
		. . R . . . . .
	`)
	matches := g.FindMatches()
	if len(matches) != 3 {
		t.Fatalf("expected 3 matched cells (vertical R R R), got %d", len(matches))
	}
	g.ClearMatches(matches)

	falls := g.GravityOnly()

	if len(falls) != 1 {
		t.Fatalf("expected 1 fall (Blue), got %d", len(falls))
	}
	if falls[0].Color != ColorBlue {
		t.Errorf("expected Blue to fall, got color %d", falls[0].Color)
	}
	if falls[0].ToRow != float64(GridH-1) {
		t.Errorf("Blue should land at row %d, got %v", GridH-1, falls[0].ToRow)
	}
	// Column 2 should have Blue at bottom, rest empty.
	if g.Cells[GridH-1][2] != ColorBlue {
		t.Error("Blue should be at bottom of column 2")
	}
	for r := 0; r < GridH-1; r++ {
		if g.Cells[r][2] != ColorNone {
			t.Errorf("GravityOnly must not fill: row %d col 2 is %d", r, g.Cells[r][2])
		}
	}
}

// ── NewAttackGrid ─────────────────────────────────────────────────────────────

func TestNewAttackGrid_BottomTwoRowsFilled(t *testing.T) {
	g := NewAttackGrid()
	for r := 0; r < GridH-2; r++ {
		for c := 0; c < GridW; c++ {
			if g.Cells[r][c] != ColorNone {
				t.Errorf("row %d col %d should be empty, got %d", r, c, g.Cells[r][c])
			}
		}
	}
	for r := GridH - 2; r < GridH; r++ {
		for c := 0; c < GridW; c++ {
			if g.Cells[r][c] == ColorNone {
				t.Errorf("row %d col %d should be filled, got empty", r, c)
			}
		}
	}
}

func TestNewAttackGrid_NoInitialMatches(t *testing.T) {
	for i := 0; i < 100; i++ {
		g := NewAttackGrid()
		if m := g.FindMatches(); len(m) != 0 {
			t.Fatalf("iteration %d: NewAttackGrid has %d initial matched cells", i, len(m))
		}
	}
}

func TestNewGrid_NoInitialMatches(t *testing.T) {
	for i := 0; i < 20; i++ {
		g := NewGrid()
		if m := g.FindMatches(); len(m) != 0 {
			t.Fatalf("iteration %d: NewGrid has %d initial matched cells", i, len(m))
		}
	}
}
