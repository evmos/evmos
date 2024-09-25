// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package wrappers

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

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

// GetBalance returns the balance of the given account converted to 18 decimals.
func (w BankWrapper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	// Get the balance from the BankModule. The balance returned is in the bank
	// decimals representation, which could be different than the 18 decimals
	// representation used in the evm.
	coin := w.BankKeeper.GetBalance(ctx, addr, denom)

	return mustConvertEvmCoinTo18Decimals(coin)
}

// SendCoinsFromAccountToModule wraps around the Cosmos SDK x/bank module's
// SendCoinsFromAccountToModule method to convert the evm coin, if present in
// the input, to its original representation.
func (w BankWrapper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	coins, err := convertCoinsFrom18Decimals(amt)
	if err != nil {
		return errors.Wrap(err, "failed to send coins to module in bank wrapper")
	}

	return w.BankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, coins)
}

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

// MintCoinsToAccount converts the evm coin to the original decimal
// representation if required, and mint requested coins to the provided account.
func (w BankWrapper) MintCoinsToAccount(ctx context.Context, recipientAddr sdk.AccAddress, coins sdk.Coins) error {
	convertedCoins, err := convertCoinsFrom18Decimals(coins)
	if err != nil {
		return errors.Wrap(err, "failed to mint coins to account in bank wrapper")
	}

	if err := w.MintCoins(ctx, types.ModuleName, convertedCoins); err != nil {
		return errors.Wrap(err, "failed to mint coins to account in bank wrapper")
	}

	return w.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, convertedCoins)
}

// BurnCoinsFromAccount converts the evm coin to the original decimal representation
// if required, and burn the requested coins from the given account.
func (w BankWrapper) BurnCoinsFromAccount(ctx context.Context, account sdk.AccAddress, coins sdk.Coins) error {
	convertedCoins, err := convertCoinsFrom18Decimals(coins)
	if err != nil {
		return errors.Wrap(err, "failed to burn coins from account in bank wrapper")
	}

	// NOTE: amt is already converted so we need to use the x/bank method.
	if err := w.BankKeeper.SendCoinsFromAccountToModule(ctx, account, types.ModuleName, convertedCoins); err != nil {
		return errors.Wrap(err, "failed to burn coins from account in bank wrapper")
	}
	return w.BurnCoins(ctx, types.ModuleName, convertedCoins)
}
