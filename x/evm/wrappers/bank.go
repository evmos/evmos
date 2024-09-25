// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package wrappers

import (
	"context"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v20/x/evm/config"
	"github.com/evmos/evmos/v20/x/evm/types"
)

var _ types.BankWrapper = BankWrapper{}

// BankWrapper is a wrapper around the Cosmos SDK bank keeper
// that is used to manage an evm denom with a custom decimal representation.
type BankWrapper struct {
	types.BankKeeper
}

// NewBankWrapper creates a new BankWrapper instance.
func NewBankWrapper(
	bk types.BankKeeper,
) *BankWrapper {
	return &BankWrapper{
		bk,
	}
}

// ------------------------------------------------------------------------------------------
// Bank wrapper own methods
// ------------------------------------------------------------------------------------------

// MintAmountToAccount converts the given amount into the evm coin scaling
// the amount to the original decimals, then mints that amount to the provided account.
func (w BankWrapper) MintAmountToAccount(ctx context.Context, recipientAddr sdk.AccAddress, amt math.Int) error {
	coin := sdk.NewCoin(config.GetEVMCoinDenom(), amt)

	convertedCoin, err := convertEvmCoinFrom18Decimals(coin)
	if err != nil {
		return errors.Wrap(err, "failed to mint coin to account in bank wrapper")
	}

	coinsToMint := sdk.NewCoins(convertedCoin)
	if err := w.MintCoins(ctx, types.ModuleName, coinsToMint); err != nil {
		return errors.Wrap(err, "failed to mint coins to account in bank wrapper")
	}

	return w.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, coinsToMint)
}

// BurnAmountFromAccount convert the given amount into the evm coin scaling
// the amount to the original decimals, then burns that quantity from the provided account.
func (w BankWrapper) BurnAmountFromAccount(ctx context.Context, account sdk.AccAddress, amt math.Int) error {
	coin := sdk.NewCoin(config.GetEVMCoinDenom(), amt)

	convertedCoin, err := convertEvmCoinFrom18Decimals(coin)
	if err != nil {
		return errors.Wrap(err, "failed to burn coins from account in bank wrapper")
	}

	coinsToBurn := sdk.NewCoins(convertedCoin)
	if err := w.SendCoinsFromAccountToModule(ctx, account, types.ModuleName, coinsToBurn); err != nil {
		return errors.Wrap(err, "failed to burn coins from account in bank wrapper")
	}
	return w.BurnCoins(ctx, types.ModuleName, coinsToBurn)
}

// ------------------------------------------------------------------------------------------
// Bank keeper shadowed methods
// ------------------------------------------------------------------------------------------

// GetBalance returns the balance of the given account converted to 18 decimals.
func (w BankWrapper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	// Get the balance from the BankModule. The balance returned is in the bank
	// decimals representation, which could be different than the 18 decimals
	// representation used in the evm.
	coin := w.BankKeeper.GetBalance(ctx, addr, denom)

	return mustConvertEvmCoinTo18Decimals(coin)
}

// // SendCoinsFromAccountToModule wraps around the Cosmos SDK x/bank module's
// // SendCoinsFromAccountToModule method to convert the evm coin, if present in
// // the input, to its original representation.
// func (w BankWrapper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, coins sdk.Coins) error {
// 	convertedCoins, err := convertCoinsFrom18Decimals(coins)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to send coins to module in bank wrapper")
// 	}
//
// 	return w.BankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, convertedCoins)
// }

// SendCoinsFromModuleToAccount wraps around the Cosmos SDK x/bank module's
// SendCoinsFromModuleToAccount method to convert the evm coin, if present in
// the input, to its original representation.
func (w BankWrapper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, coins sdk.Coins) error {
	convertedCoins, err := convertCoinsFrom18Decimals(coins)
	if err != nil {
		return errors.Wrap(err, "failed to send coins to account in bank wrapper")
	}

	return w.BankKeeper.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, convertedCoins)
}
