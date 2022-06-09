package types

import (
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

var (
	_ sdk.Msg = &MsgCreateClawbackVestingAccount{}
	_ sdk.Msg = &MsgClawback{}
	_ sdk.Msg = &MsgCreatePermanentLockedAccount{}
	_ sdk.Msg = &MsgCreatePeriodicVestingAccount{}
)

const (
	TypeMsgCreateClawbackVestingAccount = "create_clawback_vesting_account"
	TypeMsgClawback                     = "clawback"
	// TypeMsgCreatePermanentLockedAccount defines the type value for a MsgCreatePermanentLockedAccount.
	TypeMsgCreatePermanentLockedAccount = "msg_create_permanent_locked_account"
	// TypeMsgCreatePeriodicVestingAccount defines the type value for a MsgCreateVestingAccount.
	TypeMsgCreatePeriodicVestingAccount = "msg_create_periodic_vesting_account"
)

// NewMsgCreateClawbackVestingAccount creates new instance of MsgCreateClawbackVestingAccount
func NewMsgCreateClawbackVestingAccount(
	fromAddr, toAddr sdk.AccAddress,
	startTime time.Time,
	lockupPeriods,
	vestingPeriods sdkvesting.Periods,
	merge bool,
) *MsgCreateClawbackVestingAccount {
	return &MsgCreateClawbackVestingAccount{
		FromAddress:    fromAddr.String(),
		ToAddress:      toAddr.String(),
		StartTime:      startTime,
		LockupPeriods:  lockupPeriods,
		VestingPeriods: vestingPeriods,
		Merge:          merge,
	}
}

// Route returns the name of the module
func (msg MsgCreateClawbackVestingAccount) Route() string { return RouterKey }

// Type returns the the action
func (msg MsgCreateClawbackVestingAccount) Type() string { return TypeMsgCreateClawbackVestingAccount }

// ValidateBasic runs stateless checks on the message
func (msg MsgCreateClawbackVestingAccount) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid from address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid to address")
	}

	lockupCoins := sdk.NewCoins()
	for i, period := range msg.LockupPeriods {
		if period.Length < 1 {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid period length of %d in period %d, length must be greater than 0", period.Length, i)
		}
		lockupCoins = lockupCoins.Add(period.Amount...)
	}

	vestingCoins := sdk.NewCoins()
	for i, period := range msg.VestingPeriods {
		if period.Length < 1 {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid period length of %d in period %d, length must be greater than 0", period.Length, i)
		}
		vestingCoins = vestingCoins.Add(period.Amount...)
	}

	// If both schedules are present, the must describe the same total amount.
	// IsEqual can panic, so use (a == b) <=> (a <= b && b <= a).
	if len(msg.LockupPeriods) > 0 && len(msg.VestingPeriods) > 0 &&
		!(lockupCoins.IsAllLTE(vestingCoins) && vestingCoins.IsAllLTE(lockupCoins)) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "vesting and lockup schedules must have same total coins")
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg *MsgCreateClawbackVestingAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgCreateClawbackVestingAccount) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{from}
}

// NewMsgClawbackcreates new instance of MsgClawback. The dest_address may be
// nil - defaulting to the funder.
func NewMsgClawback(funder, addr, dest sdk.AccAddress) *MsgClawback {
	destString := ""
	if dest != nil {
		destString = dest.String()
	}
	return &MsgClawback{
		FunderAddress:  funder.String(),
		AccountAddress: addr.String(),
		DestAddress:    destString,
	}
}

// Route returns the message route for a MsgClawback.
func (msg MsgClawback) Route() string { return RouterKey }

// Type returns the message type for a MsgClawback.
func (msg MsgClawback) Type() string { return TypeMsgClawback }

