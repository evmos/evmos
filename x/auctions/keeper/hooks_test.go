// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/evmos/evmos/v19/utils"
	epochstypes "github.com/evmos/evmos/v19/x/epochs/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	testkeyring "github.com/evmos/evmos/v19/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	testutiltx "github.com/evmos/evmos/v19/testutil/tx"
	"github.com/evmos/evmos/v19/x/auctions/types"
)

func TestHookAfterEpochEnd(t *testing.T) {
	var network *testnetwork.UnitTestNetwork

	validSenderAddr, _ := testutiltx.NewAccAddressAndKey()
	// Token used in the bid.
	bidAmount := sdk.NewCoin(utils.BaseDenom, sdk.NewInt(1))
	// Tokens used to simulate auctioned tokens.
	additionalCoins := []sdk.Coin{
		sdk.NewInt64Coin("atoken", 1),
		sdk.NewInt64Coin("btoken", 1),
	}
	// Coin used to simulate accrued tokens to auction.
	feeCoin := sdk.NewInt64Coin("fee", 1)

	// Value required in the hook but not used.
	unusedEpochNumber := int64(0)

	testCases := []struct {
		name            string
		malleate        func()
		epochIdentifier string
		postCheck       func()
	}{
		{
			name: "pass with prize distributed",
			malleate: func() {
				round := network.App.AuctionsKeeper.GetRound(network.GetContext())
				assert.Equal(t, uint64(0), round)

				// Inflation should be enabled
				params := network.App.AuctionsKeeper.GetParams(network.GetContext())
				params.EnableAuction = true
				updateParamsMsg := types.MsgUpdateParams{
					Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
					Params:    params,
				}
				_, err := network.App.AuctionsKeeper.UpdateParams(network.GetContext(), &updateParamsMsg)
				assert.NoError(t, err, "failed to update auctions params")

				// There should be an highest bid.
				setHighestBid(t, network, validSenderAddr, bidAmount)
				// We have to mint the bidded coins and also additional.
				mintCoinsToModuleAccount(t, network, types.ModuleName, sdk.NewCoins(append(additionalCoins, bidAmount)...))
				// We have to mint coins for the auction collector to simulate accrued fees.
				mintCoinsToModuleAccount(t, network, types.AuctionCollectorName, sdk.NewCoins(feeCoin))
			},
			epochIdentifier: epochstypes.WeekEpochID,
			postCheck: func() {
				// Check that bid winner received the tokens.
				balanceCoins := network.App.BankKeeper.GetAllBalances(network.GetContext(), validSenderAddr)
				found := assertCoinSetCointainCoinsSet(balanceCoins, additionalCoins)
				assert.True(t, found, "expected sender to have received funds")

				moduleAccountAddress := network.App.AccountKeeper.GetModuleAddress(types.ModuleName)
				moduleAccountBalance := network.App.BankKeeper.GetAllBalances(network.GetContext(), moduleAccountAddress)
				found = assertCoinSetCointainCoinsSet(moduleAccountBalance, sdk.NewCoins(feeCoin))
				assert.True(t, found, "expected module account to have received funds")

				auctionCollectorAddress := network.App.AccountKeeper.GetModuleAddress(types.AuctionCollectorName)
				auctionCollectorBalance := network.App.BankKeeper.GetAllBalances(network.GetContext(), auctionCollectorAddress)
				assert.Equal(t, len(auctionCollectorBalance), 0, "expected auction collector to be empty")

				// Check that round increased by one.
				round := network.App.AuctionsKeeper.GetRound(network.GetContext())
				assert.Equal(t, uint64(1), round)

				// Check that the bid has been removed.
				bid := network.App.AuctionsKeeper.GetHighestBid(network.GetContext())
				assert.Equal(t, "", bid.Sender, "expected a different bid sender")
				assert.Equal(t, sdk.NewInt64Coin(utils.BaseDenom, 0), bid.Amount, "expected a different bid amount")
			},
		},
		{
			name: "no op if not weekly epoch",
			malleate: func() {
				round := network.App.AuctionsKeeper.GetRound(network.GetContext())
				assert.Equal(t, uint64(0), round)

				// Inflation should be enabled
				params := network.App.AuctionsKeeper.GetParams(network.GetContext())
				params.EnableAuction = true
				updateParamsMsg := types.MsgUpdateParams{
					Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
					Params:    params,
				}
				_, err := network.App.AuctionsKeeper.UpdateParams(network.GetContext(), &updateParamsMsg)
				assert.NoError(t, err, "failed to update auctions params")

				// There should be an highest bid.
				setHighestBid(t, network, validSenderAddr, bidAmount)
				// We have to mint the bidded coins and also additional.
				mintCoinsToModuleAccount(t, network, types.ModuleName, sdk.NewCoins(append(additionalCoins, bidAmount)...))
				// We have to mint coins for the auction collector to simulate accrued fees.
				mintCoinsToModuleAccount(t, network, types.AuctionCollectorName, sdk.NewCoins(feeCoin))
			},
			epochIdentifier: epochstypes.DayEpochID,
			postCheck: func() {
				balanceCoins := network.App.BankKeeper.GetAllBalances(network.GetContext(), validSenderAddr)
				assert.Equal(t, len(balanceCoins), 0, "expected not coins in the balance")
				round := network.App.AuctionsKeeper.GetRound(network.GetContext())
				assert.Equal(t, uint64(0), round, "expected round to not advance")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keyring := testkeyring.New(1)
			network = testnetwork.NewUnitTestNetwork(
				testnetwork.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
			)

			tc.malleate()

			network.App.AuctionsKeeper.AfterEpochEnd(network.GetContext(), tc.epochIdentifier, unusedEpochNumber)

			tc.postCheck()
		})
	}
}
