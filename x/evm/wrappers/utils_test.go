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
	baseCoinZero := sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(0)}

	testCases := []struct {
		name        string
		evmCoinInfo config.EvmCoinInfo
		coin        sdk.Coin
		expCoin     sdk.Coin
		expPanic    bool
	}{
		{
			name:        "pass - zero amount 18 decimals",
			evmCoinInfo: config.EvmCoinInfo{Denom: types.BaseDenom, Decimals: config.EighteenDecimals},
			coin:        baseCoinZero,
			expPanic:    false,
			expCoin:     baseCoinZero,
		},
		{
			name:        "pass - zero amount 6 decimals",
			evmCoinInfo: config.EvmCoinInfo{Denom: types.BaseDenom, Decimals: config.SixDecimals},
			coin:        baseCoinZero,
			expPanic:    false,
			expCoin:     baseCoinZero,
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

			config.SetEVMCoinTEST(tc.evmCoinInfo)

			coinConverted := mustConvertEvmCoinTo18Decimals(tc.coin)

			if !tc.expPanic {
				require.Equal(t, tc.expCoin, coinConverted, "expected a different coin")
			}
		})
	}
}

func TestConvertEvmCoinFrom18Decimals(t *testing.T) {
	baseCoinZero := sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(0)}

	testCases := []struct {
		name        string
		evmCoinInfo config.EvmCoinInfo
		coin        sdk.Coin
		expCoin     sdk.Coin
		expErr      bool
	}{
		{
			name:        "pass - zero amount 18 decimals",
			evmCoinInfo: config.EvmCoinInfo{Denom: types.BaseDenom, Decimals: config.EighteenDecimals},
			coin:        baseCoinZero,
			expErr:      false,
			expCoin:     baseCoinZero,
		},
		{
			name:        "pass - zero amount 6 decimals",
			evmCoinInfo: config.EvmCoinInfo{Denom: types.BaseDenom, Decimals: config.SixDecimals},
			coin:        baseCoinZero,
			expErr:      false,
			expCoin:     baseCoinZero,
		},
		{
			name:        "pass - no conversion with 18 decimals",
			evmCoinInfo: config.EvmCoinInfo{Denom: types.BaseDenom, Decimals: config.EighteenDecimals}, coin: sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(10)}, expErr: false,
			expCoin: sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(10)},
		},
		{
			name:        "pass - conversion with 6 decimals",
			evmCoinInfo: config.EvmCoinInfo{Denom: types.BaseDenom, Decimals: config.SixDecimals},
			coin:        sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(1e12)},
			expErr:      false,
			expCoin:     sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(1)},
		},
		{
			name:        "pass - conversion with amount less than conversion factor",
			evmCoinInfo: config.EvmCoinInfo{Denom: types.BaseDenom, Decimals: config.SixDecimals},
			coin:        sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(1e11)},
			expErr:      false,
			expCoin:     baseCoinZero,
		},
		{
			name:        "fail - not evm denom",
			evmCoinInfo: config.EvmCoinInfo{Denom: types.BaseDenom, Decimals: config.SixDecimals},
			coin:        sdk.Coin{Denom: "evmos", Amount: math.NewInt(1)},
			expErr:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config.SetEVMCoinTEST(tc.evmCoinInfo)

			coinConverted, err := convertEvmCoinFrom18Decimals(tc.coin)

			if !tc.expErr {
				require.NoError(t, err)
				require.Equal(t, tc.expCoin, coinConverted, "expected a different coin")
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestConvertCoinsFrom18Decimals(t *testing.T) {
	nonBaseCoin := sdk.Coin{Denom: "btc", Amount: math.NewInt(10)}
	baseCoin := sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(10)}

	testCases := []struct {
		name        string
		evmCoinInfo config.EvmCoinInfo
		coins       sdk.Coins
		expCoins    sdk.Coins
	}{
		{
			name:        "pass - no evm denom",
			evmCoinInfo: config.EvmCoinInfo{Denom: types.BaseDenom, Decimals: config.SixDecimals},
			coins:       sdk.Coins{nonBaseCoin},
			expCoins:    sdk.Coins{nonBaseCoin},
		},
		{
			name:        "pass - only base denom 18 decimals",
			evmCoinInfo: config.EvmCoinInfo{Denom: types.BaseDenom, Decimals: config.EighteenDecimals},
			coins:       sdk.Coins{baseCoin},
			expCoins:    sdk.Coins{baseCoin},
		},
		{
			name:        "pass - only base denom 6 decimals",
			evmCoinInfo: config.EvmCoinInfo{Denom: types.BaseDenom, Decimals: config.SixDecimals},
			coins:       sdk.Coins{baseCoin},
			expCoins:    sdk.Coins{sdk.Coin{Denom: types.BaseDenom, Amount: baseCoin.Amount.QuoRaw(1e12)}},
		},
		{
			name:        "pass - multiple coins and base denom 18 decimals",
			evmCoinInfo: config.EvmCoinInfo{Denom: types.BaseDenom, Decimals: config.EighteenDecimals},
			coins:       sdk.Coins{nonBaseCoin, baseCoin},
			expCoins:    sdk.Coins{nonBaseCoin, baseCoin},
		},
		{
			name:        "pass - multiple coins and base denom 6 decimals",
			evmCoinInfo: config.EvmCoinInfo{Denom: types.BaseDenom, Decimals: config.SixDecimals},
			coins:       sdk.Coins{nonBaseCoin, baseCoin},
			expCoins:    sdk.Coins{nonBaseCoin, sdk.Coin{Denom: types.BaseDenom, Amount: baseCoin.Amount.QuoRaw(1e12)}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config.SetEVMCoinTEST(tc.evmCoinInfo)

			coinConverted := convertCoinsFrom18Decimals(tc.coins)
			require.Equal(t, tc.expCoins, coinConverted, "expected a different coin")
		})
	}
}
