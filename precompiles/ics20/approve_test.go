// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package ics20_test

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/evmos/evmos/v18/precompiles/authorization"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	"github.com/evmos/evmos/v18/precompiles/ics20"
	"github.com/evmos/evmos/v18/utils"
)

type allowanceTestCase struct {
	name        string
	malleate    func() []interface{}
	postCheck   func(data []byte, inputArgs []interface{})
	gas         uint64
	expError    bool
	errContains string
}

func getDefaultAllowanceCases(s *PrecompileTestSuite) []allowanceTestCase {
	return []allowanceTestCase{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			func([]byte, []interface{}) {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 5, 0),
		},
		// {	// TODO uncomment when corresponding logic included
		// 	"fail - origin same as spender",
		// 	func() []interface{} {
		// 		return []interface{}{
		// 			common.BytesToAddress(s.chainA.SenderAccount.GetAddress().Bytes()),
		// 			"port-1",
		// 			"channel-1",
		// 			utils.BaseDenom,
		// 			big.NewInt(1e18),
		// 		}
		// 	},
		// 	func(data []byte, inputArgs []interface{}) {},
		// 	200000,
		// 	true,
		// 	"origin is the same as spender",
		// },
		{
			"fail - authorization does not exist",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					"port-1",
					"channel-1",
					utils.BaseDenom,
					big.NewInt(1e18),
				}
			},
			func([]byte, []interface{}) {},
			200000,
			true,
			"does not exist",
		},
		{
			"fail - allocation for specified denom does not exist",
			func() []interface{} {
				err := s.NewTransferAuthorization(s.network.GetContext(), s.network.App, s.keyring.GetAddr(0), s.keyring.GetAddr(0), s.transferPath, maxUint256Coins, nil)
				s.Require().NoError(err)
				return []interface{}{
					s.keyring.GetAddr(0),
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointB.ChannelID,
					"atom",
					big.NewInt(1e18),
				}
			},
			func([]byte, []interface{}) {
			},
			200000,
			true,
			"no matching allocation found",
		},
		{
			"fail - allocation for specified channel and port id does not exist",
			func() []interface{} {
				err := s.NewTransferAuthorization(s.network.GetContext(), s.network.App, s.keyring.GetAddr(0), s.keyring.GetAddr(0), s.transferPath, maxUint256Coins, nil)
				s.Require().NoError(err)
				return []interface{}{
					s.keyring.GetAddr(0),
					"port-1",
					"channel-1",
					utils.BaseDenom,
					big.NewInt(1e18),
				}
			},
			func([]byte, []interface{}) {
			},
			200000,
			true,
			"no matching allocation found",
		},
	}
}

