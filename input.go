package iota

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// Defines the type of inputs.
type InputType = uint64

const (
	// A type of input which references an unspent transaction output.
	InputUTXO InputType = iota

	RefUTXOIndexMin = 0
	RefUTXOIndexMax = 126
)

var (
	ErrRefUTXOIndexInvalid = errors.New(fmt.Sprintf("the referenced UTXO index must be between %d and %d (inclusive)", RefUTXOIndexMin, RefUTXOIndexMax))
)

// InputSelector implements SerializableSelectorFunc for input types.
func InputSelector(inputType uint64) (Serializable, error) {
	var seri Serializable
	switch inputType {
	case InputUTXO:
		seri = &UTXOInput{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownInputType, inputType)
	}
	return seri, nil
}

// UTXOInput references an unspent transaction output by the signed transaction payload's hash and the corresponding index of the output.
type UTXOInput struct {
	// The transaction ID of the referenced transaction.
	TransactionID [TransactionIDLength]byte `json:"transaction_id"`
	// The output index of the output on the referenced transaction.
	TransactionOutputIndex uint64 `json:"transaction_output_index"`
}

func (u *UTXOInput) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	data, typeBytesRead, err := ReadTypeAndAdvance(data, InputUTXO, deSeriMode)
	if err != nil {
		return 0, err
	}

	if len(data) < TransactionIDLength {
		return 0, fmt.Errorf("%w: unable to read transaction hash", ErrDeserializationNotEnoughData)
	}

	// transaction id
	copy(u.TransactionID[:], data[:TransactionIDLength])
	data = data[TransactionIDLength:]

	// output index
	outputIndex, outputIndexByteSize, err := Uvarint(data)
	if err != nil {
		return 0, fmt.Errorf("%w: unable to read transaction output index", err)
	}

	u.TransactionOutputIndex = outputIndex

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := utxoInputRefBoundsValidator(-1, u); err != nil {
			return 0, err
		}
	}

	return typeBytesRead + TransactionIDLength + outputIndexByteSize, nil
}

func (u *UTXOInput) Serialize(deSeriMode DeSerializationMode) (data []byte, err error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := utxoInputRefBoundsValidator(-1, u); err != nil {
			return nil, err
		}
	}

	buf, varintBuf, _ := WriteTypeHeader(InputUTXO)
	if _, err := buf.Write(u.TransactionID[:]); err != nil {
		return nil, err
	}

	bytesWritten := binary.PutUvarint(varintBuf[:], u.TransactionOutputIndex)
	if _, err := buf.Write(varintBuf[:bytesWritten]); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// InputsValidatorFunc which given the index of an input and the input itself, runs validations and returns an error if any should fail.
type InputsValidatorFunc func(index int, input *UTXOInput) error

// InputsUTXORefsUniqueValidator returns a validator which checks that every input has a unique UTXO ref.
func InputsUTXORefsUniqueValidator() InputsValidatorFunc {
	set := map[string]int{}
	return func(index int, input *UTXOInput) error {
		var b strings.Builder
		if _, err := b.Write(input.TransactionID[:]); err != nil {
			return err
		}
		if err := binary.Write(&b, binary.LittleEndian, input.TransactionOutputIndex); err != nil {
			return err
		}
		k := b.String()
		if j, has := set[k]; has {
			return fmt.Errorf("%w: input %d and %d share the same UTXO ref", ErrInputUTXORefsNotUnique, j, index)
		}
		set[k] = index
		return nil
	}
}

// InputsUTXORefIndexBoundsValidator returns a validator which checks that the UTXO ref index is within bounds.
func InputsUTXORefIndexBoundsValidator() InputsValidatorFunc {
	return func(index int, input *UTXOInput) error {
		if input.TransactionOutputIndex < RefUTXOIndexMin || input.TransactionOutputIndex > RefUTXOIndexMax {
			return fmt.Errorf("%w: input %d", ErrRefUTXOIndexInvalid, index)
		}
		return nil
	}
}

var utxoInputRefBoundsValidator = InputsUTXORefIndexBoundsValidator()

// ValidateInputs validates the inputs by running them against the given InputsValidatorFunc.
func ValidateInputs(inputs Serializables, funcs ...InputsValidatorFunc) error {
	for i, input := range inputs {
		dep, ok := input.(*UTXOInput)
		if !ok {
			return fmt.Errorf("%w: can only validate on UTXO inputs", ErrUnknownInputType)
		}
		for _, f := range funcs {
			if err := f(i, dep); err != nil {
				return err
			}
		}
	}
	return nil
}
