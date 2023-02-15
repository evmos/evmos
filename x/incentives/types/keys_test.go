package types_test

import (
	"testing"

	"github.com/evmos/evmos/v11/testutil"
	"github.com/evmos/evmos/v11/x/incentives/types"

	"github.com/stretchr/testify/require"
)

func TestSplitGasMeterKey(t *testing.T) {
	contract := testutil.GenerateAddress()
	user := testutil.GenerateAddress()

	key := types.KeyPrefixGasMeter
	key = append(key, contract.Bytes()...)
	key = append(key, user.Bytes()...)

	contract2, user2 := types.SplitGasMeterKey(key)
	require.Equal(t, contract2, contract)
	require.Equal(t, user2, user)
}
