package iotapkg

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Serializable is something which knows how to serialize/deserialize itself from/into bytes.
// This is almost analogous to BinaryMarshaler/BinaryUnmarshaler.
type Serializable interface {
	// Serialize returns a serialized byte representation.
	Serialize() ([]byte, error)
	// Deserialize deserializes the given data into the object and returns the amount of bytes consumed from data.
	// If the passed data is not big enough for deserialization, an error must be returned.
	Deserialize(data []byte) (int, error)
}

// Serializables is a slice of Serializable.
type Serializables []Serializable

// SerializableSelectorFunc is a function that given a type byte, returns an empty instance of the given underlying type.
// If the type doesn't resolve, an error is returned.
type SerializableSelectorFunc func(typeByte byte) (Serializable, error)

// DeserializeArrayOfObjects deserializes the given data into Serializables.
// The data is expected to start with the count denoting varint, followed by the actual structs.
func DeserializeArrayOfObjects(data []byte, serSel SerializableSelectorFunc) (Serializables, int, error) {
	var bytesReadTotal int
	seriCount, seriCountBytesSize, err := ReadUvarint(bytes.NewReader(data[:binary.MaxVarintLen64]))
	if err != nil {
		return nil, 0, err
	}
	bytesReadTotal += seriCountBytesSize

	// TODO: optional count check

	// advance to objects
	var seris Serializables
	data = data[seriCountBytesSize:]

	var offset int
	for i := 0; i < int(seriCount); i++ {
		seri, seriBytesConsumed, err := DeserializeObject(data[offset:], serSel)
		if err != nil {
			return nil, 0, err
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
