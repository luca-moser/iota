package iota

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// Defines the type of outputs.
type OutputType = byte

const (
	// Denotes a type of output which is locked by a signature and deposits onto a single address.
	OutputSigLockedSingleDeposit OutputType = iota

	// The size of a sig locked single deposit containing a WOTS address as its deposit address.
	SigLockedSingleDepositWOTSAddrBytesSize = OneByte + WOTSAddressSerializedBytesSize + UInt64ByteSize
	// The size of a sig locked single deposit containing an Ed25519 address as its deposit address.
	SigLockedSingleDepositEd25519AddrBytesSize = OneByte + Ed25519AddressSerializedBytesSize + UInt64ByteSize

	// Defines the minimum size a sig locked single deposit must be.
	SigLockedSingleDepositBytesMinSize = SigLockedSingleDepositEd25519AddrBytesSize
	// Defines the offset at which the address portion within a sig locked single deposit begins.
	SigLockedSingleDepositAddressOffset = 1
)

var (
	ErrDepositAmountMustBeGreaterThanZero = errors.New("deposit amount must be greater than zero")
)

// OutputSelector implements SerializableSelectorFunc for output types.
func OutputSelector(typeByte byte) (Serializable, error) {
	var seri Serializable
	switch typeByte {
	case OutputSigLockedSingleDeposit:
		seri = &SigLockedSingleDeposit{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownOutputType, typeByte)
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

func (s *SigLockedSingleDeposit) Deserialize(data []byte, skipValidation bool) (int, error) {
	if !skipValidation {
		if err := checkType(data, OutputSigLockedSingleDeposit); err != nil {
			return 0, fmt.Errorf("unable to deserialize signature locked single deposit: %w", err)
		}

		if err := checkMinByteLength(SigLockedSingleDepositBytesMinSize, len(data)); err != nil {
			return 0, err
		}
	}

	bytesReadTotal := OneByte
	data = data[OneByte:]

	addr, addrBytesRead, err := DeserializeObject(data, skipValidation, AddressSelector)
	if err != nil {
		return 0, err
	}
	bytesReadTotal += addrBytesRead
	s.Address = addr
	data = data[addrBytesRead:]

	// read amount of the deposit
	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &s.Amount); err != nil {
		return 0, fmt.Errorf("unable to deserialize deposit amount: %w", err)
	}

	if !skipValidation {
		if err := outputAmountValidator(-1, s); err != nil {
			return 0, err
		}
	}

	return OneByte + addrBytesRead + UInt64ByteSize, nil
}

func (s *SigLockedSingleDeposit) Serialize(skipValidation bool) (data []byte, err error) {
	if !skipValidation {
		if err := outputAmountValidator(-1, s); err != nil {
			return nil, err
		}
	}

	var b bytes.Buffer
	if err := b.WriteByte(OutputSigLockedSingleDeposit); err != nil {
		return nil, err
	}
	addrBytes, err := s.Address.Serialize(skipValidation)
	if err != nil {
		return nil, err
	}
	if _, err := b.Write(addrBytes); err != nil {
		return nil, err
	}
	if err := binary.Write(&b, binary.LittleEndian, s.Amount); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
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
