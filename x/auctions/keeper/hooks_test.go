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
	var (
		network         *testnetwork.UnitTestNetwork
		epochIdentifier string
	)

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
		name      string
		malleate  func()
		expUpdate bool
		postCheck func()
	}{
		{
			name: "pass with prize distributed",
			malleate: func() {
				// Initial state should pass
			},
			expUpdate: true,
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
				assert.Equal(t, sdk.NewInt64Coin(utils.BaseDenom, 0), bid.BidValue, "expected a different bid amount")
			},
		},
		{
			name: "pass if bidded against no funds to receive",
			malleate: func() {
				network.App.BankKeeper.BurnCoins(network.GetContext(), types.ModuleName, additionalCoins)
			},
			expUpdate: true,
			postCheck: func() {
				// Check that bid winner received the tokens.
				balanceCoins := network.App.BankKeeper.GetAllBalances(network.GetContext(), validSenderAddr)
				assert.True(t, len(balanceCoins) == 0, "expected sender to still have empty balance")

				moduleAccountAddress := network.App.AccountKeeper.GetModuleAddress(types.ModuleName)
				moduleAccountBalance := network.App.BankKeeper.GetAllBalances(network.GetContext(), moduleAccountAddress)
				found := assertCoinSetCointainCoinsSet(moduleAccountBalance, sdk.NewCoins(feeCoin))
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
				assert.Equal(t, sdk.NewInt64Coin(utils.BaseDenom, 0), bid.BidValue, "expected a different bid amount")
			},
		},
		{
			name: "no op if not weekly epoch",
			malleate: func() {
				epochIdentifier = epochstypes.DayEpochID
			},
			expUpdate: false,
			postCheck: func() {},
		},
		{
			name: "no op if auctions module is disabled",
			malleate: func() {
				params := network.App.AuctionsKeeper.GetParams(network.GetContext())
				params.EnableAuction = false
				updateParamsMsg := types.MsgUpdateParams{
					Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
					Params:    params,
				}
				_, err := network.App.AuctionsKeeper.UpdateParams(network.GetContext(), &updateParamsMsg)
				assert.NoError(t, err, "failed to update auctions params")
			},
			expUpdate: false,
			postCheck: func() {},
		},
		{
			name: "no op if not valid bidder address",
			malleate: func() {
				setHighestBid(t, network, "", bidAmount)
			},
			expUpdate: false,
			postCheck: func() {},
		},
		{
			name: "no op if bid amount is not positive",
			malleate: func() {
				setHighestBid(t, network, validSenderAddr.String(), sdk.NewInt64Coin(utils.BaseDenom, 0))
			},
			expUpdate: false,
			postCheck: func() {},
		},
		{
			name: "no op if auctions module does not holds the bid amount",
			malleate: func() {
				network.App.BankKeeper.BurnCoins(network.GetContext(), types.ModuleName, sdk.NewCoins(bidAmount))
			},
			expUpdate: false,
			postCheck: func() {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(_ *testing.T) {
			keyring := testkeyring.New(1)
			network = testnetwork.NewUnitTestNetwork(
				testnetwork.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
			)

			epochIdentifier = epochstypes.WeekEpochID

			// Following code define the chain state required for an auction to be
			// performed properly. malleate() function modify the state to test all
			// possible cases.
			// Check that we are at the first round.
			round := network.App.AuctionsKeeper.GetRound(network.GetContext())
			assert.Equal(t, uint64(0), round, "round different than initial one")

			// Inflation should be enabled.
			params := network.App.AuctionsKeeper.GetParams(network.GetContext())
			params.EnableAuction = true
			updateParamsMsg := types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params:    params,
			}
			_, err := network.App.AuctionsKeeper.UpdateParams(network.GetContext(), &updateParamsMsg)
			assert.NoError(t, err, "failed to update auctions params")

			// There should be an highest bid.
			setHighestBid(t, network, validSenderAddr.String(), bidAmount)

			// We have to mint the bidded coins and also additional ones that will be distributed.
			mintCoinsToModuleAccount(t, network, types.ModuleName, sdk.NewCoins(append(additionalCoins, bidAmount)...))

			// We have to mint coins for the auction collector to simulate accrued fees that
			// are then sent from the auction collector to the auctions module.k
			mintCoinsToModuleAccount(t, network, types.AuctionCollectorName, sdk.NewCoins(feeCoin))

			tc.malleate()
			ctxSnapshot := network.GetContext()
			network.App.AuctionsKeeper.AfterEpochEnd(network.GetContext(), epochIdentifier, unusedEpochNumber)

			tc.postCheck()
			if tc.expUpdate {
				tc.postCheck()
			} else {
				paramsPre := network.App.AuctionsKeeper.GetParams(ctxSnapshot)
				params := network.App.AuctionsKeeper.GetParams(network.GetContext())
				assert.Equal(t, paramsPre, params, "expected params to not have changed")

				balanceCoinsPre := network.App.BankKeeper.GetAllBalances(ctxSnapshot, validSenderAddr)
				balanceCoins := network.App.BankKeeper.GetAllBalances(network.GetContext(), validSenderAddr)
				assert.Equal(t, balanceCoinsPre, balanceCoins, "expected sender balance to not have changed")

				moduleAccountAddress := network.App.AccountKeeper.GetModuleAddress(types.ModuleName)
				moduleAccountBalancePre := network.App.BankKeeper.GetAllBalances(ctxSnapshot, moduleAccountAddress)
				moduleAccountBalance := network.App.BankKeeper.GetAllBalances(network.GetContext(), moduleAccountAddress)
				assert.Equal(t, moduleAccountBalancePre, moduleAccountBalance, "expected module account balance to not have changed")

				auctionCollectorAddress := network.App.AccountKeeper.GetModuleAddress(types.AuctionCollectorName)
				auctionCollectorBalancePre := network.App.BankKeeper.GetAllBalances(ctxSnapshot, auctionCollectorAddress)
				auctionCollectorBalance := network.App.BankKeeper.GetAllBalances(network.GetContext(), auctionCollectorAddress)
				assert.Equal(t, auctionCollectorBalancePre, auctionCollectorBalance, "expected auction collector balance to not have changed")

				roundPre := network.App.AuctionsKeeper.GetRound(ctxSnapshot)
				round := network.App.AuctionsKeeper.GetRound(network.GetContext())
				assert.Equal(t, roundPre, round, "expected round to not have changed")

				bidPre := network.App.AuctionsKeeper.GetHighestBid(ctxSnapshot)
				bid := network.App.AuctionsKeeper.GetHighestBid(network.GetContext())
				assert.Equal(t, bidPre, bid, "expected bid to not have changed")
			}
		})
	}
}
