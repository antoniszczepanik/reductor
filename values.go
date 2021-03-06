package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
)

// Value is a LZ77 sequence element - a literal or a pointer.
type Value struct {
	IsLiteral bool

	// Literal
	val byte

	// Pointer
	distance uint16
	length   byte
}

func NewValue(isLiteral bool, value, length byte, distance uint16) Value {
	return Value{
		IsLiteral: isLiteral,
		val:       value,
		distance:  distance,
		length:    length,
	}
}

func (v Value) String() string {
	if v.IsLiteral {
		return fmt.Sprintf("%c", v.val)
	}
	return fmt.Sprintf("<%d,%d>", v.distance, v.length)
}

// GetLiteralBinary returns binary representation of literal.
func (v *Value) GetLiteralBinary() byte {
	return v.val
}

// GetPointerBinary returns binary representation of pointer.
// A pointer is serialized to a few bytes, so that there are less possible
// nodes in a Huffman tree.
func (v *Value) GetPointerBinary() []byte {
	bytes := make([]byte, 3)
	// First 2 bytes encode a distance.
	binary.BigEndian.PutUint16(bytes, v.distance)
	// The last byte is length.
	bytes[2] = v.length
	return bytes
}

// BytesToValues converts input to []Value, by replacing series of
// characters with LZ77 pointers wherever possible.
func BytesToValues(input []byte, minMatchLen, maxMatchLen byte, maxSearchBuffLen uint16) []Value {
	var (
		searchBuffStart, lookaheadBuffEnd, p int
		dist                                 uint16
		l                                    byte
	)

	values := make([]Value, len(input)) // Almost always it will be less, but lets over-allocate.
	value_counter := 0
	pointer_counter := 0
	for split := 0; split < len(input); split += 1 {
		searchBuffStart = max(0, split-int(maxSearchBuffLen))
		lookaheadBuffEnd = min(len(input), split+int(maxMatchLen))

		p, l = getLongestMatchPosAndLen(input[searchBuffStart:split], input[split:lookaheadBuffEnd], minMatchLen)

		if split > int(minMatchLen) && l > 0 {
			// p is a position within searchBuff, so we need to calculate distance from the split.
			dist = uint16(split - (p + searchBuffStart))
			values[value_counter] = NewValue(false, 0, l, dist)
			value_counter += 1
			split += (int(l) - 1)
			pointer_counter += 1
		} else {
			values[value_counter] = NewValue(true, input[split], 1, 0)
			value_counter += 1
		}
	}
	log.Printf("Pointers ratio: %.2f\n", float64(pointer_counter)/float64(value_counter))
	return values[:value_counter]
}

func getLongestMatchPosAndLen(text, pattern []byte, minMatchLen byte) (int, byte) {
	if len(pattern) < int(minMatchLen) {
		return 0, 0
	}
	var (
		matchLen, maxSoFar, length byte
		position                   int
	)
	// Heuristic: get indexes at which at least minMatchLen of pattern matches.
	minMatchStarts := getMatchIndex(text, pattern[:minMatchLen])
	for _, matchStart := range minMatchStarts {
		matchLen = getMatchLen(text[matchStart:], pattern)
		if matchLen >= minMatchLen && matchLen > maxSoFar {
			position = matchStart
			length, maxSoFar = matchLen, matchLen
		}
	}
	return position, length
}

// getMatchIndex will return a slice of indexes at which pattern begins.
func getMatchIndex(text, pattern []byte) []int {
	// We surely do not have any matches.
	if len(text) == 0 || len(text) < len(pattern) {
		return []int{}
	}
	// If pattern is empty, then we have a match everywhere.
	if len(pattern) == 0 {
		matchIndicies := make([]int, len(text))
		for i := range text {
			matchIndicies[i] = i
		}
		return matchIndicies
	}

	matchIndices := make([]int, 0)
	for i := range text[:len(text)-len(pattern)+1] {
		// Just look for where pattern matches. This is hot path, and this
		// way to compare bytes turned out to be the fastest experimentally.
		if text[i] == pattern[0] {
			if text[i+1] == pattern[1] {
				if text[i+2] == pattern[2] {
					if bytes.Equal(text[i+3:i+len(pattern)], pattern[3:len(pattern)]) {
						matchIndices = append(matchIndices, i)
					}
				}
			}
		}
	}
	return matchIndices
}

// getMatchLen returns a length of a longest match between two sequences.
func getMatchLen(a, b []byte) byte {
	var matchLen byte
	maxMatchLen := min(min(len(a), len(b)), 255)
	for i := 0; i < maxMatchLen; i++ {
		if a[i] == b[i] {
			matchLen += 1
		} else {
			break
		}
	}
	return matchLen
}

// ValuesToBytes converts data from value representation back to []byte representation.
func ValuesToBytes(values []Value) []byte {
	var from int
	bytes := make([]byte, 0, len(values)) // We underallocate here.
	for _, v := range values {
		if v.IsLiteral {
			bytes = append(bytes, v.val)
		} else {
			from = len(bytes) - int(v.distance)
			bytes = append(bytes, bytes[from:from+int(v.length)]...)
		}
	}
	return bytes
}
