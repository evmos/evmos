// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/auctions/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, genesisState types.GenesisState) {
	err := k.SetParams(ctx, genesisState.Params)
	if err != nil {
		panic(errorsmod.Wrap(err, "could not set parameters at genesis"))
	}

	var bidder sdk.AccAddress
	if genesisState.Bid.Sender != "" {
		bidder, err = sdk.AccAddressFromBech32(genesisState.Bid.Sender)
		if err != nil {
			panic(errorsmod.Wrap(err, "invalid bidder address"))
		}
		if found := k.accountKeeper.HasAccount(ctx, bidder); !found {
			panic(fmt.Errorf("account associated with %s does not exist", genesisState.Bid.Sender))
		}

		bidAmount := genesisState.Bid.BidValue.Amount
		if !bidAmount.IsPositive() {
			panic(errors.New("received a bid sender but zero amount"))
		}

		auctionModuleAddress := k.accountKeeper.GetModuleAddress(types.ModuleName)
		auctionModuleBalance := k.bankKeeper.GetBalance(ctx, auctionModuleAddress, utils.BaseDenom)

		if auctionModuleBalance.Amount.LT(bidAmount) {
			panic(errors.New("auction module account does not hold enough balance"))
		}

	} else if !genesisState.Bid.BidValue.Amount.IsZero() {
		panic(errors.New("received a bid without sender but amount is non-zero"))
	}

	k.SetHighestBid(ctx, genesisState.Bid.Sender, genesisState.Bid.BidValue)
	k.SetRound(ctx, genesisState.Round)
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
		Bid:    k.GetHighestBid(ctx),
		Round:  k.GetRound(ctx),
	}
}
