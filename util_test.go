package iota_test

import (
	"bytes"
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
	// type
	wotsAddr := &iota.WOTSAddress{}
	addr := randBytes(iota.WOTSAddressBytesLength)
	copy(wotsAddr[:], addr)
	// serialized
	var b [iota.WOTSAddressSerializedBytesSize]byte
	b[0] = iota.AddressWOTS
	copy(b[iota.SmallTypeDenotationByteSize:], addr)
	return wotsAddr, b[:]
}

func randEd25519Addr() (*iota.Ed25519Address, []byte) {
	// type
	edAddr := &iota.Ed25519Address{}
	addr := randBytes(iota.Ed25519AddressBytesLength)
	copy(edAddr[:], addr)
	// serialized
	var b [iota.Ed25519AddressSerializedBytesSize]byte
	b[0] = iota.AddressEd25519
	copy(b[iota.SmallTypeDenotationByteSize:], addr)
	return edAddr, b[:]
}

func randEd25519Signature() (*iota.Ed25519Signature, []byte) {
	// type
	edSig := &iota.Ed25519Signature{}
	pub := randBytes(ed25519.PublicKeySize)
	sig := randBytes(ed25519.SignatureSize)
	copy(edSig.PublicKey[:], pub)
	copy(edSig.Signature[:], sig)
	// serialized
	var b [iota.Ed25519SignatureSerializedBytesSize]byte
	binary.LittleEndian.PutUint32(b[:iota.TypeDenotationByteSize], iota.SignatureEd25519)
	copy(b[iota.TypeDenotationByteSize:], pub)
	copy(b[iota.TypeDenotationByteSize+ed25519.PublicKeySize:], sig)
	return edSig, b[:]
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
	return block, append([]byte{iota.UnlockBlockSignature}, edSigData...)
}

func randReferenceUnlockBlock() (*iota.ReferenceUnlockBlock, []byte) {
	return referenceUnlockBlock(uint16(rand.Intn(1000)))
}

func referenceUnlockBlock(index uint16) (*iota.ReferenceUnlockBlock, []byte) {
	var b [iota.ReferenceUnlockBlockSize]byte
	b[0] = iota.UnlockBlockReference
	binary.LittleEndian.PutUint16(b[iota.SmallTypeDenotationByteSize:], index)
	return &iota.ReferenceUnlockBlock{Reference: index}, b[:]
}

func randUnsignedTransaction() (*iota.UnsignedTransaction, []byte) {
	var buf bytes.Buffer

	tx := &iota.UnsignedTransaction{}
	must(binary.Write(&buf, binary.LittleEndian, iota.TransactionUnsigned))

	inputsBytes := iota.LexicalOrderedByteSlices{}
	inputCount := rand.Intn(10) + 1
	must(binary.Write(&buf, binary.LittleEndian, uint16(inputCount)))
	for i := inputCount; i > 0; i-- {
		_, inputData := randUTXOInput()
		inputsBytes = append(inputsBytes, inputData)
	}

	sort.Sort(inputsBytes)
	for _, inputData := range inputsBytes {
		_, err := buf.Write(inputData)
		must(err)
		input := &iota.UTXOInput{}
		if _, err := input.Deserialize(inputData, iota.DeSeriModePerformValidation); err != nil {
			panic(err)
		}
		tx.Inputs = append(tx.Inputs, input)
	}

	outputsBytes := iota.LexicalOrderedByteSlices{}
	outputCount := rand.Intn(10) + 1
	must(binary.Write(&buf, binary.LittleEndian, uint16(outputCount)))
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

	// empty payload
	must(binary.Write(&buf, binary.LittleEndian, uint16(0)))

	return tx, buf.Bytes()
}

func randMilestonePayload() (*iota.MilestonePayload, []byte) {
	inclusionMerkleProof := randBytes(iota.MilestoneInclusionMerkleProofLength)
	signature := randBytes(iota.MilestoneSignatureLength)
	msPayload := &iota.MilestonePayload{
		Index:     uint64(rand.Intn(1000)),
		Timestamp: uint64(time.Now().Unix()),
		InclusionMerkleProof: func() [iota.MilestoneInclusionMerkleProofLength]byte {
			b := [iota.MilestoneInclusionMerkleProofLength]byte{}
			copy(b[:], inclusionMerkleProof)
			return b
		}(),
		Signature: func() [iota.MilestoneSignatureLength]byte {
			b := [iota.MilestoneSignatureLength]byte{}
			copy(b[:], signature)
			return b
		}(),
	}

	var b bytes.Buffer
	must(binary.Write(&b, binary.LittleEndian, iota.MilestonePayloadID))
	must(binary.Write(&b, binary.LittleEndian, msPayload.Index))
	must(binary.Write(&b, binary.LittleEndian, msPayload.Timestamp))

	if _, err := b.Write(msPayload.InclusionMerkleProof[:]); err != nil {
		panic(err)
	}

	if _, err := b.Write(msPayload.Signature[:]); err != nil {
		panic(err)
	}

	return msPayload, b.Bytes()
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

	var b bytes.Buffer
	must(binary.Write(&b, binary.LittleEndian, iota.UnsignedDataPayloadID))
	must(binary.Write(&b, binary.LittleEndian, uint16(len(unsigDataPayload.Data))))
	if _, err := b.Write(unsigDataPayload.Data); err != nil {
		panic(err)
	}

	return unsigDataPayload, b.Bytes()
}

func randMessage(withPayloadType uint32) (*iota.Message, []byte) {
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

	var b bytes.Buffer
	if err := b.WriteByte(iota.MessageVersion); err != nil {
		panic(err)
	}
	if _, err := b.Write(m.Parent1[:]); err != nil {
		panic(err)
	}
	if _, err := b.Write(m.Parent2[:]); err != nil {
		panic(err)
	}

	switch {
	case payload == nil:
		// zero length payload
		if err := binary.Write(&b, binary.LittleEndian, uint32(0)); err != nil {
			panic(err)
		}
	default:
		if err := binary.Write(&b, binary.LittleEndian, uint32(len(payloadData))); err != nil {
			panic(err)
		}
		if _, err := b.Write(payloadData); err != nil {
			panic(err)
		}
	}

	if err := binary.Write(&b, binary.LittleEndian, m.Nonce); err != nil {
		panic(err)
	}
	return m, b.Bytes()
}

func randSignedTransactionPayload() (*iota.SignedTransactionPayload, []byte) {
	var buf bytes.Buffer
	must(binary.Write(&buf, binary.LittleEndian, iota.SignedTransactionPayloadID))

	sigTxPayload := &iota.SignedTransactionPayload{}
	unTx, unTxData := randUnsignedTransaction()
	_, err := buf.Write(unTxData)
	must(err)
	sigTxPayload.Transaction = unTx

	unlockBlocksCount := len(unTx.Inputs)
	must(binary.Write(&buf, binary.LittleEndian, uint16(unlockBlocksCount)))
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
	var b [iota.UTXOInputSize]byte
	b[0] = iota.InputUTXO

	txID := randBytes(iota.TransactionIDLength)
	copy(b[iota.SmallTypeDenotationByteSize:], txID)
	copy(utxoInput.TransactionID[:], txID)

	index := uint16(rand.Intn(iota.RefUTXOIndexMax))
	binary.LittleEndian.PutUint16(b[len(b)-iota.UInt16ByteSize:], index)
	utxoInput.TransactionOutputIndex = index
	return utxoInput, b[:]
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
