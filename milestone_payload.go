package iota

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	MilestonePayloadID                  = 1
	MilestoneInclusionMerkleProofLength = 64
	MilestoneSignatureLength            = 64
	MilestoneHashLength                 = 32
)

// MilestonePayload holds the inclusion merkle proof and milestone signature.
type MilestonePayload struct {
	Index                uint64                                    `json:"index"`
	Timestamp            uint64                                    `json:"timestamp"`
	InclusionMerkleProof [MilestoneInclusionMerkleProofLength]byte `json:"inclusion_merkle_proof"`
	Signature            [MilestoneSignatureLength]byte            `json:"signature"`
}

func (m *MilestonePayload) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	data, typeBytesRead, err := ReadTypeAndAdvance(data, MilestonePayloadID, deSeriMode)
	if err != nil {
		return 0, err
	}

	index, indexBytesRead, err := Uvarint(data)
	if err != nil {
		return 0, fmt.Errorf("%w: can't read milestone index", err)
	}

	m.Index = index
	data = data[indexBytesRead:]
	if len(data) < UInt64ByteSize {
		return 0, fmt.Errorf("%w: can't read milestone timestamp", ErrDeserializationNotEnoughData)
	}
	m.Timestamp = binary.LittleEndian.Uint64(data)
	data = data[UInt64ByteSize:]

	if len(data) < MilestoneInclusionMerkleProofLength+MilestoneSignatureLength {
		return 0, fmt.Errorf("%w: for milestone inclusion merkle proof and signature", ErrDeserializationNotEnoughData)
	}

	copy(m.InclusionMerkleProof[:], data[:MilestoneInclusionMerkleProofLength])
	data = data[MilestoneInclusionMerkleProofLength:]
	copy(m.Signature[:], data[:MilestoneSignatureLength])

	return typeBytesRead + indexBytesRead + UInt64ByteSize + MilestoneInclusionMerkleProofLength + MilestoneSignatureLength, nil
}

func (m *MilestonePayload) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	var varintBuf [binary.MaxVarintLen64]byte
	bytesWritten := binary.PutUvarint(varintBuf[:], MilestonePayloadID)

	var b bytes.Buffer
	if _, err := b.Write(varintBuf[:bytesWritten]); err != nil {
		return nil, err
	}

	bytesWritten = binary.PutUvarint(varintBuf[:], m.Index)
	if _, err := b.Write(varintBuf[:bytesWritten]); err != nil {
		return nil, err
	}

	if err := binary.Write(&b, binary.LittleEndian, m.Timestamp); err != nil {
		return nil, err
	}

	if _, err := b.Write(m.InclusionMerkleProof[:]); err != nil {
		return nil, err
	}

	if _, err := b.Write(m.Signature[:]); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
