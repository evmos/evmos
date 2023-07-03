package staking_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v13/precompiles/authorization"
	cmn "github.com/evmos/evmos/v13/precompiles/common"
	"github.com/evmos/evmos/v13/precompiles/staking"
)

func (s *PrecompileTestSuite) TestApprovalEvent() {
	method := s.precompile.Methods[authorization.ApproveMethod]
	testCases := []struct {
		name        string
		malleate    func() []interface{}
		expErr      bool
		errContains string
		postCheck   func()
	}{
		{
			"success - all four methods are present in the emitted event",
			func() []interface{} {
				return []interface{}{
					s.address,
					abi.MaxUint256,
					[]string{
						staking.DelegateMsg,
						staking.UndelegateMsg,
						staking.RedelegateMsg,
						staking.CancelUnbondingDelegationMsg,
					},
				}
			},
			false,
			"",
			func() {
				log := s.stateDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[authorization.EventTypeApproval]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.ctx.BlockHeight()))

				var approvalEvent staking.EventApproval
				err := cmn.UnpackLog(s.precompile.ABI, &approvalEvent, authorization.EventTypeApproval, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.address, approvalEvent.Spender)
				s.Require().Equal(s.address, approvalEvent.Owner)
				s.Require().Equal(abi.MaxUint256, approvalEvent.Value)
				s.Require().Equal(4, len(approvalEvent.Methods))
				s.Require().Equal(staking.DelegateMsg, approvalEvent.Methods[0])
				s.Require().Equal(staking.UndelegateMsg, approvalEvent.Methods[1])
				s.Require().Equal(staking.RedelegateMsg, approvalEvent.Methods[2])
				s.Require().Equal(staking.CancelUnbondingDelegationMsg, approvalEvent.Methods[3])
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset

			err := s.CreateAuthorization(s.address, staking.DelegateAuthz, nil)
			s.Require().NoError(err)

			_, err = s.precompile.Approve(s.ctx, s.address, s.stateDB, &method, tc.malleate())

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
	approvalMethod := s.precompile.Methods[authorization.ApproveMethod]
	method := s.precompile.Methods[authorization.IncreaseAllowanceMethod]
	testCases := []struct {
		name        string
		malleate    func() []interface{}
		expErr      bool
		errContains string
		postCheck   func()
	}{
		{
			"success - increased allowance for all 3 methods by 1 evmos",
			func() []interface{} {
				return []interface{}{
					s.address,
					big.NewInt(1000000000000000000),
					[]string{
						staking.DelegateMsg,
						staking.UndelegateMsg,
						staking.RedelegateMsg,
					},
				}
			},
			false,
			"",
			func() {
				log := s.stateDB.Logs()[1]
				methods := []string{
					staking.DelegateMsg,
					staking.UndelegateMsg,
					staking.RedelegateMsg,
				}
				amounts := []*big.Int{
					big.NewInt(2000000000000000000),
					big.NewInt(2000000000000000000),
					big.NewInt(2000000000000000000),
				}
				s.CheckAllowanceChangeEvent(log, methods, amounts)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset

			err := s.CreateAuthorization(s.address, staking.DelegateAuthz, nil)
			s.Require().NoError(err)

			// Approve first with 1 evmos
			_, err = s.precompile.Approve(s.ctx, s.address, s.stateDB, &approvalMethod, tc.malleate())
			s.Require().NoError(err)

			// Increase allowance after approval
			_, err = s.precompile.IncreaseAllowance(s.ctx, s.address, s.stateDB, &method, tc.malleate())

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
	approvalMethod := s.precompile.Methods[authorization.ApproveMethod]
	method := s.precompile.Methods[authorization.DecreaseAllowanceMethod]
	testCases := []struct {
		name        string
		malleate    func() []interface{}
		expErr      bool
		errContains string
		postCheck   func()
	}{
		{
			"success - decreased allowance for all 3 methods by 1 evmos",
			func() []interface{} {
				return []interface{}{
					s.address,
					big.NewInt(1000000000000000000),
					[]string{
						staking.DelegateMsg,
						staking.UndelegateMsg,
						staking.RedelegateMsg,
					},
				}
			},
			false,
			"",
			func() {
				log := s.stateDB.Logs()[1]
				methods := []string{
					staking.DelegateMsg,
					staking.UndelegateMsg,
					staking.RedelegateMsg,
				}
				amounts := []*big.Int{
					big.NewInt(1000000000000000000),
					big.NewInt(1000000000000000000),
					big.NewInt(1000000000000000000),
				}
				s.CheckAllowanceChangeEvent(log, methods, amounts)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset

			err := s.CreateAuthorization(s.address, staking.DelegateAuthz, nil)
			s.Require().NoError(err)

			// Approve first with 2 evmos
			args := []interface{}{
				s.address,
				big.NewInt(2000000000000000000),
				[]string{
					staking.DelegateMsg,
					staking.UndelegateMsg,
					staking.RedelegateMsg,
				},
			}
			_, err = s.precompile.Approve(s.ctx, s.address, s.stateDB, &approvalMethod, args)
			s.Require().NoError(err)

			// Decrease allowance after approval
			_, err = s.precompile.DecreaseAllowance(s.ctx, s.address, s.stateDB, &method, tc.malleate())

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

func (s *PrecompileTestSuite) TestDelegateEvent() {
	var (
		delegationAmt = big.NewInt(1500000000000000000)
		newSharesExp  = delegationAmt
		method        = s.precompile.Methods[staking.DelegateMethod]
	)
	testCases := []struct {
		name        string
		malleate    func() []interface{}
		expErr      bool
		errContains string
		postCheck   func()
	}{
		{
			"success - the correct event is emitted",
			func() []interface{} {
				return []interface{}{
					s.address,
					s.validators[0].OperatorAddress,
					delegationAmt,
				}
			},
			false,
			"",
			func() {
				log := s.stateDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeDelegate]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.ctx.BlockHeight()))

				// Check the fully unpacked event matches the one emitted
				var delegationEvent staking.EventDelegate
				err := cmn.UnpackLog(s.precompile.ABI, &delegationEvent, staking.EventTypeDelegate, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.address, delegationEvent.DelegatorAddress)
				s.Require().Equal(crypto.Keccak256Hash([]byte(s.validators[0].OperatorAddress)), delegationEvent.ValidatorAddress)
				s.Require().Equal(delegationAmt, delegationEvent.Amount)
				s.Require().Equal(newSharesExp, delegationEvent.NewShares)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset

			err := s.CreateAuthorization(s.address, staking.DelegateAuthz, nil)
			s.Require().NoError(err)

			contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), 20000)
			_, err = s.precompile.Delegate(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate())

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

func (s *PrecompileTestSuite) TestUnbondEvent() {
	method := s.precompile.Methods[staking.UndelegateMethod]

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		expErr      bool
		errContains string
		postCheck   func()
	}{
		{
			"success - the correct event is emitted",
			func() []interface{} {
				return []interface{}{
					s.address,
					s.validators[0].OperatorAddress,
					big.NewInt(1000000000000000000),
				}
			},
			false,
			"",
			func() {
				log := s.stateDB.Logs()[0]
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeUnbond]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.ctx.BlockHeight()))

				// Check the fully unpacked event matches the one emitted
				var unbondEvent staking.EventUnbond
				err := cmn.UnpackLog(s.precompile.ABI, &unbondEvent, staking.EventTypeUnbond, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.address, unbondEvent.DelegatorAddress)
				s.Require().Equal(crypto.Keccak256Hash([]byte(s.validators[0].OperatorAddress)), unbondEvent.ValidatorAddress)
				s.Require().Equal(big.NewInt(1000000000000000000), unbondEvent.Amount)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset

			err := s.CreateAuthorization(s.address, staking.UndelegateAuthz, nil)
			s.Require().NoError(err)

			contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), 20000)
			_, err = s.precompile.Undelegate(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate())

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

func (s *PrecompileTestSuite) TestRedelegateEvent() {
	method := s.precompile.Methods[staking.RedelegateMethod]

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		expErr      bool
		errContains string
		postCheck   func()
	}{
		{
			"success - the correct event is emitted",
			func() []interface{} {
				return []interface{}{
					s.address,
					s.validators[0].OperatorAddress,
					s.validators[1].OperatorAddress,
					big.NewInt(1000000000000000000),
				}
			},
			false,
			"",
			func() {
				log := s.stateDB.Logs()[0]
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeRedelegate]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.ctx.BlockHeight()))

				var redelegateEvent staking.EventRedelegate
				err := cmn.UnpackLog(s.precompile.ABI, &redelegateEvent, staking.EventTypeRedelegate, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.address, redelegateEvent.DelegatorAddress)
				s.Require().Equal(crypto.Keccak256Hash([]byte(s.validators[0].OperatorAddress)), redelegateEvent.ValidatorSrcAddress)
				s.Require().Equal(crypto.Keccak256Hash([]byte(s.validators[1].OperatorAddress)), redelegateEvent.ValidatorDstAddress)
				s.Require().Equal(big.NewInt(1000000000000000000), redelegateEvent.Amount)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset

			err := s.CreateAuthorization(s.address, staking.RedelegateAuthz, nil)
			s.Require().NoError(err)

			contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), 20000)
			_, err = s.precompile.Redelegate(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate())
			s.Require().NoError(err)

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				tc.postCheck()
			}
		})
	}
}

