// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	testutil "github.com/evmos/evmos/v19/testutil"
	testnetwork "github.com/evmos/evmos/v19/testutil/integration/evmos/network"
)

// setHighestBid is an util function to call SetHighestBid to store a bid and
// vefiy that it has correctly be inserted.
func setHighestBid(t *testing.T, network *testnetwork.UnitTestNetwork, bidSender sdk.AccAddress, bidAmount sdk.Coin) {
	network.App.AuctionsKeeper.SetHighestBid(network.GetContext(), bidSender.String(), bidAmount)

	bid := network.App.AuctionsKeeper.GetHighestBid(network.GetContext())
	assert.Equal(t, bidSender.String(), bid.Sender, "expected a different bid sender")
	assert.Equal(t, bidAmount, bid.Amount, "expected a different bid amount")
}

// mintCoinsToModuleAccount is an util function to mint coins to the module account
// and verify the balance after. The verification check if the accounts balance is
// a superset of the minted coins.
func mintCoinsToModuleAccount(t *testing.T, network *testnetwork.UnitTestNetwork, moduleName string, coins sdk.Coins) {
	err := testutil.FundModuleAccount(
		network.GetContext(),
		network.App.BankKeeper,
		moduleName,
		coins,
	)
	assert.NoError(t, err, "expected no error while minting tokens for the module account")

	moduleAddress := network.App.AccountKeeper.GetModuleAddress(moduleName)
	balance := network.App.BankKeeper.GetAllBalances(network.GetContext(), moduleAddress)

	found := assertCoinSetCointainCoinsSet(balance, coins)
	assert.True(t, found, "expected a different balance for auctions module after minting tokens")
}

// assertCoinSetCointainsCoinSet check if coin set A contains coin set B.
func assertCoinSetCointainCoinsSet(coinSetA, coinSetB sdk.Coins) bool {
	expFound := len(coinSetB)
	var currFound int
	for _, coinA := range coinSetA {
		for _, coinB := range coinSetB {
			if coinA.Denom == coinB.Denom && coinA.Amount.GTE(coinB.Amount) {
				currFound++
			}
		}
	}

	return currFound == expFound
}