func (s *PrecompileTestSuite) TestApprove() {
	method := s.precompile.Methods[authorization.ApproveMethod]

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func(data []byte, inputArgs []interface{})
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			func([]byte, []interface{}) {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 2, 0),
		},
		{
			"fail - channel does not exist",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					[]cmn.ICS20Allocation{
						{
							SourcePort:    "port-1",
							SourceChannel: "channel-1",
							SpendLimit:    defaultCmnCoins,
							AllowList:     nil,
						},
					},
				}
			},
			func([]byte, []interface{}) {},
			200000,
			true,
			channeltypes.ErrChannelNotFound.Error(),
		},
		{
			"pass - MaxInt256 allocation",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					[]cmn.ICS20Allocation{
						{
							SourcePort:    s.transferPath.EndpointA.ChannelConfig.PortID,
							SourceChannel: s.transferPath.EndpointA.ChannelID,
							SpendLimit:    maxUint256CmnCoins,
							AllowList:     nil,
						},
					},
				}
			},
			func(_ []byte, _ []interface{}) {
				authz, _ := s.network.App.AuthzKeeper.GetAuthorization(s.network.GetContext(), s.keyring.GetAddr(0).Bytes(), s.keyring.GetAddr(0).Bytes(), ics20.TransferMsgURL)
				transferAuthz := authz.(*transfertypes.TransferAuthorization)
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, maxUint256Coins)
			},
			200000,
			false,
			"",
		},
		{
			"pass - create authorization with specific spend limit",
			func() []interface{} {
				return []interface{}{
					differentAddress,
					[]cmn.ICS20Allocation{
						{
							SourcePort:    s.transferPath.EndpointA.ChannelConfig.PortID,
							SourceChannel: s.transferPath.EndpointA.ChannelID,
							SpendLimit:    defaultCmnCoins,
							AllowList:     nil,
						},
					},
				}
			},
			func(_ []byte, _ []interface{}) {
				authz, _ := s.network.App.AuthzKeeper.GetAuthorization(s.network.GetContext(), differentAddress.Bytes(), s.keyring.GetAddr(0).Bytes(), ics20.TransferMsgURL)
				transferAuthz := authz.(*transfertypes.TransferAuthorization)
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, defaultCoins)
			},
			200000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			args := tc.malleate()
			bz, err := s.precompile.Approve(s.network.GetContext(), s.keyring.GetAddr(0), s.stateDB, &method, args)

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(bz, cmn.TrueValue)
				tc.postCheck(bz, args)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestRevoke() {
	method := s.precompile.Methods[authorization.RevokeMethod]

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
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 1, 0),
		},
		{
			"fail - not a correct grantee address",
			func() []interface{} {
				return []interface{}{
					"test string",
				}
			},
			func() {},
			200000,
			true,
			fmt.Sprintf(authorization.ErrInvalidGrantee, "test string"),
		},
		{
			"fail - authorization does not exist",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
				}
			},
			func() {},
			200000,
			true,
			"does not exist",
		},
		{
			"pass - deletes authorization grant",
			func() []interface{} {
				err := s.NewTransferAuthorization(s.network.GetContext(), s.network.App, differentAddress, s.keyring.GetAddr(0), s.transferPath, defaultCoins, nil)
				s.Require().NoError(err)
				authz, _ := s.network.App.AuthzKeeper.GetAuthorization(s.network.GetContext(), differentAddress.Bytes(), s.keyring.GetAddr(0).Bytes(), ics20.TransferMsgURL)
				s.Require().NotNil(authz)
				return []interface{}{
					differentAddress,
				}
			},
			func() {
				authz, _ := s.network.App.AuthzKeeper.GetAuthorization(s.network.GetContext(), differentAddress.Bytes(), s.keyring.GetAddr(0).Bytes(), ics20.TransferMsgURL)
				s.Require().Nil(authz)
			},
			200000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			args := tc.malleate()
			bz, err := s.precompile.Revoke(s.network.GetContext(), s.keyring.GetAddr(0), s.stateDB, &method, args)

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(bz, cmn.TrueValue)
				tc.postCheck()
			}
		})
	}
}

