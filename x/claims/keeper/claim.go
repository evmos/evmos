// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/evmos/evmos/v12/x/claims/types"
)

// ClaimCoinsForAction removes the claimable amount entry from a claims record
// and transfers it to the user's account
func (k Keeper) ClaimCoinsForAction(
	ctx sdk.Context,
	addr sdk.AccAddress,
	claimsRecord types.ClaimsRecord,
	action types.Action,
	params types.Params,
) (math.Int, error) {
	if action == types.ActionUnspecified || action > types.ActionIBCTransfer {
		return sdk.ZeroInt(), errorsmod.Wrapf(types.ErrInvalidAction, "%d", action)
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

	k.SetClaimsRecord(ctx, addr, claimsRecord)

	k.Logger(ctx).Info(
		"claimed action",
		"address", addr.String(),
		"action", action.String(),
	)

	return claimableAmount, nil
}

// MergeClaimsRecords merges two independent claims records (sender and
// recipient) into a new instance by summing up the initial claimable amounts
// from both records.

// This method additionally:
//   - Always claims the IBC action, assuming both record haven't claimed it.
//   - Marks an action as claimed for the new instance by performing an XOR operation between the 2 provided records: `merged completed action = sender completed action XOR recipient completed action`
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
	actions := []types.Action{types.ActionVote, types.ActionDelegate, types.ActionEVM, types.ActionIBCTransfer}
	for _, action := range actions {

		// Safety check: the sender record cannot have any claimed actions, as
		//  - the sender is not an evmos address and can't claim vote, delegation or evm actions
		//  - the first attempt to perform an ibc callback from the senders account will merge/migrate the entire claims record
		if senderClaimsRecord.HasClaimedAction(action) {
			return types.ClaimsRecord{}, errorsmod.Wrapf(errortypes.ErrNotSupported, "non-evmos sender must not have claimed action: %v", action)
		}

		recipientCompleted := recipientClaimsRecord.HasClaimedAction(action)

		if recipientCompleted {
			// claim action for sender since the recipient completed it
			amt, remainder := k.GetClaimableAmountForAction(ctx, senderClaimsRecord, action, params)
			claimedAmt = claimedAmt.Add(amt)
			remainderAmt = remainderAmt.Add(remainder)
			mergedRecord.MarkClaimed(action)
		} else {
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
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, claimedCoins); err != nil {
		return types.ClaimsRecord{}, err
	}

	remainderCoins := sdk.Coins{sdk.Coin{Denom: params.ClaimsDenom, Amount: remainderAmt}}
	// short-circuit: don't fund community pool if remainder is 0
	if remainderCoins.IsZero() {
		return mergedRecord, nil
	}

	escrowAddr := k.GetModuleAccountAddress()
	if err := k.distrKeeper.FundCommunityPool(ctx, remainderCoins, escrowAddr); err != nil {
		return types.ClaimsRecord{}, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeMergeClaimsRecords,
			sdk.NewAttribute(types.AttributeKeyRecipient, recipient.String()),
			sdk.NewAttribute(types.AttributeKeyClaimedCoins, claimedCoins.String()),
			sdk.NewAttribute(types.AttributeKeyFundCommunityPoolCoins, remainderCoins.String()),
		),
	})

	return mergedRecord, nil
}

// GetClaimableAmountForAction returns claimable amount for a specific action
// done by an address
// returns zero if airdrop didn't start, isn't enabled or has finished
func (k Keeper) GetClaimableAmountForAction(
	ctx sdk.Context,
	claimsRecord types.ClaimsRecord,
	action types.Action,
	params types.Params,
) (claimableCoins, remainder math.Int) {
	// check if the entire airdrop has completed. This shouldn't occur since at
	// the end of the airdrop, the EnableClaims param is disabled.
	if !params.IsClaimsActive(ctx.BlockTime()) {
		return sdk.ZeroInt(), sdk.ZeroInt()
	}

	return k.ClaimableAmountForAction(ctx, claimsRecord, action, params)
}

// ClaimableAmountForAction returns claimable amount for a specific action
// done by an address
func (k Keeper) ClaimableAmountForAction(
	ctx sdk.Context,
	claimsRecord types.ClaimsRecord,
	action types.Action,
	params types.Params,
) (claimableCoins, remainder math.Int) {
	// return zero if there are no coins to claim
	if claimsRecord.InitialClaimableAmount.IsNil() || claimsRecord.InitialClaimableAmount.IsZero() {
		return sdk.ZeroInt(), sdk.ZeroInt()
	}

	// check if action already completed
	if claimsRecord.HasClaimedAction(action) {
		return sdk.ZeroInt(), sdk.ZeroInt()
	}

	// NOTE: use len(actions)-1 as we don't consider the Unspecified Action
	actionsCount := int64(len(types.Action_name) - 1)
	initialClaimablePerAction := claimsRecord.InitialClaimableAmount.QuoRaw(actionsCount)

	// return full claim amount if the elapsed time <= decay start time
	decayStartTime := params.DecayStartTime()
	if !ctx.BlockTime().After(decayStartTime) {
		return initialClaimablePerAction, sdk.ZeroInt()
	}

	// Decrease claimable amount if elapsed time > decay start time.
	// The decrease is calculated proportionally to how much elapsedDeacay period
	// has passed. If you claim early in the decay period, you are entitled to
	// more coins than if you claim at the end of it.
	//
	// Claimable percent = (1 - elapsed decay) x 100
	elapsedDecay := ctx.BlockTime().Sub(decayStartTime)
	elapsedDecayRatio := sdk.NewDec(elapsedDecay.Nanoseconds()).QuoInt64(params.DurationOfDecay.Nanoseconds())
	claimableRatio := sdk.OneDec().Sub(elapsedDecayRatio)

	// calculate the claimable coins, while rounding the decimals
	claimableCoins = claimableRatio.MulInt(initialClaimablePerAction).RoundInt()
	remainder = initialClaimablePerAction.Sub(claimableCoins)
	return claimableCoins, remainder
}
