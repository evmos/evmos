package ics20_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v16/precompiles/authorization"
	cmn "github.com/evmos/evmos/v16/precompiles/common"
	"github.com/evmos/evmos/v16/precompiles/ics20"
	"github.com/evmos/evmos/v16/utils"
)

func (s *PrecompileTestSuite) TestTransferEvent() {
	method := s.precompile.Methods[ics20.TransferMethod]
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
				path := NewTransferPath(s.chainA, s.chainB)
				s.coordinator.Setup(path)
				err := s.NewTransferAuthorization(s.ctx, s.app, common.BytesToAddress(sender), common.BytesToAddress(sender), path, defaultCoins, nil)
				s.Require().NoError(err)
				return []interface{}{
					path.EndpointA.ChannelConfig.PortID,
					path.EndpointA.ChannelID,
					utils.BaseDenom,
					big.NewInt(1e18),
					common.BytesToAddress(sender.Bytes()),
					receiver.String(),
					s.chainB.GetTimeoutHeight(),
					uint64(0),
					"memo",
				}
			},
			false,
			"",
			func(sender, receiver sdk.AccAddress) {
				log := s.stateDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[ics20.EventTypeIBCTransfer]
				s.Require().Equal(event.ID, common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.ctx.BlockHeight()))

				var ibcTransferEvent ics20.EventIBCTransfer
				err := cmn.UnpackLog(s.precompile.ABI, &ibcTransferEvent, ics20.EventTypeIBCTransfer, *log)
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

			sender := s.chainA.SenderAccount.GetAddress()
			receiver := s.chainB.SenderAccount.GetAddress()
			contract := vm.NewContract(vm.AccountRef(sender), s.precompile, big.NewInt(0), 20000)
			_, err := s.precompile.Transfer(s.ctx, common.BytesToAddress(sender), contract, s.stateDB, &method, tc.malleate(sender, receiver))

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
	method := s.precompile.Methods[authorization.ApproveMethod]
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
				path := NewTransferPath(s.chainA, s.chainB)
				s.coordinator.Setup(path)
				return []interface{}{
					s.address,
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
			false,
			"",
			func() {
				log := s.stateDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[authorization.EventTypeIBCTransferAuthorization]
				s.Require().Equal(event.ID, common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.ctx.BlockHeight()))

				var transferAuthorizationEvent ics20.EventTransferAuthorization
				err := cmn.UnpackLog(s.precompile.ABI, &transferAuthorizationEvent, authorization.EventTypeIBCTransferAuthorization, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.address, transferAuthorizationEvent.Granter)
				s.Require().Equal(s.address, transferAuthorizationEvent.Grantee)
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

			_, err := s.precompile.Approve(s.ctx, s.address, s.stateDB, &method, tc.malleate())

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
	method := s.precompile.Methods[authorization.ApproveMethod]
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
				path := NewTransferPath(s.chainA, s.chainB)
				s.coordinator.Setup(path)
				err := s.NewTransferAuthorization(s.ctx, s.app, s.address, s.address, path, defaultCoins, nil)
				s.Require().NoError(err)
				return []interface{}{
					s.address,
				}
			},
			false,
			"",
			func() {
				log := s.stateDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[authorization.EventTypeIBCTransferAuthorization]
				s.Require().Equal(event.ID, common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.ctx.BlockHeight()))

				var transferRevokeAuthorizationEvent ics20.EventTransferAuthorization
				err := cmn.UnpackLog(s.precompile.ABI, &transferRevokeAuthorizationEvent, authorization.EventTypeIBCTransferAuthorization, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.address, transferRevokeAuthorizationEvent.Grantee)
				s.Require().Equal(s.address, transferRevokeAuthorizationEvent.Granter)
				s.Require().Equal(0, len(transferRevokeAuthorizationEvent.Allocations))
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset

			_, err := s.precompile.Revoke(s.ctx, s.address, s.stateDB, &method, tc.malleate())

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
	method := s.precompile.Methods[authorization.IncreaseAllowanceMethod]
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
				path := NewTransferPath(s.chainA, s.chainB)
				s.coordinator.Setup(path)
				err := s.NewTransferAuthorization(s.ctx, s.app, s.address, s.address, path, defaultCoins, nil)
				s.Require().NoError(err)
				return []interface{}{
					s.address,
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

			_, err := s.precompile.IncreaseAllowance(s.ctx, s.address, s.stateDB, &method, tc.malleate())

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
	method := s.precompile.Methods[authorization.DecreaseAllowanceMethod]
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
				path := NewTransferPath(s.chainA, s.chainB)
				s.coordinator.Setup(path)
				err := s.NewTransferAuthorization(s.ctx, s.app, s.address, s.address, path, defaultCoins, nil)
				s.Require().NoError(err)
				return []interface{}{
					s.address,
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

			_, err := s.precompile.DecreaseAllowance(s.ctx, s.address, s.stateDB, &method, tc.malleate())

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
