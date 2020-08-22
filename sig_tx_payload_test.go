package iotapkg_test

import (
	"errors"
	"testing"

	"github.com/luca-moser/iotapkg"
	"github.com/stretchr/testify/assert"
)

func TestSignedTransactionPayload_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target iotapkg.Serializable
		err    error
	}
	tests := []test{
		func() test {
			sigTxPay, sigTxPayData := randSignedTransactionPayload()
			return test{"ok", sigTxPayData, sigTxPay, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &iotapkg.SignedTransactionPayload{}
			bytesRead, err := tx.Deserialize(tt.source)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.source), bytesRead)
			assert.EqualValues(t, tt.target, tx)
		})
	}
}

func TestSignedTransactionPayload_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iotapkg.SignedTransactionPayload
		target []byte
	}
	tests := []test{
		func() test {
			sigTxPay, sigTxPayData := randSignedTransactionPayload()
			return test{"ok", sigTxPay, sigTxPayData}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize()
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}

func BenchmarkDeserializeOneInputOutputSignedTransactionPayload(b *testing.B) {
	data, err := oneInputOutputSignedTransactionPayload().Serialize()
	if err != nil {
		b.Fatal(err)
	}

	target := &iotapkg.SignedTransactionPayload{}
	_, err = target.Deserialize(data)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		target.Deserialize(data)
	}
}

func BenchmarkSerializeOneInputOutputSignedTransactionPayload(b *testing.B) {
	sigTxPayload := oneInputOutputSignedTransactionPayload()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sigTxPayload.Serialize()
	}
}
