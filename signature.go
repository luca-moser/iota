package iota

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
)

// Defines the type of signature.
type SignatureType = uint64

const (
	// Denotes a WOTS a signature.
	SignatureWOTS SignatureType = iota
	// Denotes an Ed25519 signature.
	SignatureEd25519

	Ed25519SignatureAndPubKeyByteSize = ed25519.PublicKeySize + ed25519.SignatureSize
)

// SignatureSelector implements SerializableSelectorFunc for signature types.
func SignatureSelector(sigType uint64) (Serializable, error) {
	var seri Serializable
	switch sigType {
	case SignatureWOTS:
		seri = &WOTSSignature{}
	case SignatureEd25519:
		seri = &Ed25519Signature{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownSignatureType, sigType)
	}
	return seri, nil
}

type WOTSSignature struct{}

func (w *WOTSSignature) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	_, _, err := ReadTypeAndAdvance(data, SignatureWOTS, deSeriMode)
	if err != nil {
		return 0, err
	}
	panic("implement me")
}

func (w *WOTSSignature) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	panic("implement me")
}

type Ed25519Signature struct {
	PublicKey [ed25519.PublicKeySize]byte `json:"public_key"`
	Signature [ed25519.SignatureSize]byte `json:"signature"`
}

func (e *Ed25519Signature) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	data, typeBytesRead, err := ReadTypeAndAdvance(data, SignatureEd25519, deSeriMode)
	if err != nil {
		return 0, err
	}

	if len(data) < Ed25519SignatureAndPubKeyByteSize {
		return 0, fmt.Errorf("%w: can not deserialize Ed25519 public key and signature", ErrDeserializationNotEnoughData)
	}

	copy(e.PublicKey[:], data[:ed25519.PublicKeySize])
	copy(e.Signature[:], data[ed25519.PublicKeySize:])
	return typeBytesRead + Ed25519SignatureAndPubKeyByteSize, nil
}

func (e *Ed25519Signature) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	var varintBuf [binary.MaxVarintLen64]byte
	bytesWritten := binary.PutUvarint(varintBuf[:], SignatureEd25519)
	var b bytes.Buffer
	if _, err := b.Write(varintBuf[:bytesWritten]); err != nil {
		return nil, err
	}
	if _, err := b.Write(e.PublicKey[:]); err != nil {
		return nil, err
	}
	if _, err := b.Write(e.Signature[:]); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