// ValidateBasic runs stateless checks on the message
func (msg MsgClawback) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.GetFunderAddress()); err != nil {
		return sdkerrors.Wrapf(err, "invalid funder address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.GetAccountAddress()); err != nil {
		return sdkerrors.Wrapf(err, "invalid account address")
	}

	if msg.GetDestAddress() != "" {
		if _, err := sdk.AccAddressFromBech32(msg.GetDestAddress()); err != nil {
			return sdkerrors.Wrapf(err, "invalid dest address")
		}
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg *MsgClawback) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgClawback) GetSigners() []sdk.AccAddress {
	funder, err := sdk.AccAddressFromBech32(msg.FunderAddress)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{funder}
}

// NewMsgCreatePermanentLockedAccount returns a reference to a new MsgCreatePermanentLockedAccount.
//nolint:interfacer
func NewMsgCreatePermanentLockedAccount(fromAddr, toAddr sdk.AccAddress, amount sdk.Coins) *MsgCreatePermanentLockedAccount {
	return &MsgCreatePermanentLockedAccount{
		FromAddress: fromAddr.String(),
		ToAddress:   toAddr.String(),
		Amount:      amount,
	}
}

// Route returns the message route for a MsgCreatePermanentLockedAccount.
func (msg MsgCreatePermanentLockedAccount) Route() string { return RouterKey }

// Type returns the message type for a MsgCreatePermanentLockedAccount.
func (msg MsgCreatePermanentLockedAccount) Type() string { return TypeMsgCreatePermanentLockedAccount }

// ValidateBasic Implements Msg.
func (msg MsgCreatePermanentLockedAccount) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid recipient address: %s", err)
	}

	if len(msg.Amount) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "cannot submit empty coins")
	}

	if err := msg.Amount.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid coins amount")
	}

	return nil
}

// GetSignBytes returns the bytes all expected signers must sign over for a
// MsgCreatePermanentLockedAccount.
func (msg MsgCreatePermanentLockedAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgCreatePermanentLockedAccount.
func (msg MsgCreatePermanentLockedAccount) GetSigners() []sdk.AccAddress {
	from, _ := sdk.AccAddressFromBech32(msg.FromAddress)
	return []sdk.AccAddress{from}
}

// NewMsgCreatePeriodicVestingAccount returns a reference to a new MsgCreatePeriodicVestingAccount.
//nolint:interfacer
func NewMsgCreatePeriodicVestingAccount(fromAddr, toAddr sdk.AccAddress, startTime int64, periods sdkvesting.Periods) *MsgCreatePeriodicVestingAccount {
	return &MsgCreatePeriodicVestingAccount{
		FromAddress:    fromAddr.String(),
		ToAddress:      toAddr.String(),
		StartTime:      startTime,
		VestingPeriods: periods,
	}
}

// Route returns the message route for a MsgCreatePeriodicVestingAccount.
func (msg MsgCreatePeriodicVestingAccount) Route() string { return RouterKey }

// Type returns the message type for a MsgCreatePeriodicVestingAccount.
func (msg MsgCreatePeriodicVestingAccount) Type() string { return TypeMsgCreatePeriodicVestingAccount }

// GetSigners returns the expected signers for a MsgCreatePeriodicVestingAccount.
func (msg MsgCreatePeriodicVestingAccount) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

// GetSignBytes returns the bytes all expected signers must sign over for a
// MsgCreatePeriodicVestingAccount.
func (msg MsgCreatePeriodicVestingAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic Implements Msg.
func (msg MsgCreatePeriodicVestingAccount) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid recipient address: %s", err)
	}

	if msg.StartTime < 1 {
		return fmt.Errorf("invalid start time of %d, length must be greater than 0", msg.StartTime)
	}

	for i, period := range msg.VestingPeriods {
		if period.Length < 1 {
			return fmt.Errorf("invalid period length of %d in period %d, length must be greater than 0", period.Length, i)
		}
		if len(period.Amount) == 0 {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "cannot submit empty amount in period %d", i)
		}
		if err := period.Amount.Validate(); err != nil {
			return sdkerrors.Wrapf(err, "invalid amount in period %d", i)
		}
	}

	return nil
}
