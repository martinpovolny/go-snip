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
