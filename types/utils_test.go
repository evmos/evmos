package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
)

func init() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("evmos", "evmospub")
}

func TestIsSupportedKeys(t *testing.T) {
	testCases := []struct {
		name        string
		pk          cryptotypes.PubKey
		isSupported bool
	}{
		{
			"nil key",
			nil,
			false,
		},
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

func TestGetEvmosAddressFromBech32(t *testing.T) {
	testCases := []struct {
		name       string
		address    string
		expAddress string
		expError   bool
	}{
		{
			"blank bech32 address",
			" ",
			"",
			true,
		},
		{
			"invalid bech32 address",
			"evmos",
			"",
			true,
		},
		{
			"invalid address bytes",
			"evmos1123",
			"",
			true,
		},
		{
			"evmos address",
			"evmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueuafmxps",
			"evmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueuafmxps",
			false,
		},
		{
			"cosmos address",
			"cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
			"evmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueuafmxps",
			false,
		},
		{
			"osmosis address",
			"osmo1qql8ag4cluz6r4dz28p3w00dnc9w8ueuhnecd2",
			"evmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueuafmxps",
			false,
		},
	}

	for _, tc := range testCases {
		addr, err := GetEvmosAddressFromBech32(tc.address)
		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
			require.Equal(t, tc.expAddress, addr.String(), tc.name)
		}
	}
}
