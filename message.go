package iota

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	MessageVersion    = 1
	MessageHashLength = 32
	// version + 2 msg hashes + uint16 payload length + nonce
	MessageMinSize = TypeDenotationByteSize + 2*MessageHashLength + UInt16ByteSize + UInt64ByteSize
)

// PayloadSelector implements SerializableSelectorFunc for payload types.
func PayloadSelector(payloadType uint32) (Serializable, error) {
	var seri Serializable
	switch payloadType {
	case SignedTransactionPayloadID:
		seri = &SignedTransactionPayload{}
	case UnsignedDataPayloadID:
		seri = &UnsignedDataPayload{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownPayloadType, payloadType)
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
	data = data[TypeDenotationByteSize:]
	copy(m.Parent1[:], data[:MessageHashLength])
	data = data[MessageHashLength:]
	copy(m.Parent2[:], data[:MessageHashLength])
	data = data[MessageHashLength:]

	// read payload
	payloadLength := binary.LittleEndian.Uint16(data)
	data = data[UInt16ByteSize:]

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: validate payload length
	}

	var payloadBytesConsumed int
	if payloadLength != 0 {
		payload, err := PayloadSelector(binary.LittleEndian.Uint32(data))
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

	m.Nonce = binary.LittleEndian.Uint64(data)
	return l, nil
}

func (m *Message) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	if m.Payload == nil {
		var b [MessageMinSize]byte
		binary.LittleEndian.PutUint32(b[:TypeDenotationByteSize], MessageVersion)
		copy(b[TypeDenotationByteSize:], m.Parent1[:])
		copy(b[TypeDenotationByteSize+MessageHashLength:], m.Parent2[:])
		binary.LittleEndian.PutUint16(b[TypeDenotationByteSize+MessageHashLength*2:], 0)
		binary.LittleEndian.PutUint64(b[len(b)-UInt64ByteSize:], m.Nonce)
		return b[:], nil
	}

	b := bytes.NewBuffer(make([]byte, 0, MessageMinSize))
	binary.LittleEndian.PutUint32(b.Next(UInt32ByteSize), MessageVersion)
	if _, err := b.Write(m.Parent1[:]); err != nil {
		return nil, err
	}
	if _, err := b.Write(m.Parent2[:]); err != nil {
		return nil, err
	}

	payloadData, err := m.Payload.Serialize(deSeriMode)
	if err != nil {
		return nil, err
	}
	payloadLength := uint16(len(payloadData))

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: check payload length
	}

	binary.LittleEndian.PutUint16(b.Next(UInt16ByteSize), payloadLength)
	if _, err := b.Write(payloadData); err != nil {
		return nil, err
	}

	binary.LittleEndian.PutUint64(b.Next(UInt64ByteSize), m.Nonce)

	return b.Bytes(), nil
}
