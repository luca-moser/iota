package iota

import (
	"fmt"
)

// Defines the type of addresses.
type AddressType = byte

const (
	// Denotes a WOTS address.
	AddressWOTS AddressType = iota
	// Denotes a Ed25510 address.
	AddressEd25519

	// The length of a WOTS address.
	WOTSAddressBytesLength = 49
	// The size of a serialized WOTS address with its type denoting byte.
	WOTSAddressSerializedBytesSize = OneByte + WOTSAddressBytesLength

	// The length of a Ed25519 address
	Ed25519AddressBytesLength = 32
	// The size of a serialized Ed25519 address with its type denoting byte.
	Ed25519AddressSerializedBytesSize = OneByte + Ed25519AddressBytesLength
)

// AddressSelector implements SerializableSelectorFunc for address types.
func AddressSelector(typeByte byte) (Serializable, error) {
	var seri Serializable
	switch typeByte {
	case AddressWOTS:
		seri = &WOTSAddress{}
	case AddressEd25519:
		seri = &Ed25519Address{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownAddrType, typeByte)
	}
	return seri, nil
}

// Defines a WOTS address.
type WOTSAddress [WOTSAddressBytesLength]byte

func (wotsAddr *WOTSAddress) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkType(data, AddressWOTS); err != nil {
			return 0, fmt.Errorf("unable to deserialize WOTS address: %w", err)
		}
		if err := checkMinByteLength(WOTSAddressSerializedBytesSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid WOTS address bytes: %w", err)
		}
		// TODO: check T5B1 encoding
	}
	copy(wotsAddr[:], data[OneByte:])
	return WOTSAddressSerializedBytesSize, nil
}

func (wotsAddr *WOTSAddress) Serialize(deSeriMode DeSerializationMode) (data []byte, err error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: check T5B1 encoding
	}
	var b [OneByte + WOTSAddressBytesLength]byte
	b[0] = AddressWOTS
	copy(b[OneByte:], wotsAddr[:])
	return b[:], nil
}

// Defines an Ed25519 address.
type Ed25519Address [Ed25519AddressBytesLength]byte

func (edAddr *Ed25519Address) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkType(data, AddressEd25519); err != nil {
			return 0, fmt.Errorf("unable to deserialize Ed25519 address: %w", err)
		}
		if err := checkMinByteLength(Ed25519AddressSerializedBytesSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid Ed25519 address bytes: %w", err)
		}
	}
	copy(edAddr[:], data[OneByte:])
	return Ed25519AddressSerializedBytesSize, nil
}

func (edAddr *Ed25519Address) Serialize(deSeriMode DeSerializationMode) (data []byte, err error) {
	var b [OneByte + Ed25519AddressBytesLength]byte
	b[0] = AddressEd25519
	copy(b[OneByte:], edAddr[:])
	return b[:], nil
}
