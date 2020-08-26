package iota

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
)

// LSFile defines local snapshot file content.
// The de/serialization functions should only be used for testing.
type LSFile struct {
	MilestoneIndex uint64                         `json:"milestone_index"`
	MilestoneHash  [32]byte                       `json:"milestone_hash"`
	Timestamp      uint64                         `json:"timestamp"`
	SEPs           [][32]byte                     `json:"seps"`
	UTXOs          []*LSTransactionUnspentOutputs `json:"utxos"`
}

func (s *LSFile) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	panic("implement me")
}

func (s *LSFile) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	var b bytes.Buffer
	if err := binary.Write(&b, binary.LittleEndian, s.MilestoneIndex); err != nil {
		return nil, err
	}

	if _, err := b.Write(s.MilestoneHash[:]); err != nil {
		return nil, err
	}

	if err := binary.Write(&b, binary.LittleEndian, s.Timestamp); err != nil {
		return nil, err
	}

	varIntBuf := make([]byte, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(varIntBuf, uint64(len(s.SEPs)))
	if _, err := b.Write(varIntBuf[:bytesWritten]); err != nil {
		return nil, err
	}

	for _, sep := range s.SEPs {
		if _, err := b.Write(sep[:]); err != nil {
			return nil, err
		}
	}

	bytesWritten = binary.PutUvarint(varIntBuf, uint64(len(s.SEPs)))
	if _, err := b.Write(varIntBuf[:bytesWritten]); err != nil {
		return nil, err
	}

	for _, txOuts := range s.UTXOs {
		txOutsData, err := txOuts.Serialize(deSeriMode)
		if err != nil {
			return nil, err
		}
		if _, err := b.Write(txOutsData); err != nil {
			return nil, err
		}
	}

	sha256Hash := sha256.Sum256(b.Bytes())
	if _, err := b.Write(sha256Hash[:]); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// LSTransactionUnspentOutputs are the unspent outputs under the same transaction hash.
type LSTransactionUnspentOutputs struct {
	TransactionHash [32]byte           `json:"transaction_hash"`
	UnspentOutputs  []*LSUnspentOutput `json:"unspent_outputs"`
}

func (s *LSTransactionUnspentOutputs) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	panic("implement me")
}

func (s *LSTransactionUnspentOutputs) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	var b bytes.Buffer
	if _, err := b.Write(s.TransactionHash[:]); err != nil {
		return nil, err
	}

	// write count of outputs
	varIntBuf := make([]byte, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(varIntBuf, uint64(len(s.UnspentOutputs)))
	if _, err := b.Write(varIntBuf[:bytesWritten]); err != nil {
		return nil, err
	}

	for _, out := range s.UnspentOutputs {
		outData, err := out.Serialize(deSeriMode)
		if err != nil {
			return nil, err
		}
		if _, err := b.Write(outData); err != nil {
			return nil, err
		}
	}

	return b.Bytes(), nil
}

// LSUnspentOutput defines an unspent output.
type LSUnspentOutput struct {
	Index   byte         `json:"index"`
	Address Serializable `json:"address"`
	Value   uint64       `json:"value"`
}

func (s *LSUnspentOutput) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	panic("implement me")
}

func (s *LSUnspentOutput) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	var b bytes.Buffer
	if err := b.WriteByte(s.Index); err != nil {
		return nil, err
	}
	addrData, err := s.Address.Serialize(deSeriMode)
	if err != nil {
		return nil, err
	}
	if _, err := b.Write(addrData); err != nil {
		return nil, err
	}
	if err := binary.Write(&b, binary.LittleEndian, s.Value); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
