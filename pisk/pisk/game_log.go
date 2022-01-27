package pisk

import (
	"fmt"
	"log"
	"os"
)

type Move struct {
	X uint8
	Y uint8
}

type GameLog struct {
	XStarts bool
	Moves   []Move
}

func NewGameLog(xstarts bool) *GameLog {
	return &GameLog{
		XStarts: xstarts,
		Moves:   make([]Move, 0),
	}
}

func (gl *GameLog) Add(move Move) {
	gl.Moves = append(gl.Moves, move)
}

/* SaveToFile stores moves in a file */
func (gl *GameLog) SaveToFile(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	for _, move := range gl.Moves {
		fmt.Fprintf(f, "%d %d\n", move.X, move.Y)
	}
}
