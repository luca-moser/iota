package iotapkg

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Defines a type of unlock block.
type UnlockBlockType = byte

const (
	// Denotes a signature unlock block.
	UnlockBlockSignature UnlockBlockType = iota
	// Denotes a reference unlock block.
	UnlockBlockReference
)

// UnlockBlockSelector implements SerializableSelectorFunc for unlock block types.
func UnlockBlockSelector(typeByte byte) (Serializable, error) {
	var seri Serializable
	switch typeByte {
	case UnlockBlockSignature:
		seri = &SignatureUnlockBlock{}
	case UnlockBlockReference:
		seri = &ReferenceUnlockBlock{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownUnlockBlockType, typeByte)
	}
	return seri, nil
}

// SignatureUnlockBlock holds a signature which unlocks inputs.
type SignatureUnlockBlock struct {
	Signature Serializable `json:"signature"`
}

func (s *SignatureUnlockBlock) Deserialize(data []byte) (int, error) {
	var totalBytesRead int

	// skip type byte
	data = data[OneByte:]
	totalBytesRead += OneByte

	sig, sigBytesRead, err := DeserializeObject(data, SignatureSelector)
	if err != nil {
		return 0, err
	}
	totalBytesRead += sigBytesRead
	s.Signature = sig

	return totalBytesRead, nil
}

func (s *SignatureUnlockBlock) Serialize() ([]byte, error) {
	sigBytes, err := s.Signature.Serialize()
	if err != nil {
		return nil, err
	}
	return append([]byte{UnlockBlockSignature}, sigBytes...), nil
}

// ReferenceUnlockBlock is an unlock block which references a previous unlock block.
type ReferenceUnlockBlock struct {
	Reference uint64 `json:"reference"`
}

func (r *ReferenceUnlockBlock) Deserialize(data []byte) (int, error) {
	data = data[OneByte:]
	reference, referenceByteSize, err := ReadUvarint(bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	r.Reference = reference
	return OneByte + referenceByteSize, nil
}

func (r *ReferenceUnlockBlock) Serialize() ([]byte, error) {
	varIntBuf := make([]byte, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(varIntBuf, r.Reference)
	return append([]byte{UnlockBlockReference}, varIntBuf[:bytesWritten]...), nil
}
