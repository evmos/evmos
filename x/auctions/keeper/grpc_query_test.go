package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testutil "github.com/evmos/evmos/v19/testutil"
	testnetwork "github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/auctions/types"
)

func TestAuctionInfo(t *testing.T) {
	var network *testnetwork.UnitTestNetwork
	ibcCoin := sdk.NewCoin("ibc/D219F3A490310B65BDC312B5A644B0D56FFF1789D894B902A49FBF9D2F560B32", sdk.NewInt(1))
	evmosCoin := sdk.NewCoin("aevmos", sdk.NewInt(1))
	anotherNativeCoin := sdk.NewCoin("aeth", sdk.NewInt(1))

	testCases := []struct {
		name        string
		malleate    func()
		expResp     func() *types.QueryCurrentAuctionInfoResponse
		expPass     bool
		errContains string
	}{
		{
			name: "pass with default genesis state",
			malleate: func() {
			},
			expResp: func() *types.QueryCurrentAuctionInfoResponse {
				return &types.QueryCurrentAuctionInfoResponse{
					Tokens:        nil,
					CurrentRound:  0,
					HighestBid:    sdk.NewCoin(utils.BaseDenom, sdk.NewInt(0)),
					BidderAddress: "",
				}
			},
			expPass:     true,
			errContains: "",
		},
		{
			name: "pass module with tokens",
			malleate: func() {
				// Tokens to mint to the module account. We are mint both EVMOS coin, which
				// should not appear in the response, and other coins which should appear.
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
			expResp: func() *types.QueryCurrentAuctionInfoResponse {
				return &types.QueryCurrentAuctionInfoResponse{
					Tokens:        sdk.NewCoins(ibcCoin, anotherNativeCoin),
					CurrentRound:  0,
					HighestBid:    sdk.NewCoin(utils.BaseDenom, sdk.NewInt(0)),
					BidderAddress: "",
				}
			},
			expPass:     true,
			errContains: "",
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
				expResp := tc.expResp()
				assert.Equal(t, expResp.CurrentRound, resp.CurrentRound, "expected a different current round")
				assert.Equal(t, expResp.HighestBid, resp.HighestBid, "expected a different highest bid")
				assert.Equal(t, expResp.BidderAddress, resp.BidderAddress, "expected a different bidder address")
				assert.Equal(t, expResp.Tokens, resp.Tokens, "expected a different tokens value")
			} else {
				assert.NoError(t, err, "expected no error during query execution")
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
	require.Equal(t, defaultParams, resp.Params)
}
