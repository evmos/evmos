// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"context"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/evmos/evmos/v20/utils"
	"github.com/evmos/evmos/v20/x/vesting/types"
)

var _ types.MsgServer = &Keeper{}

// CreateClawbackVestingAccount creates a new ClawbackVestingAccount.
//
// Checks performed on the ValidateBasic include:
//   - funder and vesting addresses are correct bech32 format
//   - funder and vesting addresses are not the zero address
func (k Keeper) CreateClawbackVestingAccount(
	goCtx context.Context,
	msg *types.MsgCreateClawbackVestingAccount,
) (*types.MsgCreateClawbackVestingAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ak := k.accountKeeper
	bk := k.bankKeeper
	ek := k.evmKeeper

	// Error checked during msg validation
	funderAddress := sdk.MustAccAddressFromBech32(msg.FunderAddress)
	vestingAddress := sdk.MustAccAddressFromBech32(msg.VestingAddress)

	if bk.BlockedAddr(vestingAddress) {
		return nil, errorsmod.Wrapf(errortypes.ErrUnauthorized,
			"%s is a blocked address and cannot be converted in a clawback vesting account", msg.VestingAddress,
		)
	}

	// A clawback vesting account can only be created when the account exists
	acc := ak.GetAccount(ctx, vestingAddress)
	if acc == nil {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest,
			"account %s does not exist", msg.VestingAddress,
		)
	}

	// Check if existing account already is a clawback vesting account
	_, isClawback := acc.(*types.ClawbackVestingAccount)
	if isClawback {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest,
			"%s is already a clawback vesting account", msg.VestingAddress,
		)
	}

	// Check for contract account (code hash is not empty)
	if ek.IsContract(ctx, utils.CosmosToEthAddr(acc.GetAddress())) {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest,
			"account %s is a contract account and cannot be converted in a clawback vesting account", msg.VestingAddress,
		)
	}

	baseAcc, ok := acc.(*authtypes.BaseAccount)
	if !ok {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest,
			"account %s could not be converted to a base account", msg.VestingAddress,
		)
	}
	baseVestingAcc := &sdkvesting.BaseVestingAccount{BaseAccount: baseAcc}
	vestingAcc := &types.ClawbackVestingAccount{
		BaseVestingAccount: baseVestingAcc,
		FunderAddress:      funderAddress.String(),
	}
	ak.SetAccount(ctx, vestingAcc)

	if !msg.EnableGovClawback {
		k.SetGovClawbackDisabled(ctx, vestingAcc.GetAddress())
	}

	telemetry.IncrCounter(
		float32(ctx.GasMeter().GasConsumed()),
		"tx", "create_clawback_vesting_account", "gas_used",
	)

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeCreateClawbackVestingAccount,
				sdk.NewAttribute(types.AttributeKeyFunder, msg.FunderAddress),
				sdk.NewAttribute(sdk.AttributeKeySender, msg.VestingAddress),
			),
		},
	)

	return &types.MsgCreateClawbackVestingAccountResponse{}, nil
}

