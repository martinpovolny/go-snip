package pisk

import (
	"fmt"
	"log"
	"math/rand"
)

const MaxValue = 255
const MustDefend = 100

type Depth1Strategy struct{}

var ThreatPatterns []Pattern = []Pattern{
	{
		Pat:     0b001110,
		Space:   0b110001, // fixme: need 2 spaces on either side
		NShifts: 28,
		Value:   2,
		Defense: []uint8{0, 4},
	},
	{
		Pat:     0b011100,
		Space:   0b100011, // fixme: need 2 spaces on either side
		NShifts: 28,
		Value:   2,
		Defense: []uint8{1, 5},
	},
	{
		Pat:     0b010110,
		Space:   0b101001,
		NShifts: 26,
		Value:   2,
		Defense: []uint8{0, 3},
	},
	{
		Pat:     0b011010,
		Space:   0b100101,
		NShifts: 26,
		Value:   2,
		Defense: []uint8{2, 0, 5},
	},
	{
		Pat:     0b011110,
		Space:   0b100001,
		NShifts: 26,
		Value:   128, // >> MustDefend
		Defense: []uint8{0, 5},
	},
	{
		Pat:     0b01111, // only one space, FIXME? check that the 2nd space is not present?
		Space:   0b10000,
		NShifts: 27,
		Value:   101, // >> MustDefend
		Defense: []uint8{4},
	},
	{
		Pat:     0b11110,
		Space:   0b00001,
		NShifts: 27,
		Value:   101, // >> MustDefend
		Defense: []uint8{0},
	},
	{
		Pat:     0b11111,
		Space:   0b00000,
		NShifts: 27,
		Value:   MaxValue,
		Defense: []uint8{},
	},
}

func (s Depth1Strategy) AttackMove(gb *GameBoard, player uint8) (Move, uint8) {
	moves := gb.PossibleMoves()
	//moves := []Move{{7, 4}}
	fmt.Printf("Possible move for player %v : %v\n", player, moves)
	if len(moves) == 0 {
		return Move{gb.size / 2, gb.size / 2}, 0
	}

	var bestMove Move
	var bestScore, score, bestValue uint8

	var testGb GameBoard = gb.Copy()

	for _, move := range moves {
		//fmt.Println("trying move: ", move)
		testGb.Place(move.X, move.Y, player)
		//testGb.Print()

		matches := testGb.SearchThreats(ThreatPatterns, player)
		bestValue = 0
		for _, match := range matches {
			match.Print()
			if match.Pattern.Value > bestValue {
				bestValue = match.Pattern.Value
			}
			score = bestValue
		}

		if score > bestScore {
			bestScore = score
			bestMove = move
			fmt.Println("bestMove: ", bestMove, bestScore)
			if score == MaxValue { // no point searching further
				return bestMove, bestScore
			}
		}

		testGb.Unplace(move.X, move.Y)
	}
	if bestScore > 0 {
		return bestMove, bestScore
	}

	// FIXME: We would like a better naive selection here, such as playon on a diagonal.
	fmt.Println("No attack move found, returning random move")
	return moves[rand.Intn(len(moves))], 0
}

func (s Depth1Strategy) NextMove(gb *GameBoard, player uint8) (Move, uint8) {
	attackMove, attackScore := s.AttackMove(gb, player)
	if attackScore == MaxValue {
		return attackMove, MaxValue // the winning move, no thinking needed
	}
	//threats := []PatternMatch{} // gb.SearchThreats(ThreatPatterns, player)
	b := gb.Copy() // copy should not be necessary, but it is --> there's a bug in the searchThreats function
	threats := b.SearchThreats(ThreatPatterns, 1-player)
	fmt.Printf("Threats by player %v : %v\n", 1-player, len(threats))
	for _, t := range threats {
		t.Print()
	}

	/* Evaluate threats */
	var defensiveMoves []Move = []Move{}
	var worstThreat PatternMatch
	var defenseValue uint8

	for _, match := range threats {
		// fmt.Println("evaluating threat")
		// match.Print()
		if match.Pattern.Value > defenseValue {
			defenseValue = match.Pattern.Value
			worstThreat = match

			if defenseValue == MaxValue { // the ultimate thread, a 5, needs to be defended
				// assuming there's only one worst threat
				defensiveMoves = worstThreat.Defense(gb.size)

				// There should be only one defensive move here.
				if len(defensiveMoves) != 1 {
					log.Println("Error: number of defensiveMoves is not 1")
				}
				return defensiveMoves[rand.Intn(len(defensiveMoves))], MaxValue
			}
		}
	}

	// We are going to defend the worst threat
	// disregarding the fact that there may be multiple threats of the same
	// value that share the defensive move.
	// FIXME: we should consider all threats of the same value and find a common defense if it exists.
	// We should also consider the attack value of the defensive move.
	defensiveMoves = worstThreat.Defense(gb.size)
	fmt.Println("defensiveMoves: ", defensiveMoves)

	if attackScore == MaxValue || (attackScore > defenseValue && defenseValue < MustDefend) || len(defensiveMoves) == 0 {
		// our attack has higher value than the worst threat, so we will play it
		return attackMove, attackScore
	} else {
		// goint to play a defensive move that prevents a threat of a highter value that is the best value of our move
		return defensiveMoves[rand.Intn(len(defensiveMoves))], defenseValue
	}
}
