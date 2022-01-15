package main

import (
	"crypto/rand"
	"fmt"
	"testing"
)

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

func Test_getMatchIndex(t *testing.T) {
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
			gotMatches := getMatchIndex(tt.text, tt.pattern)
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

func Test_bytesToValues(t *testing.T) {
	var tests = []struct {
		name             string
		input            []byte
		minMatchLen      byte
		maxMatchLen      byte
		maxSearchBuffLen uint16
		wantValuesRepr   string
	}{
		{
			name:             "No matches at all",
			input:            []byte("abcd"),
			minMatchLen:      0,
			maxMatchLen:      255,
			maxSearchBuffLen: 255,
			wantValuesRepr:   "abcd",
		},
		{
			name:             "Match at the end",
			input:            []byte("abcd abcd"),
			minMatchLen:      0,
			maxMatchLen:      255,
			maxSearchBuffLen: 255,
			wantValuesRepr:   "abcd <5,4>",
		},
		{
			name:             "Match in the middle",
			input:            []byte("abcd abcd ghij"),
			minMatchLen:      0,
			maxMatchLen:      255,
			maxSearchBuffLen: 255,
			wantValuesRepr:   "abcd <5,5>ghij",
		},
		{
			name:             "Two matches",
			input:            []byte("XXabXXcdXX"),
			minMatchLen:      2,
			maxMatchLen:      255,
			maxSearchBuffLen: 255,
			// The matches are the same length - first one is selected.
			wantValuesRepr: "XXab<4,2>cd<8,2>",
		},
		{
			name:             "Three matches",
			input:            []byte("XXabXXcdXXijXX"),
			minMatchLen:      2,
			maxMatchLen:      255,
			maxSearchBuffLen: 255,
			// The matches are the same length - first one is selected.
			wantValuesRepr: "XXab<4,2>cd<8,2>ij<12,2>",
		},
		{
			name:             "A match, almost too long",
			input:            []byte("XXXabcdXXX"),
			minMatchLen:      3,
			maxMatchLen:      3,
			maxSearchBuffLen: 255,
			wantValuesRepr:   "XXXabcd<7,3>",
		},
		{
			name:             "A match, too long but is not consumed",
			input:            []byte("XXXXabcdXXXX"),
			minMatchLen:      3,
			maxMatchLen:      3,
			maxSearchBuffLen: 255,
			wantValuesRepr:   "XXXXabcd<8,3>X",
		},
		{
			name:             "A match, outside search buffer",
			input:            []byte("XXXabcdefXXX"),
			minMatchLen:      3,
			maxMatchLen:      255,
			maxSearchBuffLen: 4,
			wantValuesRepr:   "XXXabcdefXXX",
		},
		{
			name:             "A match, almost outside search buffer",
			input:            []byte("XXXaXXX"),
			minMatchLen:      3,
			maxMatchLen:      255,
			maxSearchBuffLen: 4,
			wantValuesRepr:   "XXXa<4,3>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := BytesToValues(tt.input, tt.minMatchLen, tt.maxMatchLen, tt.maxSearchBuffLen)
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

func Test_ValuesToBytes(t *testing.T) {
	var tests = []struct {
		name  string
		input []byte
	}{
		{
			name:  "Only literals",
			input: []byte("abcdefghijkl"),
		},
		{
			name:  "Empty input",
			input: []byte(""),
		},
		{
			name:  "Single match",
			input: []byte("XXXaaaXXX"), // "XXXaaa<6,3>"
		},
		{
			name:  "Multiple matches",
			input: []byte("XXXabXXXcdXXXijXXX"),
		},
		{
			name:  "Reapeated character",
			input: []byte("XXXXXXXXXXXXXXXXXXXXXXX"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := BytesToValues(tt.input, 255, 255, 3)
			got := ValuesToBytes(values)
			if string(tt.input) != string(got) {
				t.Errorf("got '%s' want '%s'", got, tt.input)
			}
		})
	}
}

var Values []Value

func Benchmark_ValuesToBytes(b *testing.B) {
	randomBytes := make([]byte, 1000)
	rand.Read(randomBytes)
	for n := 0; n < b.N; n++ {
		Values = BytesToValues(randomBytes, 4, 255, 4096)
	}
}
