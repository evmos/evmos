package keeper

import (
	"context"
	"math"
	"time"

	"github.com/armon/go-metrics"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/tharsis/evmos/x/vesting/types"
)

var _ types.MsgServer = &Keeper{}

// CreateClawbackVestingAccount creates a new ClawbackVestingAccount, or merges
// a grant into an existing one.
func (k Keeper) CreateClawbackVestingAccount(
	goCtx context.Context,
	msg *types.MsgCreateClawbackVestingAccount,
) (*types.MsgCreateClawbackVestingAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ak := k.accountKeeper
	bk := k.bankKeeper

	// Error checked during msg validation
	from, _ := sdk.AccAddressFromBech32(msg.FromAddress)
	to, _ := sdk.AccAddressFromBech32(msg.ToAddress)

	if bk.BlockedAddr(to) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized,
			"%s is not allowed to receive funds", msg.ToAddress,
		)
	}

	vestingCoins := sdk.NewCoins()
	for _, period := range msg.VestingPeriods {
		vestingCoins = vestingCoins.Add(period.Amount...)
	}

	lockupCoins := sdk.NewCoins()
	for _, period := range msg.LockupPeriods {
		lockupCoins = lockupCoins.Add(period.Amount...)
	}

	// If lockup absent, default to an instant unlock schedule
	if !vestingCoins.IsZero() && len(msg.LockupPeriods) == 0 {
		msg.LockupPeriods = []sdkvesting.Period{
			{Length: 0, Amount: vestingCoins},
		}
		lockupCoins = vestingCoins
	}

	// If vesting absent, default to an instant vesting schedule
	if !lockupCoins.IsZero() && len(msg.VestingPeriods) == 0 {
		msg.VestingPeriods = []sdkvesting.Period{
			{Length: 0, Amount: lockupCoins},
		}
		vestingCoins = lockupCoins
	}

	// The vesting and lockup schedules must describe the same total amount.
	// IsEqual can panic, so use (a == b) <=> (a <= b && b <= a).
	if !(vestingCoins.IsAllLTE(lockupCoins) && lockupCoins.IsAllLTE(vestingCoins)) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest,
			"lockup and vesting amounts must be equal",
		)
	}

	// Add Grant if vesting account exists, "merge" is true and funder is correct.
	// Otherwise create a new Clawback Vesting Account
	madeNewAcc := false
	acc := ak.GetAccount(ctx, to)
	var vestingAcc *types.ClawbackVestingAccount

	if acc != nil {
		var isClawback bool
		vestingAcc, isClawback = acc.(*types.ClawbackVestingAccount)
		switch {
		case !msg.Merge && isClawback:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s already exists; consider using --merge", msg.ToAddress)
		case !msg.Merge && !isClawback:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s already exists", msg.ToAddress)
		case msg.Merge && !isClawback:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrNotSupported, "account %s must be a clawback vesting account", msg.ToAddress)
		case msg.FromAddress != vestingAcc.FunderAddress:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s can only accept grants from account %s", msg.ToAddress, vestingAcc.FunderAddress)
		}
		k.addGrant(ctx, vestingAcc, msg.GetStartTime().Unix(), msg.GetLockupPeriods(), msg.GetVestingPeriods(), vestingCoins)
		ak.SetAccount(ctx, vestingAcc)

	} else {
		baseAcc := authtypes.NewBaseAccountWithAddress(to)
		vestingAcc = types.NewClawbackVestingAccount(
			baseAcc,
			from,
			vestingCoins,
			msg.StartTime,
			msg.LockupPeriods,
			msg.VestingPeriods,
		)
		acc := ak.NewAccount(ctx, vestingAcc)
		ak.SetAccount(ctx, acc)
		madeNewAcc = true
	}

	if madeNewAcc {
		defer func() {
			telemetry.IncrCounter(1, "new", "account")

			for _, a := range vestingCoins {
				if a.Amount.IsInt64() {
					telemetry.SetGaugeWithLabels(
						[]string{"tx", "msg", "create_clawback_vesting_account"},
						float32(a.Amount.Int64()),
						[]metrics.Label{telemetry.NewLabel("denom", a.Denom)},
					)
				}
			}
		}()
	}

	// Send coins from the funder to vesting account
	if err := bk.SendCoins(ctx, from, to, vestingCoins); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.StoreKey),
		),
	)

	return &types.MsgCreateClawbackVestingAccountResponse{}, nil
}