func (s *PrecompileTestSuite) TestIncreaseAllowance() {
	method := s.precompile.Methods[authorization.IncreaseAllowanceMethod]

	testCases := []allowanceTestCase{
		{
			"fail - the new spend limit overflows the maxUint256",
			func() []interface{} {
				overflowTestCoins := maxUint256Coins.Sub(sdk.NewInt64Coin(utils.BaseDenom, 1))
				err := s.NewTransferAuthorization(s.network.GetContext(), s.network.App, differentAddress, s.keyring.GetAddr(0), s.transferPath, overflowTestCoins, nil)
				s.Require().NoError(err)
				transferAuthz := s.GetTransferAuthorization(s.network.GetContext(), differentAddress, s.keyring.GetAddr(0))
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, overflowTestCoins)
				return []interface{}{
					differentAddress,
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointB.ChannelID,
					utils.BaseDenom,
					big.NewInt(2e18),
				}
			},
			func([]byte, []interface{}) {},
			200000,
			true,
			cmn.ErrIntegerOverflow,
		},
		{
			"pass - increase allowance by 1 EVMOS for a single allocation with a single coin denomination",
			func() []interface{} {
				err := s.NewTransferAuthorization(s.network.GetContext(), s.network.App, differentAddress, s.keyring.GetAddr(0), s.transferPath, defaultCoins, nil)
				s.Require().NoError(err)
				transferAuthz := s.GetTransferAuthorization(s.network.GetContext(), differentAddress, s.keyring.GetAddr(0))
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, defaultCoins)
				return []interface{}{
					differentAddress,
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointB.ChannelID,
					utils.BaseDenom,
					big.NewInt(1e18),
				}
			},
			func([]byte, []interface{}) {
				transferAuthz := s.GetTransferAuthorization(s.network.GetContext(), differentAddress, s.keyring.GetAddr(0))
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit[0].Amount, math.NewInt(2e18))
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit[0].Denom, utils.BaseDenom)
			},
			200000,
			false,
			"",
		},
		{
			"pass - increase allowance by 1 Atom for single allocation with a multiple coin denomination",
			func() []interface{} {
				err := s.NewTransferAuthorization(s.network.GetContext(), s.network.App, differentAddress, s.keyring.GetAddr(0), s.transferPath, mutliSpendLimit, nil)
				s.Require().NoError(err)
				transferAuthz := s.GetTransferAuthorization(s.network.GetContext(), differentAddress, s.keyring.GetAddr(0))
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, mutliSpendLimit)
				return []interface{}{
					differentAddress,
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointB.ChannelID,
					"uatom",
					big.NewInt(1e18),
				}
			},
			func([]byte, []interface{}) {
				transferAuthz := s.GetTransferAuthorization(s.network.GetContext(), differentAddress, s.keyring.GetAddr(0))
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit[1].Amount, math.NewInt(2e18))
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit[1].Denom, "uatom")
			},
			200000,
			false,
			"",
		},
		{
			"pass - increase allowance by 1 Evmos for multiple allocations with a single coin denomination",
			func() []interface{} {
				allocations := []transfertypes.Allocation{
					{
						SourcePort:    "port-01",
						SourceChannel: "channel-03",
						SpendLimit:    atomCoins,
						AllowList:     nil,
					},
					{
						SourcePort:    s.transferPath.EndpointA.ChannelConfig.PortID,
						SourceChannel: s.transferPath.EndpointA.ChannelID,
						SpendLimit:    defaultCoins,
						AllowList:     nil,
					},
				}
				err := s.NewTransferAuthorizationWithAllocations(s.network.GetContext(), s.network.App, differentAddress, s.keyring.GetAddr(0), allocations)
				s.Require().NoError(err)
				transferAuthz := s.GetTransferAuthorization(s.network.GetContext(), differentAddress, s.keyring.GetAddr(0))
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, atomCoins)
				s.Require().Equal(transferAuthz.Allocations[1].SpendLimit, defaultCoins)
				return []interface{}{
					differentAddress,
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointB.ChannelID,
					utils.BaseDenom,
					big.NewInt(1e18),
				}
			},
			func([]byte, []interface{}) {
				transferAuthz := s.GetTransferAuthorization(s.network.GetContext(), differentAddress, s.keyring.GetAddr(0))
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, atomCoins)
				s.Require().Equal(transferAuthz.Allocations[1].SpendLimit[0].Amount, math.NewInt(2e18))
				s.Require().Equal(transferAuthz.Allocations[1].SpendLimit[0].Denom, utils.BaseDenom)
			},
			200000,
			false,
			"",
		},
	}
	testCases = append(testCases, getDefaultAllowanceCases(s)...)

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			args := tc.malleate()
			bz, err := s.precompile.IncreaseAllowance(s.network.GetContext(), s.keyring.GetAddr(0), s.stateDB, &method, args)

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(bz, cmn.TrueValue)
				tc.postCheck(bz, args)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestDecreaseAllowance() {
	method := s.precompile.Methods[authorization.DecreaseAllowanceMethod]

	testCases := []allowanceTestCase{
		{
			"fail - the new spend limit is negative",
			func() []interface{} {
				err := s.NewTransferAuthorization(s.network.GetContext(), s.network.App, differentAddress, s.keyring.GetAddr(0), s.transferPath, defaultCoins, nil)
				s.Require().NoError(err)
				transferAuthz := s.GetTransferAuthorization(s.network.GetContext(), differentAddress, s.keyring.GetAddr(0))
				s.Require().NotNil(transferAuthz)
				s.Require().Len(transferAuthz.Allocations, 1)
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, defaultCoins)
				return []interface{}{
					differentAddress,
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointB.ChannelID,
					utils.BaseDenom,
					big.NewInt(2e18),
				}
			},
			func([]byte, []interface{}) {},
			200000,
			true,
			cmn.ErrNegativeAmount,
		},
		{
			"pass - decrease allowance by 1 EVMOS for a single allocation with a single coin denomination",
			func() []interface{} {
				err := s.NewTransferAuthorization(s.network.GetContext(), s.network.App, differentAddress, s.keyring.GetAddr(0), s.transferPath, defaultCoins, nil)
				s.Require().NoError(err)
				transferAuthz := s.GetTransferAuthorization(s.network.GetContext(), differentAddress, s.keyring.GetAddr(0))
				s.Require().NotNil(transferAuthz)
				s.Require().Len(transferAuthz.Allocations, 1)
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, defaultCoins)
				return []interface{}{
					differentAddress,
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointB.ChannelID,
					utils.BaseDenom,
					big.NewInt(500000000000000000),
				}
			},
			func(_ []byte, _ []interface{}) {
				transferAuthz := s.GetTransferAuthorization(s.network.GetContext(), differentAddress, s.keyring.GetAddr(0))
				s.Require().NotNil(transferAuthz)
				s.Require().Len(transferAuthz.Allocations, 1, "should have at least one allocation", transferAuthz)
				s.Require().Len(transferAuthz.Allocations[0].SpendLimit, 1, "should have at least one coin; allocation %s", transferAuthz.Allocations[0])
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit[0].Denom, utils.BaseDenom)
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit[0].Amount, math.NewInt(500000000000000000))
			},
			200000,
			false,
			"",
		},
		{
			"pass - decrease allowance by 1 Atom for single allocation with a multiple coin denomination",
			func() []interface{} {
				err := s.NewTransferAuthorization(s.network.GetContext(), s.network.App, differentAddress, s.keyring.GetAddr(0), s.transferPath, mutliSpendLimit, nil)
				s.Require().NoError(err)
				transferAuthz := s.GetTransferAuthorization(s.network.GetContext(), differentAddress, s.keyring.GetAddr(0))
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, mutliSpendLimit)
				return []interface{}{
					differentAddress,
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointB.ChannelID,
					"uatom",
					big.NewInt(500000000000000000),
				}
			},
			func(_ []byte, _ []interface{}) {
				transferAuthz := s.GetTransferAuthorization(s.network.GetContext(), differentAddress, s.keyring.GetAddr(0))
				s.Require().NotNil(transferAuthz)
				s.Require().Len(transferAuthz.Allocations, 1, "should have at least one allocation")
				s.Require().Len(transferAuthz.Allocations[0].SpendLimit, len(mutliSpendLimit), "should have two coins; allocation %s", transferAuthz.Allocations[0])
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit[1].Amount, math.NewInt(500000000000000000))
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit[1].Denom, "uatom")
				// other denom should remain unchanged
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit[0], defaultCoins[0])
			},
			200000,
			false,
			"",
		},
		{
			"pass - decrease allowance by 0.5 Evmos for multiple allocations with a single coin denomination",
			func() []interface{} {
				allocations := []transfertypes.Allocation{
					{
						SourcePort:    "port-01",
						SourceChannel: "channel-03",
						SpendLimit:    atomCoins,
						AllowList:     nil,
					},
					{
						SourcePort:    s.transferPath.EndpointA.ChannelConfig.PortID,
						SourceChannel: s.transferPath.EndpointA.ChannelID,
						SpendLimit:    defaultCoins,
						AllowList:     nil,
					},
				}
				err := s.NewTransferAuthorizationWithAllocations(s.network.GetContext(), s.network.App, differentAddress, s.keyring.GetAddr(0), allocations)
				s.Require().NoError(err)
				transferAuthz := s.GetTransferAuthorization(s.network.GetContext(), differentAddress, s.keyring.GetAddr(0))
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, atomCoins)
				s.Require().Equal(transferAuthz.Allocations[1].SpendLimit, defaultCoins)
				return []interface{}{
					differentAddress,
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointB.ChannelID,
					utils.BaseDenom,
					big.NewInt(1e18 / 2),
				}
			},
			func(_ []byte, _ []interface{}) {
				transferAuthz := s.GetTransferAuthorization(s.network.GetContext(), differentAddress, s.keyring.GetAddr(0))
				s.Require().Equal(transferAuthz.Allocations[1].SpendLimit[0].Amount, math.NewInt(1e18/2))
				s.Require().Equal(transferAuthz.Allocations[1].SpendLimit[0].Denom, utils.BaseDenom)
			},
			200000,
			false,
			"",
		},
	}

	testCases = append(testCases, getDefaultAllowanceCases(s)...)

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			args := tc.malleate()
			bz, err := s.precompile.DecreaseAllowance(s.network.GetContext(), s.keyring.GetAddr(0), s.stateDB, &method, args)

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(bz, cmn.TrueValue)
				tc.postCheck(bz, args)
			}
		})
	}
}
