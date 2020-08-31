package iota_test

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/luca-moser/iota"
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

func randTxHash() [iota.TransactionIDLength]byte {
	var h [iota.TransactionIDLength]byte
	b := randBytes(32)
	copy(h[:], b)
	return h
}

func randWOTSAddr() (*iota.WOTSAddress, []byte) {
	wotsAddr := &iota.WOTSAddress{}
	addr := randBytes(iota.WOTSAddressBytesLength)
	copy(wotsAddr[:], addr)

	var varintBuf [binary.MaxVarintLen64]byte
	bytesWritten := binary.PutUvarint(varintBuf[:], iota.AddressWOTS)
	b := make([]byte, bytesWritten+iota.WOTSAddressBytesLength)
	copy(b[:bytesWritten], varintBuf[:bytesWritten])
	copy(b[bytesWritten:], wotsAddr[:])

	return wotsAddr, b[:]
}

func randEd25519Addr() (*iota.Ed25519Address, []byte) {
	edAddr := &iota.Ed25519Address{}
	addr := randBytes(iota.Ed25519AddressBytesLength)
	copy(edAddr[:], addr)

	var varintBuf [binary.MaxVarintLen64]byte
	bytesWritten := binary.PutUvarint(varintBuf[:], iota.AddressEd25519)
	b := make([]byte, bytesWritten+iota.Ed25519AddressBytesLength)
	copy(b[:bytesWritten], varintBuf[:bytesWritten])
	copy(b[bytesWritten:], edAddr[:])

	return edAddr, b[:]
}

func randEd25519Signature() (*iota.Ed25519Signature, []byte) {
	edSig := &iota.Ed25519Signature{}
	pub := randBytes(ed25519.PublicKeySize)
	sig := randBytes(ed25519.SignatureSize)
	copy(edSig.PublicKey[:], pub)
	copy(edSig.Signature[:], sig)

	buf, _, _ := iota.WriteTypeHeader(iota.SignatureEd25519)
	if _, err := buf.Write(edSig.PublicKey[:]); err != nil {
		panic(err)
	}
	if _, err := buf.Write(edSig.Signature[:]); err != nil {
		panic(err)
	}

	return edSig, buf.Bytes()
}

func randLSTransactionUnspentOutputs(outputsCount int) *iota.LSTransactionUnspentOutputs {
	return &iota.LSTransactionUnspentOutputs{
		TransactionHash: randTxHash(),
		UnspentOutputs: func() []*iota.LSUnspentOutput {
			outputs := make([]*iota.LSUnspentOutput, outputsCount)
			for i := 0; i < outputsCount; i++ {
				addr, _ := randEd25519Addr()
				outputs[i] = &iota.LSUnspentOutput{
					Index:   uint16(i),
					Address: addr,
					Value:   uint64(rand.Intn(1000000) + 1),
				}
			}
			return outputs
		}(),
	}
}

func randEd25519SignatureUnlockBlock() (*iota.SignatureUnlockBlock, []byte) {
	edSig, edSigData := randEd25519Signature()
	block := &iota.SignatureUnlockBlock{Signature: edSig}

	buf, _, _ := iota.WriteTypeHeader(iota.UnlockBlockSignature)
	if _, err := buf.Write(edSigData); err != nil {
		panic(err)
	}

	return block, buf.Bytes()
}

func randReferenceUnlockBlock() (*iota.ReferenceUnlockBlock, []byte) {
	return referenceUnlockBlock(uint64(rand.Intn(1000)))
}

func referenceUnlockBlock(index uint64) (*iota.ReferenceUnlockBlock, []byte) {
	block := &iota.ReferenceUnlockBlock{Reference: index}

	buf, varintBuf, _ := iota.WriteTypeHeader(iota.UnlockBlockReference)
	bytesWritten := binary.PutUvarint(varintBuf[:], block.Reference)
	if _, err := buf.Write(varintBuf[:bytesWritten]); err != nil {
		panic(err)
	}

	return block, buf.Bytes()
}

