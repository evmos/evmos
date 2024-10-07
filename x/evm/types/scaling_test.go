// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types_test

import (
	"fmt"
	"math/big"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v20/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"

	"github.com/stretchr/testify/require"
)

func TestMustConvertEvmCoinTo18Decimals(t *testing.T) {
	baseCoinZero := sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(0)}

	testCases := []struct {
		name        string
		evmCoinInfo evmtypes.EvmCoinInfo
		coin        sdk.Coin
		expCoin     sdk.Coin
		expPanic    bool
	}{
		{
			name:        "pass - zero amount 18 decimals",
			evmCoinInfo: evmtypes.EvmCoinInfo{Denom: types.BaseDenom, Decimals: evmtypes.EighteenDecimals},
			coin:        baseCoinZero,
			expPanic:    false,
			expCoin:     baseCoinZero,
		},
		{
			name:        "pass - zero amount 6 decimals",
			evmCoinInfo: evmtypes.EvmCoinInfo{Denom: types.BaseDenom, Decimals: evmtypes.SixDecimals},
			coin:        baseCoinZero,
			expPanic:    false,
			expCoin:     baseCoinZero,
		},
		{
			name:        "pass - no conversion with 18 decimals",
			evmCoinInfo: evmtypes.EvmCoinInfo{Denom: types.BaseDenom, Decimals: evmtypes.EighteenDecimals},
			coin:        sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(10)},
			expPanic:    false,
			expCoin:     sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(10)},
		},
		{
			name:        "pass - conversion with 6 decimals",
			evmCoinInfo: evmtypes.EvmCoinInfo{Denom: types.BaseDenom, Decimals: evmtypes.SixDecimals},
			coin:        sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(1)},
			expPanic:    false,
			expCoin:     sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(1e12)},
		},
		{
			name:        "panic - not evm denom",
			evmCoinInfo: evmtypes.EvmCoinInfo{Denom: types.BaseDenom, Decimals: evmtypes.SixDecimals},
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

			configurator := evmtypes.NewEVMConfigurator()
			configurator.ResetTestChainConfig()
			require.NoError(t, configurator.WithEVMCoinInfo(tc.evmCoinInfo.Denom, uint8(tc.evmCoinInfo.Decimals)).Configure())

			coinConverted := evmtypes.MustConvertEvmCoinTo18Decimals(tc.coin)

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
		evmCoinInfo evmtypes.EvmCoinInfo
		coin        sdk.Coin
		expCoin     sdk.Coin
		expErr      bool
	}{
		{
			name:        "pass - zero amount 18 decimals",
			evmCoinInfo: evmtypes.EvmCoinInfo{Denom: types.BaseDenom, Decimals: evmtypes.EighteenDecimals},
			coin:        baseCoinZero,
			expErr:      false,
			expCoin:     baseCoinZero,
		},
		{
			name:        "pass - zero amount 6 decimals",
			evmCoinInfo: evmtypes.EvmCoinInfo{Denom: types.BaseDenom, Decimals: evmtypes.SixDecimals},
			coin:        baseCoinZero,
			expErr:      false,
			expCoin:     baseCoinZero,
		},
		{
			name:        "pass - no conversion with 18 decimals",
			evmCoinInfo: evmtypes.EvmCoinInfo{Denom: types.BaseDenom, Decimals: evmtypes.EighteenDecimals}, coin: sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(10)}, expErr: false,
			expCoin: sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(10)},
		},
		{
			name:        "pass - conversion with 6 decimals",
			evmCoinInfo: evmtypes.EvmCoinInfo{Denom: types.BaseDenom, Decimals: evmtypes.SixDecimals},
			coin:        sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(1e12)},
			expErr:      false,
			expCoin:     sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(1)},
		},
		{
			name:        "pass - conversion with amount less than conversion factor",
			evmCoinInfo: evmtypes.EvmCoinInfo{Denom: types.BaseDenom, Decimals: evmtypes.SixDecimals},
			coin:        sdk.Coin{Denom: types.BaseDenom, Amount: math.NewInt(1e11)},
			expErr:      false,
			expCoin:     baseCoinZero,
		},
		{
			name:        "fail - not evm denom",
			evmCoinInfo: evmtypes.EvmCoinInfo{Denom: types.BaseDenom, Decimals: evmtypes.SixDecimals},
			coin:        sdk.Coin{Denom: "evmos", Amount: math.NewInt(1)},
			expErr:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configurator := evmtypes.NewEVMConfigurator()
			configurator.ResetTestChainConfig()
			require.NoError(t, configurator.WithEVMCoinInfo(tc.evmCoinInfo.Denom, uint8(tc.evmCoinInfo.Decimals)).Configure())

			coinConverted, err := evmtypes.ConvertEvmCoinFrom18Decimals(tc.coin)

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
		evmCoinInfo evmtypes.EvmCoinInfo
		coins       sdk.Coins
		expCoins    sdk.Coins
	}{
		{
			name:        "pass - no evm denom",
			evmCoinInfo: evmtypes.EvmCoinInfo{Denom: types.BaseDenom, Decimals: evmtypes.SixDecimals},
			coins:       sdk.Coins{nonBaseCoin},
			expCoins:    sdk.Coins{nonBaseCoin},
		},
		{
			name:        "pass - only base denom 18 decimals",
			evmCoinInfo: evmtypes.EvmCoinInfo{Denom: types.BaseDenom, Decimals: evmtypes.EighteenDecimals},
			coins:       sdk.Coins{baseCoin},
			expCoins:    sdk.Coins{baseCoin},
		},
		{
			name:        "pass - only base denom 6 decimals",
			evmCoinInfo: evmtypes.EvmCoinInfo{Denom: types.BaseDenom, Decimals: evmtypes.SixDecimals},
			coins:       sdk.Coins{baseCoin},
			expCoins:    sdk.Coins{sdk.Coin{Denom: types.BaseDenom, Amount: baseCoin.Amount.QuoRaw(1e12)}},
		},
		{
			name:        "pass - multiple coins and base denom 18 decimals",
			evmCoinInfo: evmtypes.EvmCoinInfo{Denom: types.BaseDenom, Decimals: evmtypes.EighteenDecimals},
			coins:       sdk.Coins{nonBaseCoin, baseCoin},
			expCoins:    sdk.Coins{nonBaseCoin, baseCoin},
		},
		{
			name:        "pass - multiple coins and base denom 6 decimals",
			evmCoinInfo: evmtypes.EvmCoinInfo{Denom: types.BaseDenom, Decimals: evmtypes.SixDecimals},
			coins:       sdk.Coins{nonBaseCoin, baseCoin},
			expCoins:    sdk.Coins{nonBaseCoin, sdk.Coin{Denom: types.BaseDenom, Amount: baseCoin.Amount.QuoRaw(1e12)}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configurator := evmtypes.NewEVMConfigurator()
			configurator.ResetTestChainConfig()
			require.NoError(t, configurator.WithEVMCoinInfo(tc.evmCoinInfo.Denom, uint8(tc.evmCoinInfo.Decimals)).Configure())

			coinConverted := evmtypes.ConvertCoinsFrom18Decimals(tc.coins)
			require.Equal(t, tc.expCoins, coinConverted, "expected a different coin")
		})
	}
}

