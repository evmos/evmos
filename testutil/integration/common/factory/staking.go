// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package factory

import (
	"fmt"

	"cosmossdk.io/math"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type StakingTxFactory interface {
	// Delegate is a method to create and broadcast a MsgDelegate
	Delegate(delegatorPriv cryptotypes.PrivKey, validatorAddr string, amount sdk.Coin) error
	// CreateValidator is a method to create and broadcast a MsgCreateValidator
	CreateValidator(operatorPriv cryptotypes.PrivKey, pubKey cryptotypes.PubKey, selfDelegation sdk.Coin, description stakingtypes.Description, commission stakingtypes.CommissionRates, minSelfDelegation math.Int) error
}

type stakingTxFactory struct {
	BaseTxFactory
}

func newStakingTxFactory(bf BaseTxFactory) StakingTxFactory {
	return &stakingTxFactory{bf}
}

// Delegate on behalf of the account associated with the given private key.
// The defined amount will delegated to the specified validator.
// The validator address should be in the format `evmosvaloper1...`.
func (tf *stakingTxFactory) Delegate(delegatorPriv cryptotypes.PrivKey, validatorAddr string, amount sdk.Coin) error {
	delegatorAccAddr := sdk.AccAddress(delegatorPriv.PubKey().Address())

	msgDelegate := stakingtypes.NewMsgDelegate(
		delegatorAccAddr.String(),
		validatorAddr,
		amount,
	)

	resp, err := tf.ExecuteCosmosTx(delegatorPriv, CosmosTxArgs{
		Msgs: []sdk.Msg{msgDelegate},
	})

	if resp.Code != 0 {
		err = fmt.Errorf("received error code %d on Delegate transaction. Logs: %s", resp.Code, resp.Log)
	}

	return err
}

// CreateValidator executes the transaction to create a validator
// with the parameters specified
func (tf *stakingTxFactory) CreateValidator(operatorPriv cryptotypes.PrivKey, pubKey cryptotypes.PubKey, selfDelegation sdk.Coin, description stakingtypes.Description, commission stakingtypes.CommissionRates, minSelfDelegation math.Int) error {
	operatorAccAddr := sdk.ValAddress(operatorPriv.PubKey().Address())

	msgCreateValidator, err := stakingtypes.NewMsgCreateValidator(
		operatorAccAddr.String(),
		pubKey,
		selfDelegation,
		description,
		commission,
		minSelfDelegation,
	)
	if err != nil {
		return err
	}

	resp, err := tf.ExecuteCosmosTx(operatorPriv, CosmosTxArgs{
		Msgs: []sdk.Msg{msgCreateValidator},
	})

	if resp.Code != 0 {
		err = fmt.Errorf("received error code %d on CreateValidator transaction. Logs: %s", resp.Code, resp.Log)
	}

	return err
}
