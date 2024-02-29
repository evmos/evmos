// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package utils

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"golang.org/x/exp/slices"

	cmnfactory "github.com/evmos/evmos/v16/testutil/integration/common/factory"
	"github.com/evmos/evmos/v16/testutil/integration/common/grpc"
	cmnnet "github.com/evmos/evmos/v16/testutil/integration/common/network"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
)

// FundAccount funds the given account with the given amount of coins.
func FundAccount(tf cmnfactory.TxFactory, nw cmnnet.Network, sender keyring.Key, receiver sdk.AccAddress, coins sdk.Coins) error {
	// validate that required coins are supported in the test network
	if err := validateDenoms(nw, coins); err != nil {
		return err
	}

	bankmsg := banktypes.NewMsgSend(
		sender.AccAddr,
		receiver,
		coins,
	)
	txArgs := cmnfactory.CosmosTxArgs{Msgs: []sdk.Msg{bankmsg}}
	txRes, err := tf.ExecuteCosmosTx(sender.Priv, txArgs)
	if err != nil {
		return err
	}

	if txRes.Code != 0 {
		return fmt.Errorf("transaction returned non-zero code %d", txRes.Code)
	}

	// commit the changes
	return nw.NextBlock()
}

// FundAccountWithBaseDenom funds the given account with the given amount of the network's
// base denomination.
func FundAccountWithBaseDenom(tf cmnfactory.TxFactory, nw cmnnet.Network, sender keyring.Key, receiver sdk.AccAddress, amount math.Int) error {
	return FundAccount(tf, nw, sender, receiver, sdk.NewCoins(sdk.NewCoin(nw.GetDenom(), amount)))
}

// CheckBalances checks that the given accounts have the expected balances and
// returns an error if that is not the case.
func CheckBalances(handler grpc.Handler, balances []banktypes.Balance) error {
	for _, balance := range balances {
		addr := balance.GetAddress()
		for _, coin := range balance.GetCoins() {
			balance, err := handler.GetBalance(sdk.AccAddress(addr), coin.Denom)
			if err != nil {
				return err
			}

			if !balance.Balance.IsEqual(coin) {
				return fmt.Errorf(
					"expected balance %s, got %s for address %s",
					coin, balance.Balance, addr,
				)
			}
		}
	}

	return nil
}

func validateDenoms(nw cmnnet.Network, coins sdk.Coins) error {
	for _, c := range coins {
		if c.Denom == nw.GetDenom() {
			continue
		}
		if slices.Contains(nw.GetOtherDenoms(), c.Denom) {
			continue
		}
		return fmt.Errorf("denomination %s does not exist in the testing network", c.Denom)
	}
	return nil
}
