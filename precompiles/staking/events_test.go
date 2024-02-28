package staking_test

import (
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v16/precompiles/authorization"
	cmn "github.com/evmos/evmos/v16/precompiles/common"
	"github.com/evmos/evmos/v16/precompiles/staking"
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
					s.keyring.GetAddr(0),
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
				log := s.network.GetStateDB().Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[authorization.EventTypeApproval]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.network.GetContext().BlockHeight()))

				var approvalEvent authorization.EventApproval
				err := cmn.UnpackLog(s.precompile.ABI, &approvalEvent, authorization.EventTypeApproval, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.keyring.GetAddr(0), approvalEvent.Grantee)
				s.Require().Equal(s.keyring.GetAddr(0), approvalEvent.Granter)
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

			err := s.CreateAuthorization(s.keyring.GetAddr(0), staking.DelegateAuthz, nil)
			s.Require().NoError(err)

			_, err = s.precompile.Approve(s.network.GetContext(), s.keyring.GetAddr(0), s.network.GetStateDB(), &method, tc.malleate())

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
					s.keyring.GetAddr(0),
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
				log := s.network.GetStateDB().Logs()[1]
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

			err := s.CreateAuthorization(s.keyring.GetAddr(0), staking.DelegateAuthz, nil)
			s.Require().NoError(err)

			// Approve first with 1 evmos
			_, err = s.precompile.Approve(s.network.GetContext(), s.keyring.GetAddr(0), s.network.GetStateDB(), &approvalMethod, tc.malleate())
			s.Require().NoError(err)

			// Increase allowance after approval
			_, err = s.precompile.IncreaseAllowance(s.network.GetContext(), s.keyring.GetAddr(0), s.network.GetStateDB(), &method, tc.malleate())

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
					s.keyring.GetAddr(0),
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
				log := s.network.GetStateDB().Logs()[1]
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

			err := s.CreateAuthorization(s.keyring.GetAddr(0), staking.DelegateAuthz, nil)
			s.Require().NoError(err)

			// Approve first with 2 evmos
			args := []interface{}{
				s.keyring.GetAddr(0),
				big.NewInt(2000000000000000000),
				[]string{
					staking.DelegateMsg,
					staking.UndelegateMsg,
					staking.RedelegateMsg,
				},
			}
			_, err = s.precompile.Approve(s.network.GetContext(), s.keyring.GetAddr(0), s.network.GetStateDB(), &approvalMethod, args)
			s.Require().NoError(err)

			// Decrease allowance after approval
			_, err = s.precompile.DecreaseAllowance(s.network.GetContext(), s.keyring.GetAddr(0), s.network.GetStateDB(), &method, tc.malleate())

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

