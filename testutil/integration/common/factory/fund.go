// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package factory

import (
	"fmt"
	"slices"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// FundAccount funds the given account with the given amount of coins.
func (tf *IntegrationTxFactory) FundAccount(addr sdk.AccAddress, coins sdk.Coins) error {
	// validate that required coins are supported in the test network
	if err := tf.validateDenoms(coins); err != nil {
		return err
	}

	funder := tf.network.GetFunder()
	bankmsg := banktypes.NewMsgSend(
		funder.Address,
		addr,
		coins,
	)
	txArgs := CosmosTxArgs{Msgs: []sdk.Msg{bankmsg}}
	txRes, err := tf.ExecuteCosmosTx(funder.PrivKey, txArgs)

	if err != nil {
		return err
	}

	if txRes.Code != 0 {
		return fmt.Errorf("transaction returned non-zero code %d", txRes.Code)
	}

	// commit the changes
	return tf.network.NextBlock()
}

// FundAccountWithBaseDenom funds the given account with the given amount of the network's
// base denomination.
func (tf *IntegrationTxFactory) FundAccountWithBaseDenom(addr sdk.AccAddress, amount sdkmath.Int) error {
	return tf.FundAccount(addr, sdk.NewCoins(sdk.NewCoin(tf.network.GetDenom(), amount)))
}

func (tf *IntegrationTxFactory) validateDenoms(coins sdk.Coins) error {
	for _, c := range coins {
		if c.Denom == tf.network.GetDenom() {
			continue
		}
		if slices.Contains(tf.network.GetOtherDenoms(), c.Denom) {
			continue
		}
		return fmt.Errorf("denomination %s does not exist in the testing network", c.Denom)
	}
	return nil
}
