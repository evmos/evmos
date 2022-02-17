package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

var (
	_ sdk.Msg = &MsgCreateClawbackVestingAccount{}
	_ sdk.Msg = &MsgClawback{}
)

const (
	// TypeMsgCreateClawbackVestingAccount defines the type value for a MsgCreateClawbackVestingAcount.
	TypeMsgCreateClawbackVestingAccount = "msg_create_clawback_vesting_account"

	// TypeMsgClawback defines the type value for a MsgClawback.
	TypeMsgClawback = "msg_clawback"
)

// NewMsgCreateClawbackVestingAccount returns a reference to a new MsgCreateClawbackVestingAccount.
//nolint:interfacer
func NewMsgCreateClawbackVestingAccount(fromAddr, toAddr sdk.AccAddress, startTime int64, lockupPeriods, vestingPeriods []sdkvesting.Period, merge bool) *MsgCreateClawbackVestingAccount {
	return &MsgCreateClawbackVestingAccount{
		FromAddress:    fromAddr.String(),
		ToAddress:      toAddr.String(),
		StartTime:      startTime,
		LockupPeriods:  lockupPeriods,
		VestingPeriods: vestingPeriods,
		Merge:          merge,
	}
}

// Route returns the message route for a MsgCreateClawbackVestingAccount.
func (msg MsgCreateClawbackVestingAccount) Route() string { return RouterKey }

// Type returns the message type for a MsgCreateClawbackVestingAccount.
func (msg MsgCreateClawbackVestingAccount) Type() string { return TypeMsgCreateClawbackVestingAccount }

// GetSigners returns the expected signers for a MsgCreateClawbackVestingAccount.
func (msg MsgCreateClawbackVestingAccount) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

// GetSignBytes returns the bytes all expected signers must sign over for a
// MsgCreateClawbackVestingAccount.
func (msg MsgCreateClawbackVestingAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(amino.MustMarshalJSON(&msg))
}

// ValidateBasic Implements Msg.
func (msg MsgCreateClawbackVestingAccount) ValidateBasic() error {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return err
	}
	to, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(from); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %s", err)
	}

	if err := sdk.VerifyAddressFormat(to); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid recipient address: %s", err)
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

// NewMsgClawback returns a reference to a new MsgClawback.
// The dest address may be nil - defaulting to the funder.
//nolint:interfacer
func NewMsgClawback(funder, addr, dest sdk.AccAddress) *MsgClawback {
	destString := ""
	if dest != nil {
		destString = dest.String()
	}
	return &MsgClawback{
		FunderAddress: funder.String(),
		Address:       addr.String(),
		DestAddress:   destString,
	}
}

// Route returns the message route for a MsgClawback.
func (msg MsgClawback) Route() string { return RouterKey }

// Type returns the message type for a MsgClawback.
func (msg MsgClawback) Type() string { return TypeMsgClawback }

// GetSigners returns the expected signers for a MsgClawback.
func (msg MsgClawback) GetSigners() []sdk.AccAddress {
	funder, err := sdk.AccAddressFromBech32(msg.FunderAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{funder}
}

// GetSignBytes returns the bytes all expected signers must sign over for a
// MsgClawback.
func (msg MsgClawback) GetSignBytes() []byte {
	return sdk.MustSortJSON(amino.MustMarshalJSON(&msg))
}

// ValidateBasic Implements Msg.
func (msg MsgClawback) ValidateBasic() error {
	funder, err := sdk.AccAddressFromBech32(msg.GetFunderAddress())
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(funder); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid funder address: %s", err)
	}

	addr, err := sdk.AccAddressFromBech32(msg.GetAddress())
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(addr); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid account address: %s", err)
	}

	if msg.GetDestAddress() != "" {
		dest, err := sdk.AccAddressFromBech32(msg.GetDestAddress())
		if err != nil {
			return err
		}
		if err := sdk.VerifyAddressFormat(dest); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid destination address: %s", err)
		}
	}

	return nil
}