func (s *PrecompileTestSuite) TestCreateValidatorEvent() {
	var (
		delegationValue = big.NewInt(1205000000000000000)
		method          = s.precompile.Methods[staking.CreateValidatorMethod]
		pubkey          = "nfJ0axJC9dhta1MAE1EBFaVdxxkYzxYrBaHuJVjG//M="
	)

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		expErr      bool
		errContains string
		postCheck   func()
	}{
		{
			name: "success - the correct event is emitted",
			malleate: func() []interface{} {
				return []interface{}{
					staking.Description{
						Moniker:         "node0",
						Identity:        "",
						Website:         "",
						SecurityContact: "",
						Details:         "",
					},
					staking.Commission{
						Rate:          math.LegacyOneDec().BigInt(),
						MaxRate:       math.LegacyOneDec().BigInt(),
						MaxChangeRate: math.LegacyOneDec().BigInt(),
					},
					big.NewInt(1),
					s.keyring.GetAddr(0),
					pubkey,
					delegationValue,
				}
			},
			postCheck: func() {
				log := s.network.GetStateDB().Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeCreateValidator]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.network.GetContext().BlockHeight()))

				// Check the fully unpacked event matches the one emitted
				var createValidatorEvent staking.EventCreateValidator
				err := cmn.UnpackLog(s.precompile.ABI, &createValidatorEvent, staking.EventTypeCreateValidator, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.keyring.GetAddr(0), createValidatorEvent.ValidatorAddress)
				s.Require().Equal(delegationValue, createValidatorEvent.Value)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset

			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), 200000)
			_, err := s.precompile.CreateValidator(s.network.GetContext(), s.keyring.GetAddr(0), contract, s.network.GetStateDB(), &method, tc.malleate())

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
					s.keyring.GetAddr(0),
					s.network.GetValidators()[0].OperatorAddress,
					delegationAmt,
				}
			},
			false,
			"",
			func() {
				log := s.network.GetStateDB().Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeDelegate]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.network.GetContext().BlockHeight()))

				optAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].OperatorAddress)
				s.Require().NoError(err)
				optHexAddr := common.BytesToAddress(optAddr)

				// Check the fully unpacked event matches the one emitted
				var delegationEvent staking.EventDelegate
				err = cmn.UnpackLog(s.precompile.ABI, &delegationEvent, staking.EventTypeDelegate, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.keyring.GetAddr(0), delegationEvent.DelegatorAddress)
				s.Require().Equal(optHexAddr, delegationEvent.ValidatorAddress)
				s.Require().Equal(delegationAmt, delegationEvent.Amount)
				s.Require().Equal(newSharesExp, delegationEvent.NewShares)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset

			err := s.CreateAuthorization(s.keyring.GetAddr(0), staking.DelegateAuthz, nil)
			s.Require().NoError(err)

			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), 20000)
			_, err = s.precompile.Delegate(s.network.GetContext(), s.keyring.GetAddr(0), contract, s.network.GetStateDB(), &method, tc.malleate())

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
					s.keyring.GetAddr(0),
					s.network.GetValidators()[0].OperatorAddress,
					big.NewInt(1000000000000000000),
				}
			},
			false,
			"",
			func() {
				log := s.network.GetStateDB().Logs()[0]
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeUnbond]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.network.GetContext().BlockHeight()))

				optAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].OperatorAddress)
				s.Require().NoError(err)
				optHexAddr := common.BytesToAddress(optAddr)

				// Check the fully unpacked event matches the one emitted
				var unbondEvent staking.EventUnbond
				err = cmn.UnpackLog(s.precompile.ABI, &unbondEvent, staking.EventTypeUnbond, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.keyring.GetAddr(0), unbondEvent.DelegatorAddress)
				s.Require().Equal(optHexAddr, unbondEvent.ValidatorAddress)
				s.Require().Equal(big.NewInt(1000000000000000000), unbondEvent.Amount)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset

			err := s.CreateAuthorization(s.keyring.GetAddr(0), staking.UndelegateAuthz, nil)
			s.Require().NoError(err)

			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), 20000)
			_, err = s.precompile.Undelegate(s.network.GetContext(), s.keyring.GetAddr(0), contract, s.network.GetStateDB(), &method, tc.malleate())

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
					s.keyring.GetAddr(0),
					s.network.GetValidators()[0].OperatorAddress,
					s.network.GetValidators()[1].OperatorAddress,
					big.NewInt(1000000000000000000),
				}
			},
			false,
			"",
			func() {
				log := s.network.GetStateDB().Logs()[0]
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeRedelegate]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.network.GetContext().BlockHeight()))

				optSrcAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].OperatorAddress)
				s.Require().NoError(err)
				optSrcHexAddr := common.BytesToAddress(optSrcAddr)

				optDstAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[1].OperatorAddress)
				s.Require().NoError(err)
				optDstHexAddr := common.BytesToAddress(optDstAddr)

				var redelegateEvent staking.EventRedelegate
				err = cmn.UnpackLog(s.precompile.ABI, &redelegateEvent, staking.EventTypeRedelegate, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.keyring.GetAddr(0), redelegateEvent.DelegatorAddress)
				s.Require().Equal(optSrcHexAddr, redelegateEvent.ValidatorSrcAddress)
				s.Require().Equal(optDstHexAddr, redelegateEvent.ValidatorDstAddress)
				s.Require().Equal(big.NewInt(1000000000000000000), redelegateEvent.Amount)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset

			err := s.CreateAuthorization(s.keyring.GetAddr(0), staking.RedelegateAuthz, nil)
			s.Require().NoError(err)

			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), 20000)
			_, err = s.precompile.Redelegate(s.network.GetContext(), s.keyring.GetAddr(0), contract, s.network.GetStateDB(), &method, tc.malleate())
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
				err := s.CreateAuthorization(s.keyring.GetAddr(0), staking.UndelegateAuthz, nil)
				s.Require().NoError(err)
				undelegateArgs := []interface{}{
					s.keyring.GetAddr(0),
					s.network.GetValidators()[0].OperatorAddress,
					big.NewInt(1000000000000000000),
				}
				_, err = s.precompile.Undelegate(s.network.GetContext(), s.keyring.GetAddr(0), contract, s.network.GetStateDB(), &methodUndelegate, undelegateArgs)
				s.Require().NoError(err)

				return []interface{}{
					s.keyring.GetAddr(0),
					s.network.GetValidators()[0].OperatorAddress,
					big.NewInt(1000000000000000000),
					big.NewInt(2),
				}
			},
			false,
			"",
			func() {
				log := s.network.GetStateDB().Logs()[1]

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeCancelUnbondingDelegation]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.network.GetContext().BlockHeight()))

				optAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].OperatorAddress)
				s.Require().NoError(err)
				optHexAddr := common.BytesToAddress(optAddr)

				// Check event fields match the ones emitted
				var cancelUnbondEvent staking.EventCancelUnbonding
				err = cmn.UnpackLog(s.precompile.ABI, &cancelUnbondEvent, staking.EventTypeCancelUnbondingDelegation, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.keyring.GetAddr(0), cancelUnbondEvent.DelegatorAddress)
				s.Require().Equal(optHexAddr, cancelUnbondEvent.ValidatorAddress)
				s.Require().Equal(big.NewInt(1000000000000000000), cancelUnbondEvent.Amount)
				s.Require().Equal(big.NewInt(2), cancelUnbondEvent.CreationHeight)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset

			err := s.CreateAuthorization(s.keyring.GetAddr(0), staking.CancelUnbondingDelegationAuthz, nil)
			s.Require().NoError(err)

			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), 20000)
			callArgs := tc.malleate(contract)
			_, err = s.precompile.CancelUnbondingDelegation(s.network.GetContext(), s.keyring.GetAddr(0), contract, s.network.GetStateDB(), &methodCancelUnbonding, callArgs)
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