// FundVestingAccount funds a ClawbackVestingAccount with the provided amount.
// This can only be executed by the funder of the vesting account.
//
// Checks performed on the ValidateBasic include:
//   - funder and vesting addresses are correct bech32 format
//   - vesting address is not the zero address
//   - both vesting and lockup periods are non-empty
//   - both lockup and vesting periods contain valid amounts and lengths
//   - both vesting and lockup periods describe the same total amount
func (k Keeper) FundVestingAccount(goCtx context.Context, msg *types.MsgFundVestingAccount) (*types.MsgFundVestingAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ak := k.accountKeeper
	bk := k.bankKeeper

	// Error checked during msg validation
	funderAddr := sdk.MustAccAddressFromBech32(msg.FunderAddress)
	vestingAddr := sdk.MustAccAddressFromBech32(msg.VestingAddress)

	if bk.BlockedAddr(vestingAddr) {
		return nil, errorsmod.Wrapf(errortypes.ErrUnauthorized,
			"%s is not allowed to receive funds", msg.VestingAddress,
		)
	}

	// Check if vesting account exists
	vestingAcc, err := k.GetClawbackVestingAccount(ctx, vestingAddr)
	if err != nil {
		return nil, err
	}

	vestingCoins := msg.VestingPeriods.TotalAmount()
	lockupCoins := msg.LockupPeriods.TotalAmount()

	// If lockup absent, default to an instant unlock schedule
	if !vestingCoins.IsZero() && len(msg.LockupPeriods) == 0 {
		msg.LockupPeriods = sdkvesting.Periods{
			{Length: 0, Amount: vestingCoins},
		}
		lockupCoins = vestingCoins
	}

	// If vesting absent, default to an instant vesting schedule
	if !lockupCoins.IsZero() && len(msg.VestingPeriods) == 0 {
		msg.VestingPeriods = sdkvesting.Periods{
			{Length: 0, Amount: lockupCoins},
		}
		vestingCoins = lockupCoins
	}

	if msg.FunderAddress != vestingAcc.FunderAddress {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "account %s can only accept grants from account %s", msg.VestingAddress, vestingAcc.FunderAddress)
	}

	err = k.addGrant(ctx, vestingAcc, msg.GetStartTime().Unix(), msg.GetLockupPeriods(), msg.GetVestingPeriods(), vestingCoins)
	if err != nil {
		return nil, err
	}
	ak.SetAccount(ctx, vestingAcc)

	// Send coins from the funder to vesting account
	if err = bk.SendCoins(ctx, funderAddr, vestingAddr, vestingCoins); err != nil {
		return nil, err
	}

	telemetry.IncrCounter(
		float32(ctx.GasMeter().GasConsumed()),
		"tx", "fund_vesting_account", "gas_used",
	)
	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeFundVestingAccount,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.FunderAddress),
				sdk.NewAttribute(types.AttributeKeyCoins, vestingCoins.String()),
				sdk.NewAttribute(types.AttributeKeyStartTime, msg.StartTime.String()),
				sdk.NewAttribute(types.AttributeKeyAccount, msg.VestingAddress),
			),
		},
	)

	return &types.MsgFundVestingAccountResponse{}, nil
}

// Clawback removes the unvested amount from a ClawbackVestingAccount.
// The destination defaults to the funder address, but can be overridden.
//
// Checks performed on the ValidateBasic include:
//   - funder and vesting addresses are correct bech32 format
//   - if destination address is not empty it is also correct bech32 format
func (k Keeper) Clawback(
	goCtx context.Context,
	msg *types.MsgClawback,
) (*types.MsgClawbackResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ak := k.accountKeeper
	bk := k.bankKeeper

	// NOTE: errors checked during msg validation
	addr := sdk.MustAccAddressFromBech32(msg.AccountAddress)
	funder := sdk.MustAccAddressFromBech32(msg.FunderAddress)

	// NOTE: ignore error in case dest address is not defined and default to funder address
	//#nosec G703 -- error is checked during ValidateBasic already.
	dest, _ := sdk.AccAddressFromBech32(msg.DestAddress)
	if msg.DestAddress == "" {
		dest = funder
	}

	if msg.FunderAddress != k.authority.String() {
		if k.HasActiveClawbackProposal(ctx, addr) {
			return nil, errorsmod.Wrapf(errortypes.ErrUnauthorized,
				"clawback is disabled while there is an active clawback proposal for account %s",
				msg.AccountAddress,
			)
		}

		// NOTE: we check the destination address only for the case where it's not sent from the
		// authority account, because in that case the destination address is hardcored to the
		// community pool address anyway (see further below).
		if bk.BlockedAddr(dest) {
			return nil, errorsmod.Wrapf(errortypes.ErrUnauthorized,
				"%s is a blocked address and not allowed to receive funds", msg.DestAddress,
			)
		}
	}

	// Get clawback vesting account
	va, err := k.GetClawbackVestingAccount(ctx, addr)
	if err != nil {
		return nil, err
	}

	// Check if account has any vesting or lockup periods
	if len(va.VestingPeriods) == 0 && len(va.LockupPeriods) == 0 {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "account %s has no vesting or lockup periods", msg.AccountAddress)
	}

	// Check to see if it's a governance proposal clawback
	if k.authority.String() == msg.FunderAddress {
		if k.HasGovClawbackDisabled(ctx, addr) {
			return nil, errorsmod.Wrap(types.ErrNotSubjectToGovClawback, addr.String())
		}

		dest = ak.GetModuleAddress(distributiontypes.ModuleName)

		// Check if account funder is same as in msg
	} else if va.FunderAddress != msg.FunderAddress {
		return nil, errorsmod.Wrapf(errortypes.ErrUnauthorized, "clawback can only be requested by original funder: %s", va.FunderAddress)
	}

	// Perform clawback transfer
	clawedBack, err := k.transferClawback(ctx, *va, dest)
	if err != nil {
		return nil, err
	}

	telemetry.IncrCounter(
		float32(ctx.GasMeter().GasConsumed()),
		"tx", "clawback", "gas_used",
	)

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeClawback,
				sdk.NewAttribute(types.AttributeKeyFunder, msg.FunderAddress),
				sdk.NewAttribute(types.AttributeKeyAccount, msg.AccountAddress),
				sdk.NewAttribute(types.AttributeKeyDestination, dest.String()),
			),
		},
	)

	return &types.MsgClawbackResponse{
		Coins: clawedBack,
	}, nil
}