// Clawback removes the unvested amount from a ClawbackVestingAccount.
// The destination defaults to the funder address, but can be overridden.
func (k Keeper) Clawback(
	goCtx context.Context,
	msg *types.MsgClawback,
) (*types.MsgClawbackResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ak := k.accountKeeper
	bk := k.bankKeeper

	// Errors checked during msg validation
	dest, _ := sdk.AccAddressFromBech32(msg.DestAddress)
	addr, _ := sdk.AccAddressFromBech32(msg.AccountAddress)

	// Default destination to funder address
	if msg.DestAddress == "" {
		dest, _ = sdk.AccAddressFromBech32(msg.FunderAddress)
	}

	if bk.BlockedAddr(dest) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized,
			"%s is not allowed to receive funds", msg.DestAddress,
		)
	}

	// Chech if account exists
	acc := ak.GetAccount(ctx, addr)
	if acc == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "account %s does not exist", msg.AccountAddress)
	}

	// Check if account has a clawback account
	va, ok := acc.(*types.ClawbackVestingAccount)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account not subject to clawback: %s", msg.AccountAddress)
	}

	// Check if account funder is same as in msg
	if va.FunderAddress != msg.FunderAddress {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "clawback can only be requested by original funder %s", va.FunderAddress)
	}

	// Perform clawback transfer
	if err := k.transferClawback(ctx, *va, dest); err != nil {
		return nil, err
	}

	return &types.MsgClawbackResponse{}, nil
}

// addGrant merges a new clawback vesting grant into an existing
// ClawbackVestingAccount.
func (k Keeper) addGrant(
	ctx sdk.Context,
	va *types.ClawbackVestingAccount,
	grantStartTime int64,
	grantLockupPeriods, grantVestingPeriods []sdkvesting.Period,
	grantCoins sdk.Coins,
) {
	// how much is really delegated?
	bondedAmt := k.GetDelegatorBonded(ctx, va.GetAddress())
	unbondingAmt := k.GetDelegatorUnbonding(ctx, va.GetAddress())
	delegatedAmt := bondedAmt.Add(unbondingAmt)
	delegated := sdk.NewCoins(sdk.NewCoin(k.stakingKeeper.BondDenom(ctx), delegatedAmt))

	// modify schedules for the new grant
	newLockupStart, newLockupEnd, newLockupPeriods := types.DisjunctPeriods(va.GetStartTime(), grantStartTime, va.LockupPeriods, grantLockupPeriods)
	newVestingStart, newVestingEnd, newVestingPeriods := types.DisjunctPeriods(va.GetStartTime(), grantStartTime,
		va.GetVestingPeriods(), grantVestingPeriods)
	if newLockupStart != newVestingStart {
		panic("bad start time calculation")
	}
	va.StartTime = time.Unix(newLockupStart, 0)
	va.EndTime = types.Max64(newLockupEnd, newVestingEnd)
	va.LockupPeriods = newLockupPeriods
	va.VestingPeriods = newVestingPeriods
	va.OriginalVesting = va.OriginalVesting.Add(grantCoins...)

	// cap DV at the current unvested amount, DF rounds out to current delegated
	unvested := va.GetVestingCoins(ctx.BlockTime())
	va.DelegatedVesting = types.CoinsMin(delegated, unvested)
	va.DelegatedFree = delegated.Sub(va.DelegatedVesting)
}

