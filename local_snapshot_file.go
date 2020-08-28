package iota

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

const (
	LSFormatVersion byte = 1

	SolidEntryPointHashLength = 32
)

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
	if err := binary.Write(&b, binary.LittleEndian, uint16(len(s.UnspentOutputs))); err != nil {
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
	Index   uint16       `json:"index"`
	Address Serializable `json:"address"`
	Value   uint64       `json:"value"`
}

func (s *LSUnspentOutput) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	panic("implement me")
}

func (s *LSUnspentOutput) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	var b bytes.Buffer
	if err := binary.Write(&b, binary.LittleEndian, s.Index); err != nil {
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

// LSSEPIteratorFunc yields a solid entry point to be written to a local snapshot or nil if no more is available.
type LSSEPIteratorFunc func() *[SolidEntryPointHashLength]byte

// LSSEPConsumerFunc consumes the given solid entry point.
// A returned error signals to cancel further reading.
type LSSEPConsumerFunc func([SolidEntryPointHashLength]byte) error

// LSHeaderConsumerFunc consumes the local snapshot file header.
// A returned error signals to cancel further reading.
type LSHeaderConsumerFunc func(*LSFileHeader) error

// LSUTXOIteratorFunc yields a transaction and its outputs to be written to a local snapshot or nil if no more is available.
type LSUTXOIteratorFunc func() *LSTransactionUnspentOutputs

// LSUTXOConsumerFunc consumes the given transaction and its outputs.
// A returned error signals to cancel further reading.
type LSUTXOConsumerFunc func(*LSTransactionUnspentOutputs) error

// LSFileHeader implements the WriteTo interface.
type LSFileHeader struct {
	Version        byte
	MilestoneIndex uint64
	MilestoneHash  [MilestoneHashLength]byte
	Timestamp      uint64
}

// A WriteFlusher writes data and has a Flush method.
type WriteFlusher interface {
	io.Writer
	Flush() error
}

// CompressionReaderInitFunc wraps a reader with a compression reader.
type CompressionReaderInitFunc func(io.Reader) (io.Reader, error)

// StreamLocalSnapshotDataTo streams local snapshot data into the given io.WriteSeeker.
func StreamLocalSnapshotDataTo(writeSeeker io.WriteSeeker, compWriter WriteFlusher, header *LSFileHeader,
	sepIter LSSEPIteratorFunc, utxoIter LSUTXOIteratorFunc) error {

	// version, seps count, utxo count
	// timestamp, milestone index, milestone hash, seps, utxos
	var sepsCount, utxoCount uint64

	if _, err := writeSeeker.Write([]byte{header.Version}); err != nil {
		return err
	}

	if err := binary.Write(writeSeeker, binary.LittleEndian, header.Timestamp); err != nil {
		return err
	}

	if err := binary.Write(writeSeeker, binary.LittleEndian, header.MilestoneIndex); err != nil {
		return err
	}

	if _, err := writeSeeker.Write(header.MilestoneHash[:]); err != nil {
		return err
	}

	// write count and hash place holders
	if _, err := writeSeeker.Write(make([]byte, UInt64ByteSize*2)); err != nil {
		return err
	}

	var maybeComprWriter io.Writer
	if compWriter != nil {
		maybeComprWriter = compWriter
	} else {
		maybeComprWriter = writeSeeker
	}

	for sep := sepIter(); sep != nil; sep = sepIter() {
		_, err := maybeComprWriter.Write(sep[:])
		if err != nil {
			return err
		}
		sepsCount++
	}

	for utxo := utxoIter(); utxo != nil; utxo = utxoIter() {
		utxoData, err := utxo.Serialize(DeSeriModeNoValidation)
		if err != nil {
			return err
		}
		if _, err := maybeComprWriter.Write(utxoData); err != nil {
			return err
		}
		utxoCount++
	}

	// flush content of compression writer
	if compWriter != nil {
		if err := compWriter.Flush(); err != nil {
			return err
		}
	}

	// seek back to counts version+timestamp+msindex+mshash and write element counts
	if _, err := writeSeeker.Seek(OneByte+UInt64ByteSize+UInt64ByteSize+MilestoneHashLength, io.SeekStart); err != nil {
		return err
	}

	if err := binary.Write(writeSeeker, binary.LittleEndian, sepsCount); err != nil {
		return err
	}

	if err := binary.Write(writeSeeker, binary.LittleEndian, utxoCount); err != nil {
		return err
	}

	return nil
}

// autocopyreader auto. writes content read from the reader to the writer on every Read call.
type autocopyreader struct {
	reader io.Reader
	writer io.Writer
}

func (a *autocopyreader) Read(p []byte) (n int, err error) {
	read, err := a.reader.Read(p)
	if err != nil {
		return read, err
	}

	if _, err := a.writer.Write(p[:read]); err != nil {
		return 0, err
	}

	return read, nil
}

// StreamLocalSnapshotDataFrom consumes local snapshot data from the given reader.
func StreamLocalSnapshotDataFrom(reader io.Reader, compReaderInit CompressionReaderInitFunc,
	headerConsumer LSHeaderConsumerFunc, sepConsumer LSSEPConsumerFunc, utxoConsumer LSUTXOConsumerFunc) error {
	header := &LSFileHeader{}

	if err := binary.Read(reader, binary.LittleEndian, &header.Version); err != nil {
		return err
	}

	if err := binary.Read(reader, binary.LittleEndian, &header.Timestamp); err != nil {
		return err
	}

	if err := binary.Read(reader, binary.LittleEndian, &header.MilestoneIndex); err != nil {
		return err
	}

	if _, err := io.ReadFull(reader, header.MilestoneHash[:]); err != nil {
		return err
	}

	if err := headerConsumer(header); err != nil {
		return err
	}

	var sepsCount uint64
	if err := binary.Read(reader, binary.LittleEndian, &sepsCount); err != nil {
		return err
	}

	var utxoCount uint64
	if err := binary.Read(reader, binary.LittleEndian, &utxoCount); err != nil {
		return err
	}

	readerToUse, err := compReaderInit(reader)
	if err != nil {
		return err
	}

	for i := uint64(0); i < sepsCount; i++ {
		var sep [SolidEntryPointHashLength]byte
		if _, err := io.ReadFull(readerToUse, sep[:]); err != nil {
			return err
		}

		// sep gets copied
		if err := sepConsumer(sep); err != nil {
			return err
		}
	}

	for i := uint64(0); i < utxoCount; i++ {
		utxo := &LSTransactionUnspentOutputs{}

		// read tx hash
		if _, err := io.ReadFull(readerToUse, utxo.TransactionHash[:]); err != nil {
			return err
		}

		var outputsCount uint16
		if err := binary.Read(readerToUse, binary.LittleEndian, &outputsCount); err != nil {
			return err
		}

		for j := uint16(0); j < outputsCount; j++ {
			output := &LSUnspentOutput{}

			if err := binary.Read(readerToUse, binary.LittleEndian, &output.Index); err != nil {
				return err
			}

			// look ahead address type
			var addrType [OneByte]byte
			if _, err := io.ReadFull(readerToUse, addrType[:]); err != nil {
				return err
			}

			addr, err := AddressSelector(addrType[0])
			if err != nil {
				return err
			}

			var addrDataWithoutType []byte
			switch addr.(type) {
			case *WOTSAddress:
				addrDataWithoutType = make([]byte, WOTSAddressBytesLength)
			case *Ed25519Address:
				addrDataWithoutType = make([]byte, Ed25519AddressBytesLength)
			default:
				panic("unknown address type")
			}

			// read the rest of the address
			if _, err := io.ReadFull(readerToUse, addrDataWithoutType); err != nil {
				return err
			}

			if _, err := addr.Deserialize(append([]byte{addrType[0]}, addrDataWithoutType...), DeSeriModePerformValidation); err != nil {
				return err
			}
			output.Address = addr

			if err := binary.Read(readerToUse, binary.LittleEndian, &output.Value); err != nil {
				return err
			}

			utxo.UnspentOutputs = append(utxo.UnspentOutputs, output)
		}

		if err := utxoConsumer(utxo); err != nil {
			return err
		}
	}

	return nil
}
