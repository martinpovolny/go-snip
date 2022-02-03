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
				{6, 6}, {5, 5}, {6, 7}, {5, 7}, {6, 8}, {5, 8}, {5, 6}, {6, 5}, {4, 5}, {3, 4}, {7, 8},
				{8, 9}, {6, 9}, {6, 10}, {3, 6}, {4, 6}, {5, 4},
			},
			moves: []pisk.Move{{2, 7}, {6, 3}},
		},
		{
			game: []pisk.Move{
				{6, 6}, {5, 5}, {6, 7}, {5, 7}, {6, 8}, {5, 8}, {5, 6}, {6, 5}, {4, 5}, {3, 4}, {7, 8},
				{8, 9}, {6, 9}, {6, 10}, {3, 6}, {4, 6}, {5, 4}, {6, 3}, {2, 7}, {1, 8}, {8, 7}, {9, 6},
				{5, 10}, {4, 11}, {3, 8}, {4, 9}, {3, 7}, {3, 5}, {3, 9}, {3, 10}, {0, 7}, {2, 4}, {1, 3},
				{2, 9}, {5, 12}, {2, 11}, {1, 12}, {4, 8}, {1, 7}, {4, 7}, {4, 10}, {2, 8}, {7, 7}, {2, 10},
			},
			moves: []pisk.Move{{2, 12}},
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
