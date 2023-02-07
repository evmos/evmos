//go:build gofuzz || go1.18

package types

import (
	"testing"

	"github.com/evmos/evmos/v11/tests"
)

func FuzzSplitGasMeterKey(f *testing.F) {
	contract := tests.GenerateAddress()
	user := tests.GenerateAddress()

	key := KeyPrefixGasMeter
	key = append(key, contract.Bytes()...)
	key = append(key, user.Bytes()...)
	f.Add(key)
	f.Fuzz(func(t *testing.T, key []byte) {
		SplitGasMeterKey(key)
	})
}
