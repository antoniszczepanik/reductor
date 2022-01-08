package main

import (
	"container/heap"
	"fmt"
	"io/ioutil"
	"os"
)

type Node struct {
	value       byte
	freq        int
	Left, Right *Node
	// Used only to represent nodes in graphviz format
	id int
}

func NewNode(id int, val byte, freq int, l, r *Node) Node {
	return Node{
		id:    id,
		value: val,
		freq:  freq,
		Left:  l,
		Right: r,
	}
}

// Stringer implementation allows to dump graphviz plot in this case.
func (n Node) String() string {
	repr := fmt.Sprintf("%d[label=\"value=%d freq=%d\"]\n", n.id, n.value, n.freq)
	if n.Left != nil {
		repr += fmt.Sprintf("%d -> %d\n", n.id, n.Left.id)
		repr += n.Left.String()
	}
	if n.Right != nil {
		repr += fmt.Sprintf("%d -> %d\n", n.id, n.Right.id)
		repr += n.Right.String()
	}
	return repr
}

type PriorityQueue []Node

func (pq PriorityQueue) Len() int           { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool { return pq[i].freq < pq[j].freq }
func (pq PriorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i] }
func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(Node))
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	x := old[n-1]
	*pq = old[0 : n-1]
	return x
}

func (pq *PriorityQueue) RemoveEmpty() PriorityQueue {
	not_empty_pq := make(PriorityQueue, 0, 255)
	for _, n := range *pq {
		if n.freq != 0 {
			not_empty_pq = append(not_empty_pq, n)
		}
	}
	return not_empty_pq
}

// constructHuffmanTree creates a tree for values and returns root *Node.
func constructHuffmanTree(values []Value) Node {
	freqs := make(PriorityQueue, 256)
	// Id counter for the nodes.
	var i int
	// Copy values into corresponding spots.
	for i = 0; i < len(freqs); i++ {
		freqs[i].value, freqs[i].id = byte(i), i
	}
	for _, v := range values {
		if v.IsLiteral {
			freqs[v.GetLiteralBinary()].freq += 1
		} else {
			for _, b := range v.GetPointerBinary() {
				freqs[b].freq += 1
			}
		}
	}
	freqs = freqs.RemoveEmpty()
	heap.Init(&freqs)
	for freqs.Len() > 1 {
		r, l := heap.Pop(&freqs).(Node), heap.Pop(&freqs).(Node)
		heap.Push(&freqs, NewNode(i, 0, r.freq+l.freq, &l, &r))
		i += 1
	}
	return heap.Pop(&freqs).(Node)
}

// encodeHuffmanTree converts a tree into a translation unit.
// Then it encodes it into binary, and appends encoded values after that.
func encodeHuffmanTree(root *Node, values []Value) []uint64 {
	return []uint64{1, 2, 3}
}

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
	//fmt.Println("File contents:")
	//fmt.Println(string(input))
	//fmt.Println("Values represantation:")
	values := BytesToValues(input, 6, 128, 512) // LZ77
	//for _, v := range values {
	//	fmt.Printf("%v", v)
	//}
	root := constructHuffmanTree(values)
	fmt.Printf("%v", root)

	//root := constructHuffmanTree(values) // Create huffman tree out of values
	//binary := encodeHuffmanTree(root, values) // Create a translation map, parse into binary.
}
