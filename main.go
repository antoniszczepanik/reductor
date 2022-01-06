package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

// For now this is 4 byte pointer: |1|2|3|4|
// 1,2,3 - distance
// 4     - length
const pointerBytesNumber = 4
const lengthBytesNumber = 1
const distanceBytesNumber = pointerBytesNumber - lengthBytesNumber
const maxSearchBufferLength = (1 << (distanceBytesNumber * 8)) - 1
const maxMatchLength = (1 << (lengthBytesNumber * 8)) - 1
const minMatchLength = 4

const windowSize = 32768

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
	return Value{
		IsLiteral: isLiteral,
		val:       value,
		distance:  distance,
		length:    length,
	}
}

func (v *Value) SetLength(length int) error {
	if length < 0 || length > maxMatchLength {
		return fmt.Errorf("length our of range: %d", length)
	}
	v.length = length
	return nil
}

func (v Value) String() string {
	if v.IsLiteral {
		return fmt.Sprintf("%c", v.val)
	}
	return fmt.Sprintf("<%d,%d>", v.distance, v.length)
}

// bytesToValues converts input to []Value, by replacing series of
// characters with LZ77 pointers wherever possible.
func bytesToValues(input []byte, maxMatchLen, maxSearchBuffLen, minMatchLen int) []Value {
	var searchBuffStart, lookaheadBuffEnd, p, l, dist int
	values := make([]Value, 0, len(input)) // It will hopefully be less, but let's over-allocate.
	for split := 0; split < len(input); split += 1 {
		searchBuffStart = max(0, split-maxSearchBuffLen)
		lookaheadBuffEnd = min(len(input), split+maxMatchLen)
		p, l = getLongestMatchPosAndLen(input[searchBuffStart:split], input[split:lookaheadBuffEnd], minMatchLen)
		if split > minMatchLen && l > 0 {
			// p is a position within searchBuff, so we need to calculate distance from the split.
			dist = split - p
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
	var matchLen, maxSoFar int
	for i := range searchBuff {
		matchLen = getMatchLen(searchBuff[i:], lookaheadBuff)
		if matchLen >= minMatchLen && matchLen > maxSoFar {
			position = i
			length, maxSoFar = matchLen, matchLen
		}
	}
	//fmt.Printf("Matching '%s' in '%s'\n", string(lookaheadBuff), string(searchBuff))
	//fmt.Printf("Match: pos=%d, len=%d\n", position, length)
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

// valuesToBytes converts data from value representation backto []byte representation.
func valuesToBytes(values []Value) []byte {
	return []byte{}
}

type Node struct {
	value int32
	freq  int64
	left  *Node
	right *Node
}

// constructHuffmanTree creates a tree for values and returns root *Node.
func constructHuffmanTree(values []Value) *Node {
	return &Node{}
}

// encodeHuffmanTree converts a tree into a translation unit.
// Then it encodes it into binary, and appends encoded values after that.
func encodeHuffmanTree(root *Node, values []Value) []uint64 {
	return []uint64{1, 2, 3}
}

// Should take a binary and return some sort of translator (a map[binary]*Value) which will
// allow to translate following string.
//
// We could create a new struct which would be called compressedreader, that
// would wrap bitio reader and would allow to call "consume header".
// methods on it.
//func decodeHuffmanTree(binary uint64) {
//
//}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func main() {
	// TODO: Get input file name from flags.
	const filePath = "data.txt"
	input, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		os.Exit(1)
	}
	fmt.Println("File contents:")
	fmt.Println(string(input))
	fmt.Println("Values represantation:")
	values := bytesToValues(input, maxMatchLength, maxSearchBufferLength, minMatchLength) // LZ77
	for _, v := range values {
		fmt.Printf("%v", v)
	}
	fmt.Println("")
	fmt.Println("Back to original represantation:")
	fmt.Println(string(valuesToBytes(values)))
	//root := constructHuffmanTree(values) // Create huffman tree out of values
	//binary := encodeHuffmanTree(root, values) // Create a translation map, parse into binary.
}
