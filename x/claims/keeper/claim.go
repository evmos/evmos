package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/tharsis/evmos/v2/x/claims/types"
)

var actions = []types.Action{types.ActionVote, types.ActionDelegate, types.ActionEVM, types.ActionIBCTransfer}

// GetClaimableAmountForAction returns claimable amount for a specific action done by an address
// and the remainder amount to be claimed by the community pool
func (k Keeper) GetClaimableAmountForAction(
	ctx sdk.Context,
	claimsRecord types.ClaimsRecord,
	action types.Action,
	params types.Params,
) (claimableCoins, remainder sdk.Int) {
	// return zero if there are no coins to claim
	if claimsRecord.InitialClaimableAmount.IsNil() || claimsRecord.InitialClaimableAmount.IsZero() {
		return sdk.ZeroInt(), sdk.ZeroInt()
	}

	// Safety check: the entire airdrop has completed
	// NOTE: This shouldn't occur since at the end of the airdrop, the EnableClaims
	// param is disabled.
	if !params.IsClaimsActive(ctx.BlockTime()) {
		return sdk.ZeroInt(), sdk.ZeroInt()
	}

	if claimsRecord.HasClaimedAction(action) {
		return sdk.ZeroInt(), sdk.ZeroInt()
	}

	// NOTE: use len(actions)-1 we don't consider the Unspecified Action
	initialClaimablePerAction := claimsRecord.InitialClaimableAmount.QuoRaw(int64(len(types.Action_name) - 1))

	decayStartTime := params.DecayStartTime()
	// claim amount is the full amount if the elapsed time is before or equal
	// the decay start time
	if !ctx.BlockTime().After(decayStartTime) {
		return initialClaimablePerAction, sdk.ZeroInt()
	}

	// calculate the claimable percent based on the elapsed time since the decay period started

	elapsedDecay := ctx.BlockTime().Sub(decayStartTime)

	elapsedDecayRatio := sdk.NewDec(elapsedDecay.Nanoseconds()).QuoInt64(params.DurationOfDecay.Nanoseconds())

	// claimable percent is (1 - elapsed decay) x 100

	// NOTE: the idea is that if you claim early in the decay period, you should
	// be entitled to more coins than if you claim at the end of it.
	claimableRatio := sdk.OneDec().Sub(elapsedDecayRatio)

	// calculate the claimable coins, while rounding the decimals
	claimableCoins = initialClaimablePerAction.ToDec().Mul(claimableRatio).RoundInt()
	remainder = initialClaimablePerAction.Sub(claimableCoins)
	return claimableCoins, remainder
}

// GetUserTotalClaimable returns claimable amount for a specific action done by an address
// at a given block time
func (k Keeper) GetUserTotalClaimable(ctx sdk.Context, addr sdk.AccAddress) sdk.Int {
	totalClaimable := sdk.ZeroInt()

	claimsRecord, found := k.GetClaimsRecord(ctx, addr)
	if !found {
		return sdk.ZeroInt()
	}

	params := k.GetParams(ctx)

	actions := []types.Action{types.ActionVote, types.ActionDelegate, types.ActionEVM, types.ActionIBCTransfer}

	for _, action := range actions {
		claimableForAction, _ := k.GetClaimableAmountForAction(ctx, claimsRecord, action, params)
		totalClaimable = totalClaimable.Add(claimableForAction)
	}

	return totalClaimable
}

// ClaimCoinsForAction remove claimable amount entry and transfer it to user's account
func (k Keeper) ClaimCoinsForAction(
	ctx sdk.Context,
	addr sdk.AccAddress,
	claimsRecord types.ClaimsRecord,
	action types.Action,
	params types.Params,
) (sdk.Int, error) {
	if action == types.ActionUnspecified || action > types.ActionIBCTransfer {
		return sdk.ZeroInt(), sdkerrors.Wrapf(types.ErrInvalidAction, "%d", action)
	}

	// If we are before the start time, after end time, or claims are disabled, do nothing.
	if !params.IsClaimsActive(ctx.BlockTime()) {
		return sdk.ZeroInt(), nil
	}

	// if action already completed, nothing is claimable
	if claimsRecord.HasClaimedAction(action) {
		return sdk.ZeroInt(), nil
	}

	claimableAmount, remainderAmount := k.GetClaimableAmountForAction(ctx, claimsRecord, action, params)

	if claimableAmount.IsZero() {
		return sdk.ZeroInt(), nil
	}

	claimedCoins := sdk.Coins{sdk.Coin{Denom: params.ClaimsDenom, Amount: claimableAmount}}
	remainderCoins := sdk.Coins{sdk.Coin{Denom: params.ClaimsDenom, Amount: remainderAmount}}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, claimedCoins); err != nil {
		return sdk.ZeroInt(), err
	}

	// fund community pool if remainder is not 0
	if !remainderAmount.IsZero() {
		escrowAddr := k.GetModuleAccountAddress()

		if err := k.distrKeeper.FundCommunityPool(ctx, remainderCoins, escrowAddr); err != nil {
			return sdk.ZeroInt(), err
		}
	}

	claimsRecord.MarkClaimed(action)

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

	k.Logger(ctx).Info(
		"claimed action",
		"address", addr.String(),
		"action", action.String(),
	)

	return claimableAmount, nil
}

