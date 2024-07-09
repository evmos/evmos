// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"context"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/evmos/evmos/v18/utils"
	"github.com/evmos/evmos/v18/x/auctions/types"
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
	if bid.Amount.Amount.LT(lastBid.Amount.Amount) {
		return nil, errors.Wrapf(types.ErrBidMustBeHigherThanCurrent, "bid amount %s is less than last bid %s", bid.Amount, lastBid.Amount)
	}

	senderAddr, err := sdk.AccAddressFromBech32(bid.Sender)
	if err != nil {
		return nil, errors.Wrap(err, "invalid sender address")
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, sdk.NewCoins(bid.Amount)); err != nil {
		return nil, errors.Wrap(err, "deposit failed")
	}

	if err := k.refundLastBid(ctx); err != nil {
		return nil, errors.Wrap(err, "refund failed")
	}

	k.SetHighestBid(ctx, bid.Sender, bid.Amount)

	// TODO: emit events

	return &types.MsgBidResponse{}, nil
}

// UpdateParams defines a method for updating inflation params
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

// refundLastBid refunds the last bid placed on an auction
func (k Keeper) refundLastBid(ctx sdk.Context) error {

	previousBid := k.GetHighestBid(ctx)
	previousBidAmount := previousBid.Amount.Amount
	lastBidder, err := sdk.AccAddressFromBech32(previousBid.Sender)
	if err != nil {
		return err
	}

	bidAmount := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, previousBidAmount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lastBidder, bidAmount); err != nil {
		return err
	}

	return nil
}
