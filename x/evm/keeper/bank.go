// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/x/evm/types"
)

// BankWrapper is a wrapper around the Cosmos SDK bank keeper
// that is used to manage an evm denom with 6 or 18 decimals.
// The wrapper makes the corresponding conversions to achieve:
//   - With the EVM, the wrapper works always with 18 decimals.
//   - With the Cosmos bank module, the wrapper works always
//     with the bank module decimals (either 6 or 18).
type BankWrapper struct {
	types.BankKeeper
	// decimals is the number of decimals used by the bank module
	decimals uint32
}

// NewBankWrapper creates a new bank Keeper wrapper instance.
// The BankWrapper is used to manage an evm denom with 6 or 18 decimals
func NewBankWrapper(
	bk types.BankKeeper,
) *BankWrapper {
	return &BankWrapper{
		bk,
		types.DefaultDenomDecimals,
	}
}

// WithDecimals function updates the decimals on the bank wrapper
// This function is useful when updating the evm params (denomDecimals)
func (w *BankWrapper) WithDecimals(decimals uint32) error {
	if decimals != types.Denom18Dec && decimals != types.Denom6Dec {
		return fmt.Errorf("decimals = %d not supported. Valid values are %d and %d", decimals, types.Denom18Dec, types.Denom6Dec)
	}
	w.decimals = decimals
	return nil
}

// GetBalance returns the balance converted to 18 decimals
func (w BankWrapper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	coin := w.BankKeeper.GetBalance(ctx, addr, denom)
	if w.decimals == types.Denom18Dec {
		return coin
	}
	return types.Convert6To18DecimalsCoin(coin)
}

// MintCoinsToAccount scales down from 18 decimals to 6 decimals the coins amount provided
// and mints that to the provided account
func (w BankWrapper) MintCoinsToAccount(ctx sdk.Context, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	convertAmtTo18Decimals(w.decimals, amt)
	if err := w.BankKeeper.MintCoins(ctx, types.ModuleName, amt); err != nil {
		return err
	}
	return w.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, amt)
}

// BurnAccountCoins scales down from 18 decimals to 6 decimals the coins amount provided
// and burns that coins of the provided account
func (w BankWrapper) BurnAccountCoins(ctx sdk.Context, account sdk.AccAddress, amt sdk.Coins) error {
	convertAmtTo18Decimals(w.decimals, amt)
	if err := w.BankKeeper.SendCoinsFromAccountToModule(ctx, account, types.ModuleName, amt); err != nil {
		return err
	}
	return w.BankKeeper.BurnCoins(ctx, types.ModuleName, amt)
}

// SendCoinsFromAccountToModule scales down
// from 18 decimals to 6 decimals the coins amount provided
// and sends the coins
func (w BankWrapper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	convertAmtTo18Decimals(w.decimals, amt)
	return w.BankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, amt)
}

// SendCoinsFromModuleToAccount scales down
// from 18 decimals to 6 decimals the coins amount provided
// and sends the coins from the module to the account
func (w BankWrapper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	convertAmtTo18Decimals(w.decimals, amt)
	return w.BankKeeper.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, amt)
}

// convertAmtTo18Decimals is a helper function to convert the amount of coins
func convertAmtTo18Decimals(decimals uint32, amt sdk.Coins) {
	if decimals == types.Denom6Dec {
		for i := range amt {
			amt[i] = types.Convert18To6DecimalsCoin(amt[i])
		}
	}
}