// UpdateVestingFunder updates the funder account of a ClawbackVestingAccount.
//
// Checks performed on the ValidateBasic include:
//   - new funder and vesting addresses are correct bech32 format
//   - new funder address is not the zero address
//   - new funder address is not the same as the current funder address
func (k Keeper) UpdateVestingFunder(
	goCtx context.Context,
	msg *types.MsgUpdateVestingFunder,
) (*types.MsgUpdateVestingFunderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ak := k.accountKeeper
	bk := k.bankKeeper

	// NOTE: errors checked during msg validation
	newFunder := sdk.MustAccAddressFromBech32(msg.NewFunderAddress)
	vestingAccAddr := sdk.MustAccAddressFromBech32(msg.VestingAddress)

	// Check if there is an active clawback proposal for the given account
	if k.HasActiveClawbackProposal(ctx, vestingAccAddr) {
		return nil, errorsmod.Wrapf(errortypes.ErrUnauthorized,
			"cannot update funder while there is an active clawback proposal for account %s",
			vestingAccAddr.String(),
		)
	}

	// Need to check if new funder can receive funds because in
	// Clawback function, destination defaults to funder address
	if bk.BlockedAddr(newFunder) {
		return nil, errorsmod.Wrapf(errortypes.ErrUnauthorized,
			"%s is a blocked address and not allowed to fund vesting accounts", msg.NewFunderAddress,
		)
	}

	// Check if vesting account exists
	va, err := k.GetClawbackVestingAccount(ctx, vestingAccAddr)
	if err != nil {
		return nil, err
	}

	// Check if current funder is same as in msg
	if va.FunderAddress != msg.FunderAddress {
		return nil, errorsmod.Wrapf(errortypes.ErrUnauthorized, "%s is not the current funder and cannot update the funder address", va.FunderAddress)
	}

	// Perform clawback account update
	va.FunderAddress = msg.NewFunderAddress
	ak.SetAccount(ctx, va)

	telemetry.IncrCounter(
		float32(ctx.GasMeter().GasConsumed()),
		"tx", "update_vesting_funder", "gas_used",
	)

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeUpdateVestingFunder,
				sdk.NewAttribute(types.AttributeKeyFunder, msg.FunderAddress),
				sdk.NewAttribute(types.AttributeKeyAccount, msg.VestingAddress),
				sdk.NewAttribute(types.AttributeKeyNewFunder, msg.NewFunderAddress),
			),
		},
	)

	return &types.MsgUpdateVestingFunderResponse{}, nil
}

// ConvertVestingAccount converts a ClawbackVestingAccount to the default chain account
// after its lockup and vesting periods have concluded.
func (k Keeper) ConvertVestingAccount(
	goCtx context.Context,
	msg *types.MsgConvertVestingAccount,
) (*types.MsgConvertVestingAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	address := sdk.MustAccAddressFromBech32(msg.VestingAddress)

	vestingAcc, err := k.GetClawbackVestingAccount(ctx, address)
	if err != nil {
		return nil, err
	}

	// check if account has any vesting coins left
	if !vestingAcc.GetVestingCoins(ctx.BlockTime()).IsZero() {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "vesting coins still left in account: %s", msg.VestingAddress)
	}

	// check if account has any locked up coins left
	if vestingAcc.HasLockedCoins(ctx.BlockTime()) {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "locked up coins still left in account: %s", msg.VestingAddress)
	}

	// if gov clawback is disabled, remove the entry from the store.
	// if no entry is found for the address, this will no-op
	k.DeleteGovClawbackDisabled(ctx, address)

	baseAcc := vestingAcc.BaseAccount
	k.accountKeeper.SetAccount(ctx, baseAcc)

	return &types.MsgConvertVestingAccountResponse{}, nil
}

