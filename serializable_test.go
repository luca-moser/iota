package iota_test

import (
	"bytes"
	"errors"
	"sort"
	"testing"

	"github.com/luca-moser/iota"
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

func DummyTypeSelector(typeByte byte) (iota.Serializable, error) {
	var seri iota.Serializable
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

func (a *A) Deserialize(data []byte, deSeriMode iota.DeSerializationMode) (int, error) {
	data = data[iota.OneByte:]
	copy(a.Key[:], data[:aKeyLength])
	return iota.OneByte + aKeyLength, nil
}

func (a *A) Serialize(deSeriMode iota.DeSerializationMode) ([]byte, error) {
	var bytes [iota.OneByte + aKeyLength]byte
	bytes[0] = TypeA
	copy(bytes[iota.OneByte:], a.Key[:])
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

func (b *B) Deserialize(data []byte, deSeriMode iota.DeSerializationMode) (int, error) {
	data = data[iota.OneByte:]
	copy(b.Name[:], data[:bNameLength])
	return iota.OneByte + bNameLength, nil
}

func (b *B) Serialize(deSeriMode iota.DeSerializationMode) ([]byte, error) {
	var bytes [iota.OneByte + bNameLength]byte
	bytes[0] = TypeB
	copy(bytes[iota.OneByte:], b.Name[:])
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
	bytesRead, err := objA.Deserialize(seriA, iota.DeSeriModePerformValidation)
	assert.NoError(t, err)
	assert.Equal(t, len(seriA), bytesRead)
	assert.Equal(t, seriA[iota.OneByte:], objA.Key[:])
}

func TestDeserializeObject(t *testing.T) {
	seriA := randSerializedA()
	objA, bytesRead, err := iota.DeserializeObject(seriA, iota.DeSeriModePerformValidation, DummyTypeSelector)
	assert.NoError(t, err)
	assert.Equal(t, len(seriA), bytesRead)
	assert.IsType(t, &A{}, objA)
	assert.Equal(t, seriA[iota.OneByte:], objA.(*A).Key[:])
}

func TestDeserializeArrayOfObjects(t *testing.T) {
	var buf bytes.Buffer
	originObjs := []iota.Serializable{
		randA(), randA(), randB(), randA(), randB(), randB(),
	}
	assert.NoError(t, buf.WriteByte(byte(len(originObjs))))

	for _, seri := range originObjs {
		seriBytes, err := seri.Serialize(iota.DeSeriModePerformValidation)
		assert.NoError(t, err)
		written, err := buf.Write(seriBytes)
		assert.NoError(t, err)
		assert.Equal(t, len(seriBytes), written)
	}

	data := buf.Bytes()
	seris, serisByteRead, err := iota.DeserializeArrayOfObjects(data, iota.DeSeriModePerformValidation, DummyTypeSelector, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), serisByteRead)
	assert.EqualValues(t, originObjs, seris)
}

func TestLexicalOrderedByteSlices(t *testing.T) {
	type test struct {
		name   string
		source iota.LexicalOrderedByteSlices
		target iota.LexicalOrderedByteSlices
	}
	tests := []test{
		{
			name: "ok - order by first ele",
			source: iota.LexicalOrderedByteSlices{
				{3, 2, 1},
				{2, 3, 1},
				{1, 2, 3},
			},
			target: iota.LexicalOrderedByteSlices{
				{1, 2, 3},
				{2, 3, 1},
				{3, 2, 1},
			},
		},
		{
			name: "ok - order by last ele",
			source: iota.LexicalOrderedByteSlices{
				{1, 1, 3},
				{1, 1, 2},
				{1, 1, 1},
			},
			target: iota.LexicalOrderedByteSlices{
				{1, 1, 1},
				{1, 1, 2},
				{1, 1, 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Sort(tt.source)
			assert.Equal(t, tt.target, tt.source)
		})
	}
}

func TestSerializationMode_HasMode(t *testing.T) {
	type args struct {
		mode iota.DeSerializationMode
	}
	tests := []struct {
		name string
		sm   iota.DeSerializationMode
		args args
		want bool
	}{
		{
			"has no validation",
			iota.DeSeriModeNoValidation,
			args{mode: iota.DeSeriModePerformValidation},
			false,
		},
		{
			"has validation",
			iota.DeSeriModePerformValidation,
			args{mode: iota.DeSeriModePerformValidation},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sm.HasMode(tt.args.mode); got != tt.want {
				t.Errorf("HasMode() = %v, want %v", got, tt.want)
			}
		})
	}
}
