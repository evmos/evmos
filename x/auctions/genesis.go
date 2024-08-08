// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package auctions

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/x/auctions/keeper"
	"github.com/evmos/evmos/v19/x/auctions/types"
)

func InitGenesis(ctx sdk.Context, k keeper.Keeper, data types.GenesisState) {
	err := k.SetParams(ctx, data.Params)
	if err != nil {
		panic(errorsmod.Wrap(err, "could not set parameters at genesis"))
	}

	// Set the highest bid
	k.SetHighestBid(ctx, data.Bid.Sender, data.Bid.Amount)

	// Set the current round
	k.SetRound(ctx, data.Round)
}

func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
		Bid:    *k.GetHighestBid(ctx),
		Round:  k.GetRound(ctx),
	}
}
