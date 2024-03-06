// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package factory

import (
	"fmt"
	"time"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	commonfactory "github.com/evmos/evmos/v16/testutil/integration/common/factory"
	vestingtypes "github.com/evmos/evmos/v16/x/vesting/types"
)

type VestingTxFactory interface {
	// CreateClawbackVestingAccount is a method to create and broadcast a MsgCreateClawbackVestingAccount
	CreateClawbackVestingAccount(vestingPriv cryptotypes.PrivKey, funderAddr sdk.AccAddress, enableGovClawback bool) error
	// FundVestingAccount is a method to create and broadcast a MsgFundVestingAccount
	FundVestingAccount(funderPriv cryptotypes.PrivKey, vestingAddr sdk.AccAddress, startTime time.Time, lockupPeriods, vestingPeriods sdkvesting.Periods) error
}

type vestingTxFactory struct {
	commonfactory.BaseTxFactory
}

func newVestingTxFactory(bf commonfactory.BaseTxFactory) VestingTxFactory {
	return &vestingTxFactory{bf}
}

// CreateClawbackVestingAccount in the provided address, with the provided
// funder address
func (tf *vestingTxFactory) CreateClawbackVestingAccount(vestingPriv cryptotypes.PrivKey, funderAddr sdk.AccAddress, enableGovClawback bool) error {
	vestingAccAddr := sdk.AccAddress(vestingPriv.PubKey().Address())

	msg := vestingtypes.NewMsgCreateClawbackVestingAccount(
		funderAddr,
		vestingAccAddr,
		enableGovClawback,
	)

	resp, err := tf.ExecuteCosmosTx(vestingPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{msg},
	})

	if resp.Code != 0 {
		err = fmt.Errorf("received error code %d on CreateClawbackVestingAccount transaction. Logs: %s", resp.Code, resp.Log)
	}

	return err
}

// FundVestingAccount at the provided address with the given vesting schedules.
func (tf *vestingTxFactory) FundVestingAccount(funderPriv cryptotypes.PrivKey, vestingAddr sdk.AccAddress, startTime time.Time, lockupPeriods, vestingPeriods sdkvesting.Periods) error {
	funderAccAddr := sdk.AccAddress(funderPriv.PubKey().Address())

	msg := vestingtypes.NewMsgFundVestingAccount(
		funderAccAddr,
		vestingAddr,
		startTime,
		lockupPeriods,
		vestingPeriods,
	)

	resp, err := tf.ExecuteCosmosTx(funderPriv, commonfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{msg},
	})

	if resp.Code != 0 {
		err = fmt.Errorf("received error code %d on FundVestingAccount transaction. Logs: %s", resp.Code, resp.Log)
	}

	return err
}
