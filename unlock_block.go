package iotapkg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// Defines a type of unlock block.
type UnlockBlockType = byte

const (
	// Denotes a signature unlock block.
	UnlockBlockSignature UnlockBlockType = iota
	// Denotes a reference unlock block.
	UnlockBlockReference
)

var (
	ErrSigUnlockBlocksNotUnique = errors.New("signature unlock blocks must be unique")
	// TODO: might also reference something else in the future than just signature unlock blocks
	ErrRefUnlockBlockInvalidRef = errors.New("reference unlock block must point to a previous signature unlock block")
)

// UnlockBlockSelector implements SerializableSelectorFunc for unlock block types.
func UnlockBlockSelector(typeByte byte) (Serializable, error) {
	var seri Serializable
	switch typeByte {
	case UnlockBlockSignature:
		seri = &SignatureUnlockBlock{}
	case UnlockBlockReference:
		seri = &ReferenceUnlockBlock{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownUnlockBlockType, typeByte)
	}
	return seri, nil
}

// SignatureUnlockBlock holds a signature which unlocks inputs.
type SignatureUnlockBlock struct {
	Signature Serializable `json:"signature"`
}

func (s *SignatureUnlockBlock) Deserialize(data []byte) (int, error) {
	if err := checkType(data, UnlockBlockSignature); err != nil {
		return 0, fmt.Errorf("unable to deserialize signature unlock block: %w", err)
	}

	// skip type byte
	bytesReadTotal := OneByte
	data = data[OneByte:]

	sig, sigBytesRead, err := DeserializeObject(data, SignatureSelector)
	if err != nil {
		return 0, err
	}
	bytesReadTotal += sigBytesRead
	s.Signature = sig

	return bytesReadTotal, nil
}

func (s *SignatureUnlockBlock) Serialize() ([]byte, error) {
	sigBytes, err := s.Signature.Serialize()
	if err != nil {
		return nil, err
	}
	return append([]byte{UnlockBlockSignature}, sigBytes...), nil
}

// ReferenceUnlockBlock is an unlock block which references a previous unlock block.
type ReferenceUnlockBlock struct {
	Reference uint64 `json:"reference"`
}

func (r *ReferenceUnlockBlock) Deserialize(data []byte) (int, error) {
	if err := checkType(data, UnlockBlockReference); err != nil {
		return 0, fmt.Errorf("unable to deserialize reference unlock block: %w", err)
	}
	data = data[OneByte:]
	reference, referenceByteSize, err := ReadUvarint(bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	r.Reference = reference
	return OneByte + referenceByteSize, nil
}

func (r *ReferenceUnlockBlock) Serialize() ([]byte, error) {
	varIntBuf := make([]byte, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(varIntBuf, r.Reference)
	return append([]byte{UnlockBlockReference}, varIntBuf[:bytesWritten]...), nil
}

// UnlockBlockValidatorFunc which given the index of an unlock block and the unlock block itself, runs validations and returns an error if any should fail.
type UnlockBlockValidatorFunc func(index int, unlockBlock Serializable) error

// UnlockBlocksSigUniqueAndRefValidator returns a validator which checks that:
//	1. signature unlock blocks are unique
//	2. reference unlock blocks reference a previous signature unlock block
func UnlockBlocksSigUniqueAndRefValidator() UnlockBlockValidatorFunc {
	seenEdPubKeys := map[string]int{}
	seenSigBlocks := map[int]struct{}{}
	return func(index int, unlockBlock Serializable) error {
		switch x := unlockBlock.(type) {
		case *SignatureUnlockBlock:
			switch y := x.Signature.(type) {
			case *WOTSSignature:
				// TODO: implement
			case *Ed25519Signature:
				k := string(y.PublicKey[:])
				j, has := seenEdPubKeys[k]
				if has {
					return fmt.Errorf("%w: unlock block %d has the same Ed25519 public key as %d", ErrSigUnlockBlocksNotUnique, index, j)
				}
				seenEdPubKeys[k] = index
				seenSigBlocks[index] = struct{}{}
			}
		case *ReferenceUnlockBlock:
			reference := int(x.Reference)
			if _, has := seenSigBlocks[reference]; !has {
				return fmt.Errorf("%w: %d references non existent unlock block %d", ErrRefUnlockBlockInvalidRef, index, reference)
			}
		default:
			return fmt.Errorf("%w: %T", ErrUnknownUnlockBlockType, x)
		}

		return nil
	}
}

// ValidateUnlockBlocks validates the unlock blocks by running them against the given UnlockBlockValidatorFunc.
func ValidateUnlockBlocks(unlockBlocks Serializables, funcs []UnlockBlockValidatorFunc) error {
	for i, unlockBlock := range unlockBlocks {
		switch unlockBlock.(type) {
		case *SignatureUnlockBlock:
		case *ReferenceUnlockBlock:
		default:
			return fmt.Errorf("%w: can only validate signature or reference unlock blocks", ErrUnknownInputType)
		}
		for _, f := range funcs {
			if err := f(i, unlockBlock); err != nil {
				return err
			}
		}
	}
	return nil
}
