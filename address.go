package iota

import (
	"encoding/binary"
	"fmt"
)

// Defines the type of addresses.
type AddressType = uint64

const (
	// Denotes a WOTS address.
	AddressWOTS AddressType = iota
	// Denotes a Ed25519 address.
	AddressEd25519

	// The length of a WOTS address.
	WOTSAddressBytesLength = 49

	// The length of a Ed25519 address
	Ed25519AddressBytesLength = 32
)

// AddressSelector implements SerializableSelectorFunc for address types.
func AddressSelector(addrType uint64) (Serializable, error) {
	var seri Serializable
	switch addrType {
	case AddressWOTS:
		seri = &WOTSAddress{}
	case AddressEd25519:
		seri = &Ed25519Address{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownAddrType, addrType)
	}
	return seri, nil
}

// Defines a WOTS address.
type WOTSAddress [WOTSAddressBytesLength]byte

func (wotsAddr *WOTSAddress) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	data, typeBytesRead, err := ReadTypeAndAdvance(data, AddressWOTS, deSeriMode)
	if err != nil {
		return 0, err
	}

	if len(data) < WOTSAddressBytesLength {
		return 0, fmt.Errorf("%w: can't read WOTS address, need %d, have %d", ErrDeserializationNotEnoughData, WOTSAddressBytesLength, len(data))
	}

	copy(wotsAddr[:], data)
	return typeBytesRead + WOTSAddressBytesLength, nil
}

func (wotsAddr *WOTSAddress) Serialize(deSeriMode DeSerializationMode) (data []byte, err error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: check T5B1 encoding
	}

	var varintBuf [binary.MaxVarintLen64]byte
	bytesWritten := binary.PutUvarint(varintBuf[:], AddressWOTS)
	b := make([]byte, bytesWritten+WOTSAddressBytesLength)
	copy(b[:bytesWritten], varintBuf[:bytesWritten])
	copy(b[bytesWritten:], wotsAddr[:])
	return b[:], nil
}

// Defines an Ed25519 address.
type Ed25519Address [Ed25519AddressBytesLength]byte

func (edAddr *Ed25519Address) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	data, typeBytesRead, err := ReadTypeAndAdvance(data, AddressEd25519, deSeriMode)
	if err != nil {
		return 0, err
	}

	if len(data) < Ed25519AddressBytesLength {
		return 0, fmt.Errorf("%w: can't read Ed25519 address", ErrDeserializationNotEnoughData)
	}

	copy(edAddr[:], data)
	return typeBytesRead + Ed25519AddressBytesLength, nil
}

func (edAddr *Ed25519Address) Serialize(deSeriMode DeSerializationMode) (data []byte, err error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO:
	}
	var varintBuf [binary.MaxVarintLen64]byte
	bytesWritten := binary.PutUvarint(varintBuf[:], AddressEd25519)

	b := make([]byte, bytesWritten+Ed25519AddressBytesLength)
	copy(b[:bytesWritten], varintBuf[:bytesWritten])
	copy(b[bytesWritten:], edAddr[:])
	return b[:], nil
}
