// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/auctions/types"
)

// Keeper of the auction store
type Keeper struct {
	storeKey      storetypes.StoreKey
	cdc           codec.BinaryCodec
	authority     sdk.AccAddress
	bankKeeper    bankkeeper.Keeper
	accountKeeper authkeeper.AccountKeeper
}

// NewKeeper creates a new auction Keeper instance
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	authority sdk.AccAddress,
	bankKeeper bankkeeper.Keeper,
	accountKeeper authkeeper.AccountKeeper,
) Keeper {
	return Keeper{
		storeKey:      storeKey,
		cdc:           cdc,
		authority:     authority,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
	}
}

// refundLastBid refunds the last bid placed on an auction
func (k Keeper) refundLastBid(ctx sdk.Context) error {
	lastBid := k.getHighestBid(ctx)
	lastBidAmount := lastBid.Amount.Amount
	lastBidder, err := sdk.AccAddressFromBech32(lastBid.Sender)
	if err != nil {
		return err
	}

	bidAmount := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, lastBidAmount))
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lastBidder, bidAmount)
}
