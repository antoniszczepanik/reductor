package main

import (
	"errors"
	"fmt"
)

// Value is a LZ77 sequence element - a literal or a pointer.
type Value struct {
	IsLiteral bool

	// Literal
	val byte

	// Pointer
	// TODO: This could be packed into smaller values to improve performance later.
	distance int
	length   int
}

func NewValue(isLiteral bool, value byte, distance, length int) Value {
	// TODO: Guard limits of value, distance and length.
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
func (v *Value) GetLiteralBinary() (byte, error) {
	if !v.IsLiteral {
		return 0, errors.New("tried to get literal binary for non-literal")
	}
	return v.val, nil
}

// GetPointerBinary returns binary representation of pointer.
// This are a few values, so that there are less possible values in a huffman
// tree.
func (v *Value) GetPointerBinary() (byte, error) {
	if v.IsLiteral {
		return 0, errors.New("tried to get pointer binary for literal")
	}
	// TODO: How would you do that? What would be the return type?
	return v.val, nil
}

// bytesToValues converts input to []Value, by replacing series of
// characters with LZ77 pointers wherever possible.
func BytesToValues(input []byte, maxMatchLen, maxSearchBuffLen, minMatchLen int) []Value {
	var searchBuffStart, lookaheadBuffEnd, p, l, dist int

	values := make([]Value, 0, len(input)) // Almost always it will be less, but lets over-allocate.
	for split := 0; split < len(input); split += 1 {
		searchBuffStart = max(0, split-maxSearchBuffLen)
		lookaheadBuffEnd = min(len(input), split+maxMatchLen)

		p, l = getLongestMatchPosAndLen(input[searchBuffStart:split], input[split:lookaheadBuffEnd], minMatchLen)

		if split > minMatchLen && l > 0 {
			// p is a position within searchBuff, so we need to calculate distance from the split.
			dist = split - (p + searchBuffStart)
			values = append(values, NewValue(false, 0, dist, l))
			split += (l - 1)
		} else {
			values = append(values, NewValue(true, input[split], 0, 0))
		}
	}
	return values
}

func getLongestMatchPosAndLen(
	searchBuff []byte,
	lookaheadBuff []byte,
	minMatchLen int,
) (position int, length int) {
	// TODO: This is O(n*m) and could be O(n+m)
	var matchLen, maxSoFar int
	for i := range searchBuff {
		matchLen = getMatchLen(searchBuff[i:], lookaheadBuff)
		if matchLen >= minMatchLen && matchLen > maxSoFar {
			position = i
			length, maxSoFar = matchLen, matchLen
		}
	}
	return
}

// getMatchLen returns a length of a match between two sequences.
func getMatchLen(a, b []byte) int {
	var matchLen int
	maxMatchLen := min(len(a), len(b))
	for i := 0; i < maxMatchLen; i++ {
		if a[i] == b[i] {
			matchLen += 1
		} else {
			break
		}
	}
	return matchLen
}

// valuesToBytes converts data from value representation back to []byte representation.
func ValuesToBytes(values []Value) []byte {
	var from int
	bytes := make([]byte, 0, len(values)) // We underallocate here.
	for _, v := range values {
		if v.IsLiteral {
			bytes = append(bytes, v.val)
		} else {
			from = len(bytes) - v.distance
			bytes = append(bytes, bytes[from:from+v.length]...)
		}
	}
	return bytes
}
