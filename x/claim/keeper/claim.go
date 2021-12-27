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
) (sdk.Int, error) {
	if action == types.ActionUnspecified || action > types.ActionIBCTransfer {
		return sdk.ZeroInt(), sdkerrors.Wrapf(types.ErrInvalidAction, "%d", action)
	}

	// If we are before the start time, do nothing.
	// This case _shouldn't_ occur on chain, since the
	// start time ought to be chain start time.
	if !params.EnableClaim || ctx.BlockTime().Before(params.AirdropStartTime) {
		return sdk.ZeroInt(), nil
	}

	// return zero if there are no coins to claim
	if claimRecord.InitialClaimableAmount.IsZero() {
		return sdk.ZeroInt(), nil
	}

	// if action already completed, nothing is claimable

	// NOTE: slice size validated during genesis
	if claimRecord.ActionsCompleted[action-1] {
		return sdk.ZeroInt(), nil
	}

	// TODO: update this and explicitely define the % instead of assuming each action
	// has the same weight
	initialClaimablePerAction := claimRecord.InitialClaimableAmount.Add(
		claimRecord.InitialClaimableAmount.QuoRaw(int64(len(types.Action_name))),
	)

	elapsedAirdropTime := ctx.BlockTime().Sub(params.AirdropStartTime)

	// Are we early enough in the airdrop s.t. theres no decay?
	if elapsedAirdropTime <= params.DurationUntilDecay {
		return initialClaimablePerAction, nil
	}

	// The entire airdrop has completed
	if elapsedAirdropTime > params.DurationUntilDecay+params.DurationOfDecay {
		return sdk.ZeroInt(), nil
	}

	// Positive, since goneTime > params.DurationUntilDecay
	decayTime := elapsedAirdropTime - params.DurationUntilDecay
	decayPercent := sdk.NewDec(decayTime.Nanoseconds()).QuoInt64(params.DurationOfDecay.Nanoseconds())
	claimablePercent := sdk.OneDec().Sub(decayPercent)

	// TODO: define claimable percent per action
	claimableCoins := initialClaimablePerAction.ToDec().Mul(claimablePercent).RoundInt()

	return claimableCoins, nil
}

// GetClaimable returns claimable amount for a specific action done by an address
func (k Keeper) GetUserTotalClaimable(ctx sdk.Context, addr sdk.AccAddress) (sdk.Int, error) {
	totalClaimable := sdk.ZeroInt()

	claimRecord, found := k.GetClaimRecord(ctx, addr)
	if !found {
		return sdk.ZeroInt(), sdkerrors.Wrap(types.ErrClaimRecordNotFound, addr.String())
	}

	params := k.GetParams(ctx)

	actions := []types.Action{types.ActionVote, types.ActionDelegate, types.ActionEVM, types.ActionIBCTransfer}

	for _, action := range actions {
		claimableForAction, err := k.GetClaimableAmountForAction(ctx, addr, claimRecord, action, params)
		if err != nil {
			return sdk.ZeroInt(), err
		}

		totalClaimable = totalClaimable.Add(claimableForAction)
	}

	return totalClaimable, nil
}

// ClaimCoins remove claimable amount entry and transfer it to user's account
func (k Keeper) ClaimCoinsForAction(ctx sdk.Context, addr sdk.AccAddress, action types.Action) (sdk.Int, error) {
	if action == types.ActionUnspecified || action > types.ActionIBCTransfer {
		return sdk.ZeroInt(), sdkerrors.Wrapf(types.ErrInvalidAction, "%d", action)
	}

	params := k.GetParams(ctx)

	claimRecord, found := k.GetClaimRecord(ctx, addr)
	if !found {
		return sdk.ZeroInt(), sdkerrors.Wrap(types.ErrClaimRecordNotFound, addr.String())
	}

	claimableAmount, err := k.GetClaimableAmountForAction(ctx, addr, claimRecord, action, params)
	if err != nil {
		return sdk.ZeroInt(), err
	}

	if claimableAmount.IsZero() {
		return sdk.ZeroInt(), nil
	}

	claimedCoins := sdk.Coins{{Denom: params.ClaimDenom, Amount: claimableAmount}}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, claimedCoins); err != nil {
		return sdk.ZeroInt(), err
	}

	claimRecord.ActionsCompleted[action-1] = true
	k.SetClaimRecord(ctx, addr, claimRecord)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(sdk.AttributeKeySender, addr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, claimedCoins.String()),
			sdk.NewAttribute(types.AttributeKeyActionType, action.String()),
		),
	})

	return claimableAmount, nil
}
