// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package utils

import (
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	commonfactory "github.com/evmos/evmos/v16/testutil/integration/common/factory"
	"github.com/evmos/evmos/v16/testutil/integration/common/grpc"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
)

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

// FundAccount is a helper function that funds a new account using
// the private key of an existing account with funds
func FundAccount(tf factory.TxFactory, funderPriv cryptotypes.PrivKey, addressToFund string, amount sdk.Coins) error {
	funder := sdk.AccAddress(funderPriv.PubKey().Address())

	msg := &banktypes.MsgSend{
		FromAddress: funder.String(),
		ToAddress:   addressToFund,
		Amount:      amount,
	}
	res, err := tf.ExecuteCosmosTx(
		funderPriv,
		commonfactory.CosmosTxArgs{
			Msgs: []sdk.Msg{msg},
		},
	)
	if err != nil {
		return err
	}

	if res.Code != 0 {
		return fmt.Errorf("recevied an error code %d when funding account %s. Logs: %s", res.Code, addressToFund, res.Log)
	}
	return nil
}
