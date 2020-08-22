package iota

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidBytes                = errors.New("invalid bytes")
	ErrDeserializationTypeMismatch = errors.New("data type is invalid for deserialization")
	ErrUnknownAddrType             = errors.New("unknown address type")
	ErrUnknownInputType            = errors.New("unknown input type")
	ErrUnknownOutputType           = errors.New("unknown output type")
	ErrUnknownTransactionType      = errors.New("unknown transaction type")
	ErrUnknownUnlockBlockType      = errors.New("unknown unlock block type")
	ErrUnknownSignatureType        = errors.New("unknown signature type")
	ErrDeserializationDataTooSmall = errors.New("not enough data for deserialization")
)

func checkType(data []byte, should byte) error {
	if data == nil || len(data) == 0 {
		return fmt.Errorf("%w: can not evaluate type", ErrDeserializationDataTooSmall)
	}
	if data[0] != should {
		return fmt.Errorf("%w: type denotation must be %d but is %d", ErrDeserializationTypeMismatch, should, data[0])
	}
	return nil
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
