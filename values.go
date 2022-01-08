package main

import (
	"encoding/binary"
	"fmt"
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

// TODO: How would you do that? What would be the return type?
// How to get from Values representation to binary?
//
// - Figure out a code for each value, then huffman encode each code.
//
// How to get from binary to Values representation?
//
// - You read Huffman translation table.
// - You can now decode all of the values:
//      - So, you start reading a single bit prefix.
//      - Now, you know if the element is a pointer, if so, you
//        read and translate next 3/4 elements (they are not prefixed)
//        and construct a pointer value for them
//      - If not, then you read and translate a single huffman node. And
//        create a literal value for in.

// GetPointerBinary returns binary representation of pointer.
// A pointer is serialized to a few bytes, so that there are less possible
// nodes in a Huffman tree.
func (v *Value) GetPointerBinary() []byte {
	bytes := make([]byte, 3)
	// First 2 bytes encode a distance.
	binary.BigEndian.PutUint16(bytes, v.distance)
	// The last byte is length
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

	values := make([]Value, 0, len(input)) // Almost always it will be less, but lets over-allocate.
	for split := 0; split < len(input); split += 1 {
		searchBuffStart = max(0, split-int(maxSearchBuffLen))
		lookaheadBuffEnd = min(len(input), split+int(maxMatchLen))

		p, l = getLongestMatchPosAndLen(input[searchBuffStart:split], input[split:lookaheadBuffEnd], minMatchLen)

		if split > int(minMatchLen) && l > 0 {
			// p is a position within searchBuff, so we need to calculate distance from the split.
			dist = uint16(split - (p + searchBuffStart))
			values = append(values, NewValue(false, 0, l, dist))
			split += (int(l) - 1)
		} else {
			values = append(values, NewValue(true, input[split], 0, 0))
		}
	}
	return values
}

func getLongestMatchPosAndLen(
	searchBuff []byte,
	lookaheadBuff []byte,
	minMatchLen byte,
) (position int, length byte) {
	// TODO: This is O(n*m) and could be O(n+m)
	var matchLen, maxSoFar byte
	for i := range searchBuff {
		matchLen = getMatchLen(searchBuff[i:], lookaheadBuff)
		if matchLen >= minMatchLen && matchLen > maxSoFar {
			position = i
			length, maxSoFar = matchLen, matchLen
		}
	}
	return
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
