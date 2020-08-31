package iota

import (
	"encoding/binary"
	"fmt"
)

// Uvarint calls binary.Uvarint but ensures that only max
// binary.MaxVarintLen64 bytes are passed from data to it. This, because binary.Uvarint
// continuously reads bytes until a less than 0x80/128 byte is seen.
func Uvarint(data []byte) (uint64, int, error) {
	l := len(data)
	if l == 0 {
		return 0, 0, fmt.Errorf("%w: can't extract varint from zero length byte slice", ErrInvalidVarint)
	}

	var num uint64
	var bytesRead int
	switch {
	case l >= binary.MaxVarintLen64:
		// limit so that only max binary.MaxVarintLen64 are read
		num, bytesRead = binary.Uvarint(data[:binary.MaxVarintLen64])
	default:
		num, bytesRead = binary.Uvarint(data)
	}

	switch {
	case bytesRead < 0:
		return 0, 0, fmt.Errorf("%w: varint value overflows binary.MaxVarintLen64", ErrInvalidVarint)
	case bytesRead == 0:
		return 0, 0, fmt.Errorf("%w: insufficient bytes to extract varint", ErrInvalidVarint)
	}
	return num, bytesRead, nil
}
