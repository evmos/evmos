// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package vesting

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"

	cosmosvestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
	vestingtypes "github.com/evmos/evmos/v15/x/vesting/types"
)

// LockupPeriods is a struct used to parse the LockupPeriods parameter
// used as input in the MsgCreateClawbackVestingAccount
type LockupPeriods struct {
	LockupPeriods []Period
}

// VestingPeriods is a  struct used to parse the VestingPeriods parameter
// used as input in the MsgCreateClawbackVestingAccount
//
//nolint:revive
type VestingPeriods struct {
	VestingPeriods []Period
}

// Period represents a period of time with a specific amount of coins
type Period struct {
	Length int64
	Amount []cmn.Coin
}

// NewMsgCreateClawbackVestingAccount creates a new MsgCreateClawbackVestingAccount instance.
func NewMsgCreateClawbackVestingAccount(args []interface{}) (*vestingtypes.MsgCreateClawbackVestingAccount, common.Address, common.Address, error) {
	funderAddress, vestingAddress, err := validateBasicArgs(args, 3)
	if err != nil {
		return nil, common.Address{}, common.Address{}, err
	}

	enableGovClawback, ok := args[2].(bool)
	if !ok {
		return nil, common.Address{}, common.Address{}, fmt.Errorf(cmn.ErrInvalidType, "enableGovClawback", true, args[2])
	}

	msg := &vestingtypes.MsgCreateClawbackVestingAccount{
		FunderAddress:     sdk.AccAddress(funderAddress.Bytes()).String(),
		VestingAddress:    sdk.AccAddress(vestingAddress.Bytes()).String(),
		EnableGovClawback: enableGovClawback,
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, common.Address{}, common.Address{}, err
	}

	return msg, funderAddress, vestingAddress, nil
}

// NewMsgFundVestingAccount creates a new MsgFundVestingAccount instance.
func NewMsgFundVestingAccount(args []interface{}, method *abi.Method) (*vestingtypes.MsgFundVestingAccount, common.Address, common.Address, *LockupPeriods, *VestingPeriods, error) {
	funderAddress, vestingAddress, err := validateBasicArgs(args, 5)
	if err != nil {
		return nil, common.Address{}, common.Address{}, nil, nil, err
	}

	startTime, ok := args[2].(uint64)
	if !ok {
		return nil, common.Address{}, common.Address{}, nil, nil, fmt.Errorf(cmn.ErrInvalidType, "startTime", uint64(0), args[2])
	}

	startTimeTimestamp := time.Unix(int64(startTime), 0)

	var lockupPeriodsInput LockupPeriods
	lockupPeriod := abi.Arguments{method.Inputs[3]}
	if err := lockupPeriod.Copy(&lockupPeriodsInput, []interface{}{args[3]}); err != nil {
		return nil, common.Address{}, common.Address{}, nil, nil, fmt.Errorf("error while unpacking args to lockupPeriods struct: %s", err)
	}

	var vestingPeriodsInput VestingPeriods
	vestingPeriod := abi.Arguments{method.Inputs[4]}
	if err := vestingPeriod.Copy(&vestingPeriodsInput, []interface{}{args[4]}); err != nil {
		return nil, common.Address{}, common.Address{}, nil, nil, fmt.Errorf("error while unpacking args to vestingPeriods struct: %s", err)
	}

	vestingCosmosPeriods := createCosmosPeriodsFromPeriod(vestingPeriodsInput.VestingPeriods)
	lockupCosmosPeriods := createCosmosPeriodsFromPeriod(lockupPeriodsInput.LockupPeriods)
	msg := &vestingtypes.MsgFundVestingAccount{
		FunderAddress:  sdk.AccAddress(funderAddress.Bytes()).String(),
		VestingAddress: sdk.AccAddress(vestingAddress.Bytes()).String(),
		StartTime:      startTimeTimestamp,
		LockupPeriods:  lockupCosmosPeriods,
		VestingPeriods: vestingCosmosPeriods,
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, common.Address{}, common.Address{}, nil, nil, err
	}

	return msg, funderAddress, vestingAddress, &lockupPeriodsInput, &vestingPeriodsInput, nil
}

// NewMsgClawback creates a new MsgClawback instance.
func NewMsgClawback(args []interface{}) (*vestingtypes.MsgClawback, common.Address, common.Address, common.Address, error) {
	funderAddress, accountAddress, err := validateBasicArgs(args, 3)
	if err != nil {
		return nil, common.Address{}, common.Address{}, common.Address{}, err
	}

	destAddress, ok := args[2].(common.Address)
	if !ok {
		return nil, common.Address{}, common.Address{}, common.Address{}, fmt.Errorf(cmn.ErrInvalidType, "startTime", "Address", args[2])
	}

	msg := &vestingtypes.MsgClawback{
		FunderAddress:  sdk.AccAddress(funderAddress.Bytes()).String(),
		AccountAddress: sdk.AccAddress(accountAddress.Bytes()).String(),
		DestAddress:    sdk.AccAddress(destAddress.Bytes()).String(),
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, common.Address{}, common.Address{}, common.Address{}, err
	}

	return msg, funderAddress, accountAddress, destAddress, nil
}

