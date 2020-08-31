package iota_test

import (
	"compress/zlib"
	"fmt"
	"io"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/blang/vfs/memfs"
	"github.com/luca-moser/iota"
	"github.com/stretchr/testify/require"
)

type test struct {
	name             string
	snapshotFileName string
	originHeader     *iota.LSFileHeader
	compWriter       compWriterInit
	sepGenerator     iota.LSSEPIteratorFunc
	sepGenRetriever  sepRetrieverFunc
	utxoGenerator    iota.LSUTXOIteratorFunc
	utxoGenRetriever utxoRetrieverFunc
	headerConsumer   iota.LSHeaderConsumerFunc
	sepConsumer      iota.LSSEPConsumerFunc
	sepConRetriever  sepRetrieverFunc
	utxoConsumer     iota.LSUTXOConsumerFunc
	utxoConRetriever utxoRetrieverFunc
	compReadInit     iota.CompressionReaderInitFunc
}

func TestStreamLocalSnapshotDataToAndFrom(t *testing.T) {
	if testing.Short() {
		return
	}

	rand.Seed(346587549867)

	originHeader := &iota.LSFileHeader{
		Version: iota.LSFormatVersion, MilestoneIndex: uint64(rand.Intn(10000)),
		MilestoneHash: randTxHash(), Timestamp: uint64(time.Now().Unix()),
	}

	testCases := []test{
		func() test {

			// create generators and consumers
			utxoIterFunc, utxosGenRetriever := newUTXOGenerator(1000000, 4)
			sepIterFunc, sepGenRetriever := newSEPGenerator(150)
			sepConsumerfunc, sepsCollRetriever := newSEPCollector()
			utxoConsumerFunc, utxoCollRetriever := newUTXOCollector()

			t := test{
				name:             "150 seps, 1 mil txs, uncompressed",
				snapshotFileName: "uncompressed_snapshot.bin",
				originHeader:     originHeader,
				compWriter: func(writer io.Writer) iota.WriteFlusher {
					return nil
				},
				sepGenerator:     sepIterFunc,
				sepGenRetriever:  sepGenRetriever,
				utxoGenerator:    utxoIterFunc,
				utxoGenRetriever: utxosGenRetriever,
				headerConsumer:   headerEqualFunc(t, originHeader),
				sepConsumer:      sepConsumerfunc,
				sepConRetriever:  sepsCollRetriever,
				utxoConsumer:     utxoConsumerFunc,
				utxoConRetriever: utxoCollRetriever,
				compReadInit: func(reader io.Reader) (io.Reader, error) {
					return reader, nil
				},
			}
			return t
		}(),
		func() test {

			// create generators and consumers
			utxoIterFunc, utxosGenRetriever := newUTXOGenerator(1000000, 4)
			sepIterFunc, sepGenRetriever := newSEPGenerator(150)
			sepConsumerfunc, sepsCollRetriever := newSEPCollector()
			utxoConsumerFunc, utxoCollRetriever := newUTXOCollector()

			t := test{
				name:             "150 seps, 1 mil txs, compressed (zlib, high)",
				snapshotFileName: "zlib_snapshot.bin",
				originHeader:     originHeader,
				compWriter: func(writer io.Writer) iota.WriteFlusher {
					zlibWriter, err := zlib.NewWriterLevel(writer, zlib.BestCompression)
					must(err)
					return zlibWriter
				},
				sepGenerator:     sepIterFunc,
				sepGenRetriever:  sepGenRetriever,
				utxoGenerator:    utxoIterFunc,
				utxoGenRetriever: utxosGenRetriever,
				headerConsumer:   headerEqualFunc(t, originHeader),
				sepConsumer:      sepConsumerfunc,
				sepConRetriever:  sepsCollRetriever,
				utxoConsumer:     utxoConsumerFunc,
				utxoConRetriever: utxoCollRetriever,
				compReadInit: func(reader io.Reader) (io.Reader, error) {
					zlibReader, err := zlib.NewReader(reader)
					require.NoError(t, err)
					return zlibReader, nil
				},
			}
			return t
		}(),
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.snapshotFileName
			fs := memfs.Create()
			snapshotFileWrite, err := fs.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0666)
			require.NoError(t, err)

			require.NoError(t, iota.StreamLocalSnapshotDataTo(snapshotFileWrite, tt.compWriter(snapshotFileWrite), tt.originHeader, tt.sepGenerator, tt.utxoGenerator))
			require.NoError(t, snapshotFileWrite.Close())

			fileInfo, err := fs.Stat(filePath)
			require.NoError(t, err)
			fmt.Printf("%s: written local snapshot file size: %d MB\n", tt.name, fileInfo.Size()/1024/1024)

			// read back written data and verify that it is equal
			snapshotFileRead, err := fs.OpenFile(filePath, os.O_RDONLY, 0666)
			require.NoError(t, err)

			require.NoError(t, iota.StreamLocalSnapshotDataFrom(snapshotFileRead, tt.compReadInit, tt.headerConsumer, tt.sepConsumer, tt.utxoConsumer))

			utxoGenerated, _ := tt.utxoGenRetriever()
			utxoConsumed, _ := tt.utxoConRetriever()
			require.EqualValues(t, utxoGenerated, utxoConsumed)
			require.EqualValues(t, tt.sepGenRetriever(), tt.sepConRetriever())
		})
	}

}

