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

func NewGameBoard(size uint8) GameBoard {
	return GameBoard{
		size:      size,
		XBoard:    NewBoard(size),
		OBoard:    NewBoard(size),
		nextMoves: NewBoard(size),
	}
}

func (gb *GameBoard) IsEmpty(x, y uint8) bool {
	return gb.XBoard.IsEmpty(x, y) && gb.OBoard.IsEmpty(x, y)
}

func (gb *GameBoard) PossibleMoves() []Move {
	var moves []Move
	var pattern uint64
	var i, j uint8

	for i = 0; i < gb.size; i++ {
		if gb.nextMoves.vertical[i] != 0 { // any value is present on this column
			// test all rows until the pattern > vertical[i]
			for j, pattern = 0, 1; pattern <= gb.nextMoves.vertical[i] && j < gb.size; j, pattern = j+1, pattern<<1 {
				if gb.nextMoves.vertical[i]&pattern == pattern {
					moves = append(moves, Move{X: j, Y: i})
				}
			}
		}
	}

	return moves
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
			// moves[i] = Move{pm.Shift + defense - pm.Index + boardSize - 1, pm.Shift + defense}
			moves[i] = Move{pm.Shift + defense, pm.Shift + defense - pm.Index + boardSize - 1}
		}
	}
	return moves
}

func (pm *PatternMatch) Print() {
	fmt.Printf("Pattern: %v (%v), direction: %v, index: %v, shift: %v\n",
		strconv.FormatUint(uint64(pm.Pattern.Pat), 2),
		pm.Pattern.Value,
		directionName(pm.Direction),
		pm.Index,
		pm.Shift,
	)
}

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
		// fmt.Println("Searching threat:", threat)
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

			j := i + board1.size
			found, shift = threat.MatchWithSpace(board1.mainDiagonal[j], board2.mainDiagonal[j])
			if found {
				results = append(results, PatternMatch{threat, j, shift, 2})
			}
			found, shift = threat.MatchWithSpace(board1.antiDiagonal[j], board2.antiDiagonal[j])
			if found {
				results = append(results, PatternMatch{threat, j, shift, 3})
			}
		}
	}
	return results
}

func (gb *GameBoard) Print() {
	fmt.Println(". 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1")
	for i := uint8(0); i < gb.size; i++ {
		fmt.Printf("%v ", i%10)
		for j := uint8(0); j < gb.size; j++ {
			//if gb.XBoard.horizontal[j]&(1<<i) != 0 {
			if gb.XBoard.vertical[i]&(1<<j) != 0 {
				fmt.Print("X ")
				//} else if gb.OBoard.horizontal[j]&(1<<i) != 0 {
			} else if gb.OBoard.vertical[i]&(1<<j) != 0 {
				fmt.Print("O ")
			} else if gb.nextMoves.vertical[i]&(1<<j) != 0 {
				//} else if gb.nextMoves.horizontal[j]&(1<<i) != 0 {
				fmt.Print("_ ")
			} else {
				fmt.Print(". ")
			}
		}
		fmt.Println("")
	}
}

func (gb *GameBoard) Place(x, y uint8, player uint8) {
	if player == 0 {
		gb.XBoard.Place(x, y)
	} else {
		gb.OBoard.Place(x, y)
	}
	gb.nextMoves.Unplace(x, y)

	var tries [][2]uint8 = [][2]uint8{{x + 1, y}, {x - 1, y}, {x, y + 1}, {x, y - 1},
		{x + 1, y + 1}, {x - 1, y - 1}, {x - 1, y + 1}, {x + 1, y - 1}}

	for _, t := range tries {
		if t[0] == 255 || t[1] == 255 {
			continue // overflow
		}
		if !gb.XBoard.Taken(t[0], t[1]) && !gb.OBoard.Taken(t[0], t[1]) {
			gb.nextMoves.Place(t[0], t[1])
		}
	}
}

/* Unplace removes a move from the board but leaves the nextMoves intact making it invalid */
func (gb *GameBoard) Unplace(x uint8, y uint8) {
	gb.XBoard.Unplace(x, y)
	gb.OBoard.Unplace(x, y)
}

func (gb *GameBoard) Copy() GameBoard {
	//return NewGameBoard(gb.size)
	return GameBoard{gb.size, gb.XBoard.Copy(), gb.OBoard.Copy(), gb.nextMoves.Copy()}
}

/*
	gb := pisk.NewGameBoard(32)
	gb.Place(10, 10, 0)
	gb.Place(11, 10, 0)
	gb.Place(12, 10, 0)
	gb.Place(10, 11, 1)
	gb.Print()
*/
