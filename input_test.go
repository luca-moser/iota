package iotapkg_test

import (
	"errors"
	"testing"

	"github.com/luca-moser/iotapkg"
	"github.com/stretchr/testify/assert"
)

func TestInputSelector(t *testing.T) {
	_, err := iotapkg.InputSelector(100)
	assert.True(t, errors.Is(err, iotapkg.ErrUnknownInputType))
}

func TestUTXOInput_Deserialize(t *testing.T) {
	randUTXOInput, randSerializedUTXOInput := randUTXOInput()
	tests := []struct {
		name   string
		data   []byte
		target *iotapkg.UTXOInput
		err    error
	}{
		{"ok", randSerializedUTXOInput, randUTXOInput, nil},
		{"not enough data", randSerializedUTXOInput[:iotapkg.UTXOInputMinSize-1], randUTXOInput, iotapkg.ErrInvalidBytes},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &iotapkg.UTXOInput{}
			bytesRead, err := u.Deserialize(tt.data)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.Equal(t, len(tt.data), bytesRead)
			assert.EqualValues(t, tt.target, u)
		})
	}
}

func TestUTXOInput_Serialize(t *testing.T) {
	randUTXOInput, randSerializedUTXOInput := randUTXOInput()
	tests := []struct {
		name   string
		source *iotapkg.UTXOInput
		target []byte
		err    error
	}{
		{"ok", randUTXOInput, randSerializedUTXOInput, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.source.Serialize()
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.EqualValues(t, tt.target, data)
		})
	}
}
