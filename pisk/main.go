package main

import (
	"bufio"
	"fmt"
	"log"
	"martinp/piskvorky/client"
	"martinp/piskvorky/pisk"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

const boardSize = 32

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

func printMatches(threats []pisk.PatternMatch, player uint8) {
	if len(threats) > 0 {
		fmt.Println("Threats by player", player)
		for _, t := range threats {
			t.Print()
			defenses := t.Defense(boardSize)
			fmt.Printf("\tDefenses: %v\n", defenses)
		}
	}
}

var strategy pisk.Depth1Strategy = pisk.Depth1Strategy{}

func interactiveGameRound(game *pisk.Game, player uint8) (bool, uint8) {
	game.Board.Print()
	threats := game.Board.SearchThreats(pisk.ThreatPatterns, player)
	printMatches(threats, player)
	threats = game.Board.SearchThreats(pisk.ThreatPatterns, 1-player)
	printMatches(threats, 1-player)
	//move, score := strategy.NextMove(&game.Board, player)
	//move, score := strategy.AttackMove(&game.Board, player)
	//fmt.Printf("Computer move: %v, score: %v\n", move, score)

	for {
		move := readMoveFromInput(player)
		if game.Play(move, player) {
			break
		}
		fmt.Println("Invalid move:", move)
	}

	if player == 0 {
		if game.Board.XWon() {
			game.Board.Print()
			fmt.Println("X won!")
			return true, 0
		}
	} else {
		if game.Board.OWon() {
			game.Board.Print()
			fmt.Println("O won!")
			return true, 1
		}
	}
	return false, 0
}

func saveAndExit(game *pisk.Game, initialMoves int) {
	if len(game.Log.Moves) <= initialMoves {
		os.Exit(0)
	}
	filename := fmt.Sprintf("./games/game-%v.log", time.Now().Unix())
	game.Log.SaveToFile(filename)
	fmt.Printf("Game saved in %s.\n", filename)
	os.Exit(0)
}

func loadCredentials(rp *client.RemotePlay) {
	err := rp.LoadCredentialsFromConfigFile("./credentials")
	if err != nil {
		fmt.Println("Error loading credentials:", err)
		os.Exit(1)
	}
}

func main() {
	var game *pisk.Game
	loadedMoves := 0

	if len(os.Args) == 3 && os.Args[1] == "load" {
		fmt.Println("Loading game from ", os.Args[2])
		game = pisk.NewGame(boardSize, true)
		loadedMoves = game.LoadFromFile(os.Args[2])
	} else if len(os.Args) == 4 && os.Args[1] == "load-remote" {
		var err error
		var finished bool

		fmt.Println("Loading remote game", os.Args[2], os.Args[3])
		rp := client.NewRemotePlay()
		loadCredentials(&rp)
		rp.SetGame(os.Args[2], os.Args[3])

		finished, game, err = rp.LoadGame()
		if err != nil {
			fmt.Println("Error loading game: ", err)
			os.Exit(1)
		}
		if finished {
			fmt.Println("Game finished.")
		}
	} else if len(os.Args) == 2 && os.Args[1] == "new-remote" {
		fmt.Println("Starting new remote game")
		rp := client.NewRemotePlay()
		loadCredentials(&rp)
		game, err := rp.StartGame()
		if err != nil {
			log.Fatalf("failed to start a remote game: %v", err)
			os.Exit(1)
		}
		log.Println("Started game", game.GameId, game.GameToken)
		os.Exit(0)
	} else {
		fmt.Println("New local game")
		game = pisk.NewGame(boardSize, true)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			saveAndExit(game, loadedMoves)
		}
	}()

	var player uint8 = game.Log.NextPlayer()
	for {
		win, _ := interactiveGameRound(game, player)
		if win {
			break
		}
		player = 1 - player
	}
}
