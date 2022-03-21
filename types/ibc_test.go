package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func init() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("evmos", "evmospub")
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