// transferClawback transfers unvested tokens in a ClawbackVestingAccount to
// dest. Future vesting events are removed. Unstaked tokens are simply sent.
// Unbonding and staked tokens are transferred with their staking state intact.
// Account state is updated to reflect the removals.
func (k Keeper) transferClawback(
	ctx sdk.Context,
	va types.ClawbackVestingAccount,
	dest sdk.AccAddress,
) error {
	// Compute the clawback based on the account state only, and update account
	updatedAcc, toClawBack := va.ComputeClawback(ctx.BlockTime().Unix())
	if toClawBack.IsZero() {
		return nil
	}
	addr := updatedAcc.GetAddress()
	bondDenom := k.stakingKeeper.BondDenom(ctx)

	// Compute the clawback based on bank balance and delegation, and update account
	encumbered := updatedAcc.GetVestingCoins(ctx.BlockTime())
	bondedAmt := k.GetDelegatorBonded(ctx, addr)
	unbondingAmt := k.GetDelegatorUnbonding(ctx, addr)
	bonded := sdk.NewCoins(sdk.NewCoin(bondDenom, bondedAmt))
	unbonding := sdk.NewCoins(sdk.NewCoin(bondDenom, unbondingAmt))
	unbonded := k.bankKeeper.GetAllBalances(ctx, addr)
	updatedAcc, toClawBack = updatedAcc.UpdateDelegation(encumbered, toClawBack, bonded, unbonding, unbonded)

	// Write now now so that the bank module sees unvested tokens are unlocked.
	// Note that all store writes are aborted if there is a panic, so there is
	// no danger in writing incomplete results.
	k.accountKeeper.SetAccount(ctx, &updatedAcc)

	// Now that future vesting events (and associated lockup) are removed,
	// the balance of the account is unlocked and can be freely transferred.
	spendable := k.bankKeeper.SpendableCoins(ctx, addr)
	toXfer := types.CoinsMin(toClawBack, spendable)
	err := k.bankKeeper.SendCoins(ctx, addr, dest, toXfer)
	if err != nil {
		return err // shouldn't happen, given spendable check
	}
	toClawBack = toClawBack.Sub(toXfer)

	// We need to traverse the staking data structures to update the vesting
	// account bookkeeping, and to recover more funds if necessary. Staking is the
	// only way unvested tokens should be missing from the bank balance.

	// If we need more, transfer UnbondingDelegations.
	want := toClawBack.AmountOf(bondDenom)
	unbondings := k.stakingKeeper.GetUnbondingDelegations(ctx, addr, math.MaxUint16)
	for _, unbonding := range unbondings {
		valAddr, err := sdk.ValAddressFromBech32(unbonding.ValidatorAddress)
		if err != nil {
			panic(err)
		}
		transferred := k.TransferUnbonding(ctx, addr, dest, valAddr, want)
		want = want.Sub(transferred)
		if !want.IsPositive() {
			break
		}
	}

	// If we need more, transfer Delegations.
	if want.IsPositive() {
		delegations := k.stakingKeeper.GetDelegatorDelegations(ctx, addr, math.MaxUint16)
		for _, delegation := range delegations {
			validatorAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
			if err != nil {
				panic(err) // shouldn't happen
			}
			validator, found := k.stakingKeeper.GetValidator(ctx, validatorAddr)
			if !found {
				// validator has been removed
				continue
			}
			wantShares, err := validator.SharesFromTokensTruncated(want)
			if err != nil {
				// validator has no tokens
				continue
			}
			transferredShares := k.TransferDelegation(ctx, addr, dest, delegation.GetValidatorAddr(), wantShares)
			// to be conservative in what we're clawing back, round transferred shares up
			transferred := validator.TokensFromSharesRoundUp(transferredShares).RoundInt()
			want = want.Sub(transferred)
			if !want.IsPositive() {
				// Could be slightly negative, due to rounding?
				// Don't think so, due to the precautions above.
				break
			}
		}
	}

	// If we've transferred everything and still haven't transferred the desired
	// clawback amount, then the account must have most some unvested tokens from
	// slashing.
	return nil
}
