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
	emptyBid := types.Bid{
		Sender:   "",
		BidValue: sdk.Coin{Denom: utils.BaseDenom, Amount: sdk.ZeroInt()},
	}
	existingBid := types.Bid{
		Sender:   validSenderAddr.String(),
		BidValue: bidAmount,
	}

	// Tokens used to simulate auctioned tokens.
	additionalCoins := sdk.Coins{
		sdk.NewInt64Coin("atoken", 1),
		sdk.NewInt64Coin("btoken", 1),
	}
	// Coin used to simulate accrued tokens to auction.
	feeCoin := sdk.NewInt64Coin("fee", 1)
	zeroCoin := sdk.NewInt64Coin(utils.BaseDenom, 0)

	// Value required in the hook but not used.
	unusedEpochNumber := int64(0)

	testCases := []struct {
		name                           string
		malleate                       func()
		expSuccessfulBid               bool
		expSenderPostBalance           sdk.Coins
		expModuleAccountPostBalance    sdk.Coins
		expAuctionCollectorPostBalance sdk.Coins
		expBidPost                     types.Bid
		expRoundDiff                   uint64
	}{
		{
			name: "pass with prize distributed",
			malleate: func() {
				// Initial state should pass
			},
			expSuccessfulBid:               true,
			expSenderPostBalance:           additionalCoins,
			expModuleAccountPostBalance:    sdk.NewCoins(feeCoin),
			expAuctionCollectorPostBalance: sdk.Coins{},
			expBidPost:                     emptyBid,
			expRoundDiff:                   1,
		},
		{
			name: "pass if bidded against no funds to receive",
			malleate: func() {
				err := network.App.BankKeeper.BurnCoins(network.GetContext(), types.ModuleName, additionalCoins)
				assert.NoError(t, err, "failed to burn coins in malleate")
			},
			expSuccessfulBid:               true,
			expSenderPostBalance:           sdk.Coins{},
			expModuleAccountPostBalance:    sdk.NewCoins(feeCoin),
			expAuctionCollectorPostBalance: sdk.Coins{},
			expBidPost:                     emptyBid,
			expRoundDiff:                   1,
		},
		{
			name: "no op if not weekly epoch",
			malleate: func() {
				epochIdentifier = epochstypes.DayEpochID
			},
			expSuccessfulBid: false,
			expBidPost:       existingBid,
			expRoundDiff:     0,
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
			expSuccessfulBid: false,
			expBidPost:       existingBid,
			expRoundDiff:     0,
		},
		{
			name: "no op if not valid bidder address",
			malleate: func() {
				setHighestBid(t, network, "", bidAmount)
			},
			expSuccessfulBid: false,
			expBidPost: types.Bid{
				Sender:   "",
				BidValue: bidAmount,
			},
			expRoundDiff: 1,
		},
		{
			name: "no op if bid amount is not positive",
			malleate: func() {
				setHighestBid(t, network, validSenderAddr.String(), zeroCoin)
			},
			expSuccessfulBid: false,
			expBidPost: types.Bid{
				Sender:   validSenderAddr.String(),
				BidValue: zeroCoin,
			},
			expRoundDiff: 1,
		},
		{
			name: "no op if auctions module does not hold the bid amount",
			malleate: func() {
				err := network.App.BankKeeper.BurnCoins(network.GetContext(), types.ModuleName, sdk.NewCoins(bidAmount))
				assert.NoError(t, err, "failed to burn coins in malleate")
			},
			expSuccessfulBid: false,
			expBidPost:       existingBid,
			expRoundDiff:     0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keyring := testkeyring.New(1)
			network = testnetwork.NewUnitTestNetwork(
				testnetwork.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
			)

			epochIdentifier = epochstypes.WeekEpochID

			// Following code define the chain state required for an auction to be
			// performed properly. malleate() function modify the state to test all
			// possible cases.
			// Check that we are at the first round.
			roundPre := network.App.AuctionsKeeper.GetRound(network.GetContext())
			assert.Equal(t, uint64(0), roundPre, "round different than initial one")

			// Auctions should be enabled.
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
			mintCoinsToModuleAccount(t, network, types.ModuleName, additionalCoins.Add(bidAmount))

			// We have to mint coins for the auction collector to simulate accrued fees that
			// are then sent from the auction collector to the auctions module.k
			mintCoinsToModuleAccount(t, network, types.AuctionCollectorName, sdk.NewCoins(feeCoin))

			tc.malleate()
			ctxSnapshot := network.GetContext()
			network.App.AuctionsKeeper.AfterEpochEnd(network.GetContext(), epochIdentifier, unusedEpochNumber)

			paramsPre := network.App.AuctionsKeeper.GetParams(ctxSnapshot)
			params = network.App.AuctionsKeeper.GetParams(network.GetContext())
			assert.Equal(t, paramsPre, params, "expected params to not have changed")

			// Get addresses and corresponding balances
			senderPreBalance := network.App.BankKeeper.GetAllBalances(ctxSnapshot, validSenderAddr)
			balanceCoins := network.App.BankKeeper.GetAllBalances(network.GetContext(), validSenderAddr)
			moduleAccountAddress := network.App.AccountKeeper.GetModuleAddress(types.ModuleName)
			moduleAccountBalance := network.App.BankKeeper.GetAllBalances(network.GetContext(), moduleAccountAddress)
			moduleAccountPreBalance := network.App.BankKeeper.GetAllBalances(ctxSnapshot, moduleAccountAddress)
			auctionCollectorAddress := network.App.AccountKeeper.GetModuleAddress(types.AuctionCollectorName)
			auctionCollectorBalance := network.App.BankKeeper.GetAllBalances(network.GetContext(), auctionCollectorAddress)
			auctionCollectorPreBalance := network.App.BankKeeper.GetAllBalances(ctxSnapshot, auctionCollectorAddress)

			if tc.expSuccessfulBid {
				assert.Equal(t, tc.expSenderPostBalance.String(), balanceCoins.String(), "expected different sender balance")
				assert.Equal(t, tc.expModuleAccountPostBalance.String(), moduleAccountBalance.String(), "expected different module account balance")
				assert.Equal(t, tc.expAuctionCollectorPostBalance.String(), auctionCollectorBalance.String(), "expected different auction collector balance")
			} else {
				assert.Equal(t, senderPreBalance.String(), balanceCoins.String(), "expected sender balance to not have changed")
				assert.Equal(t, moduleAccountPreBalance.String(), moduleAccountBalance.String(), "expected module account balance to not have changed")
				// NOTE: the auction collector balance should have been depleted
				assert.Equal(t, auctionCollectorPreBalance.String(), auctionCollectorBalance.String(), "expected auction collector balance to not have changed")
			}

			round := network.App.AuctionsKeeper.GetRound(network.GetContext())
			assert.Equal(t, roundPre+tc.expRoundDiff, round, "expected different round")

			bid := network.App.AuctionsKeeper.GetHighestBid(network.GetContext())
			assert.Equal(t, tc.expBidPost, bid, "expected different bid")
		})
	}
}
