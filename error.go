package iota

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidBytes                  = errors.New("invalid bytes")
	ErrDeserializationTypeMismatch   = errors.New("data type is invalid for deserialization")
	ErrUnknownPayloadType            = errors.New("unknown payload type")
	ErrUnknownAddrType               = errors.New("unknown address type")
	ErrUnknownInputType              = errors.New("unknown input type")
	ErrUnknownOutputType             = errors.New("unknown output type")
	ErrUnknownTransactionType        = errors.New("unknown transaction type")
	ErrUnknownUnlockBlockType        = errors.New("unknown unlock block type")
	ErrUnknownSignatureType          = errors.New("unknown signature type")
	ErrInvalidVarint                 = errors.New("invalid varint")
	ErrDeserializationNotEnoughData  = errors.New("not enough data for deserialization")
	ErrDeserializationNotAllConsumed = errors.New("not all data has been consumed but should have been")
)

// checks whether data is for the given type and returns the amount of bytes read to perform this operation.
func checkType(data []byte, shouldType uint64) (int, error) {
	if data == nil || len(data) == 0 {
		return 0, fmt.Errorf("%w: can not evaluate type", ErrDeserializationNotEnoughData)
	}
	typeData, bytesRead, err := Uvarint(data)
	if err != nil {
		return 0, fmt.Errorf("%w: can not evaluate type", err)
	}
	if typeData != shouldType {
		return 0, fmt.Errorf("%w: type denotation must be %d but is %d", ErrDeserializationTypeMismatch, shouldType, typeData)
	}
	return bytesRead, nil
}

func checkExactByteLength(exact int, length int) error {
	if length != exact {
		return fmt.Errorf("%w: data must be at exact %d bytes long but is %d", ErrInvalidBytes, exact, length)
	}
	return nil
}

func checkByteLengthRange(min int, max int, length int) error {
	if err := checkMinByteLength(min, length); err != nil {
		return err
	}
	if err := checkMaxByteLength(max, length); err != nil {
		return err
	}
	return nil
}

func checkMinByteLength(min int, length int) error {
	if length < min {
		return fmt.Errorf("%w: data must be at least %d bytes long but is %d", ErrInvalidBytes, min, length)
	}
	return nil
}

func checkMaxByteLength(max int, length int) error {
	if length > max {
		return fmt.Errorf("%w: data must be max %d bytes long but is %d", ErrInvalidBytes, max, length)
	}
	return nil
}
