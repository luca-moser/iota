package iotapkg

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
	// Defines the end index of the WOTS address portion within a sig locked single deposit.
	SigLockedSingleDepositWOTSAddrEnd = SigLockedSingleDepositWOTSAddrBytesSize - UInt64ByteSize
	// Defines the end index of the Ed25519 address portion within a sig locked single deposit.
	SigLockedSingleDepositEd25519AddrEnd = SigLockedSingleDepositEd25519AddrBytesSize - UInt64ByteSize
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
	// The type of the address.
	AddressType AddressType `json:"address_type"`
	// The actual address.
	Address Serializable `json:"address"`
	// The amount to deposit.
	Amount uint64 `json:"amount"`
}

// ExtractAddressFromSigLockedSingleDeposit extracts the address type and its deserialized from the given serialized sig locked single deposit data.
func ExtractAddressFromSigLockedSingleDeposit(data []byte) (Serializable, AddressType, int, error) {
	l := len(data)
	switch data[SigLockedSingleDepositAddressOffset] {
	case AddressWOTS:
		if err := checkMinByteLength(SigLockedSingleDepositWOTSAddrBytesSize, l); err != nil {
			return nil, 0, 0, err
		}
		var wotsAddr WOTSAddress
		if _, err := wotsAddr.Deserialize(data[SigLockedSingleDepositAddressOffset:SigLockedSingleDepositWOTSAddrEnd]); err != nil {
			return nil, 0, 0, err
		}
		return wotsAddr, AddressWOTS, WOTSAddressSerializedBytesSize, nil
	case AddressEd25519:
		if err := checkMinByteLength(SigLockedSingleDepositEd25519AddrBytesSize, l); err != nil {
			return nil, 0, 0, err
		}
		var ed25519Addr Ed25519Address
		if _, err := ed25519Addr.Deserialize(data[SigLockedSingleDepositAddressOffset:SigLockedSingleDepositEd25519AddrEnd]); err != nil {
			return nil, 0, 0, err
		}
		return ed25519Addr, AddressEd25519, Ed25519AddressSerializedBytesSize, nil
	default:
		return nil, 0, 0, fmt.Errorf("%w: type byte %d", ErrUnknownAddrType, data[SigLockedSingleDepositAddressOffset])
	}
}

func (s *SigLockedSingleDeposit) Deserialize(data []byte) (int, error) {
	l := len(data)
	var err error
	if err := checkMinByteLength(SigLockedSingleDepositBytesMinSize, l); err != nil {
		return 0, err
	}

	// for now we can take the 2nd byte to determine the address type
	var addrBytesRead int
	s.Address, s.AddressType, addrBytesRead, err = ExtractAddressFromSigLockedSingleDeposit(data)
	if err != nil {
		return 0, err
	}

	// read amount of the deposit
	if err := binary.Read(bytes.NewReader(data[l-UInt64ByteSize:]), binary.LittleEndian, &s.Amount); err != nil {
		return 0, fmt.Errorf("unable to deserialize deposit amount: %w", err)
	}

	return OneByte + addrBytesRead + UInt64ByteSize, nil
}

func (s *SigLockedSingleDeposit) Serialize() (data []byte, err error) {
	var b bytes.Buffer
	if err := b.WriteByte(OutputSigLockedSingleDeposit); err != nil {
		return nil, err
	}
	addrBytes, err := s.Address.Serialize()
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
