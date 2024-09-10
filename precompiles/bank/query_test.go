package bank_test

import (
	"math/big"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/precompiles/bank"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	evmosutiltx "github.com/evmos/evmos/v20/testutil/tx"
)

func (s *PrecompileTestSuite) TestBalances() {
	var ctx sdk.Context
	// setup test in order to have s.precompile, s.evmosAddr and s.xmplAddr defined
	s.SetupTest()
	method := s.precompile.Methods[bank.BalancesMethod]

	testcases := []struct {
		name        string
		malleate    func() []interface{}
		expPass     bool
		errContains string
		expBalances func(evmosAddr, xmplAddr common.Address) []bank.Balance
	}{
		{
			"fail - invalid number of arguments",
			func() []interface{} {
				return []interface{}{
					"", "",
				}
			},
			false,
			"invalid number of arguments",
			nil,
		},
		{
			"fail - invalid account address",
			func() []interface{} {
				return []interface{}{
					"random text",
				}
			},
			false,
			"invalid type for account",
			nil,
		},
		{
			"pass - empty balances for new account",
			func() []interface{} {
				return []interface{}{
					evmosutiltx.GenerateAddress(),
				}
			},
			true,
			"",
			func(common.Address, common.Address) []bank.Balance { return []bank.Balance{} },
		},
		{
			"pass - Initial balances present",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
				}
			},
			true,
			"",
			func(evmosAddr, xmplAddr common.Address) []bank.Balance {
				return []bank.Balance{
					{
						ContractAddress: evmosAddr,
						Amount:          network.PrefundedAccountInitialBalance.BigInt(),
					},
					{
						ContractAddress: xmplAddr,
						Amount:          network.PrefundedAccountInitialBalance.BigInt(),
					},
				}
			},
		},
		{
			"pass - EVMOS and XMPL balances present - mint extra XMPL",
			func() []interface{} {
				ctx = s.mintAndSendXMPLCoin(ctx, s.keyring.GetAccAddr(0), math.NewInt(1e18))
				return []interface{}{
					s.keyring.GetAddr(0),
				}
			},
			true,
			"",
			func(evmosAddr, xmplAddr common.Address) []bank.Balance {
				return []bank.Balance{{
					ContractAddress: evmosAddr,
					Amount:          network.PrefundedAccountInitialBalance.BigInt(),
				}, {
					ContractAddress: xmplAddr,
					Amount:          network.PrefundedAccountInitialBalance.Add(math.NewInt(1e18)).BigInt(),
				}}
			},
		},
	}

	for _, tc := range testcases {
		tc := tc

		s.Run(tc.name, func() {
			ctx = s.SetupTest() // reset the chain each test

			bz, err := s.precompile.Balances(
				ctx,
				nil,
				&method,
				tc.malleate(),
			)

			if tc.expPass {
				s.Require().NoError(err)
				var balances []bank.Balance
				err = s.precompile.UnpackIntoInterface(&balances, method.Name, bz)
				s.Require().NoError(err)
				s.Require().Equal(tc.expBalances(s.evmosAddr, s.xmplAddr), balances)
			} else {
				s.Require().Contains(err.Error(), tc.errContains)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestTotalSupply() {
	var ctx sdk.Context
	// setup test in order to have s.precompile, s.evmosAddr and s.xmplAddr defined
	s.SetupTest()
	method := s.precompile.Methods[bank.TotalSupplyMethod]

	totSupplRes, err := s.grpcHandler.GetTotalSupply()
	s.Require().NoError(err)
	evmosTotalSupply := totSupplRes.Supply.AmountOf(s.bondDenom)
	xmplTotalSupply := totSupplRes.Supply.AmountOf(s.tokenDenom)

	testcases := []struct {
		name      string
		malleate  func()
		expSupply func(evmosAddr, xmplAddr common.Address) []bank.Balance
	}{
		{
			"pass - EVMOS and XMPL total supply",
			func() {
				ctx = s.mintAndSendXMPLCoin(ctx, s.keyring.GetAccAddr(0), math.NewInt(1e18))
			},
			func(evmosAddr, xmplAddr common.Address) []bank.Balance {
				return []bank.Balance{{
					ContractAddress: evmosAddr,
					Amount:          evmosTotalSupply.BigInt(),
				}, {
					ContractAddress: xmplAddr,
					Amount:          xmplTotalSupply.Add(math.NewInt(1e18)).BigInt(),
				}}
			},
		},
	}

	for _, tc := range testcases {
		tc := tc

		s.Run(tc.name, func() {
			ctx = s.SetupTest()
			tc.malleate()
			bz, err := s.precompile.TotalSupply(
				ctx,
				nil,
				&method,
				nil,
			)

			s.Require().NoError(err)
			var balances []bank.Balance
			err = s.precompile.UnpackIntoInterface(&balances, method.Name, bz)
			s.Require().NoError(err)
			s.Require().Equal(tc.expSupply(s.evmosAddr, s.xmplAddr), balances)
		})
	}
}

func (s *PrecompileTestSuite) TestSupplyOf() {
	// setup test in order to have s.precompile, s.evmosAddr and s.xmplAddr defined
	s.SetupTest()
	method := s.precompile.Methods[bank.SupplyOfMethod]

	totSupplRes, err := s.grpcHandler.GetTotalSupply()
	s.Require().NoError(err)
	evmosTotalSupply := totSupplRes.Supply.AmountOf(s.bondDenom)
	xmplTotalSupply := totSupplRes.Supply.AmountOf(s.tokenDenom)

	testcases := []struct {
		name        string
		malleate    func() []interface{}
		expErr      bool
		errContains string
		expSupply   *big.Int
	}{
		{
			"fail - invalid number of arguments",
			func() []interface{} {
				return []interface{}{
					"", "", "",
				}
			},
			true,
			"invalid number of arguments",
			nil,
		},
		{
			"fail - invalid hex address",
			func() []interface{} {
				return []interface{}{
					"random text",
				}
			},
			true,
			"invalid type for erc20Address",
			nil,
		},
		{
			"pass - erc20 not registered return 0 supply",
			func() []interface{} {
				return []interface{}{
					evmosutiltx.GenerateAddress(),
				}
			},
			false,
			"",
			big.NewInt(0),
		},
		{
			"pass - XMPL total supply",
			func() []interface{} {
				return []interface{}{
					s.xmplAddr,
				}
			},
			false,
			"",
			xmplTotalSupply.BigInt(),
		},

		{
			"pass - EVMOS total supply",
			func() []interface{} {
				return []interface{}{
					s.evmosAddr,
				}
			},
			false,
			"",
			evmosTotalSupply.BigInt(),
		},
	}

	for _, tc := range testcases {
		tc := tc

		s.Run(tc.name, func() {
			ctx := s.SetupTest()

			bz, err := s.precompile.SupplyOf(
				ctx,
				nil,
				&method,
				tc.malleate(),
			)

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				out, err := method.Outputs.Unpack(bz)
				s.Require().NoError(err, "expected no error unpacking")
				supply, ok := out[0].(*big.Int)
				s.Require().True(ok, "expected output to be a big.Int")
				s.Require().NoError(err)
				s.Require().Equal(supply.Int64(), tc.expSupply.Int64())
			}
		})
	}
}
