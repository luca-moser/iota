package iota_test

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sort"

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

func randWOTSAddr() (*iota.WOTSAddress, []byte) {
	wotsAddr := &iota.WOTSAddress{}
	addr := randBytes(iota.WOTSAddressBytesLength)
	copy(wotsAddr[:], addr)
	return wotsAddr, append([]byte{iota.AddressWOTS}, addr...)
}

func randEd25519Addr() (*iota.Ed25519Address, []byte) {
	edAddr := &iota.Ed25519Address{}
	addr := randBytes(iota.Ed25519AddressBytesLength)
	copy(edAddr[:], addr)
	return edAddr, append([]byte{iota.AddressEd25519}, addr...)
}

func randEd25519Signature() (*iota.Ed25519Signature, []byte) {
	edSig := &iota.Ed25519Signature{}
	pub := randBytes(ed25519.PublicKeySize)
	sig := randBytes(ed25519.SignatureSize)
	copy(edSig.PublicKey[:], pub)
	copy(edSig.Signature[:], sig)
	b := append([]byte{iota.SignatureEd25519}, pub...)
	return edSig, append(b, sig...)
}

func randEd25519SignatureUnlockBlock() (*iota.SignatureUnlockBlock, []byte) {
	edSig, edSigData := randEd25519Signature()
	block := &iota.SignatureUnlockBlock{Signature: edSig}
	return block, append([]byte{iota.UnlockBlockSignature}, edSigData...)
}

func randReferenceUnlockBlock() (*iota.ReferenceUnlockBlock, []byte) {
	var buf bytes.Buffer
	index := rand.Intn(1000)
	varIntBuf := make([]byte, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(varIntBuf, uint64(index))
	if _, err := buf.Write(varIntBuf[:bytesWritten]); err != nil {
		panic(err)
	}
	block := &iota.ReferenceUnlockBlock{Reference: uint64(index)}
	return block, append([]byte{iota.UnlockBlockReference}, varIntBuf[:bytesWritten]...)
}

func referenceUnlockBlock(index uint64) (*iota.ReferenceUnlockBlock, []byte) {
	var buf bytes.Buffer
	varIntBuf := make([]byte, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(varIntBuf, index)
	if _, err := buf.Write(varIntBuf[:bytesWritten]); err != nil {
		panic(err)
	}
	block := &iota.ReferenceUnlockBlock{Reference: uint64(index)}
	return block, append([]byte{iota.UnlockBlockReference}, varIntBuf[:bytesWritten]...)
}

func randUnsignedTransaction() (*iota.UnsignedTransaction, []byte) {
	var buf bytes.Buffer
	tx := &iota.UnsignedTransaction{}
	must(buf.WriteByte(iota.TransactionUnsigned))

	inputsBytes := iota.LexicalOrderedByteSlices{}
	inputCount := rand.Intn(10) + 1
	must(buf.WriteByte(byte(inputCount)))
	for i := inputCount; i > 0; i-- {
		_, inputData := randUTXOInput()
		inputsBytes = append(inputsBytes, inputData)
	}

	sort.Sort(inputsBytes)
	for _, inputData := range inputsBytes {
		_, err := buf.Write(inputData)
		must(err)
		input := &iota.UTXOInput{}
		if _, err := input.Deserialize(inputData); err != nil {
			panic(err)
		}
		tx.Inputs = append(tx.Inputs, input)
	}

	outputsBytes := iota.LexicalOrderedByteSlices{}
	outputCount := rand.Intn(10) + 1
	must(buf.WriteByte(byte(outputCount)))
	for i := outputCount; i > 0; i-- {
		_, depData := randSigLockedSingleDeposit(iota.AddressEd25519)
		outputsBytes = append(outputsBytes, depData)
	}

	sort.Sort(outputsBytes)
	for _, outputData := range outputsBytes {
		_, err := buf.Write(outputData)
		must(err)
		output := &iota.SigLockedSingleDeposit{}
		if _, err := output.Deserialize(outputData); err != nil {
			panic(err)
		}
		tx.Outputs = append(tx.Outputs, output)
	}

	// empty payload
	must(buf.WriteByte(0))

	return tx, buf.Bytes()
}

func randSignedTransactionPayload() (*iota.SignedTransactionPayload, []byte) {
	var buf bytes.Buffer
	sigTxPayload := &iota.SignedTransactionPayload{}
	must(buf.WriteByte(iota.SignedTransactionPayloadID))

	unTx, unTxData := randUnsignedTransaction()
	_, err := buf.Write(unTxData)
	must(err)
	sigTxPayload.Transaction = unTx

	unlockBlocksCount := len(unTx.Inputs)
	must(buf.WriteByte(byte(unlockBlocksCount)))
	for i := unlockBlocksCount; i > 0; i-- {
		unlockBlock, unlockBlockData := randEd25519SignatureUnlockBlock()
		_, err := buf.Write(unlockBlockData)
		must(err)
		sigTxPayload.UnlockBlocks = append(sigTxPayload.UnlockBlocks, unlockBlock)
	}

	return sigTxPayload, buf.Bytes()
}

func randUTXOInput() (*iota.UTXOInput, []byte) {
	utxoInput := &iota.UTXOInput{}
	var buf bytes.Buffer
	must(buf.WriteByte(iota.InputUTXO))

	txID := randBytes(iota.TransactionIDLength)
	_, err := buf.Write(txID)
	must(err)
	copy(utxoInput.TransactionID[:], txID)

	index := rand.Intn(iota.RefUTXOIndexMax)
	varIntBuf := make([]byte, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(varIntBuf, uint64(index))
	if _, err := buf.Write(varIntBuf[:bytesWritten]); err != nil {
		panic(err)
	}
	utxoInput.TransactionOutputIndex = byte(index)
	return utxoInput, buf.Bytes()
}

func randSigLockedSingleDeposit(addrType iota.AddressType) (*iota.SigLockedSingleDeposit, []byte) {
	var buf bytes.Buffer
	must(buf.WriteByte(iota.OutputSigLockedSingleDeposit))

	dep := &iota.SigLockedSingleDeposit{}

	var addrData []byte
	switch addrType {
	case iota.AddressWOTS:
		dep.Address, addrData = randWOTSAddr()
	case iota.AddressEd25519:
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

func oneInputOutputSignedTransactionPayload() *iota.SignedTransactionPayload {
	return &iota.SignedTransactionPayload{
		Transaction: &iota.UnsignedTransaction{
			Inputs: []iota.Serializable{
				&iota.UTXOInput{
					TransactionID: func() [iota.TransactionIDLength]byte {
						var b [iota.TransactionIDLength]byte
						copy(b[:], randBytes(iota.TransactionIDLength))
						return b
					}(),
					TransactionOutputIndex: 0,
				},
			},
			Outputs: []iota.Serializable{
				&iota.SigLockedSingleDeposit{
					Address: func() iota.Serializable {
						edAddr, _ := randEd25519Addr()
						return edAddr
					}(),
					Amount: 1337,
				},
			},
			Payload: nil,
		},
		UnlockBlocks: []iota.Serializable{
			&iota.SignatureUnlockBlock{
				Signature: func() iota.Serializable {
					edSig, _ := randEd25519Signature()
					return edSig
				}(),
			},
		},
	}
}

func randEd25519Seed() [ed25519.SeedSize]byte {
	var b [ed25519.SeedSize]byte
	read, err := rand.Read(b[:])
	if read != ed25519.SeedSize {
		panic(fmt.Sprintf("could not read %d required bytes from secure RNG", ed25519.SeedSize))
	}
	if err != nil {
		panic(err)
	}
	return b
}