// NewMsgUpdateVestingFunder creates a new MsgUpdateVestingFunder instance.
func NewMsgUpdateVestingFunder(args []interface{}) (*vestingtypes.MsgUpdateVestingFunder, common.Address, common.Address, common.Address, error) {
	if len(args) != 3 {
		return nil, common.Address{}, common.Address{}, common.Address{}, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 3, len(args))
	}

	funderAddress, ok := args[0].(common.Address)
	if !ok {
		return nil, common.Address{}, common.Address{}, common.Address{}, fmt.Errorf(cmn.ErrInvalidType, "funderAddress", "Address", args[0])
	}

	newFunderAddress, ok := args[1].(common.Address)
	if !ok {
		return nil, common.Address{}, common.Address{}, common.Address{}, fmt.Errorf(cmn.ErrInvalidType, "newFunderAddress", "Address", args[1])
	}

	vestingAddress, ok := args[2].(common.Address)
	if !ok {
		return nil, common.Address{}, common.Address{}, common.Address{}, fmt.Errorf(cmn.ErrInvalidType, "vestingAddress", "Address", args[2])
	}

	msg := &vestingtypes.MsgUpdateVestingFunder{
		FunderAddress:    sdk.AccAddress(funderAddress.Bytes()).String(),
		NewFunderAddress: sdk.AccAddress(newFunderAddress.Bytes()).String(),
		VestingAddress:   sdk.AccAddress(vestingAddress.Bytes()).String(),
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, common.Address{}, common.Address{}, common.Address{}, err
	}

	return msg, funderAddress, newFunderAddress, vestingAddress, nil
}

// NewMsgConvertVestingAccount creates a new MsgConvertVestingAccount instance.
func NewMsgConvertVestingAccount(args []interface{}) (*vestingtypes.MsgConvertVestingAccount, common.Address, error) {
	if len(args) != 1 {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	vestingAddress, ok := args[0].(common.Address)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidType, "vestingAddress", "Address", args[0])
	}

	msg := &vestingtypes.MsgConvertVestingAccount{
		VestingAddress: sdk.AccAddress(vestingAddress.Bytes()).String(),
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, common.Address{}, err
	}

	return msg, vestingAddress, nil
}

// NewBalancesRequest creates a new QueryBalancesRequest instance.
func NewBalancesRequest(args []interface{}) (*vestingtypes.QueryBalancesRequest, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	address, ok := args[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "vestingAddress", "Address", args[0])
	}

	msg := &vestingtypes.QueryBalancesRequest{
		Address: sdk.AccAddress(address.Bytes()).String(),
	}

	return msg, nil
}

// validateBasicArgs validates the basic arguments and length of the provided arguments.
func validateBasicArgs(args []interface{}, expectedLength int) (common.Address, common.Address, error) {
	if len(args) != expectedLength {
		return common.Address{}, common.Address{}, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, expectedLength, len(args))
	}

	funderAddress, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, fmt.Errorf(cmn.ErrInvalidType, "funderAddress", "address", args[0])
	}

	vestingAddress, ok := args[1].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, fmt.Errorf(cmn.ErrInvalidType, "vestingAddress", "address", args[0])
	}

	return funderAddress, vestingAddress, nil
}

// createCosmosPeriodsFromPeriod creates a cosmosvestingtypes.Period slice from a Period slice.
func createCosmosPeriodsFromPeriod(inputPeriods []Period) cosmosvestingtypes.Periods {
	periods := make(cosmosvestingtypes.Periods, len(inputPeriods))
	for i, period := range inputPeriods {
		amount := make([]sdk.Coin, len(period.Amount))
		for j, coin := range period.Amount {
			amount[j] = sdk.NewCoin(coin.Denom, sdk.NewIntFromBigInt(coin.Amount))
		}

		periods[i] = cosmosvestingtypes.Period{
			Length: period.Length,
			Amount: amount,
		}
	}

	return periods
}

// BalancesOutput represents the balances of a ClawbackVestingAccount
type BalancesOutput struct {
	Locked   []cmn.Coin
	Unvested []cmn.Coin
	Vested   []cmn.Coin
}

// FromResponse populates the BalancesOutput from a QueryBalancesResponse.
func (bo *BalancesOutput) FromResponse(res *vestingtypes.QueryBalancesResponse) *BalancesOutput {
	bo.Locked = cmn.NewCoinsResponse(res.Locked)
	bo.Unvested = cmn.NewCoinsResponse(res.Unvested)
	bo.Vested = cmn.NewCoinsResponse(res.Vested)
	return bo
}

// ClawbackOutput represents the clawed back coins from a Clawback transaction.
type ClawbackOutput struct {
	Coins []cmn.Coin
}

// FromResponse populates the ClawbackOutput from a QueryClawbackResponse.
func (co *ClawbackOutput) FromResponse(res *vestingtypes.MsgClawbackResponse) *ClawbackOutput {
	co.Coins = cmn.NewCoinsResponse(res.Coins)
	return co
}
