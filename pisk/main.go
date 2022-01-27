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

func NewGame(xstarts bool) *Game {
	return &Game{
		play: pisk.NewGameLog(xstarts),
		gb:   pisk.NewGameBoard(32),
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

var threatPatterns []pisk.Pattern = []pisk.Pattern{
	{
		Pat:     0b00000000000000000000000000001110,
		Space:   0b00000000000000000000000000110001,
		NShifts: 28,
		Value:   2,
	},
	{
		Pat:     0b00000000000000000000000000010110,
		Space:   0b00000000000000000000000000101001,
		NShifts: 26,
		Value:   2,
	},
	{
		Pat:     0b00000000000000000000000000011010,
		Space:   0b00000000000000000000000000100101,
		NShifts: 26,
		Value:   2,
	},
	{
		Pat:     0b00000000000000000000000000011110,
		Space:   0b00000000000000000000000000100001,
		NShifts: 26,
		Value:   2,
	},
}

func readIntFromStdin() uint8 {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSuffix(text, "\n")
	num, _ := strconv.ParseUint(text, 10, 32)
	return uint8(num)
}

func readMoveFromInput() pisk.Move {
	fmt.Print("Enter move: ")
	x := readIntFromStdin()
	y := readIntFromStdin()
	return pisk.Move{X: x, Y: y}
}

func printMatches(threats []pisk.PatternMatch) {
	if len(threats) > 0 {
		fmt.Println("Threats:")
		for _, t := range threats {
			t.Print()
		}
	}
}

func main() {
	fmt.Println("Hello, World!")

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
	gb.Print()

	game := NewGame(true)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			fmt.Println("\nSave game y/n?")
			text := readIntFromStdin()
			if text == 'y' {
				filename := fmt.Sprintf("game-%v.log", time.Now().Unix())
				game.play.SaveToFile(filename)
				fmt.Printf("Game saved in %s.\n", filename)
			}
		}
	}()

	for {
		for {
			move := readMoveFromInput()
			if game.Play(move, 0) {
				break
			}
			fmt.Println("Invalid move:", move)
		}
		game.gb.Print()
		threats := game.gb.SearchThreats(threatPatterns, 0)
		printMatches(threats)

		if game.gb.XWon() {
			fmt.Println("X won!")
			break
		}

		for {
			move := readMoveFromInput()
			if game.Play(move, 1) {
				break
			}
			fmt.Println("Invalid move:", move)
		}
		game.gb.Print()
		threats = game.gb.SearchThreats(threatPatterns, 1)
		printMatches(threats)

		if game.gb.OWon() {
			fmt.Println("O won!")
			break
		}
	}
}
