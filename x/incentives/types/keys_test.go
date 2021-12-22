package types

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tharsis/ethermint/tests"
)

// TODO Somethimes tests fails
func TestSplitGasMeterKey(t *testing.T) {
	contract := tests.GenerateAddress()
	user := tests.GenerateAddress()
	key := append(append(KeyPrefixGasMeter, contract.Bytes()...), user.Bytes()...)

	contract2, user2 := SplitGasMeterKey(key)
	require.Equal(t, contract2, contract)
	require.Equal(t, user2, user)
}
