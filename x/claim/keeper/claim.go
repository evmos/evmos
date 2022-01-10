package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/tharsis/evmos/x/claim/types"
)

// GetClaimableAmountForAction returns claimable amount for a specific action done by an address
func (k Keeper) GetClaimableAmountForAction(
	ctx sdk.Context,
	addr sdk.AccAddress,
	claimRecord types.ClaimRecord,
	action types.Action,
	params types.Params,
) sdk.Int {
	// return zero if there are no coins to claim
	if claimRecord.InitialClaimableAmount.IsZero() {
		return sdk.ZeroInt()
	}

	elapsedAirdropTime := ctx.BlockTime().Sub(params.AirdropStartTime)

	// Safety check: the entire airdrop has completed
	// NOTE: This shouldn't occur since at the end of the airdrop, the EnableClaim
	// param is disabled.
	if elapsedAirdropTime > params.DurationUntilDecay+params.DurationOfDecay {
		return sdk.ZeroInt()
	}

	// TODO: update this and explicitly define the % instead of assuming each action
	// has the same weight

	// NOTE: use len(actions)-1 we don't consider the Unspecified Action
	initialClaimablePerAction := claimRecord.InitialClaimableAmount.QuoRaw(int64(len(types.Action_name) - 1))

	// Are we early enough in the airdrop s.t. theres no decay?
	if elapsedAirdropTime <= params.DurationUntilDecay {
		return initialClaimablePerAction
	}

	// Positive, since goneTime > params.DurationUntilDecay
	decayTime := elapsedAirdropTime - params.DurationUntilDecay
	decayPercent := sdk.NewDec(decayTime.Nanoseconds()).QuoInt64(params.DurationOfDecay.Nanoseconds())
	claimablePercent := sdk.OneDec().Sub(decayPercent)

	// TODO: define claimable percent per action
	claimableCoins := initialClaimablePerAction.ToDec().Mul(claimablePercent).RoundInt()
	return claimableCoins
}

// GetUserTotalClaimable returns claimable amount for a specific action done by an address
func (k Keeper) GetUserTotalClaimable(ctx sdk.Context, addr sdk.AccAddress) sdk.Int {
	totalClaimable := sdk.ZeroInt()

	claimRecord, found := k.GetClaimRecord(ctx, addr)
	if !found {
		return sdk.ZeroInt()
	}

	params := k.GetParams(ctx)

	actions := []types.Action{types.ActionVote, types.ActionDelegate, types.ActionEVM, types.ActionIBCTransfer}

	for _, action := range actions {
		claimableForAction := k.GetClaimableAmountForAction(ctx, addr, claimRecord, action, params)
		totalClaimable = totalClaimable.Add(claimableForAction)
	}

	return totalClaimable
}

// ClaimCoinsForAction remove claimable amount entry and transfer it to user's account
func (k Keeper) ClaimCoinsForAction(ctx sdk.Context, addr sdk.AccAddress, action types.Action) (sdk.Int, error) {
	if action == types.ActionUnspecified || action > types.ActionIBCTransfer {
		return sdk.ZeroInt(), sdkerrors.Wrapf(types.ErrInvalidAction, "%d", action)
	}

	params := k.GetParams(ctx)

	// If we are before the start time or claims are disabled, do nothing.
	if !params.EnableClaim || ctx.BlockTime().Before(params.AirdropStartTime) {
		return sdk.ZeroInt(), nil
	}

	claimRecord, found := k.GetClaimRecord(ctx, addr)
	if !found {
		// return nil if not claim record found to avoid panics
		return sdk.ZeroInt(), nil
	}

	// if action already completed, nothing is claimable
	if claimRecord.HasClaimedAction(action) {
		return sdk.ZeroInt(), nil
	}

	claimableAmount := k.GetClaimableAmountForAction(ctx, addr, claimRecord, action, params)

	if claimableAmount.IsZero() {
		return sdk.ZeroInt(), nil
	}

	claimedCoins := sdk.Coins{{Denom: params.ClaimDenom, Amount: claimableAmount}}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, claimedCoins); err != nil {
		return sdk.ZeroInt(), err
	}

	claimRecord.ClaimAction(action)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(sdk.AttributeKeySender, addr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, claimedCoins.String()),
			sdk.NewAttribute(types.AttributeKeyActionType, action.String()),
		),
	})

	if claimRecord.HasClaimedAll() {
		k.DeleteClaimRecord(ctx, addr)
	} else {
		k.SetClaimRecord(ctx, addr, claimRecord)
	}

	return claimableAmount, nil
}
