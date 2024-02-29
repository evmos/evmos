// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package utils

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	cmnfactory "github.com/evmos/evmos/v16/testutil/integration/common/factory"
	"github.com/evmos/evmos/v16/testutil/integration/common/grpc"
	cmnnet "github.com/evmos/evmos/v16/testutil/integration/common/network"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
)

// FundAccountWithBaseDenom funds the given account with the given amount of the network's
// base denomination.
func FundAccountWithBaseDenom(tf cmnfactory.CoreTxFactory, nw cmnnet.Network, sender keyring.Key, receiver sdk.AccAddress, amount math.Int) error {
	return tf.FundAccount(sender, receiver, sdk.NewCoins(sdk.NewCoin(nw.GetDenom(), amount)))
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
