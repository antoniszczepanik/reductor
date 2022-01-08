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
	values := BytesToValues(input, 128, 512, 6) // LZ77
	for _, v := range values {
		fmt.Printf("%v", v)
	}
	fmt.Println("")
	fmt.Println("Back to original represantation:")
	fmt.Println(string(ValuesToBytes(values)))
	//root := constructHuffmanTree(values) // Create huffman tree out of values
	//binary := encodeHuffmanTree(root, values) // Create a translation map, parse into binary.
}
