// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package auctions

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/x/auctions/keeper"
	"github.com/evmos/evmos/v19/x/auctions/types"
)

func InitGenesis(ctx sdk.Context, k keeper.Keeper, genesisState types.GenesisState) {
	err := k.SetParams(ctx, genesisState.Params)
	if err != nil {
		panic(errorsmod.Wrap(err, "could not set parameters at genesis"))
	}

	if genesisState.Bid.Sender != "" && genesisState.Bid.BidValue.IsPositive() {
		k.SetHighestBid(ctx, genesisState.Bid.Sender, genesisState.Bid.BidValue)
	}

	// Set the current round
	k.SetRound(ctx, genesisState.Round)
}

func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
		Bid:    k.GetHighestBid(ctx),
		Round:  k.GetRound(ctx),
	}
}
