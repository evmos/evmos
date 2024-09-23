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

// // convert6To18DecimalsCoin converts the coin's Amount to 18 decimals from 6.
// func convert6To18DecimalsCoin(coin sdk.Coin) (sdk.Coin, error) {
// 	if err := validateReceivedCoin(coin); err != nil {
// 		return coin, err
// 	}
//
// 	evmCoinDecimal := config.GetEVMCoinDecimals()
// 	if evmCoinDecimal == config.EighteenDecimals {
// 		return coin, nil
// 	}
//
// 	coin.Amount = coin.Amount.MulRaw(1e12)
// 	return coin, nil
// }
//
// // convert18To6DecimalsCoin converts the coin's Amount to 6 decimals from 18
// func convert18To6DecimalsCoin(coin sdk.Coin) (sdk.Coin, error) {
// 	if err := validateReceivedCoin(coin); err != nil {
// 		return coin, err
// 	}
//
// 	evmCoinDecimal := config.GetEVMCoinDecimals()
// 	if evmCoinDecimal == config.SixDecimals {
// 		return coin, nil
// 	}
//
// 	coin.Amount = coin.Amount.QuoRaw(1e12)
// 	return coin, nil
// }
//
// // validateReceivedCoin validate that the given coin has the same denom of the
// // expected EVM coin.
// func validateReceivedCoin(coin sdk.Coin) error {
// 	evmCoinDenom := config.GetEVMCoinDenom()
//
// 	if evmCoinDenom != coin.Denom {
// 		return fmt.Errorf("expected EVM coin denom %s but received %s", evmCoinDenom, coin.Denom)
// 	}
//
// 	return nil
// }
