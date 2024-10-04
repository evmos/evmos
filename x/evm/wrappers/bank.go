// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package wrappers

import (
	"context"
	"math/big"

	"cosmossdk.io/collections"
	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/evmos/evmos/v20/x/evm/types"
)

// BankWrapper is a wrapper around the Cosmos SDK bank keeper
// that is used to manage an evm denom with a custom decimal representation.
type BankWrapper struct {
	bankkeeper.BaseKeeper
}

// NewBankWrapper creates a new BankWrapper instance.
func NewBankWrapper(
	bk bankkeeper.BaseKeeper,
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
func (w BankWrapper) MintAmountToAccount(ctx context.Context, recipientAddr sdk.AccAddress, amt *big.Int) error {
	coin := sdk.Coin{Denom: types.GetEVMCoinDenom(), Amount: sdkmath.NewIntFromBigInt(amt)}

	convertedCoin, err := types.ConvertEvmCoinFrom18Decimals(coin)
	if err != nil {
		return errors.Wrap(err, "failed to mint coin to account in bank wrapper")
	}

	coinsToMint := sdk.Coins{convertedCoin}
	if err := w.BaseKeeper.MintCoins(ctx, types.ModuleName, coinsToMint); err != nil {
		return errors.Wrap(err, "failed to mint coins to account in bank wrapper")
	}

	return w.BaseKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, coinsToMint)
}

// BurnAmountFromAccount converts the given amount into the evm coin scaling
// the amount to the original decimals, then burns that quantity from the provided account.
func (w BankWrapper) BurnAmountFromAccount(ctx context.Context, account sdk.AccAddress, amt *big.Int) error {
	coin := sdk.Coin{Denom: types.GetEVMCoinDenom(), Amount: sdkmath.NewIntFromBigInt(amt)}

	convertedCoin, err := types.ConvertEvmCoinFrom18Decimals(coin)
	if err != nil {
		return errors.Wrap(err, "failed to burn coins from account in bank wrapper")
	}

	coinsToBurn := sdk.Coins{convertedCoin}
	if err := w.BaseKeeper.SendCoinsFromAccountToModule(ctx, account, types.ModuleName, coinsToBurn); err != nil {
		return errors.Wrap(err, "failed to burn coins from account in bank wrapper")
	}
	return w.BaseKeeper.BurnCoins(ctx, types.ModuleName, coinsToBurn)
}

// ------------------------------------------------------------------------------------------
// Bank keeper shadowed methods
// ------------------------------------------------------------------------------------------

// GetBalance returns the balance of the given account converted to 18 decimals.
func (w BankWrapper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	// Get the balance from the BankModule. The balance returned is in the bank
	// decimals representation, which could be different than the 18 decimals
	// representation used in the evm.
	coin := w.BaseKeeper.GetBalance(ctx, addr, denom)

	return types.ConvertEvmCoinTo18Decimals(coin)
}

// SendCoinsFromAccountToModule wraps around the Cosmos SDK x/bank module's
// SendCoinsFromAccountToModule method to convert the evm coin, if present in
// the input, to its original representation.
func (w BankWrapper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, coins sdk.Coins) error {
	convertedCoins := types.ConvertCoinsFrom18Decimals(coins)

	return w.BaseKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, convertedCoins)
}

// SendCoinsFromModuleToAccount wraps around the Cosmos SDK x/bank module's
// SendCoinsFromModuleToAccount method to convert the evm coin, if present in
// the input, to its original representation.
func (w BankWrapper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, coins sdk.Coins) error {
	convertedCoins := types.ConvertCoinsFrom18Decimals(coins)

	return w.BaseKeeper.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, convertedCoins)
}

// IterateTotalSupply iterates over the total supply calling the given cb (callback) function
// with the balance of each coin.
// The iteration stops if the callback returns true.
// In case of the EVM coin, scales the value to 18 decimals if corresponds
func (k BankWrapper) IterateTotalSupply(ctx context.Context, cb func(sdk.Coin) bool) {
	err := k.BaseKeeper.Supply.Walk(ctx, nil, func(s string, m math.Int) (bool, error) {
		coin := sdk.NewCoin(s, m)
		if coin.Denom == types.GetEVMCoinDenom() {
			coin = types.ConvertEvmCoinTo18Decimals(coin)
		}
		return cb(coin), nil
	})
	if err != nil {
		panic(err)
	}
}

// IterateAccountBalances iterates over the balances of a single account and
// provides the token balance to a callback. If true is returned from the
// callback, iteration is halted.
// In case of the EVM coin, scales the value to 18 decimals if corresponds
func (k BankWrapper) IterateAccountBalances(ctx context.Context, addr sdk.AccAddress, cb func(sdk.Coin) bool) {
	err := k.BaseKeeper.Balances.Walk(ctx, collections.NewPrefixedPairRange[sdk.AccAddress, string](addr), func(key collections.Pair[sdk.AccAddress, string], value math.Int) (stop bool, err error) {
		coin := sdk.NewCoin(key.K2(), value)
		if coin.Denom == types.GetEVMCoinDenom() {
			coin = types.ConvertEvmCoinTo18Decimals(coin)
		}
		return cb(coin), nil
	})
	if err != nil {
		panic(err)
	}
}

// GetSupply retrieves the Supply from store
// In case of the EVM coin, scales the value to 18 decimals if corresponds
func (k BankWrapper) GetSupply(ctx context.Context, denom string) sdk.Coin {
	amt, err := k.BaseKeeper.Supply.Get(ctx, denom)
	if err != nil {
		return sdk.NewCoin(denom, math.ZeroInt())
	}
	coin := sdk.NewCoin(denom, amt)
	if coin.Denom == types.GetEVMCoinDenom() {
		coin = types.ConvertEvmCoinTo18Decimals(coin)
	}
	return coin
}
