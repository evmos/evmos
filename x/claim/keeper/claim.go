package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/tharsis/evmos/x/claim/types"
)

// GetClaimable returns claimable amount for a specific action done by an address
func (k Keeper) GetClaimableAmountForAction(
	ctx sdk.Context,
	addr sdk.AccAddress,
	claimRecord types.ClaimRecord,
	action types.Action,
	params types.Params,
) (sdk.Coins, error) {
	// If we are before the start time, do nothing.
	// This case _shouldn't_ occur on chain, since the
	// start time ought to be chain start time.
	if !params.EnableClaim || ctx.BlockTime().Before(params.AirdropStartTime) {
		return sdk.Coins{}, nil
	}

	// return zero if there are no coins to claim
	if claimRecord.InitialClaimableAmount.IsZero() {
		return sdk.Coins{}, nil
	}

	// if action already completed, nothing is claimable
	if claimRecord.ActionsCompleted[action] {
		return sdk.Coins{}, nil
	}

	initialClaimablePerAction := sdk.Coins{}

	// TODO: update this and explicitely define the % instead of assuming each action
	// has the same weight
	for _, coin := range claimRecord.InitialClaimableAmount {
		initialClaimablePerAction = initialClaimablePerAction.Add(
			sdk.NewCoin(coin.Denom,
				coin.Amount.QuoRaw(int64(len(types.Action_name))),
			),
		)
	}

	elapsedAirdropTime := ctx.BlockTime().Sub(params.AirdropStartTime)

	// Are we early enough in the airdrop s.t. theres no decay?
	if elapsedAirdropTime <= params.DurationUntilDecay {
		return initialClaimablePerAction, nil
	}

	// The entire airdrop has completed
	if elapsedAirdropTime > params.DurationUntilDecay+params.DurationOfDecay {
		return sdk.Coins{}, nil
	}

	// Positive, since goneTime > params.DurationUntilDecay
	decayTime := elapsedAirdropTime - params.DurationUntilDecay
	decayPercent := sdk.NewDec(decayTime.Nanoseconds()).QuoInt64(params.DurationOfDecay.Nanoseconds())
	claimablePercent := sdk.OneDec().Sub(decayPercent)

	claimableCoins := sdk.Coins{}
	// TODO: define claimable percent
	for _, coin := range initialClaimablePerAction {
		claimableCoins = claimableCoins.Add(sdk.NewCoin(coin.Denom, coin.Amount.ToDec().Mul(claimablePercent).RoundInt()))
	}

	return claimableCoins, nil
}

// GetClaimable returns claimable amount for a specific action done by an address
func (k Keeper) GetUserTotalClaimable(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coins, error) {
	totalClaimable := sdk.Coins{}

	claimRecord, found := k.GetClaimRecord(ctx, addr)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrClaimRecordNotFound, addr.String())
	}

	params := k.GetParams(ctx)

	// FIXME: don't iterate over maps!
	for action := range types.Action_name {
		claimableForAction, err := k.GetClaimableAmountForAction(ctx, addr, claimRecord, types.Action(action), params)
		if err != nil {
			return nil, err
		}
		totalClaimable = totalClaimable.Add(claimableForAction...)
	}

	return totalClaimable, nil
}

// ClaimCoins remove claimable amount entry and transfer it to user's account
func (k Keeper) ClaimCoinsForAction(ctx sdk.Context, addr sdk.AccAddress, action types.Action) (sdk.Coins, error) {
	params := k.GetParams(ctx)

	claimRecord, found := k.GetClaimRecord(ctx, addr)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrClaimRecordNotFound, addr.String())
	}

	claimableAmount, err := k.GetClaimableAmountForAction(ctx, addr, claimRecord, action, params)
	if err != nil {
		return nil, err
	}

	if claimableAmount.Empty() {
		return claimableAmount, nil
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, claimableAmount)
	if err != nil {
		return nil, err
	}

	claimRecord.ActionCompleted[action] = true
	k.SetClaimRecord(ctx, addr, claimRecord)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(sdk.AttributeKeySender, addr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, claimableAmount.String()),
			sdk.NewAttribute(types.AttributeKeyActionType, action.String()),
		),
	})

	return claimableAmount, nil
}
