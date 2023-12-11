// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package utils

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
)

// ExpectedBalance is a helper struct to hold the expected balance for a given address.
type ExpectedBalance struct {
	Address sdk.AccAddress
	Coins   sdk.Coins
}

// ExpectedBalances is a slice of ExpectedBalance.
type ExpectedBalances []ExpectedBalance

// ToBalances converts the ExpectedBalances to a slice of banktypes.Balance.
func (eb ExpectedBalances) ToBalances() []banktypes.Balance {
	balances := make([]banktypes.Balance, 0, len(eb))
	for _, expectedBalance := range eb {
		balances = append(balances, banktypes.Balance{
			Address: expectedBalance.Address.String(),
			Coins:   expectedBalance.Coins,
		})
	}

	return balances
}

// CheckBalances checks that the given accounts have the expected balances.
func CheckBalances(handler grpc.Handler, expectedBalances []ExpectedBalance) error {
	for _, expectedBalance := range expectedBalances {
		for _, expectedCoin := range expectedBalance.Coins {
			balance, err := handler.GetBalance(expectedBalance.Address, expectedCoin.Denom)
			if err != nil {
				return err
			}
			if !balance.Balance.IsEqual(expectedCoin) {
				return fmt.Errorf(
					"expected different balance for address %s; expected: %s; got %s",
					expectedBalance.Address.String(), expectedCoin, balance.Balance,
				)
			}
		}
	}

	return nil
}
