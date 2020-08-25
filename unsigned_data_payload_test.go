package iota_test

import (
	"errors"
	"testing"

	"github.com/luca-moser/iota"
	"github.com/stretchr/testify/assert"
)

func TestUnsignedDataPayload_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target iota.Serializable
		err    error
	}
	tests := []test{
		func() test {
			unsigDataPayload, unsigDataPayloadData := randUnsignedDataPayload()
			return test{"ok", unsigDataPayloadData, unsigDataPayload, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unsigDataPayload := &iota.UnsignedDataPayload{}
			bytesRead, err := unsigDataPayload.Deserialize(tt.source, iota.DeSeriModePerformValidation)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.source), bytesRead)
			assert.EqualValues(t, tt.target, unsigDataPayload)
		})
	}
}

func TestUnsignedDataPayload_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iota.UnsignedDataPayload
		target []byte
	}
	tests := []test{
		func() test {
			unsigDataPayload, unsigDataPayloadData := randUnsignedDataPayload()
			return test{"ok", unsigDataPayload, unsigDataPayloadData}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize(iota.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}
