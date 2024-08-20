package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/x/evm/types"
)

var _ types.BankWrapper = BankWrapper{}

const (
	// denom18Dec specifies that the evm denomination has 18 decimals
	denom18Dec = 18
	// denom6Dec specifies that the evm denomination has 6 decimals
	denom6Dec = 6
)

// BankWrapper is a wrapper around the Cosmos SDK bank keeper
// that is used to manage an evm denom with 6 decimals.
// The wrapper makes the corresponding conversions to achieve:
// - With the EVM, the wrapper works always with 18 decimals.
// - With the Cosmos bank module, the wrapper works always with 6 decimals.
type BankWrapper struct {
	bk       types.BankKeeper
	decimals int8
}

// NewBankWrapper creates a new bank Keeper wrapper instance.
// The BankWrapper is used to manage an evm denom with 6 or 18 decimals
func NewBankWrapper(
	bk types.BankKeeper,
	decimals int8,
) *BankWrapper {
	if decimals != denom18Dec && decimals != denom6Dec {
		panic(fmt.Sprintf("decimals = %d not supported. Valid values are %d and %d", decimals, denom18Dec, denom6Dec))
	}
	return &BankWrapper{
		bk,
		decimals,
	}
}

// IsSendEnabledCoins implements types.BankWrapper.
// This is not used. Is needed to fulfill the interface required for the
// deduct fee ante handler
func (w BankWrapper) IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error {
	panic("unimplemented")
}

// SendCoins implements types.BankWrapper.
// This is not used. Is needed to fulfill the interface required for the
// deduct fee ante handler
func (w BankWrapper) SendCoins(ctx sdk.Context, from sdk.AccAddress, to sdk.AccAddress, amt sdk.Coins) error {
	panic("unimplemented")
}

// GetBalance returns the balance converted to 18 decimals
func (w BankWrapper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	coin := w.bk.GetBalance(ctx, addr, denom)
	if w.decimals == denom18Dec {
		return coin
	}
	return convert6To18DecimalsCoin(coin)
}

// MintCoinsToAccount scales down from 18 decimals to 6 decimals the coins amount provided
// and mints that to the provided account
func (w BankWrapper) MintCoinsToAccount(ctx sdk.Context, moduleName string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if w.decimals == denom6Dec {
		for i := range amt {
			amt[i] = convert18To6DecimalsCoin(amt[i])
		}
	}
	if err := w.bk.MintCoins(ctx, moduleName, amt); err != nil {
		return err
	}
	return w.bk.SendCoinsFromModuleToAccount(ctx, moduleName, recipientAddr, amt)
}

// BurnAccountCoins scales down from 18 decimals to 6 decimals the coins amount provided
// and burns that coins of the provided account
func (w BankWrapper) BurnAccountCoins(ctx sdk.Context, account sdk.AccAddress, burningModule string, amt sdk.Coins) error {
	if w.decimals == denom6Dec {
		for i := range amt {
			amt[i] = convert18To6DecimalsCoin(amt[i])
		}
	}
	if err := w.bk.SendCoinsFromAccountToModule(ctx, account, burningModule, amt); err != nil {
		return err
	}
	return w.bk.BurnCoins(ctx, burningModule, amt)
}

// SendCoinsFromAccountToModule scales down
// from 18 decimals to 6 decimals the coins amount provided
// and sends the coins
func (w BankWrapper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	if w.decimals == denom6Dec {
		for i := range amt {
			amt[i] = convert18To6DecimalsCoin(amt[i])
		}
	}
	return w.bk.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, amt)
}

// SendCoinsFromModuleToAccount scales down
// from 18 decimals to 6 decimals the coins amount provided
// and sends the coins from the module to the account
func (w BankWrapper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if w.decimals == denom6Dec {
		for i := range amt {
			amt[i] = convert18To6DecimalsCoin(amt[i])
		}
	}
	return w.bk.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, amt)
}

// convert6To18DecimalsCoin converts the coin amount to 18 decimals from 6
func convert6To18DecimalsCoin(coin sdk.Coin) sdk.Coin {
	coin.Amount = coin.Amount.MulRaw(1e12)
	return coin
}

// convert18To6DecimalsCoin converts the coin amount to 6 decimals from 18
func convert18To6DecimalsCoin(coin sdk.Coin) sdk.Coin {
	coin.Amount = coin.Amount.QuoRaw(1e12)
	return coin
}
