package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

const (
	rows    = 6
	cols    = 7
	empty   = 0
	player1 = 1
	player2 = 2
	winLen  = 3
	aiDepth = 7
)

type Board [rows][cols]int

func (b *Board) drop(col, piece int) (int, bool) {
	for r := rows - 1; r >= 0; r-- {
		if b[r][col] == empty {
			b[r][col] = piece
			return r, true
		}
	}
	return -1, false
}

func (b *Board) undo(col int) {
	for r := 0; r < rows; r++ {
		if b[r][col] != empty {
			b[r][col] = empty
			return
		}
	}
}

func (b *Board) checkWin(row, col, piece int) bool {
	dirs := [4][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}}
	for _, d := range dirs {
		count := 1
		for i := 1; i < winLen; i++ {
			r, c := row+d[0]*i, col+d[1]*i
			if r < 0 || r >= rows || c < 0 || c >= cols || b[r][c] != piece {
				break
			}
			count++
		}
		for i := 1; i < winLen; i++ {
			r, c := row-d[0]*i, col-d[1]*i
			if r < 0 || r >= rows || c < 0 || c >= cols || b[r][c] != piece {
				break
			}
			count++
		}
		if count >= winLen {
			return true
		}
	}
	return false
}

func (b *Board) isFull() bool {
	for c := 0; c < cols; c++ {
		if b[0][c] == empty {
			return false
		}
	}
	return true
}

func (b *Board) validCols() []int {
	valid := make([]int, 0, cols)
	// search from center outward for better pruning
	for i := 0; i < cols; i++ {
		c := cols/2 + (1-2*(i%2))*(i+1)/2
		if c >= 0 && c < cols && b[0][c] == empty {
			valid = append(valid, c)
		}
	}
	return valid
}

func minimax(b *Board, depth, alpha, beta int, maximizing bool) int {
	if b.isFull() {
		return 0
	}
	if depth == 0 {
		return 0
	}
	valid := b.validCols()
	if maximizing {
		best := math.MinInt32
		for _, col := range valid {
			row, _ := b.drop(col, player2)
			if b.checkWin(row, col, player2) {
				b.undo(col)
				return 1000 + depth
			}
			score := minimax(b, depth-1, alpha, beta, false)
			b.undo(col)
			if score > best {
				best = score
			}
			if best > alpha {
				alpha = best
			}
			if beta <= alpha {
				break
			}
		}
		return best
	}
	best := math.MaxInt32
	for _, col := range valid {
		row, _ := b.drop(col, player1)
		if b.checkWin(row, col, player1) {
			b.undo(col)
			return -1000 - depth
		}
		score := minimax(b, depth-1, alpha, beta, true)
		b.undo(col)
		if score < best {
			best = score
		}
		if best < beta {
			beta = best
		}
		if beta <= alpha {
			break
		}
	}
	return best
}

func aiMove(b *Board) int {
	valid := b.validCols()
	// immediate win
	for _, col := range valid {
		row, _ := b.drop(col, player2)
		win := b.checkWin(row, col, player2)
		b.undo(col)
		if win {
			return col
		}
	}
	// block immediate loss
	for _, col := range valid {
		row, _ := b.drop(col, player1)
		win := b.checkWin(row, col, player1)
		b.undo(col)
		if win {
			return col
		}
	}
	bestScore := math.MinInt32
	bestCol := valid[0]
	for _, col := range valid {
		row, _ := b.drop(col, player2)
		score := minimax(b, aiDepth, math.MinInt32, math.MaxInt32, false)
		b.undo(col)
		_ = row
		if score > bestScore {
			bestScore = score
			bestCol = col
		}
	}
	return bestCol
}

func (b *Board) print() {
	fmt.Println()
	for r := 0; r < rows; r++ {
		fmt.Print("|")
		for c := 0; c < cols; c++ {
			switch b[r][c] {
			case player1:
				fmt.Print("\033[31mX\033[0m|")
			case player2:
				fmt.Print("\033[34mO\033[0m|")
			default:
				fmt.Print(" |")
			}
		}
		fmt.Println()
	}
	fmt.Print("+")
	for c := 0; c < cols; c++ {
		_ = c
		fmt.Print("-+")
	}
	fmt.Println()
	fmt.Print(" ")
	for c := 1; c <= cols; c++ {
		fmt.Printf("%d ", c)
	}
	fmt.Println("\n")
}

func readLine(r *bufio.Reader) string {
	s, _ := r.ReadString('\n')
	return strings.TrimSpace(s)
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== Connect 3 ===")
	fmt.Println("Get 3 in a row (horizontal, vertical, or diagonal) to win!")
	fmt.Println()
	fmt.Println("1. Human vs Human")
	fmt.Println("2. Human vs Computer")
	fmt.Print("\nChoose mode [1/2]: ")

	vsAI := readLine(reader) == "2"

	names := [3]string{"", "\033[31mPlayer 1 (X)\033[0m", "\033[34mPlayer 2 (O)\033[0m"}
	if vsAI {
		names[2] = "\033[34mComputer (O)\033[0m"
	}

	var board Board
	turn := player1

	for {
		board.print()

		if vsAI && turn == player2 {
			fmt.Println("Computer is thinking...")
			col := aiMove(&board)
			row, _ := board.drop(col, player2)
			fmt.Printf("Computer plays column %d\n", col+1)
			if board.checkWin(row, col, player2) {
				board.print()
				fmt.Printf("%s wins!\n", names[player2])
				return
			}
		} else {
			fmt.Printf("%s — choose column (1-%d): ", names[turn], cols)
			input := readLine(reader)
			col, err := strconv.Atoi(input)
			if err != nil || col < 1 || col > cols {
				fmt.Printf("  Invalid — enter a number between 1 and %d.\n", cols)
				continue
			}
			col--
			row, ok := board.drop(col, turn)
			if !ok {
				fmt.Println("  Column is full — choose another.")
				continue
			}
			if board.checkWin(row, col, turn) {
				board.print()
				fmt.Printf("%s wins!\n", names[turn])
				return
			}
		}

		if board.isFull() {
			board.print()
			fmt.Println("It's a draw!")
			return
		}

		if turn == player1 {
			turn = player2
		} else {
			turn = player1
		}
	}
}
