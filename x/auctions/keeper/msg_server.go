// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/evmos/evmos/v19/x/auctions/types"
	"github.com/pkg/errors"
)

var _ types.MsgServer = &Keeper{}

// Bid defines a method for placing a bid on an auction
func (k Keeper) Bid(goCtx context.Context, bid *types.MsgBid) (*types.MsgBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.GetParams(ctx)
	if !params.EnableAuction {
		return nil, errorsmod.Wrapf(types.ErrAuctionDisabled, "auction is disabled")
	}

	lastBid := k.GetHighestBid(ctx)
	if bid.Amount.Amount.LTE(lastBid.BidValue.Amount) {
		return nil, errors.Wrapf(types.ErrBidMustBeHigherThanCurrent, "bid amount %s is less than or equal to the last bid %s", bid.Amount, lastBid.BidValue)
	}

	senderAddr, err := sdk.AccAddressFromBech32(bid.Sender)
	if err != nil {
		return nil, errors.Wrap(err, "invalid sender address")
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, sdk.NewCoins(bid.Amount)); err != nil {
		return nil, errors.Wrap(err, "transfer bid coins failed")
	}

	// If there is a last bid, refund it
	if !lastBid.BidValue.IsZero() {
		if err := k.refundLastBid(ctx); err != nil {
			return nil, errors.Wrap(err, "refund failed")
		}
	}
	k.SetHighestBid(ctx, bid.Sender, bid.Amount)

	return &types.MsgBidResponse{}, nil
}

// DepositCoin defines a method for depositing coins into the auction module
func (k Keeper) DepositCoin(goCtx context.Context, deposit *types.MsgDepositCoin) (*types.MsgDepositCoinResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := k.GetParams(ctx)
	if !params.EnableAuction {
		return nil, errorsmod.Wrapf(types.ErrAuctionDisabled, "auction is disabled")
	}

	senderAddr, err := sdk.AccAddressFromBech32(deposit.Sender)
	if err != nil {
		return nil, errors.Wrap(err, "invalid sender address")
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.AuctionCollectorName, sdk.NewCoins(deposit.Amount)); err != nil {
		return nil, errors.Wrap(err, "transfer of deposit failed")
	}

	return &types.MsgDepositCoinResponse{}, nil
}

// UpdateParams defines a method for updating auction params
func (k Keeper) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.authority.String() != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority.String(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, errorsmod.Wrapf(err, "error setting params")
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
