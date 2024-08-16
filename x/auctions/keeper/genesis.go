// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/auctions/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	err := k.SetParams(ctx, data.Params)
	if err != nil {
		panic(errorsmod.Wrap(err, "could not set parameters at genesis"))
	}

	var bidder sdk.AccAddress
	if data.Bid.Sender != "" {
		bidder, err = sdk.AccAddressFromBech32(data.Bid.Sender)
		if err != nil {
			panic(errorsmod.Wrap(err, "invalid bidder address"))
		}
		if found := k.accountKeeper.HasAccount(ctx, bidder); !found {
			panic(fmt.Errorf("account associated with %s does not exist", data.Bid.Sender))
		}

		bidAmount := data.Bid.BidValue.Amount
		if !bidAmount.IsPositive() {
			panic(fmt.Errorf("received a bid sender but zero amount"))
		}

		auctionModuleAddress := k.accountKeeper.GetModuleAddress(types.ModuleName)
		auctionModuleBalance := k.bankKeeper.GetBalance(ctx, auctionModuleAddress, utils.BaseDenom)

		if auctionModuleBalance.Amount.LT(bidAmount) {
			panic(fmt.Errorf("auction module account does not hold enough balance"))
		}

	} else if !data.Bid.BidValue.Amount.IsZero() {
		panic(fmt.Errorf("received a bid without sender but different than zero"))
	}

	k.SetHighestBid(ctx, data.Bid.Sender, data.Bid.BidValue)
	k.SetRound(ctx, data.Round)
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
		Bid:    k.GetHighestBid(ctx),
		Round:  k.GetRound(ctx),
	}
}
