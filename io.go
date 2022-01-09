package main

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/icza/bitio"
)

type BinaryWriter struct {
	w         *bitio.Writer
	codeTable CodeTable
}

func NewBinaryWriter(writer io.Writer, codeTable CodeTable) BinaryWriter {
	w := bitio.NewWriter(writer)
	return BinaryWriter{
		w:         w,
		codeTable: codeTable,
	}
}

func (bw *BinaryWriter) Write(values []Value) {
	var (
		err  error
		code uint64
		l    byte
	)
	bw.writeTable()
	for _, v := range values {
		err = bw.w.WriteBool(v.IsLiteral)
		if err != nil {
			panic("write bool")
		}
		if v.IsLiteral {
			code, l = bw.getCodeForValue(v.GetLiteralBinary())
			bw.w.WriteBits(code, l)
		} else {
			for _, b := range v.GetPointerBinary() {
				code, l = bw.getCodeForValue(b)
				bw.w.WriteBits(code, l)
			}
		}
	}
	bw.w.Close()
}
func (bw *BinaryWriter) writeTable() {
	// First 8 bits denote amount of elements in the table.
	if len(bw.codeTable) == 0 {
		panic("code table of length 0")
	}
	if err := bw.w.WriteBits(uint64(len(bw.codeTable)-1), 8); err != nil {
		panic(err)
	}
	for k, v := range bw.codeTable {
		// Next, we write (byte, byte, code) triplets.
		// First byte denotes a value to be encoded/decoded.
		if err := bw.w.WriteBits(uint64(k), 8); err != nil {
			panic(err)
		}
		// Second byte denotes size of the code.
		if err := bw.w.WriteBits(uint64(v.bits), 8); err != nil {
			panic(err)
		}
		// After that, code is written.
		if err := bw.w.WriteBits(uint64(v.c), v.bits); err != nil {
			panic(err)
		}
	}
}

func (bw *BinaryWriter) getCodeForValue(val byte) (uint64, byte) {
	code := bw.codeTable[val]
	return uint64(code.c), code.bits
}

type BinaryReader struct {
	r        *bitio.Reader
	valTable map[Code]byte
}

func NewBinaryReader(reader io.Reader) BinaryReader {
	r := bitio.NewReader(reader)
	return BinaryReader{
		r: r,
	}
}

func (br *BinaryReader) Read() []Value {
	br.valTable = br.readTable()
	values := make([]Value, 0)
	for {
		val, err := br.consumeValue()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			panic(err)
		}
		values = append(values, val)
	}
	return values
}

func (br *BinaryReader) readTable() map[Code]byte {
	valTable := make(map[Code]byte)
	// First 8 bits denote amount of elements in the table.
	size, err := br.r.ReadBits(8)
	if err != nil {
		panic(err)
	}
	// We substracted 1 when writing to avoid overflowing byte.
	size += 1
	var i uint64
	for i = 0; i < size; i++ {
		// Value.
		val, err := br.r.ReadBits(8)
		if err != nil {
			panic(err)
		}
		// Code size.
		codeBits, err := br.r.ReadBits(8)
		if err != nil {
			panic(err)
		}
		// Code itself.
		code, err := br.r.ReadBits(byte(codeBits))
		if err != nil {
			panic(err)
		}
		valTable[Code{c: code, bits: byte(codeBits)}] = byte(val)
	}
	return valTable
}

func (br *BinaryReader) consumeValue() (Value, error) {
	isLiteral, err := br.r.ReadBool()
	if err != nil {
		return Value{}, err
	}
	if isLiteral {
		match, err := br.readMatch()
		if err != nil {
			return Value{}, err
		}
		return NewValue(true, match, 0, 0), nil
	}
	matches, err := br.readPointerMatches()
	if err != nil {
		return Value{}, err
	}
	return pointerMatchesToPointer(matches), nil
}

func (br *BinaryReader) readMatch() (byte, error) {
	match := Code{}
	for {
		b, err := br.r.ReadBool()
		if err != nil {
			return 0, err
		}
		match = addBit(match, b)
		if val, ok := br.valTable[match]; ok {
			return val, nil
		}
	}
}

func (br *BinaryReader) readPointerMatches() ([]byte, error) {
	var err error
	bytes := make([]byte, 3)
	for i := range bytes {
		bytes[i], err = br.readMatch()
		if err != nil {
			return nil, err
		}
	}
	return bytes, nil
}

func pointerMatchesToPointer(bytes []byte) Value {
	// First 2 bytes encode a distance.
	distance := binary.BigEndian.Uint16(bytes)
	// The last byte is length.
	length := bytes[2]
	return NewValue(false, 0, length, distance)
}
