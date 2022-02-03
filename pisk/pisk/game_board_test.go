package pisk

import (
	"testing"
)

func TestSearchThreats(t *testing.T) {
	type testCase struct {
		Name     string
		X        []Move
		O        []Move
		patterns []Pattern
		match    PatternMatch
	}

	testCases := []testCase{
		{
			Name: "4 horizontal",
			//X:    []Move{{1, 10}, {2, 10}, {3, 10}, {4, 10}},
			X: []Move{{1, 1}, {2, 1}, {3, 1}, {4, 1}},
			O: []Move{{10, 10}},
			patterns: []Pattern{{
				Pat:     0b0011110,
				Space:   0b0100001,
				NShifts: 27,
				Value:   2,
				Defense: []uint8{0, 5},
			}},
			match: PatternMatch{
				Pattern:   Pattern{},
				Index:     1,
				Shift:     0,
				Direction: 0,
			},
		},
		{
			Name: "4 diagonal",
			X:    []Move{{1, 1}, {2, 2}, {3, 3}, {4, 4}},
			O:    []Move{{11, 1}, {12, 2}, {13, 3}, {14, 4}},
			patterns: []Pattern{{
				Pat:     0b0011110,
				Space:   0b0100001,
				NShifts: 27,
				Value:   2,
				Defense: []uint8{0, 5},
			}},
			match: PatternMatch{
				Pattern:   Pattern{},
				Index:     31,
				Shift:     0,
				Direction: 3,
			},
		},
	}

	for _, tc := range testCases {
		b := NewGameBoard(32)

		for _, m := range tc.X {
			b.Place(m.X, m.Y, 0)
		}

		for _, m := range tc.O {
			b.Place(m.X, m.Y, 1)
		}

		// fmt.Println(tc.Name)
		// b.Print()

		matches := b.SearchThreats(tc.patterns, 0)
		// fmt.Println(matches)

		if len(matches) != 1 {
			t.Errorf("No match for test case %v", tc.Name)
		} else {
			match := matches[0]
			if match.Index != tc.match.Index ||
				match.Shift != tc.match.Shift ||
				match.Direction != tc.match.Direction {
				t.Errorf("Mismatch for test case %v: (%v, %v, %v) != (%v, %v, %v)",
					tc.Name,
					match.Index, match.Shift, match.Direction,
					tc.match.Index, tc.match.Shift, tc.match.Direction)
			}
		}
	}
}
