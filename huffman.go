package main

import (
	"container/heap"
	"fmt"
	"io"
)

type Node struct {
	value       byte
	freq        int
	Left, Right *Node
	isLeaf      bool
	// Used only to represent nodes in graphviz format.
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

func (n Node) DumpGraphviz(w io.Writer) {
	w.Write([]byte("Digraph g {\n"))
	w.Write([]byte(n.getGraphviz()))
	w.Write([]byte("}\n"))
}

// getGraphviz recursively gets graphviz representation of all nodes with root at n.
func (n Node) getGraphviz() string {
	repr := fmt.Sprintf("\t%d[label=\"value=%c freq=%d\"]\n", n.id, n.value, n.freq)
	if n.Left != nil {
		repr += fmt.Sprintf("\t%d -> %d[label=\"0\"]\n", n.id, n.Left.id)
		repr += n.Left.getGraphviz()
	}
	if n.Right != nil {
		repr += fmt.Sprintf("\t%d -> %d[label=\"1\"]\n", n.id, n.Right.id)
		repr += n.Right.getGraphviz()
	}
	return repr
}

type PriorityQueue []Node

func (pq PriorityQueue) Len() int            { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool  { return pq[i].freq < pq[j].freq }
func (pq PriorityQueue) Swap(i, j int)       { pq[i], pq[j] = pq[j], pq[i] }
func (pq *PriorityQueue) Push(x interface{}) { *pq = append(*pq, x.(Node)) }
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
	var i int // i is id counter.

	// Copy values into corresponding spots.
	for i = 0; i < len(freqs); i++ {
		freqs[i].value, freqs[i].id, freqs[i].isLeaf = byte(i), i, true
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

type Code struct {
	c    uint64
	bits byte
}

func (c Code) String() string {
	return fmt.Sprintf("%08b(%d)\n", c.c, c.bits)
}

func addBit(c Code, bit bool) Code {
	var b uint64
	if bit {
		b = 1
	}
	return Code{
		c:    (c.c << 1) | b,
		bits: c.bits + 1,
	}
}

type CodeTable map[byte]Code

func createCodeTable(root *Node, prefix Code) CodeTable {
	codeTable := make(CodeTable)
	lt := make(CodeTable)
	rt := make(CodeTable)
	if root.isLeaf {
		codeTable[root.value] = prefix
		return codeTable
	}
	if root.Left != nil {
		lt = createCodeTable(root.Left, addBit(prefix, false))
	}
	if root.Right != nil {
		rt = createCodeTable(root.Right, addBit(prefix, true))
	}
	return mergeTables(lt, rt)
}

func mergeTables(a, b CodeTable) CodeTable {
	for k, v := range b {
		a[k] = v
	}
	return a
}
