//go:build gofuzz || go1.18

package types

import (
	"testing"

	"github.com/evmos/evmos/v11/testutil"
)

func FuzzSplitGasMeterKey(f *testing.F) {
	contract := testutil.GenerateAddress()
	user := testutil.GenerateAddress()

	key := KeyPrefixGasMeter
	key = append(key, contract.Bytes()...)
	key = append(key, user.Bytes()...)
	f.Add(key)
	f.Fuzz(func(t *testing.T, key []byte) {
		SplitGasMeterKey(key)
	})
}
