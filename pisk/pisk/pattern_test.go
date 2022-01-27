package pisk_test

import (
	"martinp/piskvorky/pisk"
	"testing"
)

func TestMatch(t *testing.T) {
	type testCase struct {
		input  uint64
		output bool
	}

	type testPattern struct {
		pattern pisk.Pattern
		cases   []testCase
	}

	var tests []testPattern = []testPattern{{
		pattern: pisk.Pattern{
			Pat:     0b00000000000000000000000000001110,
			Space:   0b00000000000000000000000000110001,
			NShifts: 28,
			Value:   2,
		},
		cases: []testCase{
			{input: 0b00000000000000000000000000001110, output: true},
			{input: 0b00000000000000000000000000001100, output: false},
			{input: 0b00000000000000000000000011100000, output: true},
			{input: 0b01110000000000000000000000000000, output: true},
			{input: 0b11100000000000000000000000000000, output: false},
		},
	},
	}

	for _, tp := range tests {
		for _, tc := range tp.cases {
			if tp.pattern.Match(tc.input) != tc.output {
				t.Errorf("Match(%v) != %v", tc.input, tc.output)
			}
		}
	}
}
