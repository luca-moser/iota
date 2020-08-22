package iota_test

import (
	"errors"
	"testing"

	"github.com/luca-moser/iotapkg"
	"github.com/stretchr/testify/assert"
)

func TestWOTSAddress_Deserialize(t *testing.T) {
	tests := []struct {
		name         string
		wotsAddrData []byte
		err          error
	}{
		{
			"ok",
			func() []byte {
				_, wotsAddrData := randWOTSAddr()
				return wotsAddrData
			}(),
			nil,
		},
		{
			"not enough bytes",
			func() []byte {
				_, wotsAddrData := randWOTSAddr()
				return wotsAddrData[:iota.WOTSAddressSerializedBytesSize-1]
			}(),
			iota.ErrInvalidBytes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wotsAddr := &iota.WOTSAddress{}
			bytesRead, err := wotsAddr.Deserialize(tt.wotsAddrData)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.wotsAddrData), bytesRead)
			assert.Equal(t, tt.wotsAddrData[iota.OneByte:], wotsAddr[:])
		})
	}
}

func TestWOTSAddress_Serialize(t *testing.T) {
	originWOTSAddr, originData := randWOTSAddr()
	tests := []struct {
		name   string
		source *iota.WOTSAddress
		target []byte
	}{
		{
			"ok", originWOTSAddr, originData,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wotsData, err := tt.source.Serialize()
			assert.NoError(t, err)
			assert.Equal(t, tt.target, wotsData)
		})
	}
}

func TestEd25519Address_Deserialize(t *testing.T) {
	tests := []struct {
		name       string
		edAddrData []byte
		err        error
	}{
		{
			"ok",
			func() []byte {
				_, edAddrData := randEd25519Addr()
				return edAddrData
			}(),
			nil,
		},
		{
			"not enough bytes",
			func() []byte {
				_, edAddrData := randEd25519Addr()
				return edAddrData[:iota.Ed25519AddressSerializedBytesSize-1]
			}(),
			iota.ErrInvalidBytes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edAddr := &iota.Ed25519Address{}
			bytesRead, err := edAddr.Deserialize(tt.edAddrData)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.edAddrData), bytesRead)
			assert.Equal(t, tt.edAddrData[iota.OneByte:], edAddr[:])
		})
	}
}

func TestEd25519Address_Serialize(t *testing.T) {
	originEdAddr, originData := randEd25519Addr()
	tests := []struct {
		name   string
		source *iota.Ed25519Address
		target []byte
	}{
		{
			"ok", originEdAddr, originData,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize()
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}
