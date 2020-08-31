package iota

import (
	"encoding/binary"
	"fmt"
)

const (
	UnsignedDataPayloadID = 2
)

// UnsignedDataPayload is a payload which holds a blob of unspecific data.
type UnsignedDataPayload struct {
	Data []byte `json:"data"`
}

func (u *UnsignedDataPayload) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	data, typeBytesRead, err := ReadTypeAndAdvance(data, UnsignedDataPayloadID, deSeriMode)
	if err != nil {
		return 0, err
	}

	// read data length
	dataLength, dataLengthBytesRead, err := Uvarint(data)
	if err != nil {
		return 0, fmt.Errorf("%w: unable to read unsigned data payload's data length", err)
	}

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: check data length
	}

	data = data[dataLengthBytesRead:]
	bytesAvailable := uint64(len(data)) - dataLength
	if bytesAvailable < 0 {
		return 0, fmt.Errorf("%w: unsigned data payload length denotes too many bytes (%d bytes)", ErrDeserializationNotEnoughData, dataLength)
	}

	u.Data = make([]byte, dataLength)
	copy(u.Data, data[:dataLength])

	return typeBytesRead + dataLengthBytesRead + int(dataLength), nil
}

func (u *UnsignedDataPayload) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	buf, varintBuf, _ := WriteTypeHeader(UnsignedDataPayloadID)

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: check data length
	}

	bytesWritten := binary.PutUvarint(varintBuf[:], uint64(len(u.Data)))
	if _, err := buf.Write(varintBuf[:bytesWritten]); err != nil {
		return nil, err
	}

	if _, err := buf.Write(u.Data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