// addGrant merges a new clawback vesting grant into an existing
// ClawbackVestingAccount.
func (k Keeper) addGrant(
	ctx sdk.Context,
	va *types.ClawbackVestingAccount,
	grantStartTime int64,
	grantLockupPeriods, grantVestingPeriods sdkvesting.Periods,
	grantCoins sdk.Coins,
) error {
	// check if the clawback vesting account has only been initialized and not yet funded --
	// in that case it's necessary to update the vesting account with the given start time because this is set to zero in the initialization
	if len(va.LockupPeriods) == 0 && len(va.VestingPeriods) == 0 {
		va.StartTime = time.Unix(grantStartTime, 0).UTC()
	}

	// how much is really delegated?
	vestingAddr := va.GetAddress()
	bondedAmt, err := k.stakingKeeper.GetDelegatorBonded(ctx, vestingAddr)
	if err != nil {
		return err
	}
	unbondingAmt, err := k.stakingKeeper.GetDelegatorUnbonding(ctx, vestingAddr)
	if err != nil {
		return err
	}
	delegatedAmt := bondedAmt.Add(unbondingAmt)
	bondDenom, err := k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return err
	}
	delegated := sdk.NewCoins(sdk.NewCoin(bondDenom, delegatedAmt))

	// modify schedules for the new grant
	accStartTime := va.GetStartTime()
	newLockupStart, newLockupEnd, newLockupPeriods := types.DisjunctPeriods(accStartTime, grantStartTime, va.LockupPeriods, grantLockupPeriods)
	newVestingStart, newVestingEnd, newVestingPeriods := types.DisjunctPeriods(
		accStartTime,
		grantStartTime,
		va.GetVestingPeriods(),
		grantVestingPeriods,
	)

	if newLockupStart != newVestingStart {
		return errorsmod.Wrapf(
			types.ErrVestingLockup,
			"vesting start time calculation should match lockup start (%d ≠ %d)",
			newVestingStart, newLockupStart,
		)
	}

	va.StartTime = time.Unix(newLockupStart, 0).UTC()
	va.EndTime = types.Max64(newLockupEnd, newVestingEnd)
	va.LockupPeriods = newLockupPeriods
	va.VestingPeriods = newVestingPeriods
	va.OriginalVesting = va.OriginalVesting.Add(grantCoins...)

	// cap DV at the current unvested amount, DF rounds out to current delegated
	unvested := va.GetVestingCoins(ctx.BlockTime())
	va.DelegatedVesting = delegated.Min(unvested)
	va.DelegatedFree = delegated.Sub(va.DelegatedVesting...)
	return nil
}

// transferClawback transfers unvested tokens in a ClawbackVestingAccount to
// the destination address. Then, it updates the lockup schedule, removes future
// vesting events and deletes the store entry for governance clawback if it exists.
func (k Keeper) transferClawback(
	ctx sdk.Context,
	vestingAccount types.ClawbackVestingAccount,
	destinationAddr sdk.AccAddress,
) (sdk.Coins, error) {
	// Compute clawback amount, unlock unvested tokens and remove future vesting events
	updatedAcc, toClawBack := vestingAccount.ComputeClawback(ctx.BlockTime().Unix())

	// convert the account back to a normal account
	//
	// NOTE: this is necessary to allow the bank keeper to send the locked coins away to the
	// destination address. If the account is not converted, the coins will still be seen as locked,
	// and can therefore not be transferred.
	baseAcc := updatedAcc.BaseAccount

	// set the account with the updated values of the vesting schedule
	k.accountKeeper.SetAccount(ctx, baseAcc)

	address := updatedAcc.GetAddress()

	// if gov clawback is disabled, remove the entry from the store.
	// if no entry is found for the address, this will no-op
	k.DeleteGovClawbackDisabled(ctx, address)

	// In case destination is community pool (e.g. Gov Clawback)
	// call the corresponding function
	if destinationAddr.String() == authtypes.NewModuleAddress(distributiontypes.ModuleName).String() {
		return toClawBack, k.distributionKeeper.FundCommunityPool(ctx, toClawBack, address)
	}

	// NOTE: don't use `SpendableCoins` to get the minimum value to clawback since
	// the amount is retrieved from `ComputeClawback`, which ensures correctness.
	// `SpendableCoins` can result in gas exhaustion if the user has too many
	// different denoms (because of store iteration).

	// Transfer clawback to the destination (funder)
	return toClawBack, k.bankKeeper.SendCoins(ctx, address, destinationAddr, toClawBack)
}
