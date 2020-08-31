package iota

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Defines the type of transaction.
type TransactionType = uint64

const (
	// Denotes an unsigned transaction.
	TransactionUnsigned TransactionType = iota

	TransactionIDLength = 32
)

var (
	ErrInputsOrderViolatesLexicalOrder   = errors.New("inputs must be in their lexical order (byte wise) when serialized")
	ErrOutputsOrderViolatesLexicalOrder  = errors.New("outputs must be in their lexical order (byte wise) when serialized")
	ErrInputUTXORefsNotUnique            = errors.New("inputs must each reference a unique UTXO")
	ErrOutputAddrNotUnique               = errors.New("outputs must each deposit to a unique address")
	ErrOutputsSumExceedsTotalSupply      = errors.New("accumulated output balance exceeds total supply")
	ErrOutputDepositsMoreThanTotalSupply = errors.New("an output can not deposit more than the total supply")
)

// TransactionSelector implements SerializableSelectorFunc for transaction types.
func TransactionSelector(txType uint64) (Serializable, error) {
	var seri Serializable
	switch txType {
	case TransactionUnsigned:
		seri = &UnsignedTransaction{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownTransactionType, txType)
	}
	return seri, nil
}

// UnsignedTransaction is the unsigned part of a transaction.
type UnsignedTransaction struct {
	// The inputs of this transaction.
	Inputs Serializables `json:"inputs"`
	// The outputs of this transaction.
	Outputs Serializables `json:"outputs"`
	// The optional embedded payload.
	Payload Serializable `json:"payload"`
}

func (u *UnsignedTransaction) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	data, typeBytesRead, err := ReadTypeAndAdvance(data, TransactionUnsigned, deSeriMode)
	if err != nil {
		return 0, err
	}

	bytesReadTotal := typeBytesRead

	inputs, inputBytesRead, err := DeserializeArrayOfObjects(data, deSeriMode, InputSelector, &inputsArrayBound)
	if err != nil {
		return 0, err
	}
	bytesReadTotal += inputBytesRead

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := ValidateInputs(inputs, InputsUTXORefsUniqueValidator()); err != nil {
			return 0, err
		}
	}

	u.Inputs = inputs

	// advance to outputs
	data = data[inputBytesRead:]
	outputs, outputBytesRead, err := DeserializeArrayOfObjects(data, deSeriMode, OutputSelector, &outputsArrayBound)
	if err != nil {
		return 0, err
	}
	bytesReadTotal += outputBytesRead

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := ValidateOutputs(outputs, OutputsAddrUniqueValidator()); err != nil {
			return 0, err
		}
	}

	u.Outputs = outputs

	// advance to payload
	// TODO: replace with payload deserializer
	data = data[outputBytesRead:]
	payloadLength, payloadLengthByteSize, err := Uvarint(data)
	if err != nil {
		return 0, fmt.Errorf("%w: can't read inner payload length in unsigned transaction", ErrInvalidVarint)
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

func (u *UnsignedTransaction) Serialize(deSeriMode DeSerializationMode) (data []byte, err error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := ValidateInputs(u.Inputs, InputsUTXORefsUniqueValidator()); err != nil {
			return nil, err
		}
		if err := ValidateOutputs(u.Outputs, OutputsAddrUniqueValidator()); err != nil {
			return nil, err
		}
	}

	buf, varintBuf, _ := WriteTypeHeader(TransactionUnsigned)

	bytesWritten := binary.PutUvarint(varintBuf[:], uint64(len(u.Inputs)))
	if _, err := buf.Write(varintBuf[:bytesWritten]); err != nil {
		return nil, err
	}

	var inputsLexicalOrderValidator LexicalOrderFunc
	if deSeriMode.HasMode(DeSeriModePerformValidation) && inputsArrayBound.ElementBytesLexicalOrder {
		inputsLexicalOrderValidator = inputsArrayBound.LexicalOrderValidator()
	}

	for i := range u.Inputs {
		inputSer, err := u.Inputs[i].Serialize(deSeriMode)
		if err != nil {
			return nil, fmt.Errorf("unable to serialize input at index %d: %w", i, err)
		}
		if _, err := buf.Write(inputSer); err != nil {
			return nil, err
		}
		if inputsLexicalOrderValidator != nil {
			if err := inputsLexicalOrderValidator(i, inputSer); err != nil {
				return nil, err
			}
		}
	}

	bytesWritten = binary.PutUvarint(varintBuf[:], uint64(len(u.Outputs)))
	if _, err := buf.Write(varintBuf[:bytesWritten]); err != nil {
		return nil, err
	}

	var outputsLexicalOrderValidator LexicalOrderFunc
	if deSeriMode.HasMode(DeSeriModePerformValidation) && outputsArrayBound.ElementBytesLexicalOrder {
		outputsLexicalOrderValidator = outputsArrayBound.LexicalOrderValidator()
	}

	for i := range u.Outputs {
		outputSer, err := u.Outputs[i].Serialize(deSeriMode)
		if err != nil {
			return nil, fmt.Errorf("unable to serialize output at index %d: %w", i, err)
		}
		if _, err := buf.Write(outputSer); err != nil {
			return nil, err
		}
		if outputsLexicalOrderValidator != nil {
			if err := outputsLexicalOrderValidator(i, outputSer); err != nil {
				return nil, err
			}
		}
	}

	// no payload
	if u.Payload == nil {
		if err := buf.WriteByte(0); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	payloadSer, err := u.Payload.Serialize(deSeriMode)
	if _, err := buf.Write(payloadSer); err != nil {
		return nil, err
	}

	bytesWritten = binary.PutUvarint(varintBuf[:], uint64(len(payloadSer)))
	if _, err := buf.Write(varintBuf[:bytesWritten]); err != nil {
		return nil, err
	}

	if _, err := buf.Write(payloadSer); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// SyntacticallyValid checks whether the unsigned transaction is syntactically valid by checking whether:
//	1. every input references a unique UTXO and has valid UTXO index bounds
//	2. every output deposits to a unique address and deposits more than zero
//	3. the accumulated deposit output is not over the total supply
// The function does not syntactically validate the input or outputs themselves.
func (u *UnsignedTransaction) SyntacticallyValid() error {
	if err := ValidateInputs(u.Inputs,
		InputsUTXORefIndexBoundsValidator(),
		InputsUTXORefsUniqueValidator(),
	); err != nil {
		return err
	}

	if err := ValidateOutputs(u.Outputs,
		OutputsAddrUniqueValidator(),
		OutputsDepositAmountValidator(),
	); err != nil {
		return err
	}

	return nil
}
