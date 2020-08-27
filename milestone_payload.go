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
	MilestonePayloadMinSize             = OneByte + OneByte + UInt64ByteSize + MilestoneInclusionMerkleProofLength + MilestoneSignatureLength
)

// MilestonePayload holds the inclusion merkle proof and milestone signature.
type MilestonePayload struct {
	Index                uint64                                    `json:"index"`
	Timestamp            uint64                                    `json:"timestamp"`
	InclusionMerkleProof [MilestoneInclusionMerkleProofLength]byte `json:"inclusion_merkle_proof"`
	Signature            [MilestoneSignatureLength]byte            `json:"signature"`
}

func (m *MilestonePayload) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkType(data, MilestonePayloadID); err != nil {
			return 0, fmt.Errorf("unable to deserialize milestone payload: %w", err)
		}
		if err := checkMinByteLength(MilestonePayloadMinSize, len(data)); err != nil {
			return 0, err
		}
	}
	data = data[OneByte:]

	index, indexBytesRead, err := ReadUvarint(bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	m.Index = index
	data = data[indexBytesRead:]

	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &m.Timestamp); err != nil {
		return 0, err
	}
	data = data[UInt64ByteSize:]

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if len(data) < MilestoneInclusionMerkleProofLength+MilestoneSignatureLength {
			return 0, fmt.Errorf("%w: for milestone inclusion merkle proof and signature", ErrDeserializationNotEnoughData)
		}
	}

	copy(m.InclusionMerkleProof[:], data[:MilestoneInclusionMerkleProofLength])
	data = data[MilestoneInclusionMerkleProofLength:]
	copy(m.Signature[:], data[:MilestoneSignatureLength])

	return OneByte + indexBytesRead + UInt64ByteSize + MilestoneInclusionMerkleProofLength + MilestoneSignatureLength, nil
}

func (m *MilestonePayload) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	var b bytes.Buffer
	if err := b.WriteByte(MilestonePayloadID); err != nil {
		return nil, err
	}

	varIntBuf := make([]byte, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(varIntBuf, m.Index)
	if _, err := b.Write(varIntBuf[:bytesWritten]); err != nil {
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
