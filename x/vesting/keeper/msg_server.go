// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"context"
	"time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/cosmos/cosmos-sdk/telemetry"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	vestingexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evmostypes "github.com/evmos/evmos/v14/types"
	"github.com/evmos/evmos/v14/x/vesting/types"
)

var _ types.MsgServer = &Keeper{}

// CreateClawbackVestingAccount creates a new ClawbackVestingAccount
// Checks performed on the ValidateBasic include:
// - funder and vesting addresses are correct bech32 format
// - funder and vesting addresses are not the zero address
func (k Keeper) CreateClawbackVestingAccount(
	goCtx context.Context,
	msg *types.MsgCreateClawbackVestingAccount,
) (*types.MsgCreateClawbackVestingAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ak := k.accountKeeper
	bk := k.bankKeeper

	// Error checked during msg validation
	funderAddress := sdk.MustAccAddressFromBech32(msg.FunderAddress)
	vestingAddress := sdk.MustAccAddressFromBech32(msg.VestingAddress)

	if bk.BlockedAddr(vestingAddress) {
		return nil, errorsmod.Wrapf(errortypes.ErrUnauthorized,
			"%s is not allowed to be a clawback vesting account", msg.VestingAddress,
		)
	}

	// Create clawback vesting account if the account is not already one
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

	// Initialize the vesting account
	ethAcc, ok := acc.(*evmostypes.EthAccount)

	if !ok {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest,
			"account %s is not an Ethereum account", msg.VestingAddress,
		)
	}
	baseAcc := ethAcc.GetBaseAccount()
	baseVestingAcc := &sdkvesting.BaseVestingAccount{BaseAccount: baseAcc}
	vestingAcc := &types.ClawbackVestingAccount{
		BaseVestingAccount: baseVestingAcc,
		FunderAddress:      funderAddress.String(),
	}
	ak.SetAccount(ctx, vestingAcc)

	if msg.EnableGovClawback {
		k.SetGovClawbackEnabled(ctx, vestingAcc.GetAddress())
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
// Checks performed on the ValidateBasic include:
// - funder and vesting addresses are correct bech32 format
// - vesting address is not the zero address
// - both vesting and lockup periods are non-empty
// - both lockup and vesting periods contain valid amounts and lengths
// - both vesting and lockup periods describe the same total amount
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

	// Check if vesting account exists
	acc := ak.GetAccount(ctx, vestingAddr)
	if acc == nil {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "account %s does not exist", msg.VestingAddress)
	}

	vestingAcc, isClawback := acc.(*types.ClawbackVestingAccount)
	if !isClawback {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "account %s must be a clawback vesting account", msg.VestingAddress)
	}

	// NOTE: Add grant only if vesting account is empty or "merge" is true and the funder is correct.
	if msg.FunderAddress != vestingAcc.FunderAddress {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "account %s can only accept grants from account %s", msg.VestingAddress, vestingAcc.FunderAddress)
	}

	err := k.addGrant(ctx, vestingAcc, msg.GetStartTime().Unix(), msg.GetLockupPeriods(), msg.GetVestingPeriods(), vestingCoins)
	if err != nil {
		return nil, err
	}
	ak.SetAccount(ctx, vestingAcc)

	// Send coins from the funder to vesting account
	if err := bk.SendCoins(ctx, funderAddr, vestingAddr, vestingCoins); err != nil {
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
// Checks performed on the ValidateBasic include:
// - funder and vesting addresses are correct bech32 format
// - if destination address is not empty it is also correct bech32 format
func (k Keeper) Clawback(
	goCtx context.Context,
	msg *types.MsgClawback,
) (*types.MsgClawbackResponse, error) {
	// Check if governance clawback is enabled
	params := k.GetParams(sdk.UnwrapSDKContext(goCtx))
	if !params.EnableGovClawback && msg.FunderAddress == k.authority.String() {
		return nil, errorsmod.Wrapf(errortypes.ErrUnauthorized, "gov clawback is disabled")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	ak := k.accountKeeper
	bk := k.bankKeeper

	// NOTE: ignore error in case dest address is not defined
	dest, _ := sdk.AccAddressFromBech32(msg.DestAddress)

	// NOTE: error checked during msg validation
	addr := sdk.MustAccAddressFromBech32(msg.AccountAddress)

	// Default destination to funder address
	if msg.DestAddress == "" {
		dest = sdk.MustAccAddressFromBech32(msg.FunderAddress)
	}

	if k.authority.String() != msg.FunderAddress && bk.BlockedAddr(dest) {
		return nil, errorsmod.Wrapf(errortypes.ErrUnauthorized,
			"account is not allowed to receive funds: %s", msg.DestAddress,
		)
	}

	// Check if account exists
	acc := ak.GetAccount(ctx, addr)
	if acc == nil {
		return nil, errorsmod.Wrapf(errortypes.ErrNotFound, "account does not exist: %s", msg.AccountAddress)
	}

	// Check if account has a clawback account
	va, ok := acc.(*types.ClawbackVestingAccount)
	if !ok {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "account not subject to clawback: %s", msg.AccountAddress)
	}

	// Check if account has any vesting or lockup periods
	if len(va.VestingPeriods) == 0 && len(va.LockupPeriods) == 0 {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "account %s has no vesting or lockup periods", msg.AccountAddress)
	}

	// Return error if clawback is attempted before start time
	if ctx.BlockTime().Before(va.StartTime) {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "clawback can only be executed after vesting begins: %s", va.FunderAddress)
	}

	// Check to see if it's a governance proposal clawback
	if k.authority.String() == msg.FunderAddress {
		if !k.HasGovClawbackEnabled(ctx, addr) {
			return nil, errorsmod.Wrapf(errortypes.ErrUnauthorized, "account %s doesn't have governance clawback enabled", addr)
		}

		dest = ak.GetModuleAddress(distributiontypes.ModuleName)

		// Check if account funder is same as in msg
	} else if va.FunderAddress != msg.FunderAddress {
		return nil, errorsmod.Wrapf(errortypes.ErrUnauthorized, "clawback can only be requested by original funder: %s", va.FunderAddress)
	}

	// Perform clawback transfer
	if err := k.transferClawback(ctx, *va, dest); err != nil {
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

	return &types.MsgClawbackResponse{}, nil
}