func (s *PrecompileTestSuite) TestCancelUnbondingDelegationEvent() {
	methodCancelUnbonding := s.precompile.Methods[staking.CancelUnbondingDelegationMethod]
	methodUndelegate := s.precompile.Methods[staking.UndelegateMethod]

	testCases := []struct {
		name        string
		malleate    func(contract *vm.Contract) []interface{}
		expErr      bool
		errContains string
		postCheck   func()
	}{
		{
			"success - the correct event is emitted",
			func(contract *vm.Contract) []interface{} {
				err := s.CreateAuthorization(s.address, staking.UndelegateAuthz, nil)
				s.Require().NoError(err)
				undelegateArgs := []interface{}{
					s.address,
					s.validators[0].OperatorAddress,
					big.NewInt(1000000000000000000),
				}
				_, err = s.precompile.Undelegate(s.ctx, s.address, contract, s.stateDB, &methodUndelegate, undelegateArgs)
				s.Require().NoError(err)

				return []interface{}{
					s.address,
					s.validators[0].OperatorAddress,
					big.NewInt(1000000000000000000),
					big.NewInt(2),
				}
			},
			false,
			"",
			func() {
				log := s.stateDB.Logs()[1]

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeCancelUnbondingDelegation]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.ctx.BlockHeight()))

				// Check event fields match the ones emitted
				var cancelUnbondEvent staking.EventCancelUnbonding
				err := cmn.UnpackLog(s.precompile.ABI, &cancelUnbondEvent, staking.EventTypeCancelUnbondingDelegation, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.address, cancelUnbondEvent.DelegatorAddress)
				s.Require().Equal(crypto.Keccak256Hash([]byte(s.validators[0].OperatorAddress)), cancelUnbondEvent.ValidatorAddress)
				s.Require().Equal(big.NewInt(1000000000000000000), cancelUnbondEvent.Amount)
				s.Require().Equal(big.NewInt(2), cancelUnbondEvent.CreationHeight)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset

			err := s.CreateAuthorization(s.address, staking.CancelUnbondingDelegationAuthz, nil)
			s.Require().NoError(err)

			contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), 20000)
			callArgs := tc.malleate(contract)
			_, err = s.precompile.CancelUnbondingDelegation(s.ctx, s.address, contract, s.stateDB, &methodCancelUnbonding, callArgs)
			s.Require().NoError(err)

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				tc.postCheck()
			}
		})
	}
}