func randUnsignedTransaction() (*iota.UnsignedTransaction, []byte) {
	tx := &iota.UnsignedTransaction{}
	buf, varintBuf, _ := iota.WriteTypeHeader(iota.TransactionUnsigned)

	inputsBytes := iota.LexicalOrderedByteSlices{}
	inputCount := rand.Intn(10) + 1
	bytesWritten := binary.PutUvarint(varintBuf[:], uint64(inputCount))
	if _, err := buf.Write(varintBuf[:bytesWritten]); err != nil {
		panic(err)
	}
	for i := inputCount; i > 0; i-- {
		_, inputData := randUTXOInput()
		inputsBytes = append(inputsBytes, inputData)
	}

	sort.Sort(inputsBytes)
	for _, inputData := range inputsBytes {
		if _, err := buf.Write(inputData); err != nil {
			panic(err)
		}
		input := &iota.UTXOInput{}
		if _, err := input.Deserialize(inputData, iota.DeSeriModePerformValidation); err != nil {
			panic(err)
		}
		tx.Inputs = append(tx.Inputs, input)
	}

	outputsBytes := iota.LexicalOrderedByteSlices{}
	outputCount := rand.Intn(10) + 1
	bytesWritten = binary.PutUvarint(varintBuf[:], uint64(outputCount))
	if _, err := buf.Write(varintBuf[:bytesWritten]); err != nil {
		panic(err)
	}
	for i := outputCount; i > 0; i-- {
		_, depData := randSigLockedSingleDeposit(iota.AddressEd25519)
		outputsBytes = append(outputsBytes, depData)
	}

	sort.Sort(outputsBytes)
	for _, outputData := range outputsBytes {
		_, err := buf.Write(outputData)
		must(err)
		output := &iota.SigLockedSingleDeposit{}
		if _, err := output.Deserialize(outputData, iota.DeSeriModePerformValidation); err != nil {
			panic(err)
		}
		tx.Outputs = append(tx.Outputs, output)
	}

	// empty payload length
	bytesWritten = binary.PutUvarint(varintBuf[:], 0)
	if _, err := buf.Write(varintBuf[:bytesWritten]); err != nil {
		panic(err)
	}

	return tx, buf.Bytes()
}

func randMilestonePayload() (*iota.MilestonePayload, []byte) {
	msPayload := &iota.MilestonePayload{
		Index:     uint64(rand.Intn(1000)),
		Timestamp: uint64(time.Now().Unix()),
		InclusionMerkleProof: func() [iota.MilestoneInclusionMerkleProofLength]byte {
			b := [iota.MilestoneInclusionMerkleProofLength]byte{}
			copy(b[:], randBytes(iota.MilestoneInclusionMerkleProofLength))
			return b
		}(),
		Signature: func() [iota.MilestoneSignatureLength]byte {
			b := [iota.MilestoneSignatureLength]byte{}
			copy(b[:], randBytes(iota.MilestoneSignatureLength))
			return b
		}(),
	}

	buf, varintBuf, _ := iota.WriteTypeHeader(iota.MilestonePayloadID)
	bytesWritten := binary.PutUvarint(varintBuf[:], msPayload.Index)
	if _, err := buf.Write(varintBuf[:bytesWritten]); err != nil {
		panic(err)
	}

	if err := binary.Write(buf, binary.LittleEndian, msPayload.Timestamp); err != nil {
		panic(err)
	}

	if _, err := buf.Write(msPayload.InclusionMerkleProof[:]); err != nil {
		panic(err)
	}

	if _, err := buf.Write(msPayload.Signature[:]); err != nil {
		panic(err)
	}

	return msPayload, buf.Bytes()
}

func randUnsignedDataPayload(dataLength ...int) (*iota.UnsignedDataPayload, []byte) {
	var data []byte
	switch {
	case len(dataLength) > 0:
		data = randBytes(dataLength[0])
	default:
		data = randBytes(rand.Intn(200) + 1)
	}
	unsigDataPayload := &iota.UnsignedDataPayload{Data: data}

	buf, varintBuf, _ := iota.WriteTypeHeader(iota.UnsignedDataPayloadID)
	bytesWritten := binary.PutUvarint(varintBuf[:], uint64(len(unsigDataPayload.Data)))
	if _, err := buf.Write(varintBuf[:bytesWritten]); err != nil {
		panic(err)
	}

	if _, err := buf.Write(unsigDataPayload.Data); err != nil {
		panic(err)
	}

	return unsigDataPayload, buf.Bytes()
}

