package pisk

type Pattern struct {
	Pat     uint64
	Space   uint64
	NShifts uint8
	Value   uint8
}

func (p Pattern) Match(xs uint64) bool {
	match, _ := p.MatchIndex(xs)
	return match
}

func (p Pattern) MatchIndex(xs uint64) (bool, uint8) {
	for i := 0; i < int(p.NShifts); i++ {
		if (xs & p.Pat) == p.Pat {
			return true, uint8(i)
		}
		p.Pat = p.Pat << 1
	}
	return false, 0
}

func (p Pattern) MatchWithSpace(xs uint64, os uint64) (bool, uint8) {
	for i := 0; i < int(p.NShifts); i++ {
		if (xs&p.Pat) == p.Pat && // crosses are where expected
			((^os)&p.Space) == p.Space { // spaces (not os) is where expected

			return true, uint8(i)
		}
		p.Pat = p.Pat << 1
		p.Space = p.Space << 1
	}
	return false, 0
}

var WinningPattern = Pattern{
	Pat:     uint64(0b11111),
	NShifts: 26,
	Value:   100,
}
