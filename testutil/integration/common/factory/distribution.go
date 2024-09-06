// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package factory

import (
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

type DistributionTxFactory interface {
	// SetWithdrawAddress is a method to create and broadcast a MsgSetWithdrawAddress
	SetWithdrawAddress(delegatorPriv cryptotypes.PrivKey, withdrawerAddr sdk.AccAddress) error
	// WithdrawDelegationRewards is a method to create and broadcast a MsgWithdrawDelegationRewards
	WithdrawDelegationRewards(delegatorPriv cryptotypes.PrivKey, validatorAddr string) error
	// WithdrawValidatorCommission is a method to create and broadcast a MsgWithdrawValidatorCommission
	WithdrawValidatorCommission(validatorPriv cryptotypes.PrivKey) error
}

type distributionTxFactory struct {
	BaseTxFactory
}

func newDistrTxFactory(bf BaseTxFactory) DistributionTxFactory {
	return &distributionTxFactory{bf}
}

func (tf *distributionTxFactory) SetWithdrawAddress(delegatorPriv cryptotypes.PrivKey, withdrawerAddr sdk.AccAddress) error {
	delegatorAccAddr := sdk.AccAddress(delegatorPriv.PubKey().Address())

	msg := distrtypes.NewMsgSetWithdrawAddress(
		delegatorAccAddr,
		withdrawerAddr,
	)

	resp, err := tf.ExecuteCosmosTx(delegatorPriv, CosmosTxArgs{
		Msgs: []sdk.Msg{msg},
	})

	if resp.Code != 0 {
		err = fmt.Errorf("received error code %d on SetWithdrawAddress transaction. Logs: %s", resp.Code, resp.Log)
	}

	return err
}

// WithdrawDelegationRewards will withdraw any unclaimed staking rewards for the delegator associated with
// the given private key from the validator.
// The validator address should be in the format `evmosvaloper1...`.
func (tf *distributionTxFactory) WithdrawDelegationRewards(delegatorPriv cryptotypes.PrivKey, validatorAddr string) error {
	delegatorAccAddr := sdk.AccAddress(delegatorPriv.PubKey().Address())

	msg := distrtypes.NewMsgWithdrawDelegatorReward(
		delegatorAccAddr.String(),
		validatorAddr,
	)

	resp, err := tf.ExecuteCosmosTx(delegatorPriv, CosmosTxArgs{
		Msgs: []sdk.Msg{msg},
	})

	if resp.Code != 0 {
		err = fmt.Errorf("received error code %d on WithdrawDelegationRewards transaction. Logs: %s", resp.Code, resp.Log)
	}

	return err
}

func (tf *distributionTxFactory) WithdrawValidatorCommission(validatorPriv cryptotypes.PrivKey) error {
	validatorAddr := sdk.ValAddress(validatorPriv.PubKey().Address())

	msg := distrtypes.NewMsgWithdrawValidatorCommission(
		validatorAddr.String(),
	)

	resp, err := tf.ExecuteCosmosTx(validatorPriv, CosmosTxArgs{
		Msgs: []sdk.Msg{msg},
	})

	if resp.Code != 0 {
		err = fmt.Errorf("received error code %d on WithdrawValidatorCommission transaction. Logs: %s", resp.Code, resp.Log)
	}

	return err
}
