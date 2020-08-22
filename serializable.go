package iota

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Serializable is something which knows how to serialize/deserialize itself from/into bytes.
// This is almost analogous to BinaryMarshaler/BinaryUnmarshaler.
type Serializable interface {
	// Deserialize deserializes the given data into the object and returns the amount of bytes consumed from data.
	// If the passed data is not big enough for deserialization, an error must be returned.
	// During deserialization the data is checked for validity.
	Deserialize(data []byte) (int, error)
	// Serialize returns a serialized byte representation.
	// This function does not check the serialized data for validity.
	Serialize() ([]byte, error)
}

// Serializables is a slice of Serializable.
type Serializables []Serializable

// SerializableSelectorFunc is a function that given a type byte, returns an empty instance of the given underlying type.
// If the type doesn't resolve, an error is returned.
type SerializableSelectorFunc func(typeByte byte) (Serializable, error)

// ArrayRules defines rules around a to be deserialized array.
// Min and Max at 0 define an unbounded array.
type ArrayRules struct {
	// The min array bound.
	Min uint64
	// The max array bound.
	Max uint64
	// The error returned if the min bound is violated.
	MinErr error
	// The error returned if the max bound is violated.
	MaxErr error
	// Whether the bytes of the elements have to be in lexical order.
	ElementBytesLexicalOrder bool
	// The error returned if the element bytes lexical order is violated.
	ElementBytesLexicalOrderErr error
}

// LexicalOrderedByteSlices are byte slices ordered in lexical order.
type LexicalOrderedByteSlices [][]byte

func (l LexicalOrderedByteSlices) Len() int {
	return len(l)
}

func (l LexicalOrderedByteSlices) Less(i, j int) bool {
	return bytes.Compare(l[i], l[j]) < 0
}

func (l LexicalOrderedByteSlices) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// DeserializeArrayOfObjects deserializes the given data into Serializables.
// The data is expected to start with the count denoting varint, followed by the actual structs.
// An optional ArrayRules can be passed in to return an error in case it is violated.
func DeserializeArrayOfObjects(data []byte, serSel SerializableSelectorFunc, arrayBounds *ArrayRules) (Serializables, int, error) {
	var bytesReadTotal int
	seriCount, seriCountBytesSize, err := ReadUvarint(bytes.NewReader(data[:binary.MaxVarintLen64]))
	if err != nil {
		return nil, 0, err
	}
	bytesReadTotal += seriCountBytesSize

	if arrayBounds != nil {
		if arrayBounds.Min != 0 && seriCount < arrayBounds.Min {
			return nil, 0, fmt.Errorf("%w: min is %d but count is %d", arrayBounds.MinErr, arrayBounds.Min, seriCount)
		}
		if arrayBounds.Max != 0 && seriCount > arrayBounds.Max {
			return nil, 0, fmt.Errorf("%w: max is %d but count is %d", arrayBounds.MaxErr, arrayBounds.Max, seriCount)
		}
	}

	// advance to objects
	var seris Serializables
	data = data[seriCountBytesSize:]

	var prevEleBytes []byte

	var offset int
	for i := 0; i < int(seriCount); i++ {
		seri, seriBytesConsumed, err := DeserializeObject(data[offset:], serSel)
		if err != nil {
			return nil, 0, err
		}
		// check lexical order against previous element
		if arrayBounds != nil && arrayBounds.ElementBytesLexicalOrder {
			eleBytes := data[offset : offset+seriBytesConsumed]
			switch {
			case prevEleBytes == nil:
				prevEleBytes = eleBytes
			case bytes.Compare(prevEleBytes, eleBytes) > 0:
				return nil, 0, fmt.Errorf("%w: element %d should have been before element %d", arrayBounds.ElementBytesLexicalOrderErr, i, i-1)
			default:
				prevEleBytes = eleBytes
			}
		}
		seris = append(seris, seri)
		offset += seriBytesConsumed
	}
	bytesReadTotal += offset

	return seris, bytesReadTotal, nil
}

// DeserializeObject deserializes the given data into a Serializable.
// The data is expected to start with the type denoting byte.
func DeserializeObject(data []byte, serSel SerializableSelectorFunc) (Serializable, int, error) {
	if len(data) < 2 {
		return nil, 0, ErrDeserializationDataTooSmall
	}
	seri, err := serSel(data[0])
	if err != nil {
		return nil, 0, err
	}
	seriBytesConsumed, err := seri.Deserialize(data)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to deserialize %T: %w", seri, err)
	}
	return seri, seriBytesConsumed, nil
}
