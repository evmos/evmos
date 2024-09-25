// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package wrappers

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v20/x/evm/types"
)

var _ types.BankWrapper = BankWrapper{}

// BankWrapper is a wrapper around the Cosmos SDK bank keeper
// that is used to manage an evm denom a custom decimal representation.
// The wrapper makes the corresponding conversions to achieve:
//   - With the EVM, the wrapper works always with 18 decimals.
//   - With the Cosmos bank module, the wrapper works always
//     with the bank module decimals.
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
	coin := w.BankKeeper.GetBalance(ctx, addr, denom)

	convertedCoin := mustConvertEvmCoinTo18DecimalsUnchecked(coin)

	return convertedCoin
}

// SendCoinsFromAccountToModule wraps around the Cosmos SDK x/bank module's
// SendCoinsFromAccountToModule method to convert the evm coin, if present in
// the input, to its original representation.
func (w BankWrapper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	coins, err := convertCoinsFrom18Decimals(amt)
	if err != nil {
		return err
	}

	return w.BankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, coins)
}

// SendCoinsFromModuleToAccount wraps around the Cosmos SDK x/bank module's
// SendCoinsFromModuleToAccount method to convert the evm coin, if present in
// the input, to its original representation.
func (w BankWrapper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	coins, err := convertCoinsFrom18Decimals(amt)
	if err != nil {
		return err
	}

	// NOTE: amt is already converted so we need to use the x/bank method.
	return w.BankKeeper.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, coins)
}

// MintCoinsToAccount converts the evm coin to the original decimal
// representation if required, and mint requested coins to the provided account.
func (w BankWrapper) MintCoinsToAccount(ctx context.Context, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	coins, err := convertCoinsFrom18Decimals(amt)
	if err != nil {
		return err
	}

	if err := w.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return err
	}
	// NOTE: amt is already converted so we need to use the x/bank method.
	return w.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, coins)
}

// BurnCoinsFromAccount converts the evm coin to the original decimal representation
// if required, and burn the requested coins from the given account.
func (w BankWrapper) BurnCoinsFromAccount(ctx context.Context, account sdk.AccAddress, amt sdk.Coins) error {
	coins, err := convertCoinsFrom18Decimals(amt)
	if err != nil {
		return err
	}

	// NOTE: amt is already converted so we need to use the x/bank method.
	if err := w.BankKeeper.SendCoinsFromAccountToModule(ctx, account, types.ModuleName, coins); err != nil {
		return err
	}
	return w.BurnCoins(ctx, types.ModuleName, amt)
}
