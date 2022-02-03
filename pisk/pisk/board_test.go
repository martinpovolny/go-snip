package pisk

import (
	"testing"
)

func TestPlace(t *testing.T) {

	testCases := []Move{
		{X: 10, Y: 9},
		{X: 24, Y: 30},
	}

	for _, testCase := range testCases {
		b := NewBoard(32)
		b.Place(testCase.X, testCase.Y)

		if b.vertical[testCase.Y]&(1<<testCase.X) == 0 {
			t.Errorf("vertical array bit missing (%v, %v): %v", testCase.X, testCase.Y, b.vertical[testCase.Y])
		}

		if b.horizontal[testCase.X]&(1<<testCase.Y) == 0 {
			t.Errorf("horizontal array bit missing (%v, %v): %v", testCase.X, testCase.Y, b.horizontal[testCase.X])
		}

		if b.mainDiagonal[testCase.X+testCase.Y]&(1<<testCase.X) == 0 {
			t.Errorf("mainDiagonal array bit missing (%v, %v): %v", testCase.X, testCase.Y, b.mainDiagonal[testCase.X+testCase.Y])
		}

		if b.antiDiagonal[testCase.X-testCase.Y+b.size-1]&(1<<testCase.X) == 0 {
			t.Errorf("antiDiagonal array bit missing (%v, %v): %v", testCase.X, testCase.Y,
				b.antiDiagonal[testCase.X-testCase.Y+b.size-1])
		}
	}
}
