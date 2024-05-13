package ics20_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	evmosibc "github.com/evmos/evmos/v18/ibc/testing"
	"github.com/evmos/evmos/v18/precompiles/authorization"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	"github.com/evmos/evmos/v18/precompiles/ics20"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v18/testutil/integration/ibc/coordinator"
	"github.com/evmos/evmos/v18/utils"
)

func (s *PrecompileTestSuite) TestTransferEvent() {
	var (
		ctx        sdk.Context
		nw         *network.UnitTestNetwork
		path       *evmosibc.Path
		coord      *coordinator.IntegrationCoordinator
		method     abi.Method
		precompile *ics20.Precompile
	)
	testCases := []struct {
		name        string
		malleate    func(sender, receiver sdk.AccAddress) []interface{}
		expErr      bool
		errContains string
		postCheck   func(sender, receiver sdk.AccAddress)
	}{
		{
			"success - transfer event emitted",
			func(sender, receiver sdk.AccAddress) []interface{} {
				err := s.NewTransferAuthorization(ctx, nw.App, common.BytesToAddress(sender), common.BytesToAddress(sender), path, defaultCoins, nil)
				s.Require().NoError(err)
				return []interface{}{
					path.EndpointA.ChannelConfig.PortID,
					path.EndpointA.ChannelID,
					utils.BaseDenom,
					big.NewInt(1e18),
					common.BytesToAddress(sender.Bytes()),
					receiver.String(),
					coord.GetChain(s.chainB).GetTimeoutHeight(),
					uint64(0),
					"memo",
				}
			},
			false,
			"",
			func(sender, receiver sdk.AccAddress) {
				log := s.stateDB.Logs()[0]
				s.Require().Equal(log.Address, precompile.Address())
				// Check event signature matches the one emitted
				event := precompile.ABI.Events[ics20.EventTypeIBCTransfer]
				s.Require().Equal(event.ID, common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight()))

				var ibcTransferEvent ics20.EventIBCTransfer
				err := cmn.UnpackLog(precompile.ABI, &ibcTransferEvent, ics20.EventTypeIBCTransfer, *log)
				s.Require().NoError(err)
				s.Require().Equal(common.BytesToAddress(sender.Bytes()), ibcTransferEvent.Sender)
				s.Require().Equal(crypto.Keccak256Hash([]byte(receiver.String())), ibcTransferEvent.Receiver)
				s.Require().Equal("transfer", ibcTransferEvent.SourcePort)
				s.Require().Equal("channel-0", ibcTransferEvent.SourceChannel)
				s.Require().Equal(big.NewInt(1e18), ibcTransferEvent.Amount)
				s.Require().Equal(utils.BaseDenom, ibcTransferEvent.Denom)
				s.Require().Equal("memo", ibcTransferEvent.Memo)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			nw = s.network
			ctx = nw.GetContext()
			path = s.transferPath
			coord = s.coordinator
			precompile = s.precompile
			method = s.precompile.Methods[ics20.TransferMethod]

			sender := coord.GetChainSenderAcc(s.chainA).GetAddress()
			receiver := coord.GetChainSenderAcc(s.chainB).GetAddress()
			contract := vm.NewContract(vm.AccountRef(sender), s.precompile, big.NewInt(0), 20000)
			_, err := s.precompile.Transfer(ctx, common.BytesToAddress(sender), contract, s.stateDB, &method, tc.malleate(sender, receiver))

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck(sender, receiver)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestApproveTransferAuthorizationEvent() {
	var (
		ctx        sdk.Context
		keys       keyring.Keyring
		precompile *ics20.Precompile
	)
	testCases := []struct {
		name        string
		malleate    func() []interface{}
		expErr      bool
		errContains string
		postCheck   func()
	}{
		{
			"success - transfer authorization event emitted with default coins ",
			func() []interface{} {
				return []interface{}{
					keys.GetAddr(0),
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
			false,
			"",
			func() {
				log := s.stateDB.Logs()[0]
				s.Require().Equal(log.Address, precompile.Address())
				// Check event signature matches the one emitted
				event := precompile.ABI.Events[authorization.EventTypeIBCTransferAuthorization]
				s.Require().Equal(event.ID, common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight()))

				var transferAuthorizationEvent ics20.EventTransferAuthorization
				err := cmn.UnpackLog(precompile.ABI, &transferAuthorizationEvent, authorization.EventTypeIBCTransferAuthorization, *log)
				s.Require().NoError(err)
				s.Require().Equal(keys.GetAddr(0), transferAuthorizationEvent.Granter)
				s.Require().Equal(keys.GetAddr(0), transferAuthorizationEvent.Grantee)
				s.Require().Equal("transfer", transferAuthorizationEvent.Allocations[0].SourcePort)
				s.Require().Equal("channel-0", transferAuthorizationEvent.Allocations[0].SourceChannel)
				abiCoins := cmn.NewCoinsResponse(defaultCoins)
				s.Require().Equal(abiCoins, transferAuthorizationEvent.Allocations[0].SpendLimit)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()
			keys = s.keyring
			precompile = s.precompile
			method := precompile.Methods[authorization.ApproveMethod]

			_, err := s.precompile.Approve(ctx, keys.GetAddr(0), s.stateDB, &method, tc.malleate())

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}

func (s *PrecompileTestSuite) TestRevokeTransferAuthorizationEvent() {
	var (
		ctx        sdk.Context
		nw         *network.UnitTestNetwork
		keys       keyring.Keyring
		path       *evmosibc.Path
		precompile *ics20.Precompile
	)
	testCases := []struct {
		name        string
		malleate    func() []interface{}
		expErr      bool
		errContains string
		postCheck   func()
	}{
		{
			"success - transfer revoke authorization event emitted",
			func() []interface{} {
				err := s.NewTransferAuthorization(ctx, nw.App, keys.GetAddr(0), keys.GetAddr(0), path, defaultCoins, nil)
				s.Require().NoError(err)
				return []interface{}{
					keys.GetAddr(0),
				}
			},
			false,
			"",
			func() {
				log := s.stateDB.Logs()[0]
				s.Require().Equal(log.Address, precompile.Address())
				// Check event signature matches the one emitted
				event := precompile.ABI.Events[authorization.EventTypeIBCTransferAuthorization]
				s.Require().Equal(event.ID, common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight()))

				var transferRevokeAuthorizationEvent ics20.EventTransferAuthorization
				err := cmn.UnpackLog(precompile.ABI, &transferRevokeAuthorizationEvent, authorization.EventTypeIBCTransferAuthorization, *log)
				s.Require().NoError(err)
				s.Require().Equal(keys.GetAddr(0), transferRevokeAuthorizationEvent.Grantee)
				s.Require().Equal(keys.GetAddr(0), transferRevokeAuthorizationEvent.Granter)
				s.Require().Equal(0, len(transferRevokeAuthorizationEvent.Allocations))
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			nw = s.network
			ctx = nw.GetContext()
			keys = s.keyring
			precompile = s.precompile
			path = s.transferPath

			method := precompile.Methods[authorization.ApproveMethod]

			_, err := s.precompile.Revoke(ctx, keys.GetAddr(0), s.stateDB, &method, tc.malleate())

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}

func (s *PrecompileTestSuite) TestIncreaseAllowanceEvent() {
	var (
		ctx  sdk.Context
		nw   *network.UnitTestNetwork
		keys keyring.Keyring
		path *evmosibc.Path
	)
	testCases := []struct {
		name        string
		malleate    func() []interface{}
		expErr      bool
		errContains string
		postCheck   func()
	}{
		{
			"success - increased allowance by 1 Evmos",
			func() []interface{} {
				err := s.NewTransferAuthorization(ctx, nw.App, keys.GetAddr(0), keys.GetAddr(0), path, defaultCoins, nil)
				s.Require().NoError(err)
				return []interface{}{
					keys.GetAddr(0),
					path.EndpointA.ChannelConfig.PortID,
					path.EndpointA.ChannelID,
					utils.BaseDenom,
					big.NewInt(1e18),
				}
			},
			false,
			"",
			func() {
				log := s.stateDB.Logs()[0]
				amount := big.NewInt(1e18)
				s.CheckAllowanceChangeEvent(log, amount, true)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			nw = s.network
			ctx = nw.GetContext()
			keys = s.keyring
			path = s.transferPath

			method := s.precompile.Methods[authorization.IncreaseAllowanceMethod]
			_, err := s.precompile.IncreaseAllowance(ctx, keys.GetAddr(0), s.stateDB, &method, tc.malleate())

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}

func (s *PrecompileTestSuite) TestDecreaseAllowanceEvent() {
	var (
		ctx  sdk.Context
		nw   *network.UnitTestNetwork
		keys keyring.Keyring
		path *evmosibc.Path
	)
	testCases := []struct {
		name        string
		malleate    func() []interface{}
		expErr      bool
		errContains string
		postCheck   func()
	}{
		{
			"success - decrease allowance by 0.5 Evmos",
			func() []interface{} {
				err := s.NewTransferAuthorization(ctx, nw.App, keys.GetAddr(0), keys.GetAddr(0), path, defaultCoins, nil)
				s.Require().NoError(err)
				return []interface{}{
					keys.GetAddr(0),
					path.EndpointA.ChannelConfig.PortID,
					path.EndpointA.ChannelID,
					utils.BaseDenom,
					big.NewInt(1e18 / 2),
				}
			},
			false,
			"",
			func() {
				log := s.stateDB.Logs()[0]
				amount := big.NewInt(1e18 / 2)
				s.CheckAllowanceChangeEvent(log, amount, false)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			nw = s.network
			ctx = nw.GetContext()
			keys = s.keyring
			path = s.transferPath

			method := s.precompile.Methods[authorization.DecreaseAllowanceMethod]

			_, err := s.precompile.DecreaseAllowance(ctx, keys.GetAddr(0), s.stateDB, &method, tc.malleate())

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}