// UpdateVestingFunder updates the funder account of a ClawbackVestingAccount.
// Checks performed on the ValidateBasic include:
// - new funder and vesting addresses are correct bech32 format
// - new funder address is not the zero address
// - new funder address is not the same as the current funder address
func (k Keeper) UpdateVestingFunder(
	goCtx context.Context,
	msg *types.MsgUpdateVestingFunder,
) (*types.MsgUpdateVestingFunderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ak := k.accountKeeper
	bk := k.bankKeeper

	// NOTE: errors checked during msg validation
	newFunder := sdk.MustAccAddressFromBech32(msg.NewFunderAddress)
	vesting := sdk.MustAccAddressFromBech32(msg.VestingAddress)

	// Need to check if new funder can receive funds because in
	// Clawback function, destination defaults to funder address
	if bk.BlockedAddr(newFunder) {
		return nil, errorsmod.Wrapf(errortypes.ErrUnauthorized,
			"%s is not allowed to receive funds", msg.NewFunderAddress,
		)
	}

	// Check if vesting account exists
	vestingAcc := ak.GetAccount(ctx, vesting)
	if vestingAcc == nil {
		return nil, errorsmod.Wrapf(errortypes.ErrNotFound, "account %s does not exist", msg.VestingAddress)
	}

	// Check if account is a clawback vesting account
	va, ok := vestingAcc.(*types.ClawbackVestingAccount)
	if !ok {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "account not subject to clawback: %s", msg.VestingAddress)
	}

	// Check if account current funder is same as in msg
	if va.FunderAddress != msg.FunderAddress {
		return nil, errorsmod.Wrapf(errortypes.ErrUnauthorized, "clawback can only be requested by original funder %s", va.FunderAddress)
	}

	// Perform clawback account update
	va.FunderAddress = msg.NewFunderAddress
	// set the account with the updated funder
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
// after its lock and vesting periods have concluded.
func (k Keeper) ConvertVestingAccount(
	goCtx context.Context,
	msg *types.MsgConvertVestingAccount,
) (*types.MsgConvertVestingAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	address := sdk.MustAccAddressFromBech32(msg.VestingAddress)

	account := k.accountKeeper.GetAccount(ctx, address)
	if account == nil {
		return nil, errorsmod.Wrapf(errortypes.ErrNotFound, "account %s does not exist", msg.VestingAddress)
	}

	// Check if account is of VestingAccount interface
	if _, ok := account.(vestingexported.VestingAccount); !ok {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "account not subject to vesting: %s", msg.VestingAddress)
	}

	// check if account is of type ClawbackVestingAccount
	vestingAcc, ok := account.(*types.ClawbackVestingAccount)
	if !ok {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "account %s is not a ClawbackVestingAccount", msg.VestingAddress)
	}

	// check if account  has any vesting coins left
	if vestingAcc.GetVestingCoins(ctx.BlockTime()) != nil {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "vesting coins still left in account: %s", msg.VestingAddress)
	}

	ethAccount := evmostypes.ProtoAccount().(*evmostypes.EthAccount)
	ethAccount.BaseAccount = vestingAcc.BaseAccount
	k.accountKeeper.SetAccount(ctx, ethAccount)

	return &types.MsgConvertVestingAccountResponse{}, nil
}

