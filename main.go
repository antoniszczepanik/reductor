package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"
)

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

func revTable(m map[byte]string) map[string]byte {
	rev := make(map[string]byte, len(m))
	for k, v := range m {
		rev[v] = k
	}
	return rev
}

func compress(r io.Reader, w io.Writer) map[byte]string {
	input, err := ioutil.ReadAll(r)
	fmt.Printf("Original input len: %d\n", len(input))
	if err != nil {
		panic(err)
	}
	values := BytesToValues(input, 6, 255, 1024) // LZ77
	root := constructHuffmanTree(values)
	//root.DumpGraphviz(os.Stdout)
	codeTable := createCodeTable(&root, "")
	bw := NewBinaryWriter(w, codeTable)
	bw.Write(values)
	return codeTable
}

func decompress(r io.Reader, revTable map[string]byte) []byte {
	br := NewBinaryReader(r, revTable)
	new_values := br.Read()
	return ValuesToBytes(new_values)
}

func main() {
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	// TODO: Parse all params from flags.
	const filePath = "data.txt"
	f, err = os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		os.Exit(1)
	}
	b := &bytes.Buffer{}
	// TODO: Code table still needs to be written and read.
	// TODO: Error handling does not exist.
	codeTable := compress(f, b)
	fmt.Printf("Compressed input len: %d\n", b.Len())
	decompress(b, revTable(codeTable))
}
