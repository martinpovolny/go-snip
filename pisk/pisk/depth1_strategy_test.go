package pisk_test

import (
	"martinp/piskvorky/pisk"
	"testing"
)

func TestNextMove(t *testing.T) {
	type testCase struct {
		game  []pisk.Move // moves of the game
		moves []pisk.Move // expected moves (one of)
	}

	var tests []testCase = []testCase{
		{
			game: []pisk.Move{
				{2, 3}, {3, 3}, {4, 3}, {5, 3}, {6, 3}, {2, 4}, {4, 4}, {6, 4},
				{7, 4}, {2, 5}, {3, 5}, {7, 5}, {2, 6}, {7, 6}, {2, 7}, {3, 7},
				{4, 7}, {7, 7}, {8, 7}, {4, 8}, {8, 8}, {9, 8}, {4, 9}, {5, 9},
				{7, 9}, {9, 9}, {5, 10}, {7, 10}, {8, 10}, {9, 10}, {5, 11}, {6, 11}, {7, 11},
			},
			moves: []pisk.Move{{2, 7}, {6, 3}},
		},
	}

	var strategy pisk.Depth1Strategy = pisk.Depth1Strategy{}

	for _, tc := range tests {
		game := pisk.NewGame(32, true)
		numMoves := game.LoadFromArray(tc.game)
		played, _ := strategy.NextMove(&game.Board, uint8(numMoves%2))

		found := false
		for _, move := range tc.moves {
			if move == played {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Strategy failed: %v not in %v", played, tc.moves)
		}
	}
}
