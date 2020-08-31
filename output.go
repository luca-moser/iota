package iota

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// Defines the type of outputs.
type OutputType = uint64

const (
	// Denotes a type of output which is locked by a signature and deposits onto a single address.
	OutputSigLockedSingleDeposit OutputType = iota
)

var (
	ErrDepositAmountMustBeGreaterThanZero = errors.New("deposit amount must be greater than zero")
)

// OutputSelector implements SerializableSelectorFunc for output types.
func OutputSelector(outputType uint64) (Serializable, error) {
	var seri Serializable
	switch outputType {
	case OutputSigLockedSingleDeposit:
		seri = &SigLockedSingleDeposit{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownOutputType, outputType)
	}
	return seri, nil
}

// SigLockedSingleDeposit is an output type which can be unlocked via a signature. It deposits onto one single address.
type SigLockedSingleDeposit struct {
	// The actual address.
	Address Serializable `json:"address"`
	// The amount to deposit.
	Amount uint64 `json:"amount"`
}

func (s *SigLockedSingleDeposit) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	data, typeBytesRead, err := ReadTypeAndAdvance(data, OutputSigLockedSingleDeposit, deSeriMode)
	if err != nil {
		return 0, err
	}

	addr, addrBytesRead, err := DeserializeObject(data, deSeriMode, AddressSelector)
	if err != nil {
		return 0, err
	}

	s.Address = addr
	data = data[addrBytesRead:]

	// read amount of the deposit
	if len(data) < UInt64ByteSize {
		return 0, fmt.Errorf("%w: can't read deposit amount", ErrDeserializationNotEnoughData)
	}

	s.Amount = binary.LittleEndian.Uint64(data)

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := outputAmountValidator(-1, s); err != nil {
			return 0, err
		}
	}

	return typeBytesRead + addrBytesRead + UInt64ByteSize, nil
}

func (s *SigLockedSingleDeposit) Serialize(deSeriMode DeSerializationMode) (data []byte, err error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := outputAmountValidator(-1, s); err != nil {
			return nil, err
		}
	}

	buf, _, _ := WriteTypeHeader(OutputSigLockedSingleDeposit)
	addrBytes, err := s.Address.Serialize(deSeriMode)
	if err != nil {
		return nil, err
	}

	if _, err := buf.Write(addrBytes); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.LittleEndian, s.Amount); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// OutputsValidatorFunc which given the index of an output and the output itself, runs validations and returns an error if any should fail.
type OutputsValidatorFunc func(index int, output *SigLockedSingleDeposit) error

// OutputsAddrUniqueValidator returns a validator which checks that all addresses are unique.
func OutputsAddrUniqueValidator() OutputsValidatorFunc {
	set := map[string]int{}
	return func(index int, dep *SigLockedSingleDeposit) error {
		var b strings.Builder
		// can't be reduced to one b.Write()
		switch addr := dep.Address.(type) {
		case *WOTSAddress:
			if _, err := b.Write(addr[:]); err != nil {
				return err
			}
		case *Ed25519Address:
			if _, err := b.Write(addr[:]); err != nil {
				return err
			}
		}
		k := b.String()
		if j, has := set[k]; has {
			return fmt.Errorf("%w: output %d and %d share the same  address", ErrOutputAddrNotUnique, j, index)
		}
		set[k] = index
		return nil
	}
}

// OutputsDepositAmountValidator returns a validator which checks that:
//	1. every output deposits more than zero
//	2. every output deposits less than the total supply
//	3. the sum of deposits does not exceed the total supply
// If -1 is passed to the validator func, then the sum is not aggregated over multiple calls.
func OutputsDepositAmountValidator() OutputsValidatorFunc {
	var sum uint64
	return func(index int, dep *SigLockedSingleDeposit) error {
		if dep.Amount == 0 {
			return fmt.Errorf("%w: output %d", ErrDepositAmountMustBeGreaterThanZero, index)
		}
		if dep.Amount > TokenSupply {
			return fmt.Errorf("%w: output %d", ErrOutputDepositsMoreThanTotalSupply, index)
		}
		if sum+dep.Amount > TokenSupply {
			return fmt.Errorf("%w: output %d", ErrOutputsSumExceedsTotalSupply, index)
		}
		if index != -1 {
			sum += dep.Amount
		}
		return nil
	}
}

// supposed to be called with -1 as input in order to be used over multiple calls.
var outputAmountValidator = OutputsDepositAmountValidator()

// ValidateOutputs validates the outputs by running them against the given OutputsValidatorFunc.
func ValidateOutputs(outputs Serializables, funcs ...OutputsValidatorFunc) error {
	for i, output := range outputs {
		dep, ok := output.(*SigLockedSingleDeposit)
		if !ok {
			return fmt.Errorf("%w: can only validate on signature locked single deposits", ErrUnknownOutputType)
		}
		for _, f := range funcs {
			if err := f(i, dep); err != nil {
				return err
			}
		}
	}
	return nil
}
