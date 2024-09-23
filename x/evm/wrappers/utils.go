// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package wrappers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v20/x/evm/config"
)

// convertTo18DecimalsCoin converts the coin's Amount from its original
// representation to 18 decimals.
func convertTo18DecimalsCoin(coin sdk.Coin) sdk.Coin {
	evmCoinDecimal := config.GetEVMCoinDecimals()

	coin.Amount = coin.Amount.MulRaw(evmCoinDecimal.ConversionFactor())
	return coin
}

// convertFrom18DecimalsCoin converts the coin's Amount from 18 decimals to its
// original representation.
func convertFrom18DecimalsCoin(coin sdk.Coin) sdk.Coin {
	evmCoinDecimal := config.GetEVMCoinDecimals()

	coin.Amount = coin.Amount.QuoRaw(evmCoinDecimal.ConversionFactor())
	return coin
}