// UpdateParams defines a method for updating vesting params
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
		va.StartTime = time.Unix(grantStartTime, 0)
	}

	// how much is really delegated?
	bondedAmt := k.stakingKeeper.GetDelegatorBonded(ctx, va.GetAddress())
	unbondingAmt := k.stakingKeeper.GetDelegatorUnbonding(ctx, va.GetAddress())
	delegatedAmt := bondedAmt.Add(unbondingAmt)
	delegated := sdk.NewCoins(sdk.NewCoin(k.stakingKeeper.BondDenom(ctx), delegatedAmt))

	// modify schedules for the new grant
	newLockupStart, newLockupEnd, newLockupPeriods := types.DisjunctPeriods(va.GetStartTime(), grantStartTime, va.LockupPeriods, grantLockupPeriods)
	newVestingStart, newVestingEnd, newVestingPeriods := types.DisjunctPeriods(
		va.GetStartTime(),
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

	va.StartTime = time.Unix(newLockupStart, 0)
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
// the destination address. Then it, updates the lockup schedule, removes future
// vesting events and disables clawback vesting from governance.
func (k Keeper) transferClawback(
	ctx sdk.Context,
	vestingAccount types.ClawbackVestingAccount,
	destinationAddr sdk.AccAddress,
) error {
	// Compute clawback amount, unlock unvested tokens and remove future vesting events
	updatedAcc, toClawBack := vestingAccount.ComputeClawback(ctx.BlockTime().Unix())
	if toClawBack.IsZero() {
		// no-op, nothing to transfer
		return nil
	}

	// convert the account back to a normal EthAccount
	ethAccount := evmostypes.ProtoAccount().(*evmostypes.EthAccount)
	ethAccount.BaseAccount = updatedAcc.BaseAccount

	// set the account with the updated values of the vesting schedule
	k.accountKeeper.SetAccount(ctx, ethAccount)

	address := updatedAcc.GetAddress()

	// Disable governance clawback for vesting account. If the account has this
	// functionality disabled, this will no-op
	k.DeleteGovClawbackEnabled(ctx, address)

	// In case destination is community pool (e.g. Gov Clawback)
	// call the corresponding function
	if destinationAddr.String() == authtypes.NewModuleAddress(distributiontypes.ModuleName).String() {
		return k.distributionKeeper.FundCommunityPool(ctx, toClawBack, address)
	}

	// NOTE: don't use `SpendableCoins` to get the minimum value to clawback since
	// the amount is retrieved from `ComputeClawback`, which ensures correctness.
	// `SpendableCoins` can result in gas exhaustion if the user has too many
	// different denoms (because of store iteration).

	// Transfer clawback to the destination (funder)
	return k.bankKeeper.SendCoins(ctx, address, destinationAddr, toClawBack)
}
