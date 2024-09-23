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
// that is used to manage an evm denom with 6 or 18 decimals.
// The wrapper makes the corresponding conversions to achieve:
//   - With the EVM, the wrapper works always with 18 decimals.
//   - With the Cosmos bank module, the wrapper works always
//     with the bank module decimals (either 6 or 18).
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

// GetBalance returns the balance converted to 18 decimals
func (w BankWrapper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	coin := w.BankKeeper.GetBalance(ctx, addr, denom)

	return convertTo18DecimalsCoin(coin)
}

// SendCoinsFromAccountToModule scales down
// from 18 decimals to 6 decimals the coins amount provided
// and sends the coins
func (w BankWrapper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	//  NOTE: are  we sure we can handle all coins in amt? We know the decimals
	//  only for the evm coin registered.
	for i := range amt {
		amt[i] = convertFrom18DecimalsCoin(amt[i])
	}
	return w.BankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, amt)
}

// SendCoinsFromModuleToAccount scales down
// from 18 decimals to 6 decimals the coins amount provided
// and sends the coins from the module to the account
func (w BankWrapper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	//  NOTE: are  we sure we can handle all coins in amt? We know the decimals
	//  only for the evm coin registered.
	for i := range amt {
		amt[i] = convertFrom18DecimalsCoin(amt[i])
	}
	return w.BankKeeper.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, amt)
}

// MintCoinsToAccount check if the provided amount has to be converted from 18
// decimals to the original coin representation and mints that amount to to the
// provided account.
func (w BankWrapper) MintCoinsToAccount(ctx context.Context, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	//  NOTE: are  we sure we can handle all coins in amt? We know the decimals
	//  only for the evm coin registered.
	for i := range amt {
		amt[i] = convertFrom18DecimalsCoin(amt[i])
	}

	if err := w.MintCoins(ctx, types.ModuleName, amt); err != nil {
		return err
	}
	return w.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, amt)
}

// BurnAccountCoins scales down from 18 decimals to 6 decimals the coins amount provided
// and burns that coins of the provided account
func (w BankWrapper) BurnAccountCoins(ctx context.Context, account sdk.AccAddress, amt sdk.Coins) error {
	//  NOTE: are  we sure we can handle all coins in amt? We know the decimals
	//  only for the evm coin registered.
	for i := range amt {
		amt[i] = convertFrom18DecimalsCoin(amt[i])
	}
	if err := w.BankKeeper.SendCoinsFromAccountToModule(ctx, account, types.ModuleName, amt); err != nil {
		return err
	}
	return w.BurnCoins(ctx, types.ModuleName, amt)
}
