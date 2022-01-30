package pisk

type Game struct {
	Log   *GameLog
	Board GameBoard
}

func NewGame(boardSize uint8, xstarts bool) *Game {
	return &Game{
		Log:   NewGameLog(xstarts),
		Board: NewGameBoard(boardSize),
	}
}

func (g *Game) Play(move Move, player uint8) bool {
	if move.X >= g.Board.size || move.Y >= g.Board.size {
		return false
	}
	if !g.Board.IsEmpty(move.X, move.Y) {
		return false
	}
	g.Log.Add(move)
	g.Board.Place(move.X, move.Y, player)
	return true
}

func (g *Game) LoadFromFile(filename string) int {
	g.Log.LoadFromFile(filename)
	var player uint8 = 0
	for _, move := range g.Log.Moves {
		g.Board.Place(move.X, move.Y, player)
		player = 1 - player
	}
	return len(g.Log.Moves)
}
