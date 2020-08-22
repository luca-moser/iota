package iotapkg_test

import (
	"errors"
	"testing"

	"github.com/luca-moser/iotapkg"
	"github.com/stretchr/testify/assert"
)

func TestUnlockBlockSelector(t *testing.T) {
	_, err := iotapkg.UnlockBlockSelector(100)
	assert.True(t, errors.Is(err, iotapkg.ErrUnknownUnlockBlockType))
}

func TestSignatureUnlockBlock_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target iotapkg.Serializable
		err    error
	}
	tests := []test{
		func() test {
			edSigBlock, edSigBlockData := randEd25519SignatureUnlockBlock()
			return test{"ok", edSigBlockData, edSigBlock, nil}
		}(),
		func() test {
			edSigBlock, edSigBlockData := randEd25519SignatureUnlockBlock()
			return test{"not enough data", edSigBlockData[:5], edSigBlock, iotapkg.ErrInvalidBytes}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edSig := &iotapkg.SignatureUnlockBlock{}
			bytesRead, err := edSig.Deserialize(tt.source)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.source), bytesRead)
			assert.EqualValues(t, tt.target, edSig)
		})
	}
}

func TestUnlockBlockSignature_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iotapkg.SignatureUnlockBlock
		target []byte
	}
	tests := []test{
		func() test {
			edSigBlock, edSigBlockData := randEd25519SignatureUnlockBlock()
			return test{"ok", edSigBlock, edSigBlockData}
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

func TestReferenceUnlockBlock_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target iotapkg.Serializable
		err    error
	}
	tests := []test{
		func() test {
			refBlock, refBlockData := randReferenceUnlockBlock()
			return test{"ok", refBlockData, refBlock, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edSig := &iotapkg.ReferenceUnlockBlock{}
			bytesRead, err := edSig.Deserialize(tt.source)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.source), bytesRead)
			assert.EqualValues(t, tt.target, edSig)
		})
	}
}

func TestUnlockBlockReference_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iotapkg.ReferenceUnlockBlock
		target []byte
	}
	tests := []test{
		func() test {
			refBlock, refBlockData := randReferenceUnlockBlock()
			return test{"ok", refBlock, refBlockData}
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

func TestUnlockBlockValidatorFunc(t *testing.T) {
	type args struct {
		inputs []iotapkg.Serializable
		funcs  []iotapkg.UnlockBlockValidatorFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"ok",
			args{inputs: []iotapkg.Serializable{
				func() iotapkg.Serializable {
					block, _ := randEd25519SignatureUnlockBlock()
					return block
				}(),
				func() iotapkg.Serializable {
					block, _ := randEd25519SignatureUnlockBlock()
					return block
				}(),
				func() iotapkg.Serializable {
					return &iotapkg.ReferenceUnlockBlock{Reference: 0}
				}(),
			}, funcs: []iotapkg.UnlockBlockValidatorFunc{iotapkg.UnlockBlocksSigUniqueAndRefValidator()}}, false,
		},
		{
			"duplicate ed25519 sig block",
			args{inputs: []iotapkg.Serializable{
				func() iotapkg.Serializable {
					return &iotapkg.SignatureUnlockBlock{Signature: &iotapkg.Ed25519Signature{
						PublicKey: [32]byte{},
						Signature: [64]byte{},
					}}
				}(),
				func() iotapkg.Serializable {
					return &iotapkg.SignatureUnlockBlock{Signature: &iotapkg.Ed25519Signature{
						PublicKey: [32]byte{},
						Signature: [64]byte{},
					}}
				}(),
			}, funcs: []iotapkg.UnlockBlockValidatorFunc{iotapkg.UnlockBlocksSigUniqueAndRefValidator()}}, true,
		},
		{
			"invalid ref",
			args{inputs: []iotapkg.Serializable{
				func() iotapkg.Serializable {
					block, _ := randEd25519SignatureUnlockBlock()
					return block
				}(),
				func() iotapkg.Serializable {
					block, _ := randEd25519SignatureUnlockBlock()
					return block
				}(),
				func() iotapkg.Serializable {
					return &iotapkg.ReferenceUnlockBlock{Reference: 2}
				}(),
			}, funcs: []iotapkg.UnlockBlockValidatorFunc{iotapkg.UnlockBlocksSigUniqueAndRefValidator()}}, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := iotapkg.ValidateUnlockBlocks(tt.args.inputs, tt.args.funcs); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
