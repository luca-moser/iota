package iota_test

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/luca-moser/iota"
)

func TestLSFile_Compression(t *testing.T) {
	rand.Seed(1346587549867)

	const sepsCount = 150
	const txCount = 1000000
	const maxRandOutputsPerTx = 4

	var outputsTotal uint64
	lsFile := &iota.LSFile{
		MilestoneIndex: uint64(rand.Intn(1000000)),
		MilestoneHash:  randTxHash(),
		Timestamp:      uint64(time.Now().Unix()),
		SEPs: func() [][32]byte {
			seps := make([][32]byte, sepsCount)
			for i := 0; i < sepsCount; i++ {
				copy(seps[i][:], randBytes(iota.TransactionIDLength))
			}
			return seps
		}(),
		// generate UTXOs
		UTXOs: func() []*iota.LSTransactionUnspentOutputs {
			txs := make([]*iota.LSTransactionUnspentOutputs, txCount)
			for i := 0; i < txCount; i++ {
				tx := &iota.LSTransactionUnspentOutputs{TransactionHash: randTxHash()}
				tx.UnspentOutputs = func() []*iota.LSUnspentOutput {
					outputsCount := rand.Intn(maxRandOutputsPerTx + 1)
					outputsTotal += uint64(outputsCount)
					outputs := make([]*iota.LSUnspentOutput, outputsCount)
					for i := 0; i < outputsCount; i++ {
						addr, _ := randEd25519Addr()
						outputs[i] = &iota.LSUnspentOutput{
							Index:   byte(i),
							Address: addr,
							Value:   uint64(rand.Intn(1000000 + 1)),
						}
					}
					return outputs
				}()
				txs[i] = tx
			}
			return txs
		}(),
	}

	lsFileData, err := lsFile.Serialize(iota.DeSeriModeNoValidation)
	must(err)

	fmt.Printf("raw ls file size with %d SEPs, %d outputs (from %d txs): %d MB\n", sepsCount, outputsTotal, txCount, len(lsFileData)/1024/1024)

	var zlibBuf bytes.Buffer
	zlibWriter := zlib.NewWriter(&zlibBuf)

	_, err = zlibWriter.Write(lsFileData)
	must(err)
	must(zlibWriter.Close())

	fmt.Printf("size after zlib compression: %d MB\n", len(zlibBuf.Bytes())/1024/2024)

	var gzipBuf bytes.Buffer
	gzipWriter := gzip.NewWriter(&gzipBuf)

	_, err = gzipWriter.Write(lsFileData)
	must(err)
	must(gzipWriter.Close())

	fmt.Printf("size after gzip compression: %d MB\n", len(gzipBuf.Bytes())/1024/2024)
}
