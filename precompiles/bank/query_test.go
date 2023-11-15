package bank_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v15/precompiles/bank"
	evmosutiltx "github.com/evmos/evmos/v15/testutil/tx"
)

func (s *PrecompileTestSuite) TestBalances() {
	method := s.precompile.Methods[bank.BalancesMethod]

	testcases := []struct {
		name        string
		malleate    func() []interface{}
		expPass     bool
		errContains string
		expBalances []bank.Balance
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
			[]bank.Balance{},
		},
		{
			"pass - EVMOS balance present",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
				}
			},
			true,
			"",
			[]bank.Balance{{
				ContractAddress: s.evmosAddr,
				Amount:          big.NewInt(4e18),
			}},
		},
		{
			"pass - EVMOS and XMPL balances present",
			func() []interface{} {
				s.mintAndSendCoin("xmpl", s.keyring.GetAccAddr(0), sdk.NewInt(1e18))
				return []interface{}{
					s.keyring.GetAddr(0),
				}
			},
			true,
			"",
			[]bank.Balance{{
				ContractAddress: s.evmosAddr,
				Amount:          big.NewInt(4e18),
			}, {
				ContractAddress: s.xmplAddr,
				Amount:          big.NewInt(1e18),
			}},
		},
	}

	for _, tc := range testcases {
		tc := tc

		s.Run(tc.name, func() {
			s.SetupTest()

			bz, err := s.precompile.Balances(
				s.network.GetContext(),
				nil,
				&method,
				tc.malleate(),
			)

			if tc.expPass {
				s.Require().NoError(err)
				var balances []bank.Balance
				err = s.precompile.UnpackIntoInterface(&balances, method.Name, bz)
				s.Require().NoError(err)
				s.Require().Equal(tc.expBalances, balances)
			} else {
				s.Require().Contains(err.Error(), tc.errContains)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestTotalSupply() {
	method := s.precompile.Methods[bank.TotalSupplyMethod]

	evmosTotalSupply, ok := new(big.Int).SetString("11000000000000000000", 10)
	s.Require().True(ok)

	testcases := []struct {
		name      string
		malleate  func()
		expSupply []bank.Balance
	}{
		{
			"pass - EVMOS and XMPL total supply",
			func() {
				s.mintAndSendCoin("xmpl", s.keyring.GetAccAddr(0), sdk.NewInt(1e18))
			},
			[]bank.Balance{{
				ContractAddress: s.evmosAddr,
				Amount:          evmosTotalSupply,
			}, {
				ContractAddress: s.xmplAddr,
				Amount:          big.NewInt(1e18),
			}},
		},
	}

	for _, tc := range testcases {
		tc := tc

		s.Run(tc.name, func() {
			s.SetupTest()

			bz, err := s.precompile.TotalSupply(
				s.network.GetContext(),
				nil,
				&method,
				nil,
			)

			s.Require().NoError(err)
			var balances []bank.Balance
			err = s.precompile.UnpackIntoInterface(&balances, method.Name, bz)
			s.Require().NoError(err)
			s.Require().Equal(tc.expSupply, balances)
		})
	}
}
