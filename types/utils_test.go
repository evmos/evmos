package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
)

func TestIsSupportedKeys(t *testing.T) {
	testCases := []struct {
		name        string
		pk          cryptotypes.PubKey
		isSupported bool
	}{
		{
			"ethsecp256k1 key",
			&ethsecp256k1.PubKey{},
			true,
		},
		{
			"ed25519 key",
			&ed25519.PubKey{},
			true,
		},
		{
			"multisig key - no pubkeys",
			&multisig.LegacyAminoPubKey{},
			false,
		},
		{
			"multisig key - valid pubkeys",
			multisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{&ed25519.PubKey{}, &ed25519.PubKey{}, &ed25519.PubKey{}}),
			true,
		},
		{
			"multisig key - nested multisig",
			multisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{&ed25519.PubKey{}, &ed25519.PubKey{}, &multisig.LegacyAminoPubKey{}}),
			false,
		},
		{
			"multisig key - invalid pubkey",
			multisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{&ed25519.PubKey{}, &ed25519.PubKey{}, &secp256k1.PubKey{}}),
			false,
		},
		{
			"cosmos secp256k1",
			&secp256k1.PubKey{},
			false,
		},
	}

	for _, tc := range testCases {
		require.Equal(t, tc.isSupported, IsSupportedKey(tc.pk), tc.name)
	}
}
