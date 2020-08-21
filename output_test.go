package iotapkg_test

import (
	"errors"
	"testing"

	"github.com/luca-moser/iotapkg"
	"github.com/stretchr/testify/assert"
)

func TestOutputSelector(t *testing.T) {
	_, err := iotapkg.OutputSelector(100)
	assert.True(t, errors.Is(err, iotapkg.ErrUnknownOutputType))
}

func TestSigLockedSingleDeposit_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target iotapkg.Serializable
		err    error
	}
	tests := []test{
		func() test {
			dep, depData := randSigLockedSingleDeposit(iotapkg.AddressWOTS)
			return test{"ok wots", depData, dep, nil}
		}(),
		func() test {
			dep, depData := randSigLockedSingleDeposit(iotapkg.AddressWOTS)
			return test{"not enough data wots", depData[:5], dep, iotapkg.ErrInvalidBytes}
		}(),
		func() test {
			dep, depData := randSigLockedSingleDeposit(iotapkg.AddressEd25519)
			return test{"ok ed25519", depData, dep, nil}
		}(),
		func() test {
			dep, depData := randSigLockedSingleDeposit(iotapkg.AddressEd25519)
			return test{"not enough data ed25519", depData[:5], dep, iotapkg.ErrInvalidBytes}
		}(),
		func() test {
			dep, depData := randSigLockedSingleDeposit(iotapkg.AddressEd25519)
			depData[iotapkg.SigLockedSingleDepositAddressOffset] = 100
			return test{"unknown addr type", depData, dep, iotapkg.ErrUnknownAddrType}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := &iotapkg.SigLockedSingleDeposit{}
			bytesRead, err := dep.Deserialize(tt.source)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.source), bytesRead)
			assert.EqualValues(t, tt.target, dep)
		})
	}
}

func TestSigLockedSingleDeposit_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iotapkg.SigLockedSingleDeposit
		target []byte
		err    error
	}
	tests := []test{
		func() test {
			dep, depData := randSigLockedSingleDeposit(iotapkg.AddressEd25519)
			return test{"ok", dep, depData, nil}
		}(),
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
