package iota

import (
	"encoding/binary"
	"errors"
	"io"
)

var overflow = errors.New("binary: varint overflows a 64-bit integer")

// ReadUvarint reads an encoded unsigned integer from r and returns it as a uint64.
// This is merely a copy of the std lib's function with the additional amount of bytes read for the varint.
func ReadUvarint(r io.ByteReader) (uint64, int, error) {
	var x uint64
	var s uint
	var bytesRead int
	for i := 0; i < binary.MaxVarintLen64; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return x, bytesRead, err
		}
		bytesRead++
		if b < 0x80 {
			if i == 9 && b > 1 {
				return x, bytesRead, overflow
			}
			return x | uint64(b)<<s, bytesRead, nil
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
	return x, bytesRead, overflow
}

// ReadVarint reads an encoded signed integer from r and returns it as an int64.
// This is merely a copy of the std lib's function with the additional amount of bytes read for the varint.
func ReadVarint(r io.ByteReader) (int64, int, error) {
	ux, bytesRead, err := ReadUvarint(r) // ok to continue in presence of error
	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}
	return x, bytesRead, err
}
