package auctions_test

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/precompiles/auctions"
	"github.com/evmos/evmos/v20/precompiles/testutil"
	auctionstypes "github.com/evmos/evmos/v20/x/auctions/types"
	"github.com/evmos/evmos/v20/x/evm/core/vm"

	cmn "github.com/evmos/evmos/v20/precompiles/common"
)

func (s *PrecompileTestSuite) TestBid() {
	method := s.precompile.Methods[auctions.BidMethod]

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func()
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			func() {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 2, 0),
		},
		{
			"fail - invalid sender address",
			func() []interface{} {
				return []interface{}{
					"",
					big.NewInt(1e18),
				}
			},
			func() {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidHexAddress, ""),
		},
		{
			"fail - invalid bid amount",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					nil,
				}
			},
			func() {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidAmount, ""),
		},
		{
			"success - bid placed",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					big.NewInt(1e18),
				}
			},
			func() {
				bid := s.network.App.AuctionsKeeper.GetHighestBid(s.network.GetContext())
				s.Require().Equal(bid.BidValue.Amount, math.NewInt(1e18))
				s.Require().Equal(bid.BidValue.Denom, "aevmos")
			},
			200000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			var contract *vm.Contract
			contract, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), s.keyring.GetAddr(0), s.precompile, tc.gas)

			bz, err := s.precompile.Bid(ctx, s.keyring.GetAddr(0), contract, s.network.GetStateDB(), &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)
				tc.postCheck()
			}
		})
	}
}

func (s *PrecompileTestSuite) TestDepositCoin() {
	method := s.precompile.Methods[auctions.DepositCoinMethod]

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func()
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			func() {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 3, 0),
		},
		{
			"fail - invalid sender address",
			func() []interface{} {
				return []interface{}{
					"",
					common.Address{},
					big.NewInt(1e18),
				}
			},
			func() {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidHexAddress, ""),
		},
		{
			"fail - invalid token address",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					"invalid_address",
					big.NewInt(1e18),
				}
			},
			func() {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidHexAddress, "invalid_address"),
		},
		{
			"fail - invalid deposit amount",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					common.Address{},
					nil,
				}
			},
			func() {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidAmount, ""),
		},
		{
			"success - coin deposited",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					s.tokenPair.GetERC20Contract(),
					big.NewInt(1e18),
				}
			},
			func() {
				// Check the auctions collector address has the deposited coins
				collectorAddr := s.network.App.AccountKeeper.GetModuleAddress(auctionstypes.AuctionCollectorName)
				deposits := s.network.App.BankKeeper.GetAllBalances(s.network.GetContext(), collectorAddr)
				s.Require().Equal(deposits, sdk.NewCoins(sdk.NewCoin("uatom", sdkmath.NewInt(1e18))))
			},
			200000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			var contract *vm.Contract
			contract, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), s.keyring.GetAddr(0), s.precompile, tc.gas)

			bz, err := s.precompile.DepositCoin(ctx, s.keyring.GetAddr(0), contract, s.network.GetStateDB(), &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)
				tc.postCheck()
			}
		})
	}
}
