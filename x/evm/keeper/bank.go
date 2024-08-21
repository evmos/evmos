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
	bk       types.BankKeeper
	decimals uint8
}

// NewBankWrapper creates a new bank Keeper wrapper instance.
// The BankWrapper is used to manage an evm denom with 6 or 18 decimals
func NewBankWrapper(
	bk types.BankKeeper,
	decimals uint8,
) *BankWrapper {
	if decimals != types.Denom18Dec && decimals != types.Denom6Dec {
		panic(fmt.Sprintf("decimals = %d not supported. Valid values are %d and %d", decimals, types.Denom18Dec, types.Denom6Dec))
	}
	return &BankWrapper{
		bk,
		decimals,
	}
}

// WithDecimals function updates the decimals on the bank wrapper
// This function is useful when updating the evm params (denomDecimals)
func (w *BankWrapper) WithDecimals(decimals uint8) {
	w.decimals = decimals
}

// IsSendEnabledCoins implements types.BankWrapper.
// This is not used. Is needed to fulfill the interface required for the
// deduct fee ante handler
func (w BankWrapper) IsSendEnabledCoins(sdk.Context, ...sdk.Coin) error {
	panic("unimplemented")
}

// SendCoins implements types.BankWrapper.
// This is not used. Is needed to fulfill the interface required for the
// deduct fee ante handler
func (w BankWrapper) SendCoins(sdk.Context, sdk.AccAddress, sdk.AccAddress, sdk.Coins) error {
	panic("unimplemented")
}

// GetBalance returns the balance converted to 18 decimals
func (w BankWrapper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	coin := w.bk.GetBalance(ctx, addr, denom)
	if w.decimals == types.Denom18Dec {
		return coin
	}
	return types.Convert6To18DecimalsCoin(coin)
}

// MintCoinsToAccount scales down from 18 decimals to 6 decimals the coins amount provided
// and mints that to the provided account
func (w BankWrapper) MintCoinsToAccount(ctx sdk.Context, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if w.decimals == types.Denom6Dec {
		for i := range amt {
			amt[i] = types.Convert18To6DecimalsCoin(amt[i])
		}
	}
	if err := w.bk.MintCoins(ctx, types.ModuleName, amt); err != nil {
		return err
	}
	return w.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, amt)
}

// BurnAccountCoins scales down from 18 decimals to 6 decimals the coins amount provided
// and burns that coins of the provided account
func (w BankWrapper) BurnAccountCoins(ctx sdk.Context, account sdk.AccAddress, amt sdk.Coins) error {
	if w.decimals == types.Denom6Dec {
		for i := range amt {
			amt[i] = types.Convert18To6DecimalsCoin(amt[i])
		}
	}
	if err := w.bk.SendCoinsFromAccountToModule(ctx, account, types.ModuleName, amt); err != nil {
		return err
	}
	return w.bk.BurnCoins(ctx, types.ModuleName, amt)
}

// SendCoinsFromAccountToModule scales down
// from 18 decimals to 6 decimals the coins amount provided
// and sends the coins
func (w BankWrapper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	if w.decimals == types.Denom6Dec {
		for i := range amt {
			amt[i] = types.Convert18To6DecimalsCoin(amt[i])
		}
	}
	return w.bk.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, amt)
}

// SendCoinsFromModuleToAccount scales down
// from 18 decimals to 6 decimals the coins amount provided
// and sends the coins from the module to the account
func (w BankWrapper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if w.decimals == types.Denom6Dec {
		for i := range amt {
			amt[i] = types.Convert18To6DecimalsCoin(amt[i])
		}
	}
	return w.bk.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, amt)
}
