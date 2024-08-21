package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/x/evm/types"
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

// GetBalance returns the balance converted to 18 decimals
func (w BankWrapper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	coin := w.BankKeeper.GetBalance(ctx, addr, denom)
	if w.decimals == types.Denom18Dec {
		return coin
	}
	return convert6To18DecimalsCoin(coin)
}

// MintCoinsToAccount scales down from 18 decimals to 6 decimals the coins amount provided
// and mints that to the provided account
func (w BankWrapper) MintCoinsToAccount(ctx sdk.Context, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if w.decimals == types.Denom6Dec {
		for i := range amt {
			amt[i] = convert18To6DecimalsCoin(amt[i])
		}
	}
	if err := w.BankKeeper.MintCoins(ctx, types.ModuleName, amt); err != nil {
		return err
	}
	return w.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, amt)
}

// BurnAccountCoins scales down from 18 decimals to 6 decimals the coins amount provided
// and burns that coins of the provided account
func (w BankWrapper) BurnAccountCoins(ctx sdk.Context, account sdk.AccAddress, amt sdk.Coins) error {
	if w.decimals == types.Denom6Dec {
		for i := range amt {
			amt[i] = convert18To6DecimalsCoin(amt[i])
		}
	}
	if err := w.BankKeeper.SendCoinsFromAccountToModule(ctx, account, types.ModuleName, amt); err != nil {
		return err
	}
	return w.BankKeeper.BurnCoins(ctx, types.ModuleName, amt)
}

// SendCoinsFromAccountToModule scales down
// from 18 decimals to 6 decimals the coins amount provided
// and sends the coins
func (w BankWrapper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	if w.decimals == types.Denom6Dec {
		for i := range amt {
			amt[i] = convert18To6DecimalsCoin(amt[i])
		}
	}
	return w.BankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, amt)
}

// SendCoinsFromModuleToAccount scales down
// from 18 decimals to 6 decimals the coins amount provided
// and sends the coins from the module to the account
func (w BankWrapper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if w.decimals == types.Denom6Dec {
		for i := range amt {
			amt[i] = convert18To6DecimalsCoin(amt[i])
		}
	}
	return w.BankKeeper.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, amt)
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
