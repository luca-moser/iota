package iotapkg

import (
	"bytes"
	"encoding/binary"
)

const TransactionIDLength = 32

const SignedTransactionPayloadID = 0

// SignedTransactionPayload is a transaction with its inputs, outputs and unlock blocks.
type SignedTransactionPayload struct {
	Transaction  Serializable   `json:"transaction"`
	UnlockBlocks []Serializable `json:"unlock_blocks"`
}

func (s *SignedTransactionPayload) Deserialize(data []byte) (int, error) {
	var bytesReadTotal int

	// skip payload type
	data = data[OneByte:]
	bytesReadTotal += OneByte

	tx, txBytesRead, err := DeserializeObject(data, TransactionSelector)
	if err != nil {
		return 0, err
	}
	bytesReadTotal += txBytesRead
	s.Transaction = tx

	// advance to unlock blocks
	data = data[txBytesRead:]
	unlockBlocks, unlockBlocksByteRead, err := DeserializeArrayOfObjects(data, UnlockBlockSelector)
	if err != nil {
		return 0, err
	}
	bytesReadTotal += unlockBlocksByteRead
	s.UnlockBlocks = unlockBlocks

	return bytesReadTotal, nil
}

func (s *SignedTransactionPayload) Serialize() ([]byte, error) {
	var b bytes.Buffer
	if err := b.WriteByte(SignedTransactionPayloadID); err != nil {
		return nil, err
	}

	// write transaction
	txBytes, err := s.Transaction.Serialize()
	if err != nil {
		return nil, err
	}
	if _, err := b.Write(txBytes); err != nil {
		return nil, err
	}

	// write unlock blocks and count
	varIntBuf := make([]byte, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(varIntBuf, uint64(len(s.UnlockBlocks)))
	if _, err := b.Write(varIntBuf[:bytesWritten]); err != nil {
		return nil, err
	}

	for i := range s.UnlockBlocks {
		unlockBlockSer, err := s.UnlockBlocks[i].Serialize()
		if err != nil {
			return nil, err
		}
		if _, err := b.Write(unlockBlockSer); err != nil {
			return nil, err
		}
	}

	return b.Bytes(), nil
}
