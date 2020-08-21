package iotapkg

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Defines the type of transaction.
type TransactionType = byte

const (
	// Denotes an unsigned transaction.
	TransactionUnsigned TransactionType = iota
)

// TransactionSelector implements SerializableSelectorFunc for transaction types.
func TransactionSelector(typeByte byte) (Serializable, error) {
	var seri Serializable
	switch typeByte {
	case TransactionUnsigned:
		seri = &UnsignedTransaction{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownTransactionType, typeByte)
	}
	return seri, nil
}

// UnsignedTransaction is the unsigned part of a transaction.
type UnsignedTransaction struct {
	// The inputs of this transaction.
	Inputs []Serializable `json:"inputs"`
	// The outputs of this transaction.
	Outputs []Serializable `json:"outputs"`
	// The optional embedded payload.
	Payload Serializable `json:"payload"`
}

func (u *UnsignedTransaction) Deserialize(data []byte) (int, error) {
	var bytesReadTotal int

	// skip type byte
	data = data[OneByte:]
	bytesReadTotal += OneByte

	inputs, inputBytesRead, err := DeserializeArrayOfObjects(data, InputSelector)
	if err != nil {
		return 0, err
	}
	bytesReadTotal += inputBytesRead
	u.Inputs = inputs

	// advance to outputs
	data = data[inputBytesRead:]
	outputs, outputBytesRead, err := DeserializeArrayOfObjects(data, OutputSelector)
	if err != nil {
		return 0, err
	}
	bytesReadTotal += outputBytesRead
	u.Outputs = outputs

	// advance to payload
	// TODO: replace with payload deserializer
	data = data[bytesReadTotal:]
	payloadLength, payloadLengthByteSize, err := ReadUvarint(bytes.NewReader(data[:binary.MaxVarintLen64]))
	if err != nil {
		return 0, err
	}
	bytesReadTotal += payloadLengthByteSize

	if payloadLength == 0 {
		return bytesReadTotal, nil
	}

	// TODO: payload extraction logic
	data = data[payloadLengthByteSize:]
	switch data[0] {

	}
	bytesReadTotal += int(payloadLength)

	return bytesReadTotal, nil
}

func (u *UnsignedTransaction) Serialize() (data []byte, err error) {
	var b bytes.Buffer
	if err := b.WriteByte(TransactionUnsigned); err != nil {
		return nil, err
	}

	varIntBuf := make([]byte, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(varIntBuf, uint64(len(u.Inputs)))

	if _, err := b.Write(varIntBuf[:bytesWritten]); err != nil {
		return nil, err
	}

	for i := range u.Inputs {
		inputSer, err := u.Inputs[i].Serialize()
		if err != nil {
			return nil, fmt.Errorf("unable to serialize input at index %d: %w", i, err)
		}
		if _, err := b.Write(inputSer); err != nil {
			return nil, err
		}
	}

	// reuse varIntBuf (this is safe as b.Write() copies the bytes)
	bytesWritten = binary.PutUvarint(varIntBuf, uint64(len(u.Outputs)))
	if _, err := b.Write(varIntBuf[:bytesWritten]); err != nil {
		return nil, err
	}

	for i := range u.Outputs {
		outputSer, err := u.Outputs[i].Serialize()
		if err != nil {
			return nil, fmt.Errorf("unable to serialize output at index %d: %w", i, err)
		}
		if _, err := b.Write(outputSer); err != nil {
			return nil, err
		}
	}

	// no payload
	if u.Payload == nil {
		if err := b.WriteByte(0); err != nil {
			return nil, err
		}
		return b.Bytes(), nil
	}

	payloadSer, err := u.Payload.Serialize()
	if _, err := b.Write(payloadSer); err != nil {
		return nil, err
	}

	bytesWritten = binary.PutUvarint(varIntBuf, uint64(len(payloadSer)))
	if _, err := b.Write(varIntBuf[:bytesWritten]); err != nil {
		return nil, err
	}

	if _, err := b.Write(payloadSer); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
