// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testutil

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	inflationtypes "github.com/evmos/evmos/v20/x/inflation/v1/types"
)

// FundAccount is a utility function that funds an account by minting and
// sending the coins to the address.
func FundAccount(ctx sdk.Context, bankKeeper bankkeeper.Keeper, addr sdk.AccAddress, amounts sdk.Coins) error {
	if err := bankKeeper.MintCoins(ctx, inflationtypes.ModuleName, amounts); err != nil {
		return err
	}

	return bankKeeper.SendCoinsFromModuleToAccount(ctx, inflationtypes.ModuleName, addr, amounts)
}

// FundAccountWithBaseDenom is a utility function that uses the FundAccount function
// to fund an account with the default Evmos denomination.
func FundAccountWithBaseDenom(ctx sdk.Context, bankKeeper bankkeeper.Keeper, addr sdk.AccAddress, amount int64) error {
	baseDenom, err := sdk.GetBaseDenom()
	if err != nil {
		return err
	}
	coins := sdk.NewCoins(
		sdk.NewCoin(baseDenom, math.NewInt(amount)),
	)
	return FundAccount(ctx, bankKeeper, addr, coins)
}

// FundModuleAccount is a utility function that funds a module account by
// minting and sending the coins to the address.
func FundModuleAccount(ctx sdk.Context, bankKeeper bankkeeper.Keeper, recipientMod string, amounts sdk.Coins) error {
	if err := bankKeeper.MintCoins(ctx, inflationtypes.ModuleName, amounts); err != nil {
		return err
	}

	return bankKeeper.SendCoinsFromModuleToModule(ctx, inflationtypes.ModuleName, recipientMod, amounts)
}
