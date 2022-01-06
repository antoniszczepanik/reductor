package main

import (
	"fmt"
	"testing"
)

func Test_getLongestMatchPosAndLen(t *testing.T) {
	var tests = []struct {
		name          string
		searchBuff    []byte
		lookaheadBuff []byte
		wantPos       int
		wantLen       int
		minMatchLen   int
	}{
		{
			name:          "Empty search buffer",
			searchBuff:    []byte(""),
			lookaheadBuff: []byte("hijklmno"),
			wantPos:       0,
			wantLen:       0,
			minMatchLen:   0,
		},
		{
			name:          "Empty lookahead buffer",
			searchBuff:    []byte("abcdefg"),
			lookaheadBuff: []byte(""),
			wantPos:       0,
			wantLen:       0,
			minMatchLen:   0,
		},
		{
			name:          "No matches in lookahead buffer",
			searchBuff:    []byte("abcdefg"),
			lookaheadBuff: []byte("hijklmno"),
			wantPos:       0,
			wantLen:       0,
			minMatchLen:   0,
		},
		{
			name:          "Full match",
			searchBuff:    []byte("abcdefg"),
			lookaheadBuff: []byte("abcdefg"),
			wantPos:       0,
			wantLen:       7,
			minMatchLen:   0,
		},
		{
			name:          "Half match in search buffer",
			searchBuff:    []byte("abc"),
			lookaheadBuff: []byte("abcdefg"),
			wantPos:       0,
			wantLen:       3,
			minMatchLen:   0,
		},
		{
			name:          "Half match in lookahed buffer",
			searchBuff:    []byte("abcdefg"),
			lookaheadBuff: []byte("abc"),
			wantPos:       0,
			wantLen:       3,
			minMatchLen:   0,
		},
		{
			name:          "Second half of search buff matches",
			searchBuff:    []byte("efgabc"),
			lookaheadBuff: []byte("abc"),
			wantPos:       3,
			wantLen:       3,
			minMatchLen:   0,
		},
		{
			name:          "Two matches, but the first one is longer",
			searchBuff:    []byte("milk milk"),
			lookaheadBuff: []byte("milk "),
			wantPos:       0,
			wantLen:       5,
			minMatchLen:   0,
		},
		{
			name:          "Full match shorter than minMatchLen",
			searchBuff:    []byte("abcdefgh"),
			lookaheadBuff: []byte("abcdefgh "),
			wantPos:       0,
			wantLen:       0,
			minMatchLen:   9,
		},
		{
			name:          "Some random match in the middle",
			searchBuff:    []byte("abcd peace efgh"),
			lookaheadBuff: []byte(" peace abcd "),
			wantPos:       4,
			wantLen:       7,
			minMatchLen:   5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, l := getLongestMatchPosAndLen(tt.searchBuff, tt.lookaheadBuff, tt.minMatchLen)
			if p != tt.wantPos {
				t.Errorf("unexpected pos (got %d want %d)", p, tt.wantPos)
			}
			if l != 0 && l < tt.minMatchLen {
				t.Errorf("found match shorter than expected (got len %d min len %d)", l, tt.minMatchLen)
			}
			if l != tt.wantLen {
				t.Errorf("unexpected len (got %d want %d)", l, tt.wantLen)
			}
		})
	}
}

func Test_bytesToValues(t *testing.T) {
	var tests = []struct {
		name             string
		input            []byte
		maxMatchLen      int
		maxSearchBuffLen int
		minMatchLen      int
		wantValuesRepr   string
	}{
		{
			name:             "No matches at all",
			input:            []byte("abcd"),
			maxMatchLen:      999,
			maxSearchBuffLen: 999,
			minMatchLen:      0,
			wantValuesRepr:   "abcd",
		},
		{
			name:             "Match at the end",
			input:            []byte("abcd abcd"),
			maxMatchLen:      999,
			maxSearchBuffLen: 999,
			minMatchLen:      0,
			wantValuesRepr:   "abcd <5,4>",
		},
		{
			name:             "Match in the middle",
			input:            []byte("abcd abcd ghij"),
			maxMatchLen:      999,
			maxSearchBuffLen: 999,
			minMatchLen:      0,
			wantValuesRepr:   "abcd <5,5>ghij",
		},
		{
			name:             "Two matches",
			input:            []byte("XXabXXcdXX"),
			maxMatchLen:      999,
			maxSearchBuffLen: 999,
			minMatchLen:      2,
			// The matches are the same length - first one is selected.
			wantValuesRepr: "XXab<4,2>cd<8,2>",
		},
		{
			name:             "Three matches",
			input:            []byte("XXabXXcdXXijXX"),
			maxMatchLen:      999,
			maxSearchBuffLen: 999,
			minMatchLen:      2,
			// The matches are the same length - first one is selected.
			wantValuesRepr: "XXab<4,2>cd<8,2>ij<12,2>",
		},
		{
			name:             "A match, almost too long",
			input:            []byte("XXXabcdXXX"),
			maxMatchLen:      3,
			maxSearchBuffLen: 999,
			minMatchLen:      3,
			wantValuesRepr:   "XXXabcd<7,3>",
		},
		{
			name:             "A match, too long but is not consumed",
			input:            []byte("XXXXabcdXXXX"),
			maxMatchLen:      3,
			maxSearchBuffLen: 999,
			minMatchLen:      3,
			wantValuesRepr:   "XXXXabcd<8,3>X",
		},
		{
			name:             "A match, outside search buffer",
			input:            []byte("XXXabcdefXXX"),
			maxMatchLen:      999,
			maxSearchBuffLen: 4,
			minMatchLen:      3,
			wantValuesRepr:   "XXXabcdefXXX",
		},
		{
			name:             "A match, almost outside search buffer",
			input:            []byte("XXXaXXX"),
			maxMatchLen:      999,
			maxSearchBuffLen: 4,
			minMatchLen:      3,
			wantValuesRepr:   "XXXa<4,3>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := bytesToValues(tt.input, tt.maxMatchLen, tt.maxSearchBuffLen, tt.minMatchLen)
			valuesRepr := ""
			for _, v := range values {
				valuesRepr += fmt.Sprintf("%v", v)
			}
			if valuesRepr != tt.wantValuesRepr {
				t.Errorf("unexpected repr (got '%s' want '%s')", valuesRepr, tt.wantValuesRepr)
			}
		})
	}
}
