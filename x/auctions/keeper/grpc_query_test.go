// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	testutil "github.com/evmos/evmos/v19/testutil"
	testnetwork "github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	testutiltx "github.com/evmos/evmos/v19/testutil/tx"
	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/auctions/types"
)

func TestAuctionInfo(t *testing.T) {
	var network *testnetwork.UnitTestNetwork
	ibcCoin := sdk.NewCoin("ibc/D219F3A490310B65BDC312B5A644B0D56FFF1789D894B902A49FBF9D2F560B32", sdk.NewInt(1))
	evmosCoin := sdk.NewCoin("aevmos", sdk.NewInt(1))
	anotherNativeCoin := sdk.NewCoin("aeth", sdk.NewInt(1))
	validBidderAddr, _ := testutiltx.NewAccAddressAndKey()

	testCases := []struct {
		name        string
		malleate    func()
		expResp     types.QueryCurrentAuctionInfoResponse
		expPass     bool
		errContains string
	}{
		{
			name: "success - with default genesis state",
			malleate: func() {
			},
			expResp: types.QueryCurrentAuctionInfoResponse{
				Tokens:        nil,
				CurrentRound:  0,
				HighestBid:    sdk.NewCoin(utils.BaseDenom, sdk.NewInt(0)),
				BidderAddress: "",
			},
			expPass: true,
		},
		{
			name: "success - with non empty bid",
			malleate: func() {
				network.App.AuctionsKeeper.SetHighestBid(network.GetContext(), validBidderAddr.String(), sdk.NewInt64Coin(utils.BaseDenom, 1))
			},
			expResp: types.QueryCurrentAuctionInfoResponse{
				Tokens:        nil,
				CurrentRound:  0,
				HighestBid:    sdk.NewCoin(utils.BaseDenom, sdk.NewInt(1)),
				BidderAddress: validBidderAddr.String(),
			},
			expPass:     true,
			errContains: "",
		},
		{
			name: "success - module with tokens",
			malleate: func() {
				// Tokens to mint to the module account. We mint both EVMOS coin, which
				// should not appear in the response, and other coins which should appear.
				// We have to mint tokens because the module accounts of x/auctions are blocked
				// but it is acceptable in unit test.
				coinsToMint := sdk.NewCoins(evmosCoin, ibcCoin, anotherNativeCoin)
				err := testutil.FundModuleAccount(
					network.GetContext(),
					network.App.BankKeeper,
					types.ModuleName,
					coinsToMint,
				)
				require.NoError(t, err, "expected no error while minting tokens for the module account")

				auctionModuleAddress := network.App.AccountKeeper.GetModuleAddress(types.ModuleName)
				bal := network.App.BankKeeper.GetAllBalances(network.GetContext(), auctionModuleAddress)
				require.Equal(t, bal, coinsToMint, "expected a different balance for auctions module after minting tokens")
			},
			expResp: types.QueryCurrentAuctionInfoResponse{
				Tokens:        sdk.NewCoins(ibcCoin, anotherNativeCoin),
				CurrentRound:  0,
				HighestBid:    sdk.NewCoin(utils.BaseDenom, sdk.NewInt(0)),
				BidderAddress: "",
			},
			expPass:     true,
			errContains: "",
		},
		{
			name: "fail - auction module is not enabled",
			malleate: func() {
				// Update params to disable the auction.
				params := network.App.AuctionsKeeper.GetParams(network.GetContext())
				params.EnableAuction = false
				updateParamsMsg := types.MsgUpdateParams{
					Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
					Params:    params,
				}
				_, err := network.App.AuctionsKeeper.UpdateParams(network.GetContext(), &updateParamsMsg)
				assert.NoError(t, err, "failed to update auctions params")
			},
			expResp:     types.QueryCurrentAuctionInfoResponse{},
			expPass:     false,
			errContains: types.ErrAuctionDisabled.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// keyring := testkeyring.New(1)
			network = testnetwork.NewUnitTestNetwork()
			auctionQueryClient := network.GetAuctionsClient()

			tc.malleate()

			resp, err := auctionQueryClient.AuctionInfo(network.GetContext(), &types.QueryCurrentAuctionInfoRequest{})

			if tc.expPass {
				assert.NoError(t, err, "expected no error during query execution")
				expResp := tc.expResp
				assert.Equal(t, expResp.CurrentRound, resp.CurrentRound, "expected a different current round")
				assert.Equal(t, expResp.HighestBid, resp.HighestBid, "expected a different highest bid")
				assert.Equal(t, expResp.BidderAddress, resp.BidderAddress, "expected a different bidder address")
				assert.Equal(t, expResp.Tokens, resp.Tokens, "expected a different tokens value")
			} else {
				assert.Error(t, err, "expected error during query execution")
				assert.Contains(t, err.Error(), tc.errContains)
			}
		})
	}
}

func TestParams(t *testing.T) {
	network := testnetwork.NewUnitTestNetwork()
	auctionQueryClient := network.GetAuctionsClient()

	defaultParams := types.DefaultParams()
	resp, err := auctionQueryClient.Params(network.GetContext(), &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &defaultParams, resp.Params)
}
