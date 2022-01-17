package main

import "testing"

func Test_getLongestMatchPosAndLen(t *testing.T) {
	var tests = []struct {
		name          string
		searchBuff    []byte
		lookaheadBuff []byte
		wantPos       int
		wantLen       byte
		minMatchLen   byte
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
func Test_getMatchIndices(t *testing.T) {
	var tests = []struct {
		name        string
		text        []byte
		pattern     []byte
		wantMatches []int
	}{
		{
			name:        "Match starts at the beginning",
			text:        []byte("hello"),
			pattern:     []byte("hel"),
			wantMatches: []int{0},
		},
		{
			name:        "Match starts somewher in the middle",
			text:        []byte("abchello"),
			pattern:     []byte("hel"),
			wantMatches: []int{3},
		},
		{
			name:        "Pattern is empty",
			text:        []byte("abchello"),
			pattern:     []byte(""),
			wantMatches: []int{0, 1, 2, 3, 4, 5, 6, 7},
		},
		{
			name:        "Text is empty",
			text:        []byte(""),
			pattern:     []byte("abc"),
			wantMatches: []int{},
		},
		{
			name:        "Several matches",
			text:        []byte("aaaabcaaaabcaaaabc"),
			pattern:     []byte("abc"),
			wantMatches: []int{3, 9, 15},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatches := getMatchIndices(tt.text, tt.pattern)
			if len(gotMatches) != len(tt.wantMatches) {
				t.Errorf("unexpected min match starts (got '%v' want '%v')", gotMatches, tt.wantMatches)
			}
			for i := range gotMatches {
				if gotMatches[i] != tt.wantMatches[i] {
					t.Errorf("unexpected min match starts (got '%v' want '%v')", gotMatches, tt.wantMatches)
				}
			}
		})
	}
}
