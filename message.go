package iota

import (
	"bytes"
	"encoding/binary"
)

const (
	MessageVersion    = 1
	MessageHashLength = 32
)

// Message is an envelope for payloads which references two other messages.
type Message struct {
	Parent1 [MessageHashLength]byte `json:"parent_1"`
	Parent2 [MessageHashLength]byte `json:"parent_2"`
	Payload Serializable            `json:"payload"`
	Nonce   uint64                  `json:"nonce"`
}

func (m Message) Deserialize(data []byte) (int, error) {
	return 0, nil
}

func (m Message) Serialize() ([]byte, error) {
	var b bytes.Buffer
	if err := b.WriteByte(MessageVersion); err != nil {
		return nil, err
	}

	if _, err := b.Write(m.Parent1[:]); err != nil {
		return nil, err
	}

	if _, err := b.Write(m.Parent2[:]); err != nil {
		return nil, err
	}

	payloadData, err := m.Payload.Serialize()
	if err != nil {
		return nil, err
	}

	// write payload length
	varIntBuf := make([]byte, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(varIntBuf, uint64(len(payloadData)))
	if _, err := b.Write(varIntBuf[:bytesWritten]); err != nil {
		return nil, err
	}

	// actual payload
	if _, err := b.Write(payloadData); err != nil {
		return nil, err
	}

	if err := binary.Write(&b, binary.LittleEndian, m.Nonce); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
