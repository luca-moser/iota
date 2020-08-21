package iotapkg_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/luca-moser/iotapkg"
	"github.com/stretchr/testify/assert"
)

const (
	TypeA       = 0
	TypeB       = 1
	aKeyLength  = 16
	bNameLength = 32
)

var (
	ErrUnknownDummyType = errors.New("unknown example type")
)

func DummyTypeSelector(typeByte byte) (iotapkg.Serializable, error) {
	var seri iotapkg.Serializable
	switch typeByte {
	case TypeA:
		seri = &A{}
	case TypeB:
		seri = &B{}
	default:
		return nil, ErrUnknownDummyType
	}
	return seri, nil
}

type A struct {
	Key [aKeyLength]byte
}

func (a *A) Deserialize(data []byte) (int, error) {
	data = data[iotapkg.OneByte:]
	copy(a.Key[:], data[:aKeyLength])
	return iotapkg.OneByte + aKeyLength, nil
}

func (a *A) Serialize() ([]byte, error) {
	var bytes [iotapkg.OneByte + aKeyLength]byte
	bytes[0] = TypeA
	copy(bytes[iotapkg.OneByte:], a.Key[:])
	return bytes[:], nil
}

func randSerializedA() []byte {
	keyData := randBytes(aKeyLength)
	return append([]byte{TypeA}, keyData...)
}

func randA() *A {
	var k [aKeyLength]byte
	copy(k[:], randBytes(aKeyLength))
	return &A{Key: k}
}

type B struct {
	Name [bNameLength]byte
}

func (b *B) Deserialize(data []byte) (int, error) {
	data = data[iotapkg.OneByte:]
	copy(b.Name[:], data[:bNameLength])
	return iotapkg.OneByte + bNameLength, nil
}

func (b *B) Serialize() ([]byte, error) {
	var bytes [iotapkg.OneByte + bNameLength]byte
	bytes[0] = TypeB
	copy(bytes[iotapkg.OneByte:], b.Name[:])
	return bytes[:], nil
}

func randSerializedB() []byte {
	nameData := randBytes(bNameLength)
	return append([]byte{TypeB}, nameData...)
}

func randB() *B {
	var n [bNameLength]byte
	copy(n[:], randBytes(bNameLength))
	return &B{Name: n}
}

func TestDeserializeA(t *testing.T) {
	seriA := randSerializedA()
	objA := &A{}
	bytesRead, err := objA.Deserialize(seriA)
	assert.NoError(t, err)
	assert.Equal(t, len(seriA), bytesRead)
	assert.Equal(t, seriA[iotapkg.OneByte:], objA.Key[:])
}

func TestDeserializeObject(t *testing.T) {
	seriA := randSerializedA()
	objA, bytesRead, err := iotapkg.DeserializeObject(seriA, DummyTypeSelector)
	assert.NoError(t, err)
	assert.Equal(t, len(seriA), bytesRead)
	assert.IsType(t, &A{}, objA)
	assert.Equal(t, seriA[iotapkg.OneByte:], objA.(*A).Key[:])
}

func TestDeserializeArrayOfObjects(t *testing.T) {
	var buf bytes.Buffer
	originObjs := []iotapkg.Serializable{
		randA(), randA(), randB(), randA(), randB(), randB(),
	}
	assert.NoError(t, buf.WriteByte(byte(len(originObjs))))

	for _, seri := range originObjs {
		seriBytes, err := seri.Serialize()
		assert.NoError(t, err)
		written, err := buf.Write(seriBytes)
		assert.NoError(t, err)
		assert.Equal(t, len(seriBytes), written)
	}

	data := buf.Bytes()
	seris, serisByteRead, err := iotapkg.DeserializeArrayOfObjects(data, DummyTypeSelector)
	assert.NoError(t, err)
	assert.Equal(t, len(data), serisByteRead)
	assert.EqualValues(t, originObjs, seris)
}
