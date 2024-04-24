// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package stride_test

import (
	"fmt"
	"math/big"

	"github.com/evmos/evmos/v18/precompiles/erc20"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v18/utils"

	common "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	"github.com/evmos/evmos/v18/precompiles/outposts/stride"
)

const (
	// RandomAddress is a token that is not supported by the Stride Outpost
	RandomAddress = "0x1FD55A1B9FC24967C4dB09C513C3BA0DFa7FF687"
)

func (s *PrecompileTestSuite) TestLiquidStake() {
	method := s.precompile.Methods[stride.LiquidStakeMethod]
	denomID := s.network.App.Erc20Keeper.GetDenomMap(s.network.GetContext(), utils.BaseDenom)
	tokenPair, ok := s.network.App.Erc20Keeper.GetTokenPair(s.network.GetContext(), denomID)
	s.Require().True(ok, "expected token pair to be found")

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 1, 0),
		},
		{
			"fail - token not found",
			func() []interface{} {
				err := s.network.App.StakingKeeper.SetParams(s.network.GetContext(), stakingtypes.DefaultParams())
				s.Require().NoError(err)
				return []interface{}{
					stride.AutopilotArgs{
						ChannelID:       channelID,
						Sender:          s.keyring.GetAddr(0),
						Receiver:        s.keyring.GetAddr(0),
						Token:           common.HexToAddress(RandomAddress),
						Amount:          big.NewInt(1e18),
						StrideForwarder: "stride1mdna37zrprxl7kn0rj4e58ndp084fzzwcxhrh2",
					},
				}
			},
			200000,
			true,
			"unsupported token",
		},
		{
			"fail - unsupported token",
			func() []interface{} {
				return []interface{}{
					stride.AutopilotArgs{
						ChannelID:       channelID,
						Sender:          s.keyring.GetAddr(0),
						Receiver:        s.keyring.GetAddr(0),
						Token:           common.HexToAddress(RandomAddress),
						Amount:          big.NewInt(1e18),
						StrideForwarder: "stride1mdna37zrprxl7kn0rj4e58ndp084fzzwcxhrh2",
					},
				}
			},
			200000,
			true,
			"unsupported token",
		},
		{
			"fail - invalid strideForwarder address (not a stride address)",
			func() []interface{} {
				return []interface{}{
					stride.AutopilotArgs{
						ChannelID:       channelID,
						Sender:          s.keyring.GetAddr(0),
						Receiver:        s.keyring.GetAddr(0),
						Token:           common.HexToAddress(tokenPair.Erc20Address),
						Amount:          big.NewInt(1e18),
						StrideForwarder: "cosmos1xv9tklw7d82sezh9haa573wufgy59vmwe6xxe5",
					},
				}
			},
			200000,
			true,
			"invalid stride bech32 address",
		},
		{
			"fail - strideForwarder address is an invalid stride bech32 address",
			func() []interface{} {
				return []interface{}{
					stride.AutopilotArgs{
						ChannelID:       channelID,
						Sender:          s.keyring.GetAddr(0),
						Receiver:        s.keyring.GetAddr(0),
						Token:           common.HexToAddress(tokenPair.Erc20Address),
						Amount:          big.NewInt(1e18),
						StrideForwarder: "stride1xv9tklw7d82sezh9haa573wufgy59vmwe6xx",
					},
				}
			},
			200000,
			true,
			"invalid stride bech32 address",
		},
		{
			"success",
			func() []interface{} {
				return []interface{}{
					stride.AutopilotArgs{
						ChannelID:       channelID,
						Sender:          s.keyring.GetAddr(0),
						Receiver:        s.keyring.GetAddr(0),
						Token:           common.HexToAddress(erc20.WEVMOSContractTestnet),
						Amount:          big.NewInt(1e18),
						StrideForwarder: "stride1rhe5leyt5w0mcwd9rpp93zqn99yktsxvyaqgd0",
					},
				}
			},
			200000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			sender := s.keyring.GetAddr(0)

			contract := vm.NewContract(vm.AccountRef(sender), s.precompile, big.NewInt(0), tc.gas)

			s.setupIBCCoordinator()

			_, err := s.precompile.LiquidStake(s.network.GetContext(), sender, s.network.GetStateDB(), contract, &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestRedeem() {
	method := s.precompile.Methods[stride.RedeemStakeMethod]
	stEvmos := utils.ComputeIBCDenom(portID, channelID, "st"+s.network.GetDenom())

	denomID := s.network.App.Erc20Keeper.GetDenomMap(s.network.GetContext(), stEvmos)
	tokenPair, ok := s.network.App.Erc20Keeper.GetTokenPair(s.network.GetContext(), denomID)
	s.Require().True(ok, "expected token pair to be found")

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 1, 0),
		},
		{
			"fail - token not found",
			func() []interface{} {
				err := s.network.App.StakingKeeper.SetParams(s.network.GetContext(), stakingtypes.DefaultParams())
				s.Require().NoError(err)
				return []interface{}{
					stride.AutopilotArgs{
						ChannelID:       channelID,
						Sender:          s.keyring.GetAddr(0),
						Receiver:        s.keyring.GetAddr(0),
						Token:           common.HexToAddress(RandomAddress),
						Amount:          big.NewInt(1e18),
						StrideForwarder: "stride1mdna37zrprxl7kn0rj4e58ndp084fzzwcxhrh2",
					},
				}
			},
			200000,
			true,
			"token pair not found",
		},
		{
			"fail - unsupported token",
			func() []interface{} {
				return []interface{}{
					stride.AutopilotArgs{
						ChannelID:       channelID,
						Sender:          s.keyring.GetAddr(0),
						Receiver:        s.keyring.GetAddr(0),
						Token:           common.HexToAddress(RandomAddress),
						Amount:          big.NewInt(1e18),
						StrideForwarder: "stride1mdna37zrprxl7kn0rj4e58ndp084fzzwcxhrh2",
					},
				}
			},
			200000,
			true,
			"The only supported token contract for Stride Outpost v1 is 0xd567B3d7B8FE3C79a1AD8dA978812cfC4Fa05e75",
		},
		{
			"fail - invalid receiver address (not a stride address)",
			func() []interface{} {
				return []interface{}{
					stride.AutopilotArgs{
						ChannelID:       channelID,
						Sender:          s.keyring.GetAddr(0),
						Receiver:        s.keyring.GetAddr(0),
						Token:           common.HexToAddress(tokenPair.Erc20Address),
						Amount:          big.NewInt(1e18),
						StrideForwarder: "cosmos1xv9tklw7d82sezh9haa573wufgy59vmwe6xxe5",
					},
				}
			},
			200000,
			true,
			"invalid stride bech32 address",
		},
		{
			"fail - stride forwarder address is an invalid stride bech32 address",
			func() []interface{} {
				return []interface{}{
					stride.AutopilotArgs{
						ChannelID:       channelID,
						Sender:          s.keyring.GetAddr(0),
						Receiver:        s.keyring.GetAddr(0),
						Token:           common.HexToAddress(tokenPair.Erc20Address),
						Amount:          big.NewInt(1e18),
						StrideForwarder: "stride1xv9tklw7d82sezh9haa573wufgy59vmwe6xxe",
					},
				}
			},
			200000,
			true,
			"invalid stride bech32 address",
		},
		{
			"success",
			func() []interface{} {
				return []interface{}{
					stride.AutopilotArgs{
						ChannelID:       channelID,
						Sender:          s.keyring.GetAddr(0),
						Receiver:        s.keyring.GetAddr(0),
						Token:           common.HexToAddress(tokenPair.GetErc20Address()),
						Amount:          big.NewInt(1e18),
						StrideForwarder: "stride1rhe5leyt5w0mcwd9rpp93zqn99yktsxvyaqgd0",
					},
				}
			},
			200000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			sender := s.keyring.GetAddr(0)
			contract := vm.NewContract(vm.AccountRef(sender), s.precompile, big.NewInt(0), tc.gas)

			s.setupIBCCoordinator()

			_, err := s.precompile.RedeemStake(s.network.GetContext(), sender, s.network.GetStateDB(), contract, &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
