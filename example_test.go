package iota_test

import (
	"fmt"
	"testing"

	"github.com/luca-moser/iota"
	"github.com/stretchr/testify/require"
)

func TestSignedTransactionPayloadSize(t *testing.T) {
	data, err := oneInputOutputSignedTransactionPayload().Serialize(iota.DeSeriModeNoValidation)
	require.NoError(t, err)
	fmt.Printf("length of signed transaction payload: %d\n", len(data))
}
