// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/x/auctions/types"
)

func InitGenesis(ctx sdk.Context, k Keeper, data types.GenesisState) {
	err := k.SetParams(ctx, data.Params)
	if err != nil {
		panic(errorsmod.Wrap(err, "could not set parameters at genesis"))
	}

	// Bidder address should exists in the account keeper.
	bidder, err := sdk.AccAddressFromBech32(data.Bid.Sender)
	if err != nil {
		panic(errorsmod.Wrap(err, "bidder address is not valid"))
	}

	if found := k.accountKeeper.HasAccount(ctx, bidder); !found {
		panic(fmt.Errorf("account associated with %s does not exist", data.Bid.Sender))
	}

	// Set the highest bid
	if data.Bid.Sender != "" && data.Bid.Amount.IsPositive() {
		k.SetHighestBid(ctx, data.Bid.Sender, data.Bid.Amount)
	}

	// Set the current round
	k.SetRound(ctx, data.Round)
}

func ExportGenesis(ctx sdk.Context, k Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
		Bid:    *k.GetHighestBid(ctx),
		Round:  k.GetRound(ctx),
	}
}
