package pisk

import (
	"fmt"
	"strconv"
)

type GameBoard struct {
	size      uint8
	XBoard    Board
	OBoard    Board
	nextMoves Board
}

func NewGameBoard(size uint8) *GameBoard {
	return &GameBoard{
		size:      size,
		XBoard:    *NewBoard(size),
		OBoard:    *NewBoard(size),
		nextMoves: *NewBoard(size),
	}
}

func (gb *GameBoard) XWon() bool {
	return gb.XBoard.Won()
}

func (gb *GameBoard) OWon() bool {
	return gb.OBoard.Won()
}

type PatternMatch struct {
	Pattern
	Index     uint8
	Shift     uint8
	Direction uint8
}

func directionName(direction uint8) string {
	switch direction {
	case 0:
		return "vertical"
	case 1:
		return "horizontal"
	case 2:
		return "mainDiagonal"
	case 3:
		return "antiDiagonal"
	default:
		return "invalid"
	}
}

func (pm *PatternMatch) Defense(boardSize uint8) []Move {
	var moves []Move = make([]Move, len(pm.Pattern.Defense))

	if pm.Direction == 0 { // vertical
		for i, defense := range pm.Pattern.Defense {
			moves[i] = Move{pm.Shift + defense, pm.Index}
		}
	} else if pm.Direction == 1 { // horizontal
		for i, defense := range pm.Pattern.Defense {
			moves[i] = Move{pm.Index, pm.Shift + defense}
		}
	} else if pm.Direction == 2 { // main diagonal
		for i, defense := range pm.Pattern.Defense {
			moves[i] = Move{pm.Shift + defense, -pm.Shift + pm.Index - defense}
		}
	} else if pm.Direction == 3 { // anti diagonal
		for i, defense := range pm.Pattern.Defense {
			moves[i] = Move{pm.Shift + defense - pm.Index + boardSize - 1, pm.Shift + defense}
		}
	}
	return moves
}

func (pm *PatternMatch) Print() {
	fmt.Printf("Pattern: %v, direction: %v, index: %v, shift: %v\n",
		strconv.FormatUint(uint64(pm.Pattern.Pat), 2),
		directionName(pm.Direction),
		pm.Index,
		pm.Shift,
	)
}

/*
func (pm *PatternMatch) Defense() []Move {
}*/

func (gb *GameBoard) SearchThreats(threats []Pattern, player uint8) []PatternMatch {
	var board1 *Board
	var board2 *Board

	var results []PatternMatch
	results = make([]PatternMatch, 0)

	if player == 0 {
		board1 = &gb.XBoard
		board2 = &gb.OBoard
	} else {
		board1 = &gb.OBoard
		board2 = &gb.XBoard
	}

	for _, threat := range threats {
		for i := uint8(0); i < gb.size; i++ {
			found, shift := threat.MatchWithSpace(board1.vertical[i], board2.vertical[i])
			if found {
				results = append(results, PatternMatch{threat, i, shift, 0})
			}
			found, shift = threat.MatchWithSpace(board1.horizontal[i], board2.horizontal[i])
			if found {
				results = append(results, PatternMatch{threat, i, shift, 1})
			}
			found, shift = threat.MatchWithSpace(board1.mainDiagonal[i], board2.mainDiagonal[i])
			if found {
				results = append(results, PatternMatch{threat, i, shift, 2})
			}
			found, shift = threat.MatchWithSpace(board1.antiDiagonal[i], board2.antiDiagonal[i])
			if found {
				results = append(results, PatternMatch{threat, i, shift, 3})
			}
		}
	}
	return results
}

func (gb *GameBoard) Print() {
	for i := uint8(0); i < gb.size; i++ {
		for j := uint8(0); j < gb.size; j++ {
			if gb.XBoard.vertical[i]&(1<<j) != 0 {
				fmt.Print("X ")
			} else if gb.OBoard.vertical[i]&(1<<j) != 0 {
				fmt.Print("O ")
			} else if gb.nextMoves.vertical[i]&(1<<j) != 0 {
				fmt.Print(". ")
			} else {
				fmt.Print("- ")
			}
		}
		fmt.Println("")
	}
}

func (gb *GameBoard) IsEmpty(x uint8, y uint8) bool {
	return gb.OBoard.IsEmpty(x, y) && gb.XBoard.IsEmpty(x, y)
}

func (gb *GameBoard) Place(x, y uint8, player uint8) {
	if player == 0 {
		gb.XBoard.Place(x, y)
	} else {
		gb.OBoard.Place(x, y)
	}
	var tries [][2]uint8 = [][2]uint8{{x + 1, y}, {x - 1, y}, {x, y + 1}, {x, y - 1},
		{x + 1, y + 1}, {x - 1, y - 1}, {x - 1, y + 1}, {x + 1, y - 1}}

	for _, t := range tries {
		gb.nextMoves.TryPlace(t[0], t[1])
	}
}
