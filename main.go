package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"
	"strings"
	"time"
)

func compress(
	source io.Reader,
	sink io.Writer,
	minMatch uint8,
	maxMatch uint8,
	searchSize uint16,

	graphf io.Writer,
	lzf io.Writer,
) {
	log.Printf("Config: min-match=%d, max-match=%d, search-size=%d\n", minMatch, maxMatch, searchSize)
	input, err := ioutil.ReadAll(source)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Input size(bytes): %d\n", len(input))
	// LZ coding.
	values := BytesToValues(input, minMatch, maxMatch, searchSize)
	for _, v := range values {
		fmt.Fprintf(lzf, "%v", v)
	}
	// Huffman coding.
	root := constructHuffmanTree(values)
	root.DumpGraphviz(graphf)
	codeTable := createCodeTable(&root, Code{})
	// Write binary representation.
	bw := NewBinaryWriter(sink, codeTable)
	bw.Write(values)
}

func decompress(source io.Reader, sink io.Writer) {
	br := NewBinaryReader(source)
	newVals := br.Read()
	_, err := sink.Write(ValuesToBytes(newVals))
	if err != nil {
		log.Fatal(err)
	}
}
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] <filename>\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	var (
		err                            error
		minMatch, maxMatch, searchSize uint
	)

	mode := flag.Bool("compress", true, "run the program in compression mode")
	name := flag.String("name", "", "name for the file with compressed data")
	flag.UintVar(&minMatch, "min-match", 4, "minimum match size for LZ algorithm")
	flag.UintVar(&maxMatch, "max-match", 255, "maximum match size for LZ algorithm (upper limit is 255)")
	flag.UintVar(&searchSize, "search-size", 4096, "size of the search window of LZ algorithm (upper limit is 65535)")

	// Diagnostic options.
	verbose := flag.Bool("verbose", false, "display log messages")
	graphvizPath := flag.String("graphviz", "", "write graphviz huffman tree representation to file")
	lzPath := flag.String("lz", "", "write lz representation to file")
	cpuProfilePath := flag.String("cpuprofile", "", "write cpu profile to file")
	flag.Parse()

	if flag.NArg() != 1 {
		Usage()
	}
	filePath := flag.Arg(0)

	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.Printf("Running %s in verbose mode\n", os.Args[0])

	// Run and save a cpuprofile.
	if *cpuProfilePath != "" {
		log.Printf("Will create cpu profile: %s\n", *cpuProfilePath)
		f, err := os.Create(*cpuProfilePath)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// Open Graphviz writer.
	var graphf io.Writer
	if *graphvizPath != "" {
		log.Printf("Will create graph of huffman tree: %s\n", *graphvizPath)
		graphf, err = os.Create(*graphvizPath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		graphf = ioutil.Discard
	}

	// Open LZ writer.
	var lzf io.Writer
	if *lzPath != "" {
		log.Printf("Will create LZ representation: %s\n", *lzPath)
		lzf, err = os.Create(*lzPath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		lzf = ioutil.Discard
	}

	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}

	if *mode {
		log.Printf("Compress: %s\n", filePath)
		fileSize := getFileSize(filePath)
		if *name == "" {
			*name = filePath + ".reduced"
		}
		target, err := os.Create(*name)
		if err != nil {
			log.Fatal(err)
		}
		start := time.Now()
		compress(f, target, byte(minMatch), byte(maxMatch), uint16(searchSize), graphf, lzf)
		compressedFileSize := getFileSize(*name)
		log.Printf("Time elapsed: %s\n", time.Since(start))
		log.Printf("Compression ratio: %.2f\n", float64(fileSize)/float64(compressedFileSize))
	} else {
		log.Printf("Decompress: %s\n", filePath)
		var sink io.Writer
		if *name == "" {
			// Strip off ".reduced" suffix if present.
			if strings.EqualFold(filePath[len(filePath)-8:], ".reduced") {
				*name = filePath[:len(filePath)-8] + ".unreduced"
			} else {
				*name = filePath + ".unreduced"
			}
			log.Printf("Writing decompressed data to %s\n", *name)
			sink, err = os.Create(*name)
			if err != nil {
				log.Fatal(err)
			}
		}
		start := time.Now()
		decompress(f, sink)
		log.Printf("Time elapsed: %s\n", time.Since(start))
	}
}
