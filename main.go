package main

import (
	"bytes"
	"fmt"
)

// For now this is 4 byte pointer: |1|2|3|4|
// 1,2,3 - distance
// 4     - length
const pointerBytesNumber = 4
const lengthBytesNumber = 1
const distanceBytesNumber = pointerBytesNumber - lengthBytesNumber
const maxWindowSize = (1 << (distanceBytesNumber * 8)) - 1
const maxMatchLength = (1 << (lengthBytesNumber * 8)) - 1

// Value is a LZ77 sequence element - a literal or a pointer.
type Value struct {
	IsLiteral bool

	// Literal
	val byte

	// Pointer
	distance int
	length   int
}

func NewValue(isLiteral bool, value byte, distance, length int) *Value {
	return &Value{
		isLiteral: isLiteral,
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

func main() {
	// First iterate over some input data and create a sequence of values (as defined above).
	// Next figure out how to serialize/deserialize all those elements into elements of future
	// huffman tree (only a single tree).
	fmt.Printf("max window size %d\nmaxMatchLength %d\n", maxWindowSize, maxMatchLength)
	input := bytes.NewBuffer([]byte("hello, gello!"))
	fmt.Printf("%b", input)
}
