// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package utils

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	cmnfactory "github.com/evmos/evmos/v16/testutil/integration/common/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
)

// Delegate on behalf of the account associated with the given private key.
// The defined amount will delegated to the specified validator.
// The validator address should be in the format `evmosvaloper1...`.
func Delegate(tf cmnfactory.TxFactory, delegatorPriv cryptotypes.PrivKey, validatorAddr string, amount sdk.Coin) error {
	delegatorAccAddr := sdk.AccAddress(delegatorPriv.PubKey().Address())

	msgDelegate := stakingtypes.NewMsgDelegate(
		delegatorAccAddr.String(),
		validatorAddr,
		amount,
	)

	resp, err := tf.ExecuteCosmosTx(delegatorPriv, cmnfactory.CosmosTxArgs{
		Msgs: []sdk.Msg{msgDelegate},
	})

	if resp.Code != 0 {
		err = fmt.Errorf("received error code %d on Delegate transaction. Logs: %s", resp.Code, resp.Log)
	}

	return err
}

// WaitToAccrueRewards is a helper function that waits for rewards to
// accumulate up to a specified expected amount
func WaitToAccrueRewards(n network.Network, gh grpc.Handler, delegatorAddr string, expRewards sdk.DecCoins) (sdk.DecCoins, error) {
	var (
		err     error
		lapse   = time.Hour * 24 * 7 // one week
		rewards = sdk.DecCoins{}
	)

	expAmt := expRewards.AmountOf(n.GetDenom())
	for rewards.AmountOf(n.GetDenom()).LT(expAmt) {
		rewards, err = checkRewardsAfter(n, gh, delegatorAddr, lapse)
		if err != nil {
			return nil, errorsmod.Wrap(err, "error checking rewards")
		}
	}

	return rewards, err
}

// checkRewardsAfter is a helper function that checks the accrued rewards
// after the provided time lapse
func checkRewardsAfter(n network.Network, gh grpc.Handler, delegatorAddr string, lapse time.Duration) (sdk.DecCoins, error) {
	err := n.NextBlockAfter(lapse)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to commit block after voting period ends")
	}

	res, err := gh.GetDelegationTotalRewards(delegatorAddr)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "error while querying for delegation rewards")
	}

	return res.Total, nil
}

// WaitToAccrueCommission is a helper function that waits for commission to
// accumulate up to a specified expected amount
func WaitToAccrueCommission(n network.Network, gh grpc.Handler, validatorAddr string, expCommission sdk.DecCoins) (sdk.DecCoins, error) {
	var (
		err        error
		lapse      = time.Hour * 24 * 7 // one week
		commission = sdk.DecCoins{}
	)

	expAmt := expCommission.AmountOf(n.GetDenom())
	for commission.AmountOf(n.GetDenom()).LT(expAmt) {
		commission, err = checkCommissionAfter(n, gh, validatorAddr, lapse)
		if err != nil {
			return nil, errorsmod.Wrap(err, "error checking comission")
		}
	}

	return commission, err
}

// checkCommissionAfter is a helper function that checks the accrued commission
// after the provided time lapse
func checkCommissionAfter(n network.Network, gh grpc.Handler, valAddr string, lapse time.Duration) (sdk.DecCoins, error) {
	err := n.NextBlockAfter(lapse)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to commit block after voting period ends")
	}

	res, err := gh.GetValidatorCommission(valAddr)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "error while querying for delegation rewards")
	}

	return res.Commission.Commission, nil
}
