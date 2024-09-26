// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package wrappers

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v20/types"
	"github.com/evmos/evmos/v20/x/evm/config"

	"github.com/stretchr/testify/require"
)

func TestMustConvertEvmCoinTo18Decimals(t *testing.T) {
	testCases := []struct {
		name        string
		evmCoinInfo config.EvmCoinInfo
		coin        sdk.Coin
		expCoin     sdk.Coin
		expPanic    bool
	}{
		{
			name:        "pass - zero amount",
			evmCoinInfo: config.EvmCoinInfo{Denom: types.BaseDenom, Decimals: config.EighteenDecimals},
			coin:        sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(0)},
			expPanic:    false,
			expCoin:     sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(0)},
		},
		{
			name:        "pass - no conversion with 18 decimals",
			evmCoinInfo: config.EvmCoinInfo{Denom: types.BaseDenom, Decimals: config.EighteenDecimals},
			coin:        sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(10)},
			expPanic:    false,
			expCoin:     sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(10)},
		},
		{
			name:        "pass - conversion with 6 decimals",
			evmCoinInfo: config.EvmCoinInfo{Denom: types.BaseDenom, Decimals: config.SixDecimals},
			coin:        sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(1)},
			expPanic:    false,
			expCoin:     sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(1e12)},
		},
		{
			name:        "panic - not evm denom",
			evmCoinInfo: config.EvmCoinInfo{Denom: types.BaseDenom, Decimals: config.SixDecimals},
			coin:        sdk.Coin{Denom: "evmos", Amount: math.NewInt(1)},
			expPanic:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if tc.expPanic {
						require.NotNil(t, r, "expected test to panic")
					} else {
						t.Errorf("unexpected panic: %v", r)
					}
				} else if tc.expPanic {
					t.Errorf("expected panic but did not occur")
				}
			}()

			config.SetEVMCoindTEST(tc.evmCoinInfo)

			coinConverted := mustConvertEvmCoinTo18Decimals(tc.coin)

			if !tc.expPanic {
				require.Equal(t, tc.expCoin, coinConverted, "expected a different coin")
			}
		})
	}
}
