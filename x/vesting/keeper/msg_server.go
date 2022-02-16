package keeper

import (
	"context"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/armon/go-metrics"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tharsis/evmos/x/vesting/types"
)

var _ types.MsgServer = &Keeper{}

// CreateVestingAccount creates a new delayed or continuous vesting account.
func (k Keeper) CreateVestingAccount(
	goCtx context.Context,
	msg *types.MsgCreateVestingAccount,
) (*types.MsgCreateVestingAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ak := k.accountKeeper
	bk := k.bankKeeper

	if err := bk.IsSendEnabledCoins(ctx, msg.Amount...); err != nil {
		return nil, err
	}

	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}
	to, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}

	if bk.BlockedAddr(to) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", msg.ToAddress)
	}

	if acc := ak.GetAccount(ctx, to); acc != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s already exists", msg.ToAddress)
	}

	baseAccount := ak.NewAccountWithAddress(ctx, to)
	if _, ok := baseAccount.(*authtypes.BaseAccount); !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid account type; expected: BaseAccount, got: %T", baseAccount)
	}

	baseVestingAccount := types.NewBaseVestingAccount(baseAccount.(*authtypes.BaseAccount), msg.Amount.Sort(), msg.EndTime)

	var acc authtypes.AccountI

	if msg.Delayed {
		acc = types.NewDelayedVestingAccountRaw(baseVestingAccount)
	} else {
		acc = types.NewContinuousVestingAccountRaw(baseVestingAccount, ctx.BlockTime().Unix())
	}

	ak.SetAccount(ctx, acc)

	defer func() {
		telemetry.IncrCounter(1, "new", "account")

		for _, a := range msg.Amount {
			if a.Amount.IsInt64() {
				telemetry.SetGaugeWithLabels(
					[]string{"tx", "msg", "create_vesting_account"},
					float32(a.Amount.Int64()),
					[]metrics.Label{telemetry.NewLabel("denom", a.Denom)},
				)
			}
		}
	}()

	err = bk.SendCoins(ctx, from, to, msg.Amount)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.StoreKey),
		),
	)

	return &types.MsgCreateVestingAccountResponse{}, nil
}

// CreatePeriodicVestingAccount creates a new periodic vesting account, or merges a grant into an existing one.
func (k Keeper) CreatePeriodicVestingAccount(
	goCtx context.Context,
	msg *types.MsgCreatePeriodicVestingAccount,
) (*types.MsgCreatePeriodicVestingAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ak := k.accountKeeper
	bk := k.bankKeeper

	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}
	to, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}

	if bk.BlockedAddr(to) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", msg.ToAddress)
	}

	var totalCoins sdk.Coins
	for _, period := range msg.VestingPeriods {
		totalCoins = totalCoins.Add(period.Amount...)
	}
	totalCoins = totalCoins.Sort()

	madeNewAcc := false
	acc := ak.GetAccount(ctx, to)

	if acc != nil {
		pva, isPeriodic := acc.(*types.PeriodicVestingAccount)
		switch {
		case !msg.Merge && isPeriodic:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s already exists; consider using --merge", msg.ToAddress)
		case !msg.Merge && !isPeriodic:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s already exists", msg.ToAddress)
		case msg.Merge && !isPeriodic:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrNotSupported, "account %s must be a periodic vesting account", msg.ToAddress)
		}
		k.AddGrantToPeriodicVestingAccount(ctx, pva, msg.GetStartTime(), msg.GetVestingPeriods(), totalCoins)
	} else {
		baseAccount := ak.NewAccountWithAddress(ctx, to)
		acc = types.NewPeriodicVestingAccount(baseAccount.(*authtypes.BaseAccount), totalCoins, msg.StartTime, msg.VestingPeriods)
		madeNewAcc = true
	}

	ak.SetAccount(ctx, acc)

	if madeNewAcc {
		defer func() {
			telemetry.IncrCounter(1, "new", "account")

			for _, a := range totalCoins {
				if a.Amount.IsInt64() {
					telemetry.SetGaugeWithLabels(
						[]string{"tx", "msg", "create_periodic_vesting_account"},
						float32(a.Amount.Int64()),
						[]metrics.Label{telemetry.NewLabel("denom", a.Denom)},
					)
				}
			}
		}()
	}

	err = bk.SendCoins(ctx, from, to, totalCoins)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.StoreKey),
		),
	)
	return &types.MsgCreatePeriodicVestingAccountResponse{}, nil
}

