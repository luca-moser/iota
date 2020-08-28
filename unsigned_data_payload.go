package iota

import (
	"encoding/binary"
	"fmt"
)

const (
	UnsignedDataPayloadID      uint32 = 2
	UnsignedDataPayloadMinSize        = TypeDenotationByteSize + ArrayLengthByteSize
)

// UnsignedDataPayload is a payload which holds a blob of unspecific data.
type UnsignedDataPayload struct {
	Data []byte `json:"data"`
}

func (u *UnsignedDataPayload) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkType(data, UnsignedDataPayloadID); err != nil {
			return 0, fmt.Errorf("unable to deserialize unsigned data payload: %w", err)
		}
		if err := checkMinByteLength(UnsignedDataPayloadMinSize, len(data)); err != nil {
			return 0, err
		}
	}
	data = data[TypeDenotationByteSize:]

	// read data length
	dataLength := binary.LittleEndian.Uint16(data)

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: check data length
	}

	data = data[ArrayLengthByteSize:]
	bytesAvailable := uint32(len(data)) - uint32(dataLength)
	if bytesAvailable < 0 {
		return 0, fmt.Errorf("%w: unsigned data payload length denotes too many bytes (%d bytes)", ErrDeserializationNotEnoughData, dataLength)
	}

	u.Data = make([]byte, dataLength)
	copy(u.Data, data[:dataLength])

	return TypeDenotationByteSize + ArrayLengthByteSize + int(dataLength), nil
}

func (u *UnsignedDataPayload) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: check data length
	}

	b := make([]byte, TypeDenotationByteSize+ArrayLengthByteSize+len(u.Data))
	binary.LittleEndian.PutUint32(b, UnsignedDataPayloadID)
	binary.LittleEndian.PutUint16(b[TypeDenotationByteSize:], uint16(len(u.Data)))
	copy(b[TypeDenotationByteSize+ArrayLengthByteSize:], u.Data)

	return b, nil
}
