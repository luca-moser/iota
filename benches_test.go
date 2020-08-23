package iota_test

import (
	"crypto/ed25519"
	"testing"

	"github.com/luca-moser/iota"
)

func BenchmarkDeserializeWithValidationOneIOSigTxPayload(b *testing.B) {
	data, err := oneInputOutputSignedTransactionPayload().Serialize(true)
	if err != nil {
		b.Fatal(err)
	}

	target := &iota.SignedTransactionPayload{}
	_, err = target.Deserialize(data, true)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		target.Deserialize(data, false)
	}
}

func BenchmarkDeserializeWithoutValidationOneIOSigTxPayload(b *testing.B) {
	data, err := oneInputOutputSignedTransactionPayload().Serialize(true)
	if err != nil {
		b.Fatal(err)
	}

	target := &iota.SignedTransactionPayload{}
	_, err = target.Deserialize(data, true)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		target.Deserialize(data, true)
	}
}

func BenchmarkSerializeWithValidationOneIOSigTxPayload(b *testing.B) {
	sigTxPayload := oneInputOutputSignedTransactionPayload()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sigTxPayload.Serialize(false)
	}
}

func BenchmarkSerializeWithoutValidationOneIOSigTxPayload(b *testing.B) {
	sigTxPayload := oneInputOutputSignedTransactionPayload()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sigTxPayload.Serialize(true)
	}
}

func BenchmarkSignEd25519OneIOUnsignedTx(b *testing.B) {
	sigTxPayload := oneInputOutputSignedTransactionPayload()
	b.ResetTimer()

	unsigTxData, err := sigTxPayload.Transaction.Serialize(true)
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

	unsigTxData, err := sigTxPayload.Transaction.Serialize(true)
	must(err)

	seed := randEd25519Seed()
	prvKey := ed25519.NewKeyFromSeed(seed[:])

	sig := ed25519.Sign(prvKey, unsigTxData)

	pubKey := prvKey.Public().(ed25519.PublicKey)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ed25519.Verify(pubKey, unsigTxData, sig)
	}
}
