package pisk

import "fmt"

type Board struct {
	size         uint8
	vertical     []uint64
	horizontal   []uint64
	mainDiagonal []uint64
	antiDiagonal []uint64
}

func NewBoard(size uint8) Board {
	b := Board{size, make([]uint64, size), make([]uint64, size), make([]uint64, size*2), make([]uint64, size*2)}
	for i := uint8(0); i < size; i++ {
		b.vertical[i] = 0
		b.horizontal[i] = 0
		b.mainDiagonal[i] = 0
		b.antiDiagonal[i] = 0
	}
	return b
}

func (b *Board) Won() bool {
	for i := uint8(0); i < b.size; i++ {
		if WinningPattern.Match(b.vertical[i]) ||
			WinningPattern.Match(b.horizontal[i]) ||
			WinningPattern.Match(b.mainDiagonal[i]) ||
			WinningPattern.Match(b.antiDiagonal[i]) {
			return true
		}
	}
	return false
}

func (b *Board) IsEmpty(x uint8, y uint8) bool {
	return b.vertical[y]&(1<<x) == 0
}

func (b *Board) Print() {
	for i := uint8(0); i < b.size; i++ {
		fmt.Printf("%d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d\n",
			b.vertical[i]>>0&1,
			b.vertical[i]>>1&1,
			b.vertical[i]>>2&1,
			b.vertical[i]>>3&1,
			b.vertical[i]>>4&1,
			b.vertical[i]>>5&1,
			b.vertical[i]>>6&1,
			b.vertical[i]>>7&1,
			b.vertical[i]>>8&1,
			b.vertical[i]>>9&1,
			b.vertical[i]>>10&1,
			b.vertical[i]>>11&1,
			b.vertical[i]>>12&1,
			b.vertical[i]>>13&1,
			b.vertical[i]>>14&1,
			b.vertical[i]>>15&1,
			b.vertical[i]>>16&1,
			b.vertical[i]>>17&1,
			b.vertical[i]>>18&1,
			b.vertical[i]>>19&1,
			b.vertical[i]>>20&1,
			b.vertical[i]>>21&1,
			b.vertical[i]>>22&1,
			b.vertical[i]>>23&1,
			b.vertical[i]>>24&1,
			b.vertical[i]>>25&1,
			b.vertical[i]>>26&1,
			b.vertical[i]>>27&1,
			b.vertical[i]>>28&1,
			b.vertical[i]>>29&1,
			b.vertical[i]>>30&1,
			b.vertical[i]>>31&1)
	}
}

func (b *Board) Place(x, y uint8) {
	b.vertical[y] |= 1 << x
	b.horizontal[x] |= 1 << y
	b.mainDiagonal[x+y] |= 1 << x
	b.antiDiagonal[x-y+b.size-1] |= 1 << x
}

func (b *Board) Unplace(x, y uint8) {
	b.vertical[y] &= ^(1 << x)
	b.horizontal[x] &= ^(1 << y)
	b.mainDiagonal[x+y] &= ^(1 << x)
	b.antiDiagonal[x-y+b.size-1] &= ^(1 << x)
}

func (b *Board) Taken(x, y uint8) bool {
	return b.vertical[y]&(1<<x) != 0
}

func (b *Board) TryPlace(x, y uint8) {
	if x > 0 && x < b.size && y > 0 && y < b.size {
		if !b.Taken(x, y) {
			b.Place(x, y)
		}
	}
}

func (b *Board) Copy() Board {
	var board = NewBoard(b.size)
	copy(board.vertical, b.vertical)
	copy(board.horizontal, b.horizontal)
	copy(board.mainDiagonal, b.mainDiagonal)
	copy(board.antiDiagonal, b.antiDiagonal)
	return board
}

/*
	b := pisk.NewBoard(32)
	b.Place(10, 10)
	b.Print()
*/
