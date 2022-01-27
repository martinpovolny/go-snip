package main

import (
	"bufio"
	"fmt"
	"martinp/piskvorky/pisk"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

type Game struct {
	play *pisk.GameLog
	gb   *pisk.GameBoard
}

const boardSize = 32

func NewGame(xstarts bool) *Game {
	return &Game{
		play: pisk.NewGameLog(xstarts),
		gb:   pisk.NewGameBoard(boardSize),
	}
}

func (g *Game) Play(move pisk.Move, player uint8) bool {
	if !g.gb.IsEmpty(move.X, move.Y) {
		return false
	}
	g.play.Add(move)
	g.gb.Place(move.X, move.Y, player)
	return true
}

func (g *Game) LoadFromFile(filename string) int {
	g.play.LoadFromFile(filename)
	var player uint8 = 0
	for _, move := range g.play.Moves {
		g.gb.Place(move.X, move.Y, player)
		player = 1 - player
	}
	return len(g.play.Moves)
}

var threatPatterns []pisk.Pattern = []pisk.Pattern{
	{
		Pat:     0b001110,
		Space:   0b110001, // fixme: need 2 spaces on either side
		NShifts: 28,
		Value:   2,
		Defense: []uint8{0, 4},
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
		Value:   10,
		Defense: []uint8{0, 5},
	},
}

func readIntFromStdin() uint8 {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSuffix(text, "\n")
	num, _ := strconv.ParseUint(text, 10, 32)
	return uint8(num)
}

func readMoveFromInput(player uint8) pisk.Move {
	fmt.Printf("Player %v, enter move: \n", player)
	x := readIntFromStdin()
	y := readIntFromStdin()
	return pisk.Move{X: x, Y: y}
}

func printMatches(threats []pisk.PatternMatch) {
	if len(threats) > 0 {
		fmt.Println("Threats:")
		for _, t := range threats {
			t.Print()
			defenses := t.Defense(boardSize)
			fmt.Println("Defenses:", defenses)
		}
	}
}

func interactiveGameRound(game *Game, player uint8) (bool, uint8) {
	for {
		move := readMoveFromInput(player)
		if game.Play(move, player) {
			break
		}
		fmt.Println("Invalid move:", move)
	}
	game.gb.Print()
	threats := game.gb.SearchThreats(threatPatterns, player)
	printMatches(threats)

	if player == 0 {
		if game.gb.XWon() {
			fmt.Println("X won!")
			return true, 0
		}
	} else {
		if game.gb.OWon() {
			fmt.Println("O won!")
			return true, 1
		}
	}
	return false, 0
}

func saveAndExit(game *Game, initialMoves int) {
	if len(game.play.Moves) <= initialMoves {
		os.Exit(0)
	}
	filename := fmt.Sprintf("./games/game-%v.log", time.Now().Unix())
	game.play.SaveToFile(filename)
	fmt.Printf("Game saved in %s.\n", filename)
	os.Exit(0)
}

func main() {
	/*
		b := pisk.NewBoard(32)
		b.Place(10, 10)
		b.Print()

		fmt.Println(strconv.FormatUint(uint64(threatPatterns[0].Pat), 2))
		fmt.Println(strconv.FormatUint(uint64(threatPatterns[0].Pat)<<1, 2))
		fmt.Println(strconv.FormatUint(uint64(threatPatterns[0].Pat)<<3, 2))
		fmt.Println(strconv.FormatUint(uint64(threatPatterns[0].Pat)<<8, 2))

		gb := pisk.NewGameBoard(32)
		gb.Place(10, 10, 0)
		gb.Place(11, 10, 0)
		gb.Place(12, 10, 0)
		gb.Place(10, 11, 1)
		gb.Print()*/

	game := NewGame(true)
	loadedMoves := 0

	if len(os.Args) == 3 && os.Args[1] == "load" {
		fmt.Println("Loading game from ", os.Args[2])
		loadedMoves = game.LoadFromFile(os.Args[2])
	}

	var player uint8 = game.play.NextPlayer()
	game.gb.Print()
	threats := game.gb.SearchThreats(threatPatterns, 1-player)
	printMatches(threats)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			saveAndExit(game, loadedMoves)
		}
	}()

	for {
		win, _ := interactiveGameRound(game, player)
		if win {
			break
		}
		player = 1 - player
	}
}