func TestZeroExtraDecimalsBigInt(t *testing.T) {
	testCases := []struct {
		name string
		amt  *big.Int
		exp  *big.Int
	}{
		{
			name: "almost 1: 0.99999...",
			amt:  big.NewInt(999999999999),
			exp:  big.NewInt(0),
		},
		{
			name: "decimal < 5: 1.4",
			amt:  big.NewInt(14e11),
			exp:  big.NewInt(1e12),
		},
		{
			name: "decimal < 5: 1.499999999999",
			amt:  big.NewInt(1499999999999),
			exp:  big.NewInt(1e12),
		},
		{
			name: "decimal == 5: 1.5",
			amt:  big.NewInt(15e11),
			exp:  big.NewInt(1e12),
		},
		{
			name: "decimal > 5: 1.9",
			amt:  big.NewInt(19e11),
			exp:  big.NewInt(1e12),
		},
		{
			name: "1 wei",
			amt:  big.NewInt(1),
			exp:  big.NewInt(0),
		},
	}

	for _, cfg := range []evmtypes.EvmCoinInfo{
		{Denom: types.BaseDenom, Decimals: evmtypes.SixDecimals},
		{Denom: types.BaseDenom, Decimals: evmtypes.EighteenDecimals},
	} {
		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%d dec - %s", cfg.Decimals, tc.name), func(t *testing.T) {
				configurator := evmtypes.NewEVMConfigurator()
				configurator.ResetTestChainConfig()
				require.NoError(t, configurator.WithEVMCoinInfo(cfg.Denom, uint8(cfg.Decimals)).Configure())
				res := evmtypes.AdjustExtraDecimalsBigInt(tc.amt)
				if cfg.Decimals == evmtypes.EighteenDecimals {
					tc.exp = tc.amt
				}
				require.Equal(t, tc.exp, res)
			})
		}
	}
}

func TestConvertBigIntFrom18DecimalsToLegacyDec(t *testing.T) {
	testCases := []struct {
		name    string
		amt     *big.Int
		exp6dec math.LegacyDec
	}{
		{
			name:    "smallest amount",
			amt:     big.NewInt(1),
			exp6dec: math.LegacyMustNewDecFromStr("0.000000000001"),
		},
		{
			name:    "almost 1: 0.99999...",
			amt:     big.NewInt(999999999999),
			exp6dec: math.LegacyMustNewDecFromStr("0.999999999999"),
		},
		{
			name:    "half of the minimum uint",
			amt:     big.NewInt(5e11),
			exp6dec: math.LegacyMustNewDecFromStr("0.5"),
		},
		{
			name:    "one int",
			amt:     big.NewInt(1e12),
			exp6dec: math.LegacyOneDec(),
		},
		{
			name:    "one 'ether'",
			amt:     big.NewInt(1e18),
			exp6dec: math.LegacyNewDec(1e6),
		},
	}

	for _, cfg := range []evmtypes.EvmCoinInfo{
		{Denom: types.BaseDenom, Decimals: evmtypes.SixDecimals},
		{Denom: types.BaseDenom, Decimals: evmtypes.EighteenDecimals},
	} {
		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%d dec - %s", cfg.Decimals, tc.name), func(t *testing.T) {
				configurator := evmtypes.NewEVMConfigurator()
				configurator.ResetTestChainConfig()
				require.NoError(t, configurator.WithEVMCoinInfo(cfg.Denom, uint8(cfg.Decimals)).Configure())
				res := evmtypes.ConvertBigIntFrom18DecimalsToLegacyDec(tc.amt)
				exp := math.LegacyNewDecFromBigInt(tc.amt)
				if cfg.Decimals == evmtypes.SixDecimals {
					exp = tc.exp6dec
				}
				require.Equal(t, exp, res)
			})
		}
	}
}
