package iotapkg_test

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"math/rand"

	"github.com/luca-moser/iotapkg"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// returns length amount random bytes
func randBytes(length int) []byte {
	var b []byte
	for i := 0; i < length; i++ {
		b = append(b, byte(rand.Intn(256)))
	}
	return b
}

func randWOTSAddr() (*iotapkg.WOTSAddress, []byte) {
	wotsAddr := &iotapkg.WOTSAddress{}
	addr := randBytes(iotapkg.WOTSAddressBytesLength)
	copy(wotsAddr[:], addr)
	return wotsAddr, append([]byte{iotapkg.AddressWOTS}, addr...)
}

func randEd25519Addr() (*iotapkg.Ed25519Address, []byte) {
	edAddr := &iotapkg.Ed25519Address{}
	addr := randBytes(iotapkg.Ed25519AddressBytesLength)
	copy(edAddr[:], addr)
	return edAddr, append([]byte{iotapkg.AddressEd25519}, addr...)
}

func randEd25519Signature() (*iotapkg.Ed25519Signature, []byte) {
	edSig := &iotapkg.Ed25519Signature{}
	pub := randBytes(ed25519.PublicKeySize)
	sig := randBytes(ed25519.SignatureSize)
	copy(edSig.PublicKey[:], pub)
	copy(edSig.Signature[:], sig)
	b := append([]byte{iotapkg.SignatureEd25519}, pub...)
	return edSig, append(b, sig...)
}

func randEd25519SignatureUnlockBlock() (*iotapkg.SignatureUnlockBlock, []byte) {
	edSig, edSigData := randEd25519Signature()
	block := &iotapkg.SignatureUnlockBlock{Signature: edSig}
	return block, append([]byte{iotapkg.UnlockBlockSignature}, edSigData...)
}

func randReferenceUnlockBlock() (*iotapkg.ReferenceUnlockBlock, []byte) {
	var buf bytes.Buffer
	index := rand.Intn(1000)
	varIntBuf := make([]byte, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(varIntBuf, uint64(index))
	if _, err := buf.Write(varIntBuf[:bytesWritten]); err != nil {
		panic(err)
	}
	block := &iotapkg.ReferenceUnlockBlock{Reference: uint64(index)}
	return block, append([]byte{iotapkg.UnlockBlockReference}, varIntBuf[:bytesWritten]...)
}

func referenceUnlockBlock(index uint64) (*iotapkg.ReferenceUnlockBlock, []byte) {
	var buf bytes.Buffer
	varIntBuf := make([]byte, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(varIntBuf, index)
	if _, err := buf.Write(varIntBuf[:bytesWritten]); err != nil {
		panic(err)
	}
	block := &iotapkg.ReferenceUnlockBlock{Reference: uint64(index)}
	return block, append([]byte{iotapkg.UnlockBlockReference}, varIntBuf[:bytesWritten]...)
}

func randUnsignedTransaction() (*iotapkg.UnsignedTransaction, []byte) {
	var buf bytes.Buffer
	tx := &iotapkg.UnsignedTransaction{}
	must(buf.WriteByte(iotapkg.TransactionUnsigned))

	inputCount := rand.Intn(10) + 1
	must(buf.WriteByte(byte(inputCount)))
	for i := inputCount; i > 0; i-- {
		input, inputData := randUTXOInput()
		_, err := buf.Write(inputData)
		must(err)
		tx.Inputs = append(tx.Inputs, input)
	}

	outputCount := rand.Intn(10) + 1
	must(buf.WriteByte(byte(outputCount)))
	for i := outputCount; i > 0; i-- {
		dep, depData := randSigLockedSingleDeposit(iotapkg.AddressEd25519)
		_, err := buf.Write(depData)
		must(err)
		tx.Outputs = append(tx.Outputs, dep)
	}

	// empty payload
	must(buf.WriteByte(0))

	return tx, buf.Bytes()
}

func randSignedTransactionPayload() (*iotapkg.SignedTransactionPayload, []byte) {
	var buf bytes.Buffer
	sigTxPayload := &iotapkg.SignedTransactionPayload{}
	must(buf.WriteByte(iotapkg.SignedTransactionPayloadID))

	unTx, unTxData := randUnsignedTransaction()
	_, err := buf.Write(unTxData)
	must(err)
	sigTxPayload.Transaction = unTx

	unlockBlocksCount := rand.Intn(10) + 1
	must(buf.WriteByte(byte(unlockBlocksCount)))
	for i := unlockBlocksCount; i > 0; i-- {
		var unlockBlock iotapkg.Serializable
		var unlockBlockData []byte
		switch rand.Intn(2) {
		case 0:
			unlockBlock, unlockBlockData = randEd25519SignatureUnlockBlock()
		case 1:
			unlockBlock, unlockBlockData = randReferenceUnlockBlock()
		default:
			panic("not all rands covered")
		}
		_, err := buf.Write(unlockBlockData)
		must(err)
		sigTxPayload.UnlockBlocks = append(sigTxPayload.UnlockBlocks, unlockBlock)
	}

	return sigTxPayload, buf.Bytes()
}

func randUTXOInput() (*iotapkg.UTXOInput, []byte) {
	utxoInput := &iotapkg.UTXOInput{}
	var buf bytes.Buffer
	must(buf.WriteByte(iotapkg.InputUTXO))

	txID := randBytes(iotapkg.TransactionIDLength)
	_, err := buf.Write(txID)
	must(err)
	copy(utxoInput.TransactionID[:], txID)

	index := rand.Intn(1000)
	varIntBuf := make([]byte, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(varIntBuf, uint64(index))
	if _, err := buf.Write(varIntBuf[:bytesWritten]); err != nil {
		panic(err)
	}
	utxoInput.TransactionOutputIndex = uint16(index)
	return utxoInput, buf.Bytes()
}

func randSigLockedSingleDeposit(addrType iotapkg.AddressType) (*iotapkg.SigLockedSingleDeposit, []byte) {
	var buf bytes.Buffer
	must(buf.WriteByte(iotapkg.OutputSigLockedSingleDeposit))

	dep := &iotapkg.SigLockedSingleDeposit{}

	var addrData []byte
	switch addrType {
	case iotapkg.AddressWOTS:
		dep.Address, addrData = randWOTSAddr()
	case iotapkg.AddressEd25519:
		dep.Address, addrData = randEd25519Addr()
	default:
		panic(fmt.Sprintf("invalid addr type: %d", addrType))
	}

	_, err := buf.Write(addrData)
	must(err)

	amount := uint64(rand.Intn(10000))
	must(binary.Write(&buf, binary.LittleEndian, amount))
	dep.Amount = amount

	return dep, buf.Bytes()
}
