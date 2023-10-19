// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package stride_test

import (
	"fmt"
	"math/big"

	common "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
	"github.com/evmos/evmos/v15/precompiles/outposts/stride"
)

func (s *PrecompileTestSuite) TestLiquidStake() {
	method := s.precompile.Methods[stride.LiquidStakeMethod]

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
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 4, 0),
		},
		{
			"fail - invalid receiver address (not a stride address)",
			func() []interface{} {
				denomID := s.app.Erc20Keeper.GetDenomMap(s.ctx, "ibc/3A5B71F2AA11D24F9688A10D4279CE71560489D7A695364FC361EC6E09D02889")
				tokenPair, ok := s.app.Erc20Keeper.GetTokenPair(s.ctx, denomID)
				s.Require().True(ok, "expected token pair to be found")
				return []interface{}{
					s.address,
					common.HexToAddress(tokenPair.Erc20Address),
					big.NewInt(1e18),
					"cosmos1xv9tklw7d82sezh9haa573wufgy59vmwe6xxe5",
				}
			},
			func() {},
			200000,
			true,
			"receiverAddress is not a stride address",
		},
		{
			"fail - receiver address is an invalid stride bech32 address",
			func() []interface{} {
				denomID := s.app.Erc20Keeper.GetDenomMap(s.ctx, "ibc/3A5B71F2AA11D24F9688A10D4279CE71560489D7A695364FC361EC6E09D02889")
				tokenPair, ok := s.app.Erc20Keeper.GetTokenPair(s.ctx, denomID)
				s.Require().True(ok, "expected token pair to be found")
				return []interface{}{
					s.address,
					common.HexToAddress(tokenPair.Erc20Address),
					big.NewInt(1e18),
					"stride1xv9tklw7d82sezh9haa573wufgy59vmwe6xxe",
				}
			},
			func() {},
			200000,
			true,
			"invalid stride bech32 address",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), tc.gas)

			_, err := s.precompile.LiquidStake(s.ctx, s.address, s.stateDB, contract, &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}

func (s *PrecompileTestSuite) TestRedeem() {

}
