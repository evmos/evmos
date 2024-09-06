// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package factory

import (
	"fmt"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/keyring"
)

// FundTxFactory is the interface that wraps the common methods to fund accounts
// via a bank send transaction
type FundTxFactory interface {
	// FundAccount funds the given account with the given amount.
	FundAccount(sender keyring.Key, receiver sdktypes.AccAddress, amount sdktypes.Coins) error
}

// baseTxFactory is the struct of the basic tx factory
// to build and broadcast transactions.
// This is to simulate the behavior of a real user.
type fundTxFactory struct {
	BaseTxFactory
}

// newBaseTxFactory instantiates a new baseTxFactory
func newFundTxFactory(bf BaseTxFactory) FundTxFactory {
	return &fundTxFactory{bf}
}

// FundAccount funds the given account with the given amount of coins.
func (tf *fundTxFactory) FundAccount(sender keyring.Key, receiver sdktypes.AccAddress, coins sdktypes.Coins) error {
	bankmsg := banktypes.NewMsgSend(
		sender.AccAddr,
		receiver,
		coins,
	)
	txArgs := CosmosTxArgs{Msgs: []sdktypes.Msg{bankmsg}}
	txRes, err := tf.ExecuteCosmosTx(sender.Priv, txArgs)
	if err != nil {
		return err
	}

	if txRes.Code != 0 {
		return fmt.Errorf("transaction returned non-zero code %d", txRes.Code)
	}

	return nil
}
