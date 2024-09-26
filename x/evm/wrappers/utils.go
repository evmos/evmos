// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package wrappers

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v20/x/evm/config"
)

// mustConvertEvmCoinTo18Decimals converts the coin's Amount from its original
// representation into a 18 decimals. The function panics if coin denom is
// not the evm denom or in case of overflow.
func mustConvertEvmCoinTo18Decimals(coin sdk.Coin) sdk.Coin {
	if coin.Denom != config.GetEVMCoinDenom() {
		panic(fmt.Sprintf("expected evm denom %s, received %s", config.GetEVMCoinDenom(), coin.Denom))
	}

	evmCoinDecimal := config.GetEVMCoinDecimals()
	newAmount := coin.Amount.Mul(evmCoinDecimal.ConversionFactor())

	return sdk.Coin{Denom: coin.Denom, Amount: newAmount}
}

// convertEvmCoinFrom18Decimals converts the coin's Amount from 18 decimals to its
// original representation. Return an error if the coin denom is not the EVM.
func convertEvmCoinFrom18Decimals(coin sdk.Coin) (sdk.Coin, error) {
	if coin.Denom != config.GetEVMCoinDenom() {
		return sdk.Coin{}, fmt.Errorf("expected coin denom %s, received %s", config.GetEVMCoinDenom(), coin.Denom)
	}

	evmCoinDecimal := config.GetEVMCoinDecimals()
	newAmount := coin.Amount.Quo(evmCoinDecimal.ConversionFactor())

	return sdk.Coin{Denom: coin.Denom, Amount: newAmount}, nil
}

// convertCoinsFrom18Decimals returns the given coins with the Amount of the evm
// coin converted from the 18 decimals representation to the original one.
func convertCoinsFrom18Decimals(coins sdk.Coins) sdk.Coins {
	evmDenom := config.GetEVMCoinDenom()

	convertedCoins := make(sdk.Coins, len(coins))
	for i, coin := range coins {
		if coin.Denom == evmDenom {
			evmCoinDecimals := config.GetEVMCoinDecimals()

			newAmount := coin.Amount.Quo(evmCoinDecimals.ConversionFactor())

			coin = sdk.Coin{Denom: coin.Denom, Amount: newAmount}
		}
		convertedCoins[i] = coin
	}
	return convertedCoins
}
