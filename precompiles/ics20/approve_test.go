// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package ics20_test

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/evmos/evmos/v16/precompiles/authorization"
	cmn "github.com/evmos/evmos/v16/precompiles/common"
	"github.com/evmos/evmos/v16/precompiles/ics20"
	"github.com/evmos/evmos/v16/utils"
)

type allowanceTestCase struct {
	name        string
	malleate    func() []interface{}
	postCheck   func(data []byte, inputArgs []interface{})
	gas         uint64
	expError    bool
	errContains string
}

var defaultAllowanceCases = []allowanceTestCase{
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
				s.address,
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
			path := NewTransferPath(s.chainA, s.chainB)
			s.coordinator.Setup(path)
			err := s.NewTransferAuthorization(s.ctx, s.app, s.address, s.address, path, maxUint256Coins, nil)
			s.Require().NoError(err)
			return []interface{}{
				s.address,
				path.EndpointA.ChannelConfig.PortID,
				path.EndpointB.ChannelID,
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
			path := NewTransferPath(s.chainA, s.chainB)
			s.coordinator.Setup(path)
			err := s.NewTransferAuthorization(s.ctx, s.app, s.address, s.address, path, maxUint256Coins, nil)
			s.Require().NoError(err)
			return []interface{}{
				s.address,
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
					s.address,
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
				path := NewTransferPath(s.chainA, s.chainB)
				s.coordinator.Setup(path)
				return []interface{}{
					s.address,
					[]cmn.ICS20Allocation{
						{
							SourcePort:    path.EndpointA.ChannelConfig.PortID,
							SourceChannel: path.EndpointA.ChannelID,
							SpendLimit:    maxUint256CmnCoins,
							AllowList:     nil,
						},
					},
				}
			},
			func(_ []byte, _ []interface{}) {
				authz, _ := s.app.AuthzKeeper.GetAuthorization(s.ctx, s.address.Bytes(), s.address.Bytes(), ics20.TransferMsgURL)
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
				path := NewTransferPath(s.chainA, s.chainB)
				s.coordinator.Setup(path)
				return []interface{}{
					differentAddress,
					[]cmn.ICS20Allocation{
						{
							SourcePort:    path.EndpointA.ChannelConfig.PortID,
							SourceChannel: path.EndpointA.ChannelID,
							SpendLimit:    defaultCmnCoins,
							AllowList:     nil,
						},
					},
				}
			},
			func(_ []byte, _ []interface{}) {
				authz, _ := s.app.AuthzKeeper.GetAuthorization(s.ctx, differentAddress.Bytes(), s.address.Bytes(), ics20.TransferMsgURL)
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
			bz, err := s.precompile.Approve(s.ctx, s.address, s.stateDB, &method, args)

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
					s.address,
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
				path := NewTransferPath(s.chainA, s.chainB)
				s.coordinator.Setup(path)
				err := s.NewTransferAuthorization(s.ctx, s.app, differentAddress, s.address, path, defaultCoins, nil)
				s.Require().NoError(err)
				authz, _ := s.app.AuthzKeeper.GetAuthorization(s.ctx, differentAddress.Bytes(), s.address.Bytes(), ics20.TransferMsgURL)
				s.Require().NotNil(authz)
				return []interface{}{
					differentAddress,
				}
			},
			func() {
				authz, _ := s.app.AuthzKeeper.GetAuthorization(s.ctx, differentAddress.Bytes(), s.address.Bytes(), ics20.TransferMsgURL)
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
			bz, err := s.precompile.Revoke(s.ctx, s.address, s.stateDB, &method, args)

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
				path := NewTransferPath(s.chainA, s.chainB)
				s.coordinator.Setup(path)
				overflowTestCoins := maxUint256Coins.Sub(sdk.NewInt64Coin(utils.BaseDenom, 1))
				err := s.NewTransferAuthorization(s.ctx, s.app, differentAddress, s.address, path, overflowTestCoins, nil)
				s.Require().NoError(err)
				transferAuthz := s.GetTransferAuthorization(s.ctx, differentAddress, s.address)
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, overflowTestCoins)
				return []interface{}{
					differentAddress,
					path.EndpointA.ChannelConfig.PortID,
					path.EndpointB.ChannelID,
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
				path := NewTransferPath(s.chainA, s.chainB)
				s.coordinator.Setup(path)
				err := s.NewTransferAuthorization(s.ctx, s.app, differentAddress, s.address, path, defaultCoins, nil)
				s.Require().NoError(err)
				transferAuthz := s.GetTransferAuthorization(s.ctx, differentAddress, s.address)
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, defaultCoins)
				return []interface{}{
					differentAddress,
					path.EndpointA.ChannelConfig.PortID,
					path.EndpointB.ChannelID,
					utils.BaseDenom,
					big.NewInt(1e18),
				}
			},
			func([]byte, []interface{}) {
				transferAuthz := s.GetTransferAuthorization(s.ctx, differentAddress, s.address)
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
				path := NewTransferPath(s.chainA, s.chainB)
				s.coordinator.Setup(path)
				err := s.NewTransferAuthorization(s.ctx, s.app, differentAddress, s.address, path, mutliSpendLimit, nil)
				s.Require().NoError(err)
				transferAuthz := s.GetTransferAuthorization(s.ctx, differentAddress, s.address)
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, mutliSpendLimit)
				return []interface{}{
					differentAddress,
					path.EndpointA.ChannelConfig.PortID,
					path.EndpointB.ChannelID,
					"uatom",
					big.NewInt(1e18),
				}
			},
			func([]byte, []interface{}) {
				transferAuthz := s.GetTransferAuthorization(s.ctx, differentAddress, s.address)
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
				path := NewTransferPath(s.chainA, s.chainB)
				s.coordinator.Setup(path)
				allocations := []transfertypes.Allocation{
					{
						SourcePort:    "port-01",
						SourceChannel: "channel-03",
						SpendLimit:    atomCoins,
						AllowList:     nil,
					},
					{
						SourcePort:    path.EndpointA.ChannelConfig.PortID,
						SourceChannel: path.EndpointA.ChannelID,
						SpendLimit:    defaultCoins,
						AllowList:     nil,
					},
				}
				err := s.NewTransferAuthorizationWithAllocations(s.ctx, s.app, differentAddress, s.address, allocations)
				s.Require().NoError(err)
				transferAuthz := s.GetTransferAuthorization(s.ctx, differentAddress, s.address)
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, atomCoins)
				s.Require().Equal(transferAuthz.Allocations[1].SpendLimit, defaultCoins)
				return []interface{}{
					differentAddress,
					path.EndpointA.ChannelConfig.PortID,
					path.EndpointB.ChannelID,
					utils.BaseDenom,
					big.NewInt(1e18),
				}
			},
			func([]byte, []interface{}) {
				transferAuthz := s.GetTransferAuthorization(s.ctx, differentAddress, s.address)
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, atomCoins)
				s.Require().Equal(transferAuthz.Allocations[1].SpendLimit[0].Amount, math.NewInt(2e18))
				s.Require().Equal(transferAuthz.Allocations[1].SpendLimit[0].Denom, utils.BaseDenom)
			},
			200000,
			false,
			"",
		},
	}
	testCases = append(testCases, defaultAllowanceCases...)

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			args := tc.malleate()
			bz, err := s.precompile.IncreaseAllowance(s.ctx, s.address, s.stateDB, &method, args)

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
				path := NewTransferPath(s.chainA, s.chainB)
				s.coordinator.Setup(path)
				err := s.NewTransferAuthorization(s.ctx, s.app, differentAddress, s.address, path, defaultCoins, nil)
				s.Require().NoError(err)
				transferAuthz := s.GetTransferAuthorization(s.ctx, differentAddress, s.address)
				s.Require().NotNil(transferAuthz)
				s.Require().Len(transferAuthz.Allocations, 1)
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, defaultCoins)
				return []interface{}{
					differentAddress,
					path.EndpointA.ChannelConfig.PortID,
					path.EndpointB.ChannelID,
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
				path := NewTransferPath(s.chainA, s.chainB)
				s.coordinator.Setup(path)
				err := s.NewTransferAuthorization(s.ctx, s.app, differentAddress, s.address, path, defaultCoins, nil)
				s.Require().NoError(err)
				transferAuthz := s.GetTransferAuthorization(s.ctx, differentAddress, s.address)
				s.Require().NotNil(transferAuthz)
				s.Require().Len(transferAuthz.Allocations, 1)
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, defaultCoins)
				return []interface{}{
					differentAddress,
					path.EndpointA.ChannelConfig.PortID,
					path.EndpointB.ChannelID,
					utils.BaseDenom,
					big.NewInt(500000000000000000),
				}
			},
			func(_ []byte, _ []interface{}) {
				transferAuthz := s.GetTransferAuthorization(s.ctx, differentAddress, s.address)
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
				path := NewTransferPath(s.chainA, s.chainB)
				s.coordinator.Setup(path)
				err := s.NewTransferAuthorization(s.ctx, s.app, differentAddress, s.address, path, mutliSpendLimit, nil)
				s.Require().NoError(err)
				transferAuthz := s.GetTransferAuthorization(s.ctx, differentAddress, s.address)
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, mutliSpendLimit)
				return []interface{}{
					differentAddress,
					path.EndpointA.ChannelConfig.PortID,
					path.EndpointB.ChannelID,
					"uatom",
					big.NewInt(500000000000000000),
				}
			},
			func(_ []byte, _ []interface{}) {
				transferAuthz := s.GetTransferAuthorization(s.ctx, differentAddress, s.address)
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
				path := NewTransferPath(s.chainA, s.chainB)
				s.coordinator.Setup(path)
				allocations := []transfertypes.Allocation{
					{
						SourcePort:    "port-01",
						SourceChannel: "channel-03",
						SpendLimit:    atomCoins,
						AllowList:     nil,
					},
					{
						SourcePort:    path.EndpointA.ChannelConfig.PortID,
						SourceChannel: path.EndpointA.ChannelID,
						SpendLimit:    defaultCoins,
						AllowList:     nil,
					},
				}
				err := s.NewTransferAuthorizationWithAllocations(s.ctx, s.app, differentAddress, s.address, allocations)
				s.Require().NoError(err)
				transferAuthz := s.GetTransferAuthorization(s.ctx, differentAddress, s.address)
				s.Require().Equal(transferAuthz.Allocations[0].SpendLimit, atomCoins)
				s.Require().Equal(transferAuthz.Allocations[1].SpendLimit, defaultCoins)
				return []interface{}{
					differentAddress,
					path.EndpointA.ChannelConfig.PortID,
					path.EndpointB.ChannelID,
					utils.BaseDenom,
					big.NewInt(1e18 / 2),
				}
			},
			func(_ []byte, _ []interface{}) {
				transferAuthz := s.GetTransferAuthorization(s.ctx, differentAddress, s.address)
				s.Require().Equal(transferAuthz.Allocations[1].SpendLimit[0].Amount, math.NewInt(1e18/2))
				s.Require().Equal(transferAuthz.Allocations[1].SpendLimit[0].Denom, utils.BaseDenom)
			},
			200000,
			false,
			"",
		},
	}

	testCases = append(testCases, defaultAllowanceCases...)

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			args := tc.malleate()
			bz, err := s.precompile.DecreaseAllowance(s.ctx, s.address, s.stateDB, &method, args)

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