// CreateClawbackVestingAccount creates a new ClawbackVestingAccount, or merges a grant into an existing one.
func (k Keeper) CreateClawbackVestingAccount(
	goCtx context.Context,
	msg *types.MsgCreateClawbackVestingAccount,
) (*types.MsgCreateClawbackVestingAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ak := k.accountKeeper
	bk := k.bankKeeper

	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}
	to, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}

	if bk.BlockedAddr(to) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", msg.ToAddress)
	}

	vestingCoins := sdk.NewCoins()
	for _, period := range msg.VestingPeriods {
		vestingCoins = vestingCoins.Add(period.Amount...)
	}

	lockupCoins := sdk.NewCoins()
	for _, period := range msg.LockupPeriods {
		lockupCoins = lockupCoins.Add(period.Amount...)
	}

	if !vestingCoins.IsZero() && len(msg.LockupPeriods) == 0 {
		// If lockup absent, default to an instant unlock schedule
		msg.LockupPeriods = []types.Period{
			{Length: 0, Amount: vestingCoins},
		}
		lockupCoins = vestingCoins
	}

	if !lockupCoins.IsZero() && len(msg.VestingPeriods) == 0 {
		// If vesting absent, default to an instant vesting schedule
		msg.VestingPeriods = []types.Period{
			{Length: 0, Amount: lockupCoins},
		}
		vestingCoins = lockupCoins
	}

	// The vesting and lockup schedules must describe the same total amount.
	// IsEqual can panic, so use (a == b) <=> (a <= b && b <= a).
	if !(vestingCoins.IsAllLTE(lockupCoins) && lockupCoins.IsAllLTE(vestingCoins)) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "lockup and vesting amounts must be equal")
	}

	madeNewAcc := false
	acc := ak.GetAccount(ctx, to)
	var va *types.ClawbackVestingAccount

	if acc != nil {
		var isClawback bool
		va, isClawback = acc.(*types.ClawbackVestingAccount)
		switch {
		case !msg.Merge && isClawback:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s already exists; consider using --merge", msg.ToAddress)
		case !msg.Merge && !isClawback:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s already exists", msg.ToAddress)
		case msg.Merge && !isClawback:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrNotSupported, "account %s must be a clawback vesting account", msg.ToAddress)
		case msg.FromAddress != va.FunderAddress:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s can only accept grants from account %s", msg.ToAddress, va.FunderAddress)
		}
		k.AddGrantToClawbackVestingAccount(ctx, va, msg.GetStartTime(), msg.GetLockupPeriods(), msg.GetVestingPeriods(), vestingCoins)
	} else {
		baseAccount := ak.NewAccountWithAddress(ctx, to)
		va = types.NewClawbackVestingAccount(baseAccount.(*authtypes.BaseAccount), from, vestingCoins, msg.StartTime, msg.LockupPeriods, msg.VestingPeriods)
		madeNewAcc = true
	}

	ak.SetAccount(ctx, va)

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

	err = bk.SendCoins(ctx, from, to, vestingCoins)
	if err != nil {
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

	funder, err := sdk.AccAddressFromBech32(msg.GetFunderAddress())
	if err != nil {
		return nil, err
	}
	addr, err := sdk.AccAddressFromBech32(msg.GetAddress())
	if err != nil {
		return nil, err
	}
	dest := funder
	if msg.GetDestAddress() != "" {
		dest, err = sdk.AccAddressFromBech32(msg.GetDestAddress())
		if err != nil {
			return nil, err
		}
	}

	if bk.BlockedAddr(dest) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", msg.DestAddress)
	}

	acc := ak.GetAccount(ctx, addr)
	if acc == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "account %s does not exist", msg.Address)
	}
	va, ok := acc.(*types.ClawbackVestingAccount)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "account not subject to clawback: %s", msg.Address)
	}

	if va.FunderAddress != msg.GetFunderAddress() {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "clawback can only be requested by original funder %s", va.FunderAddress)
	}

	err = k.ClawbackFromClawbackVestingAccount(ctx, *va, dest)
	if err != nil {
		return nil, err
	}

	return &types.MsgClawbackResponse{}, nil
}
