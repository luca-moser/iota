package iota

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	UnsignedDataPayloadID      = 2
	UnsignedDataPayloadMinSize = 2 * OneByte
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
	data = data[OneByte:]

	// read data length
	dataLength, dataLengthBytesRead, err := ReadUvarint(bytes.NewReader(data))
	if err != nil {
		return 0, err
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

	return OneByte + dataLengthBytesRead + int(dataLength), nil
}

func (u *UnsignedDataPayload) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: check data length
	}
	var b bytes.Buffer
	if err := b.WriteByte(UnsignedDataPayloadID); err != nil {
		return nil, err
	}

	// write data length
	varIntBuf := make([]byte, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(varIntBuf, uint64(len(u.Data)))
	if _, err := b.Write(varIntBuf[:bytesWritten]); err != nil {
		return nil, err
	}

	// write data
	if _, err := b.Write(u.Data); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
