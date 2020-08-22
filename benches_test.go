package iota_test

import (
	"crypto/ed25519"
	"testing"

	"github.com/luca-moser/iotapkg"
)

func BenchmarkDeserializeOneIOSigTxPayload(b *testing.B) {
	data, err := oneInputOutputSignedTransactionPayload().Serialize()
	if err != nil {
		b.Fatal(err)
	}

	target := &iota.SignedTransactionPayload{}
	_, err = target.Deserialize(data)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		target.Deserialize(data)
	}
}

func BenchmarkSerializeOneIOSigTxPayload(b *testing.B) {
	sigTxPayload := oneInputOutputSignedTransactionPayload()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sigTxPayload.Serialize()
	}
}

func BenchmarkSignEd25519OneIOUnsignedTx(b *testing.B) {
	sigTxPayload := oneInputOutputSignedTransactionPayload()
	b.ResetTimer()

	unsigTxData, err := sigTxPayload.Transaction.Serialize()
	must(err)

	seed := randEd25519Seed()
	prvKey := ed25519.NewKeyFromSeed(seed[:])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ed25519.Sign(prvKey, unsigTxData)
	}
}

func BenchmarkVerifyEd25519OneIOUnsignedTx(b *testing.B) {
	sigTxPayload := oneInputOutputSignedTransactionPayload()
	b.ResetTimer()

	unsigTxData, err := sigTxPayload.Transaction.Serialize()
	must(err)

	seed := randEd25519Seed()
	prvKey := ed25519.NewKeyFromSeed(seed[:])

	sig := ed25519.Sign(prvKey, unsigTxData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ed25519.Verify(prvKey.Public().(ed25519.PublicKey), unsigTxData, sig)
	}
}
