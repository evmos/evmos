package auctions_test

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v20/precompiles/auctions"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/precompiles/testutil"
	auctionstypes "github.com/evmos/evmos/v20/x/auctions/types"
)

func (s *PrecompileTestSuite) TestAuctionInfo() {
	method := s.precompile.Methods[auctions.AuctionInfoMethod]

	testCases := []struct {
		name        string
		malleate    func(ctx sdk.Context)
		gas         uint64
		expError    bool
		errContains string
		postCheck   func(bz []byte)
	}{
		{
			"success - get auction info",
			func(ctx sdk.Context) {
				moduleAcc := s.network.App.AccountKeeper.GetModuleAddress(auctionstypes.ModuleName)
				err := s.network.FundAccount(moduleAcc, types.NewCoins(types.NewCoin("uatom", math.NewInt(2e18))))
				s.Require().NoError(err)

				s.network.App.AuctionsKeeper.InitGenesis(ctx, auctionstypes.GenesisState{
					Params: auctionstypes.DefaultParams(),
					Bid: auctionstypes.Bid{
						BidValue: types.NewCoin("aevmos", math.NewInt(1e18)),
						Sender:   s.keyring.GetAddr(0).String(),
					},
					Round: 1,
				})

			},
			200000,
			false,
			"",
			func(bz []byte) {
				var auctionInfo auctions.AuctionInfoOutput
				err := s.precompile.UnpackIntoInterface(&auctionInfo, auctions.AuctionInfoMethod, bz)
				s.Require().NoError(err)
				s.Require().Equal(uint64(1), auctionInfo.AuctionInfo.CurrentRound)
				s.Require().Equal(big.NewInt(1e18), auctionInfo.AuctionInfo.HighestBid.Amount)
				s.Require().Equal(s.keyring.GetAddr(0), auctionInfo.AuctionInfo.BidderAddress)
				s.Require().Equal("aevmos", auctionInfo.AuctionInfo.HighestBid.Denom)
				s.Require().Equal(1, len(auctionInfo.AuctionInfo.Tokens))

			},
		},
		{
			"fail - invalid number of args",
			func(ctx sdk.Context) {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 0, 1),
			func(bz []byte) {},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Malleate the context and state
			ctx := s.network.GetContext()
			tc.malleate(ctx)

			// Create a new contract
			contract, ctx := testutil.NewPrecompileContract(s.T(), ctx, s.keyring.GetAddr(0), s.precompile, tc.gas)

			// Run the query
			bz, err := s.precompile.AuctionInfo(ctx, contract, &method, []interface{}{})

			if tc.expError {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(bz)
				tc.postCheck(bz)
			}
		})
	}
}
