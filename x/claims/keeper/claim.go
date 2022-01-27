package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/tharsis/evmos/x/claims/types"
)

// GetClaimableAmountForAction returns claimable amount for a specific action done by an address
func (k Keeper) GetClaimableAmountForAction(
	ctx sdk.Context,
	addr sdk.AccAddress,
	claimsRecord types.ClaimsRecord,
	action types.Action,
	params types.Params,
) sdk.Int {
	// return zero if there are no coins to claim
	if claimsRecord.InitialClaimableAmount.IsNil() || claimsRecord.InitialClaimableAmount.IsZero() {
		return sdk.ZeroInt()
	}

	airdropEndTime := params.AirdropEndTime()

	// Safety check: the entire airdrop has completed
	// NOTE: This shouldn't occur since at the end of the airdrop, the EnableClaims
	// param is disabled.
	if ctx.BlockTime().After(airdropEndTime) {
		return sdk.ZeroInt()
	}

	if claimsRecord.HasClaimedAction(action) {
		return sdk.ZeroInt()
	}

	// NOTE: use len(actions)-1 we don't consider the Unspecified Action
	initialClaimablePerAction := claimsRecord.InitialClaimableAmount.QuoRaw(int64(len(types.Action_name) - 1))

	decayStartTime := params.DecayStartTime()
	// claim amount is the full amount if the elapsed time is before or equal
	// the decay start time
	if !ctx.BlockTime().After(decayStartTime) {
		return initialClaimablePerAction
	}

	// calculate the claimable percent based on the elapsed time since the decay period started

	elapsedDecay := ctx.BlockTime().Sub(decayStartTime)

	elapsedDecayRatio := sdk.NewDec(elapsedDecay.Nanoseconds()).QuoInt64(params.DurationOfDecay.Nanoseconds())

	// claimable percent is (1 - elapsed decay) x 100

	// NOTE: the idea is that if you claim early in the decay period, you should
	// be entitled to more coins than if you claim at the end of it.
	claimableRatio := sdk.OneDec().Sub(elapsedDecayRatio)

	// calculate the claimable coins, while rounding the decimals
	claimableCoins := initialClaimablePerAction.ToDec().Mul(claimableRatio).RoundInt()
	return claimableCoins
}

// GetUserTotalClaimable returns claimable amount for a specific action done by an address
func (k Keeper) GetUserTotalClaimable(ctx sdk.Context, addr sdk.AccAddress) sdk.Int {
	totalClaimable := sdk.ZeroInt()

	claimsRecord, found := k.GetClaimsRecord(ctx, addr)
	if !found {
		return sdk.ZeroInt()
	}

	params := k.GetParams(ctx)

	actions := []types.Action{types.ActionVote, types.ActionDelegate, types.ActionEVM, types.ActionIBCTransfer}

	for _, action := range actions {
		claimableForAction := k.GetClaimableAmountForAction(ctx, addr, claimsRecord, action, params)
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
	if !params.EnableClaims || ctx.BlockTime().Before(params.AirdropStartTime) {
		return sdk.ZeroInt(), nil
	}

	claimsRecord, found := k.GetClaimsRecord(ctx, addr)
	if !found {
		// return nil if not claim record found to avoid panics
		return sdk.ZeroInt(), nil
	}

	// if action already completed, nothing is claimable
	if claimsRecord.HasClaimedAction(action) {
		return sdk.ZeroInt(), nil
	}

	claimableAmount := k.GetClaimableAmountForAction(ctx, addr, claimsRecord, action, params)

	if claimableAmount.IsZero() {
		return sdk.ZeroInt(), nil
	}

	claimedCoins := sdk.Coins{{Denom: params.ClaimsDenom, Amount: claimableAmount}}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, claimedCoins); err != nil {
		return sdk.ZeroInt(), err
	}

	claimsRecord.ClaimAction(action)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(sdk.AttributeKeySender, addr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, claimedCoins.String()),
			sdk.NewAttribute(types.AttributeKeyActionType, action.String()),
		),
	})

	if claimsRecord.HasClaimedAll() {
		k.DeleteClaimsRecord(ctx, addr)
	} else {
		k.SetClaimsRecord(ctx, addr, claimsRecord)
	}

	return claimableAmount, nil
}
