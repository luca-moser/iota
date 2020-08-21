package iotapkg

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Defines the type of inputs.
type InputType = byte

const (
	// A type of input which references an unspent transaction output.
	InputUTXO InputType = iota
)

// InputSelector implements SerializableSelectorFunc for input types.
func InputSelector(typeByte byte) (Serializable, error) {
	var seri Serializable
	switch typeByte {
	case InputUTXO:
		seri = &UTXOInput{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownInputType, typeByte)
	}
	return seri, nil
}

// input type + tx id + index
const UTXOInputMinSize = OneByte + TransactionIDLength + OneByte

// input type + tx id + max index
const UTXOInputMaxSize = OneByte + TransactionIDLength + binary.MaxVarintLen64

// UTXOInput references an unspent transaction output by the signed transaction payload's hash and the corresponding index of the output.
type UTXOInput struct {
	// The transaction ID of the referenced transaction.
	TransactionID [TransactionIDLength]byte `json:"transaction_id"`
	// The output index of the output on the referenced transaction.
	TransactionOutputIndex uint16 `json:"transaction_output_index"`
}

func (u *UTXOInput) Deserialize(data []byte) (int, error) {
	if err := checkMinByteLength(UTXOInputMinSize, len(data)); err != nil {
		return 0, fmt.Errorf("invalid utxo input bytes: %w", err)
	}

	// skip type
	data = data[OneByte:]

	// transaction id
	copy(u.TransactionID[:], data[:TransactionIDLength])

	// output index
	outputIndex, outputIndexByteSize, err := ReadUvarint(bytes.NewReader(data[TransactionIDLength:]))
	if err != nil {
		return 0, fmt.Errorf("%w: unable to read output index", err)
	}

	// TODO: validate index bound before truncation
	u.TransactionOutputIndex = uint16(outputIndex)
	return OneByte + TransactionIDLength + outputIndexByteSize, nil
}

func (u *UTXOInput) Serialize() (data []byte, err error) {
	var b bytes.Buffer
	if err := b.WriteByte(InputUTXO); err != nil {
		return nil, err
	}
	if _, err := b.Write(u.TransactionID[:]); err != nil {
		return nil, err
	}

	// TODO: replace with single alloc
	varIntBuf := make([]byte, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(varIntBuf, uint64(u.TransactionOutputIndex))
	if _, err := b.Write(varIntBuf[:bytesWritten]); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
