package main

import (
	"bytes"
	"container/heap"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/icza/bitio"
)

type Node struct {
	value       byte
	freq        int
	Left, Right *Node
	isLeaf      bool
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

func (n Node) DumpGraphviz(w io.Writer) {
	w.Write([]byte("Digraph g {\n"))
	w.Write([]byte(n.getGraphviz()))
	w.Write([]byte("}\n"))
}

// getGraphviz recursively gets graphviz representation of all nodes with root at n.
func (n Node) getGraphviz() string {
	repr := fmt.Sprintf("\t%d[label=\"value=%d freq=%d\"]\n", n.id, n.value, n.freq)
	if n.Left != nil {
		repr += fmt.Sprintf("\t%d -> %d\n", n.id, n.Left.id)
		repr += n.Left.getGraphviz()
	}
	if n.Right != nil {
		repr += fmt.Sprintf("\t%d -> %d\n", n.id, n.Right.id)
		repr += n.Right.getGraphviz()
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

func createCodeTable(root *Node, prefix string) map[string]string {
	codeTable := make(map[string]string)
	lt := make(map[string]string)
	rt := make(map[string]string)
	if root.isLeaf {
		codeTable[fmt.Sprintf("%08b", root.value)] = prefix
		return codeTable
	}
	if root.Left != nil {
		lt = createCodeTable(root.Left, prefix+"0")
	}
	if root.Right != nil {
		rt = createCodeTable(root.Right, prefix+"1")
	}
	return mergeTables(lt, rt)
}

type BinaryWriter struct {
	w *bitio.Writer
}

func NewBinaryWriter(writer io.Writer) BinaryWriter {
	w := bitio.NewWriter(writer)
	return BinaryWriter{
		w: w,
	}
}

func (bw *BinaryWriter) Write(codeTable map[string]string, values []Value) {
	for _, v := range values {
		if v.IsLiteral {
			//bw.w.Write()
		}
	}
	bw.w.Close()
}

func mergeTables(a, b map[string]string) map[string]string {
	for k, v := range b {
		a[k] = v
	}
	return a
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
	values := BytesToValues(input, 6, 128, 512) // LZ77
	root := constructHuffmanTree(values)
	//root.DumpGraphviz(os.Stdout)
	codeTable := createCodeTable(&root, "")
	//fmt.Printf("%+v\n", codeTable)

	b := &bytes.Buffer{} // Just some writer interface.
	bw := NewBinaryWriter(b)
	bw.Write(codeTable, values)
	fmt.Printf("% x", b)
}
