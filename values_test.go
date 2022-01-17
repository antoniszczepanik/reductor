package main

import (
	"crypto/rand"
	"fmt"
	"testing"
)

func Test_BytesToValues(t *testing.T) {
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
