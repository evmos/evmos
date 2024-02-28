package factory

import (
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func (tf *IntegrationTxFactory) SetWithdrawAddress(delegatorPriv cryptotypes.PrivKey, withdrawerAddr sdk.AccAddress) error {
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
func (tf *IntegrationTxFactory) WithdrawDelegationRewards(delegatorPriv cryptotypes.PrivKey, validatorAddr string) error {
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

func (tf *IntegrationTxFactory) WithdrawValidatorCommission(validatorPriv cryptotypes.PrivKey) error {
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
