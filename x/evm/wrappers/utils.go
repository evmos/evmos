// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package wrappers

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v20/x/evm/config"
)

// convertEvmCoinTo18Decimals converts the coin's Amount from its original
// representation to 18 decimals. Return an error if the coin denom is not the
// EVM denom or in case of overflow.
func convertEvmCoinTo18Decimals(coin sdk.Coin) (sdk.Coin, error) {
	if coin.Denom != config.GetEVMCoinDenom() {
		return sdk.Coin{}, fmt.Errorf("expected coin denom %s, received %s", config.GetEVMCoinDenom(), coin.Denom)
	}

	evmCoinDecimal := config.GetEVMCoinDecimals()
	newAmount, err := coin.Amount.SafeMul(math.NewInt(evmCoinDecimal.ConversionFactor()))
	if err != nil {
		return sdk.Coin{}, err
	}

	coin.Amount = newAmount

	return coin, nil
}

// convertEvmCoinFrom18Decimals converts the coin's Amount from 18 decimals to its
// original representation. Return an error if the coin denom is not the EVM
// denom or in case of underflow.
func convertEvmCoinFrom18Decimals(coin sdk.Coin) (sdk.Coin, error) {
	if coin.Denom != config.GetEVMCoinDenom() {
		return sdk.Coin{}, fmt.Errorf("expected coin denom %s, received %s", config.GetEVMCoinDenom(), coin.Denom)
	}

	evmCoinDecimal := config.GetEVMCoinDecimals()
	newAmount, err := coin.Amount.SafeQuo(math.NewInt(evmCoinDecimal.ConversionFactor()))
	if err != nil {
		return sdk.Coin{}, err
	}

	coin.Amount = newAmount

	return coin, nil
}

// convertCoinsFrom18Decimals returns the given coins with the evm amount
// converted from the 18 decimals representation to the original one.
func convertCoinsFrom18Decimals(coins sdk.Coins) (sdk.Coins, error) {
	evmDenom := config.GetEVMCoinDenom()
	for i, coin := range coins {
		if coin.Denom == evmDenom {
			convertedCoin, err := convertEvmCoinFrom18Decimals(coins[i])
			if err != nil {
				return sdk.Coins{}, nil
			}
			coins[i] = convertedCoin
		}
	}
	return coins, nil
}
