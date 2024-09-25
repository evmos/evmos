// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package wrappers

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v20/x/evm/config"
)

// convertEvmCoinTo18DecimalsUnchecked convert the given coin from its original
// representation into a 18 decimals one. The function panic if coin denom is
// not the evm denom or in case of overflow.
func mustConvertEvmCoinTo18DecimalsUnchecked(coin sdk.Coin) sdk.Coin {
	if coin.Denom != config.GetEVMCoinDenom() {
		panic(fmt.Sprintf("expected evm denom %s, received %s", config.GetEVMCoinDenom(), coin.Denom))
	}

	evmCoinDecimal := config.GetEVMCoinDecimals()
	newAmount := coin.Amount.Mul(math.NewInt(evmCoinDecimal.ConversionFactor()))

	return sdk.NewCoin(coin.Denom, newAmount)
}

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

	return sdk.NewCoin(coin.Denom, newAmount), nil
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

	return sdk.NewCoin(coin.Denom, newAmount), nil
}

// convertCoinsFrom18Decimals returns the given coins with the evm amount
// converted from the 18 decimals representation to the original one.
func convertCoinsFrom18Decimals(coins sdk.Coins) (sdk.Coins, error) {
	evmDenom := config.GetEVMCoinDenom()
	convertedCoins := make(sdk.Coins, len(coins))

	for i, coin := range coins {
		if coin.Denom == evmDenom {
			convertedCoin, err := convertEvmCoinFrom18Decimals(coins[i])
			if err != nil {
				return sdk.Coins{}, err
			}
			coin = convertedCoin
		}
		convertedCoins[i] = coin
	}
	return coins, nil
}