type compWriterInit func(writer io.Writer) iota.WriteFlusher

type sepRetrieverFunc func() [][iota.SolidEntryPointHashLength]byte

func newSEPGenerator(count int) (iota.LSSEPIteratorFunc, sepRetrieverFunc) {
	var generatedSEPs [][iota.SolidEntryPointHashLength]byte
	return func() *[iota.SolidEntryPointHashLength]byte {
			if count == 0 {
				return nil
			}
			count--
			x := randTxHash()
			generatedSEPs = append(generatedSEPs, x)
			return &x
		}, func() [][32]byte {
			return generatedSEPs
		}
}

func newSEPCollector() (iota.LSSEPConsumerFunc, sepRetrieverFunc) {
	var generatedSEPs [][iota.SolidEntryPointHashLength]byte
	return func(sep [iota.SolidEntryPointHashLength]byte) error {
			generatedSEPs = append(generatedSEPs, sep)
			return nil
		}, func() [][32]byte {
			return generatedSEPs
		}
}

type utxoRetrieverFunc func() ([]iota.LSTransactionUnspentOutputs, uint64)

func newUTXOGenerator(count int, maxRandOutputsPerTx int) (iota.LSUTXOIteratorFunc, utxoRetrieverFunc) {
	var generatedUTXOs []iota.LSTransactionUnspentOutputs
	var outputsTotal uint64
	return func() *iota.LSTransactionUnspentOutputs {
			if count == 0 {
				return nil
			}
			count--
			outputsCount := rand.Intn(maxRandOutputsPerTx) + 1
			tx := randLSTransactionUnspentOutputs(outputsCount)
			generatedUTXOs = append(generatedUTXOs, *tx)
			outputsTotal += uint64(len(tx.UnspentOutputs))
			return tx
		}, func() ([]iota.LSTransactionUnspentOutputs, uint64) {
			return generatedUTXOs, outputsTotal
		}
}

func newUTXOCollector() (iota.LSUTXOConsumerFunc, utxoRetrieverFunc) {
	var generatedUTXOs []iota.LSTransactionUnspentOutputs
	return func(utxo *iota.LSTransactionUnspentOutputs) error {
			generatedUTXOs = append(generatedUTXOs, *utxo)
			return nil
		}, func() ([]iota.LSTransactionUnspentOutputs, uint64) {
			return generatedUTXOs, 0
		}
}

func headerEqualFunc(t *testing.T, originHeader *iota.LSFileHeader) iota.LSHeaderConsumerFunc {
	return func(readHeader *iota.LSFileHeader) error {
		require.Equal(t, originHeader, readHeader)
		return nil
	}
}
