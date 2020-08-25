package iota

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	MessageVersion    = 1
	MessageHashLength = 32
	MessageMinSize    = 2*MessageHashLength + OneByte + UInt64ByteSize
)

// PayloadSelector implements SerializableSelectorFunc for payload types.
func PayloadSelector(typeByte byte) (Serializable, error) {
	var seri Serializable
	switch typeByte {
	case SignedTransactionPayloadID:
		seri = &SignedTransactionPayload{}
	case UnsignedDataPayloadID:
		seri = &UnsignedDataPayload{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownPayloadType, typeByte)
	}
	return seri, nil
}

// Message carries a payload and references two other messages.
type Message struct {
	Parent1 [MessageHashLength]byte `json:"parent_1"`
	Parent2 [MessageHashLength]byte `json:"parent_2"`
	Payload Serializable            `json:"payload"`
	Nonce   uint64                  `json:"nonce"`
}

func (m *Message) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkType(data, MessageVersion); err != nil {
			return 0, fmt.Errorf("unable to deserialize message: %w", err)
		}
		if err := checkMinByteLength(MessageMinSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid message bytes: %w", err)
		}
	}
	l := len(data)

	// read parents
	data = data[OneByte:]
	copy(m.Parent1[:], data[:MessageHashLength])
	data = data[MessageHashLength:]
	copy(m.Parent2[:], data[:MessageHashLength])
	data = data[MessageHashLength:]

	// read payload
	payloadLength, payloadLengthByteSize, err := ReadUvarint(bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	data = data[payloadLengthByteSize:]

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: validate payload length
	}

	var payloadBytesConsumed int
	if payloadLength != 0 {
		payload, err := PayloadSelector(data[0])
		if err != nil {
			return 0, err
		}
		payloadBytesConsumed, err = payload.Deserialize(data, deSeriMode)
		if err != nil {
			return 0, err
		}
		m.Payload = payload
	}

	// must have consumed entire data slice minus the nonce
	data = data[payloadBytesConsumed:]
	if leftOver := len(data) - UInt64ByteSize; leftOver != 0 {
		return 0, fmt.Errorf("%w: %d are still available", ErrDeserializationNotAllConsumed, leftOver)
	}

	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &m.Nonce); err != nil {
		return 0, err
	}

	return l, nil
}

func (m *Message) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
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

	switch {
	case m.Payload == nil:
		if err := b.WriteByte(0); err != nil {
			return nil, err
		}
	default:
		payloadData, err := m.Payload.Serialize(deSeriMode)
		if err != nil {
			return nil, err
		}

		if deSeriMode.HasMode(DeSeriModePerformValidation) {
			// TODO: check payload length
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
	}

	if err := binary.Write(&b, binary.LittleEndian, m.Nonce); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