// MergeClaimsRecords merges two independent claim records (sender and recipient)
// into a new instance by adding the initial the initial claimable amount from
// both records. This method additionally:
// - Always claims the IBC action, assuming both record haven't claimed it.
// - Marks an action as claimed for the new instance by performing an XOR operation between
// the 2 provided records:
// `merged completed action = sender completed action XOR recipient completed action`
func (k Keeper) MergeClaimsRecords(
	ctx sdk.Context,
	recipient sdk.AccAddress,
	senderClaimsRecord,
	recipientClaimsRecord types.ClaimsRecord,
	params types.Params,
) (mergedRecord types.ClaimsRecord, err error) {
	claimedAmt := sdk.ZeroInt()
	remainderAmt := sdk.ZeroInt()

	// new total is the sum of the sender and recipient claims records amounts
	totalClaimableAmt := senderClaimsRecord.InitialClaimableAmount.Add(recipientClaimsRecord.InitialClaimableAmount)
	mergedRecord = types.NewClaimsRecord(totalClaimableAmt)

	// iterate over all the available actions and claim the amount if
	// the recipient or sender has completed an action but the other hasn't

	for _, action := range actions {
		senderCompleted := senderClaimsRecord.HasClaimedAction(action)
		recipientCompleted := recipientClaimsRecord.HasClaimedAction(action)

		switch {
		case senderCompleted && recipientCompleted:
			// Both sender and recipient completed the action.
			// Only mark the action as completed
			mergedRecord.MarkClaimed(action)
		case recipientCompleted && !senderCompleted:
			// claim action for sender since the recipient completed it
			amt, remainder := k.GetClaimableAmountForAction(ctx, senderClaimsRecord, action, params)
			claimedAmt = claimedAmt.Add(amt)
			remainderAmt = remainderAmt.Add(remainder)
			mergedRecord.MarkClaimed(action)
		case !recipientCompleted && senderCompleted:
			// claim action for recipient since the sender completed it
			amt, remainder := k.GetClaimableAmountForAction(ctx, recipientClaimsRecord, action, params)
			claimedAmt = claimedAmt.Add(amt)
			remainderAmt = remainderAmt.Add(remainder)
			mergedRecord.MarkClaimed(action)
		case !senderCompleted && !recipientCompleted:
			// Neither sender or recipient completed the action.
			if action != types.ActionIBCTransfer {
				// No-op if the action is not IBC transfer
				continue
			}

			// claim IBC action for both sender and recipient
			amtIBCRecipient, remainderRecipient := k.GetClaimableAmountForAction(ctx, recipientClaimsRecord, action, params)
			amtIBCSender, remainderSender := k.GetClaimableAmountForAction(ctx, senderClaimsRecord, action, params)
			claimedAmt = claimedAmt.Add(amtIBCRecipient).Add(amtIBCSender)
			remainderAmt = remainderAmt.Add(remainderRecipient).Add(remainderSender)
			mergedRecord.MarkClaimed(action)
		}
	}

	// safety check to prevent error while sending coins from the module escrow balance to the recipient
	if claimedAmt.IsZero() {
		return mergedRecord, nil
	}

	claimedCoins := sdk.Coins{sdk.Coin{Denom: params.ClaimsDenom, Amount: claimedAmt}}
	remainderCoins := sdk.Coins{sdk.Coin{Denom: params.ClaimsDenom, Amount: remainderAmt}}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, claimedCoins); err != nil {
		return types.ClaimsRecord{}, err
	}

	// short-circuit: don't fund community pool if remainder is 0
	if remainderCoins.IsZero() {
		return mergedRecord, nil
	}

	escrowAddr := k.GetModuleAccountAddress()

	if err := k.distrKeeper.FundCommunityPool(ctx, remainderCoins, escrowAddr); err != nil {
		return types.ClaimsRecord{}, err
	}

	return mergedRecord, nil
}