func randMessage(withPayloadType int) (*iota.Message, []byte) {
	var payload iota.Serializable
	var payloadData []byte

	switch withPayloadType {
	case iota.SignedTransactionPayloadID:
		payload, payloadData = randSignedTransactionPayload()
	case iota.UnsignedDataPayloadID:
		payload, payloadData = randUnsignedDataPayload()
	}

	m := &iota.Message{}
	copy(m.Parent1[:], randBytes(iota.MessageHashLength))
	copy(m.Parent2[:], randBytes(iota.MessageHashLength))
	m.Payload = payload
	m.Nonce = uint64(rand.Intn(1000))

	buf, varintBuf, _ := iota.WriteTypeHeader(iota.MilestonePayloadID)
	if _, err := buf.Write(m.Parent1[:]); err != nil {
		panic(err)
	}
	if _, err := buf.Write(m.Parent2[:]); err != nil {
		panic(err)
	}

	switch {
	case payload == nil:
		// zero length payload
		must(buf.WriteByte(0))
	default:
		bytesWritten := binary.PutUvarint(varintBuf[:], uint64(len(payloadData)))
		if _, err := buf.Write(varintBuf[:bytesWritten]); err != nil {
			panic(err)
		}

		// actual payload
		if _, err := buf.Write(payloadData); err != nil {
			panic(err)
		}
	}

	must(binary.Write(buf, binary.LittleEndian, m.Nonce))

	return m, buf.Bytes()
}

func randSignedTransactionPayload() (*iota.SignedTransactionPayload, []byte) {
	sigTxPayload := &iota.SignedTransactionPayload{}
	buf, varintBuf, _ := iota.WriteTypeHeader(iota.SignedTransactionPayloadID)

	unTx, unTxData := randUnsignedTransaction()
	_, err := buf.Write(unTxData)
	must(err)
	sigTxPayload.Transaction = unTx

	unlockBlocksCount := len(unTx.Inputs)
	bytesWritten := binary.PutUvarint(varintBuf[:], uint64(unlockBlocksCount))
	if _, err := buf.Write(varintBuf[:bytesWritten]); err != nil {
		panic(err)
	}
	for i := 0; i < unlockBlocksCount; i++ {
		unlockBlock, unlockBlockData := randEd25519SignatureUnlockBlock()
		_, err := buf.Write(unlockBlockData)
		must(err)
		sigTxPayload.UnlockBlocks = append(sigTxPayload.UnlockBlocks, unlockBlock)
	}

	return sigTxPayload, buf.Bytes()
}

func randUTXOInput() (*iota.UTXOInput, []byte) {
	utxoInput := &iota.UTXOInput{}
	buf, varintBuf, _ := iota.WriteTypeHeader(iota.InputUTXO)

	txID := randBytes(iota.TransactionIDLength)
	_, err := buf.Write(txID)
	must(err)
	copy(utxoInput.TransactionID[:], txID)

	index := uint64(rand.Intn(iota.RefUTXOIndexMax))
	bytesWritten := binary.PutUvarint(varintBuf[:], index)
	if _, err := buf.Write(varintBuf[:bytesWritten]); err != nil {
		panic(err)
	}
	utxoInput.TransactionOutputIndex = index
	return utxoInput, buf.Bytes()
}

func randSigLockedSingleDeposit(addrType iota.AddressType) (*iota.SigLockedSingleDeposit, []byte) {
	dep := &iota.SigLockedSingleDeposit{}
	buf, _, _ := iota.WriteTypeHeader(iota.OutputSigLockedSingleDeposit)

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

	dep.Amount = uint64(rand.Intn(10000))
	must(binary.Write(buf, binary.LittleEndian, dep.Amount))

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
