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

func (s *SigLockedSingleDeposit) Deserialize(data []byte) (int, error) {
	if err := checkMinByteLength(SigLockedSingleDepositBytesMinSize, len(data)); err != nil {
		return 0, err
	}

	var bytesReadTotal int
	data = data[OneByte:]
	bytesReadTotal++

	addr, addrBytesRead, err := DeserializeObject(data, AddressSelector)
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
