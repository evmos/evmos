//go:build gofuzz || go1.18

package types_test

import (
	"testing"

	"github.com/evmos/evmos/v11/testutil"
	"github.com/evmos/evmos/v11/x/incentives/types"
)

func FuzzSplitGasMeterKey(f *testing.F) {
	contract := testutil.GenerateAddress()
	user := testutil.GenerateAddress()

	key := types.KeyPrefixGasMeter
	key = append(key, contract.Bytes()...)
	key = append(key, user.Bytes()...)
	f.Add(key)
	f.Fuzz(func(t *testing.T, key []byte) {
		types.SplitGasMeterKey(key)
	})
}
