// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package osmosis_test

import (
	"fmt"
	"math/big"

	//	"github.com/evmos/evmos/v15/utils"

	//	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	//	common "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
	"github.com/evmos/evmos/v15/precompiles/outposts/osmosis"
)

const (
	gas = 200000
)

func (s *PrecompileTestSuite) TestSwap() {
	method := s.precompile.Methods[osmosis.SwapMethod]

	// // Retrieve Evmos token information
	// evmosDenomID := s.app.Erc20Keeper.GetDenomMap(s.ctx, utils.BaseDenom)
	// evmosTokenPair, ok := s.app.Erc20Keeper.GetTokenPair(s.ctx, evmosDenomID)
	// s.Require().True(ok, "expected evmos token pair to be found")
	//
	// // Retrieve Osmo token information
	// osmoIBCDenom := utils.ComputeIBCDenom(portID, channelID, osmosis.OsmosisDenom)
	// osmoDenomID := s.app.Erc20Keeper.GetDenomMap(s.ctx, osmoIBCDenom)
	// osmoTokenPair, ok := s.app.Erc20Keeper.GetTokenPair(s.ctx, osmoDenomID)
	// s.Require().True(ok, "expected osmo token pair to be found")

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
			gas,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 7, 0),
		},
		// {
		// 	"fail - token not found",
		// 	func() []interface{} {
		// 		err := s.app.StakingKeeper.SetParams(s.ctx, stakingtypes.DefaultParams())
		// 		s.Require().NoError(err)
		// 		return []interface{}{
		// 			s.address,
		// 			common.HexToAddress("0x1FD55A1B9FC24967C4dB09C513C3BA0DFa7FF687"),
		// 			big.NewInt(1e18),
		// 			"stride1mdna37zrprxl7kn0rj4e58ndp084fzzwcxhrh2",
		// 		}
		// 	},
		// 	func() {},
		// 	200000,
		// 	true,
		// 	"token pair not found",
		// },
		// {
		// 	"fail - unsupported token",
		// 	func() []interface{} {
		// 		return []interface{}{
		// 			s.address,
		// 			common.HexToAddress("0x1FD55A1B9FC24967C4dB09C513C3BA0DFa7FF687"),
		// 			big.NewInt(1e18),
		// 			"stride1mdna37zrprxl7kn0rj4e58ndp084fzzwcxhrh2",
		// 		}
		// 	},
		// 	func() {},
		// 	200000,
		// 	true,
		// 	"The only supported token contract for Stride Outpost v1 is 0x80b5a32E4F032B2a058b4F29EC95EEfEEB87aDcd",
		// },
		// {
		// 	"fail - invalid strideForwarder address (not a stride address)",
		// 	func() []interface{} {
		// 		return []interface{}{
		// 			s.address,
		// 			common.HexToAddress(tokenPair.Erc20Address),
		// 			big.NewInt(1e18),
		// 			"cosmos1xv9tklw7d82sezh9haa573wufgy59vmwe6xxe5",
		// 		}
		// 	},
		// 	func() {},
		// 	200000,
		// 	true,
		// 	"receiver is not a stride address",
		// },
		// {
		// 	"fail - strideForwarder address is an invalid stride bech32 address",
		// 	func() []interface{} {
		// 		return []interface{}{
		// 			s.address,
		// 			common.HexToAddress(tokenPair.Erc20Address),
		// 			big.NewInt(1e18),
		// 			"stride1xv9tklw7d82sezh9haa573wufgy59vmwe6xxe",
		// 		}
		// 	},
		// 	func() {},
		// 	200000,
		// 	true,
		// 	"invalid stride bech32 address",
		// },
		// {
		// 	"success",
		// 	func() []interface{} {
		// 		path := NewTransferPath(s.chainA, s.chainB)
		// 		s.coordinator.Setup(path)
		// 		return []interface{}{
		// 			s.address,
		// 			common.HexToAddress(tokenPair.Erc20Address),
		// 			big.NewInt(1e18),
		// 			"stride1rhe5leyt5w0mcwd9rpp93zqn99yktsxvyaqgd0",
		// 		}
		// 	},
		// 	func() {},
		// 	200000,
		// 	false,
		// 	"",
		// },
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), tc.gas)

			_, err := s.precompile.Swap(s.ctx, s.address, s.stateDB, contract, &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}
