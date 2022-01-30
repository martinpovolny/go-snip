package pisk

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
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

func (gl *GameLog) LoadFromFile(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") { // skip comments
			continue
		}
		parts := strings.Split(line, " ") // each line has "x y"
		x, err := strconv.Atoi(parts[0])
		if err != nil {
			log.Fatal(err)
		}
		y, err := strconv.Atoi(parts[1])
		if err != nil {
			log.Fatal(err)
		}
		gl.Add(Move{uint8(x), uint8(y)})
	}
}

func (gl *GameLog) NextPlayer() uint8 {
	return uint8(len(gl.Moves) % 2)
}
