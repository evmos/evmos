package types

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tharsis/ethermint/tests"
)

func TestSplitGasMeterKey(t *testing.T) {
	contract := tests.GenerateAddress()
	user := tests.GenerateAddress()

	key := KeyPrefixGasMeter
	key = append(key, contract.Bytes()...)
	key = append(key, user.Bytes()...)

	contract2, user2 := SplitGasMeterKey(key)
	require.Equal(t, contract2, contract)
	require.Equal(t, user2, user)
}
