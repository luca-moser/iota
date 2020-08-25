package iota

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// Defines the type of inputs.
type InputType = byte

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
	TransactionOutputIndex byte `json:"transaction_output_index"`
}

func (u *UTXOInput) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkType(data, InputUTXO); err != nil {
			return 0, fmt.Errorf("unable to deserialize UTXO input: %w", err)
		}

		if err := checkMinByteLength(UTXOInputMinSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid utxo input bytes: %w", err)
		}
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

	outputIndexByte := byte(outputIndex)
	u.TransactionOutputIndex = outputIndexByte

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := utxoInputRefBoundsValidator(-1, u); err != nil {
			return 0, err
		}
	}

	return OneByte + TransactionIDLength + outputIndexByteSize, nil
}

func (u *UTXOInput) Serialize(deSeriMode DeSerializationMode) (data []byte, err error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := utxoInputRefBoundsValidator(-1, u); err != nil {
			return nil, err
		}
	}

	var b bytes.Buffer
	if err := b.WriteByte(InputUTXO); err != nil {
		return nil, err
	}

	if _, err := b.Write(u.TransactionID[:]); err != nil {
		return nil, err
	}

	if err := b.WriteByte(u.TransactionOutputIndex); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
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
		if err := b.WriteByte(input.TransactionOutputIndex); err != nil {
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
