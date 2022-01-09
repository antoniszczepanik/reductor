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
	for _, v := range values {
		err = bw.w.WriteBool(v.IsLiteral)
		if err != nil {
			panic("writing bool")
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

func (bw *BinaryWriter) getCodeForValue(val byte) (uint64, byte) {
	code := bw.codeTable[val]
	return uint64(code.c), code.bits
}

type BinaryReader struct {
	r        *bitio.Reader
	valTable map[Code]byte
}

func NewBinaryReader(reader io.Reader, valTable map[Code]byte) BinaryReader {
	r := bitio.NewReader(reader)
	return BinaryReader{
		r:        r,
		valTable: valTable,
	}
}

func (br *BinaryReader) Read() []Value {
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
