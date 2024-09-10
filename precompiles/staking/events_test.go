package staking_test

import (
	"math/big"

	testkeyring "github.com/evmos/evmos/v19/testutil/integration/evmos/keyring"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v19/precompiles/authorization"
	cmn "github.com/evmos/evmos/v19/precompiles/common"
	"github.com/evmos/evmos/v19/precompiles/staking"
	"github.com/evmos/evmos/v19/x/evm/core/vm"
	"github.com/evmos/evmos/v19/x/evm/statedb"
)

func (s *PrecompileTestSuite) TestApprovalEvent() {
	var (
		stDB *statedb.StateDB
		ctx  sdk.Context
	)
	method := s.precompile.Methods[authorization.ApproveMethod]
	testCases := []struct {
		name        string
		malleate    func(grantee common.Address) []interface{}
		expErr      bool
		errContains string
		postCheck   func(granter, grantee common.Address)
	}{
		{
			"success - all four methods are present in the emitted event",
			func(grantee common.Address) []interface{} {
				return []interface{}{
					grantee,
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
			func(granter, grantee common.Address) {
				log := stDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[authorization.EventTypeApproval]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight())) //nolint:gosec // G115

				var approvalEvent authorization.EventApproval
				err := cmn.UnpackLog(s.precompile.ABI, &approvalEvent, authorization.EventTypeApproval, *log)
				s.Require().NoError(err)
				s.Require().Equal(grantee, approvalEvent.Grantee)
				s.Require().Equal(granter, approvalEvent.Granter)
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
			ctx = s.network.GetContext()
			stDB = s.network.GetStateDB()

			granter := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			err := s.CreateAuthorization(ctx, granter.AccAddr, grantee.AccAddr, staking.DelegateAuthz, nil)
			s.Require().NoError(err)

			approveArgs := tc.malleate(grantee.Addr)
			_, err = s.precompile.Approve(ctx, granter.Addr, stDB, &method, approveArgs)

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck(granter.Addr, grantee.Addr)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestIncreaseAllowanceEvent() {
	var (
		stDB *statedb.StateDB
		ctx  sdk.Context
	)
	approvalMethod := s.precompile.Methods[authorization.ApproveMethod]
	method := s.precompile.Methods[authorization.IncreaseAllowanceMethod]
	testCases := []struct {
		name        string
		malleate    func(grantee common.Address) []interface{}
		expErr      bool
		errContains string
		postCheck   func(granter, grantee common.Address)
	}{
		{
			"success - increased allowance for all 3 methods by 1 evmos",
			func(grantee common.Address) []interface{} {
				return []interface{}{
					grantee,
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
			func(granter, grantee common.Address) {
				log := stDB.Logs()[1]
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
				s.CheckAllowanceChangeEvent(log, methods, amounts, granter, grantee)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()
			stDB = s.network.GetStateDB()

			granter := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			err := s.CreateAuthorization(ctx, granter.AccAddr, grantee.AccAddr, staking.DelegateAuthz, nil)
			s.Require().NoError(err)

			// Approve first with 1 evmos
			approveArgs := tc.malleate(grantee.Addr)
			_, err = s.precompile.Approve(ctx, granter.Addr, stDB, &approvalMethod, approveArgs)
			s.Require().NoError(err)

			// Increase allowance after approval
			_, err = s.precompile.IncreaseAllowance(ctx, granter.Addr, stDB, &method, approveArgs)

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck(granter.Addr, grantee.Addr)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestDecreaseAllowanceEvent() {
	var (
		stDB *statedb.StateDB
		ctx  sdk.Context
	)
	approvalMethod := s.precompile.Methods[authorization.ApproveMethod]
	method := s.precompile.Methods[authorization.DecreaseAllowanceMethod]
	testCases := []struct {
		name        string
		malleate    func(grantee common.Address) []interface{}
		expErr      bool
		errContains string
		postCheck   func(granter, grantee common.Address)
	}{
		{
			"success - decreased allowance for all 3 methods by 1 evmos",
			func(grantee common.Address) []interface{} {
				return []interface{}{
					grantee,
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
			func(granter, grantee common.Address) {
				log := stDB.Logs()[1]
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
				s.CheckAllowanceChangeEvent(log, methods, amounts, granter, grantee)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()
			stDB = s.network.GetStateDB()

			granter := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			err := s.CreateAuthorization(ctx, granter.AccAddr, grantee.AccAddr, staking.DelegateAuthz, nil)
			s.Require().NoError(err)

			// Approve first with 2 evmos
			args := []interface{}{
				grantee.Addr,
				big.NewInt(2000000000000000000),
				[]string{
					staking.DelegateMsg,
					staking.UndelegateMsg,
					staking.RedelegateMsg,
				},
			}
			_, err = s.precompile.Approve(ctx, granter.Addr, stDB, &approvalMethod, args)
			s.Require().NoError(err)

			// Decrease allowance after approval
			_, err = s.precompile.DecreaseAllowance(ctx, granter.Addr, stDB, &method, tc.malleate(grantee.Addr))

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck(granter.Addr, grantee.Addr)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestCreateValidatorEvent() {
	var (
		stDB            *statedb.StateDB
		ctx             sdk.Context
		delegationValue = big.NewInt(1205000000000000000)
		method          = s.precompile.Methods[staking.CreateValidatorMethod]
		pubkey          = "nfJ0axJC9dhta1MAE1EBFaVdxxkYzxYrBaHuJVjG//M="
	)

	testCases := []struct {
		name        string
		malleate    func(delegator common.Address) []interface{}
		expErr      bool
		errContains string
		postCheck   func(delegator common.Address)
	}{
		{
			name: "success - the correct event is emitted",
			malleate: func(delegator common.Address) []interface{} {
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
					delegator,
					pubkey,
					delegationValue,
				}
			},
			postCheck: func(delegator common.Address) {
				log := stDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeCreateValidator]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight())) //nolint:gosec // G115

				// Check the fully unpacked event matches the one emitted
				var createValidatorEvent staking.EventCreateValidator
				err := cmn.UnpackLog(s.precompile.ABI, &createValidatorEvent, staking.EventTypeCreateValidator, *log)
				s.Require().NoError(err)
				s.Require().Equal(delegator, createValidatorEvent.ValidatorAddress)
				s.Require().Equal(delegationValue, createValidatorEvent.Value)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()
			stDB = s.network.GetStateDB()

			delegator := s.keyring.GetKey(0)

			contract := vm.NewContract(vm.AccountRef(delegator.Addr), s.precompile, big.NewInt(0), 200000)
			_, err := s.precompile.CreateValidator(ctx, delegator.Addr, contract, stDB, &method, tc.malleate(delegator.Addr))

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck(delegator.Addr)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestEditValidatorEvent() {
	var (
		stDB        *statedb.StateDB
		ctx         sdk.Context
		valOperAddr common.Address
		method      = s.precompile.Methods[staking.EditValidatorMethod]
		minSelfDel  = big.NewInt(11)
		commRate    = math.LegacyNewDecWithPrec(5, 2).BigInt()
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
						Moniker:         "node0-edited",
						Identity:        "",
						Website:         "",
						SecurityContact: "",
						Details:         "",
					},
					valOperAddr,
					commRate,
					minSelfDel,
				}
			},
			postCheck: func() {
				s.Require().Equal(len(stDB.Logs()), 1)
				log := stDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeEditValidator]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight())) //nolint:gosec // G115

				// Check the fully unpacked event matches the one emitted
				var editValidatorEvent staking.EventEditValidator
				err := cmn.UnpackLog(s.precompile.ABI, &editValidatorEvent, staking.EventTypeEditValidator, *log)
				s.Require().NoError(err)
				s.Require().Equal(valOperAddr, editValidatorEvent.ValidatorAddress)
				s.Require().Equal(minSelfDel, editValidatorEvent.MinSelfDelegation)
				s.Require().Equal(commRate, editValidatorEvent.CommissionRate)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()
			stDB = s.network.GetStateDB()

			acc, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].GetOperator())
			s.Require().NoError(err)
			valOperAddr = common.BytesToAddress(acc.Bytes())

			contract := vm.NewContract(vm.AccountRef(valOperAddr), s.precompile, big.NewInt(0), 200000)
			_, err = s.precompile.EditValidator(ctx, valOperAddr, contract, stDB, &method, tc.malleate())

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
		stDB          *statedb.StateDB
		ctx           sdk.Context
		delegationAmt = big.NewInt(1500000000000000000)
		newSharesExp  = delegationAmt
		method        = s.precompile.Methods[staking.DelegateMethod]
	)
	testCases := []struct {
		name        string
		malleate    func(delegator common.Address) []interface{}
		expErr      bool
		errContains string
		postCheck   func(delegator common.Address)
	}{
		{
			"success - the correct event is emitted",
			func(delegator common.Address) []interface{} {
				return []interface{}{
					delegator,
					s.network.GetValidators()[0].OperatorAddress,
					delegationAmt,
				}
			},
			false,
			"",
			func(delegator common.Address) {
				log := stDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeDelegate]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight())) //nolint:gosec // G115

				optAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].OperatorAddress)
				s.Require().NoError(err)
				optHexAddr := common.BytesToAddress(optAddr)

				// Check the fully unpacked event matches the one emitted
				var delegationEvent staking.EventDelegate
				err = cmn.UnpackLog(s.precompile.ABI, &delegationEvent, staking.EventTypeDelegate, *log)
				s.Require().NoError(err)
				s.Require().Equal(delegator, delegationEvent.DelegatorAddress)
				s.Require().Equal(optHexAddr, delegationEvent.ValidatorAddress)
				s.Require().Equal(delegationAmt, delegationEvent.Amount)
				s.Require().Equal(newSharesExp, delegationEvent.NewShares)
			},
		},
	}

	for _, tc := range testCases { //nolint:dupl
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()
			stDB = s.network.GetStateDB()

			delegator := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			err := s.CreateAuthorization(ctx, delegator.AccAddr, grantee.AccAddr, staking.DelegateAuthz, nil)
			s.Require().NoError(err)

			contract := vm.NewContract(vm.AccountRef(delegator.Addr), s.precompile, big.NewInt(0), 20000)
			_, err = s.precompile.Delegate(ctx, delegator.Addr, contract, stDB, &method, tc.malleate(delegator.Addr))

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck(delegator.Addr)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestUnbondEvent() {
	var (
		stDB *statedb.StateDB
		ctx  sdk.Context
	)
	method := s.precompile.Methods[staking.UndelegateMethod]

	testCases := []struct {
		name        string
		malleate    func(delegator common.Address) []interface{}
		expErr      bool
		errContains string
		postCheck   func(delegator common.Address)
	}{
		{
			"success - the correct event is emitted",
			func(delegator common.Address) []interface{} {
				return []interface{}{
					delegator,
					s.network.GetValidators()[0].OperatorAddress,
					big.NewInt(1000000000000000000),
				}
			},
			false,
			"",
			func(delegator common.Address) {
				log := stDB.Logs()[0]
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeUnbond]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight())) //nolint:gosec // G115

				optAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].OperatorAddress)
				s.Require().NoError(err)
				optHexAddr := common.BytesToAddress(optAddr)

				// Check the fully unpacked event matches the one emitted
				var unbondEvent staking.EventUnbond
				err = cmn.UnpackLog(s.precompile.ABI, &unbondEvent, staking.EventTypeUnbond, *log)
				s.Require().NoError(err)
				s.Require().Equal(delegator, unbondEvent.DelegatorAddress)
				s.Require().Equal(optHexAddr, unbondEvent.ValidatorAddress)
				s.Require().Equal(big.NewInt(1000000000000000000), unbondEvent.Amount)
			},
		},
	}

	for _, tc := range testCases { //nolint:dupl
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()
			stDB = s.network.GetStateDB()

			delegator := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			err := s.CreateAuthorization(ctx, delegator.AccAddr, grantee.AccAddr, staking.UndelegateAuthz, nil)
			s.Require().NoError(err)

			contract := vm.NewContract(vm.AccountRef(delegator.Addr), s.precompile, big.NewInt(0), 20000)
			_, err = s.precompile.Undelegate(ctx, delegator.Addr, contract, stDB, &method, tc.malleate(delegator.Addr))

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck(delegator.Addr)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestRedelegateEvent() {
	var (
		stDB *statedb.StateDB
		ctx  sdk.Context
	)
	method := s.precompile.Methods[staking.RedelegateMethod]

	testCases := []struct {
		name        string
		malleate    func(delegator common.Address) []interface{}
		expErr      bool
		errContains string
		postCheck   func(delegator common.Address)
	}{
		{
			"success - the correct event is emitted",
			func(delegator common.Address) []interface{} {
				return []interface{}{
					delegator,
					s.network.GetValidators()[0].OperatorAddress,
					s.network.GetValidators()[1].OperatorAddress,
					big.NewInt(1000000000000000000),
				}
			},
			false,
			"",
			func(delegator common.Address) {
				log := stDB.Logs()[0]
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeRedelegate]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight())) //nolint:gosec // G115

				optSrcAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].OperatorAddress)
				s.Require().NoError(err)
				optSrcHexAddr := common.BytesToAddress(optSrcAddr)

				optDstAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[1].OperatorAddress)
				s.Require().NoError(err)
				optDstHexAddr := common.BytesToAddress(optDstAddr)

				var redelegateEvent staking.EventRedelegate
				err = cmn.UnpackLog(s.precompile.ABI, &redelegateEvent, staking.EventTypeRedelegate, *log)
				s.Require().NoError(err)
				s.Require().Equal(delegator, redelegateEvent.DelegatorAddress)
				s.Require().Equal(optSrcHexAddr, redelegateEvent.ValidatorSrcAddress)
				s.Require().Equal(optDstHexAddr, redelegateEvent.ValidatorDstAddress)
				s.Require().Equal(big.NewInt(1000000000000000000), redelegateEvent.Amount)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()
			stDB = s.network.GetStateDB()

			delegator := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			err := s.CreateAuthorization(ctx, delegator.AccAddr, grantee.AccAddr, staking.RedelegateAuthz, nil)
			s.Require().NoError(err)

			contract := vm.NewContract(vm.AccountRef(delegator.Addr), s.precompile, big.NewInt(0), 20000)
			_, err = s.precompile.Redelegate(ctx, delegator.Addr, contract, stDB, &method, tc.malleate(delegator.Addr))
			s.Require().NoError(err)

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				tc.postCheck(delegator.Addr)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestCancelUnbondingDelegationEvent() {
	var (
		stDB *statedb.StateDB
		ctx  sdk.Context
	)
	methodCancelUnbonding := s.precompile.Methods[staking.CancelUnbondingDelegationMethod]
	methodUndelegate := s.precompile.Methods[staking.UndelegateMethod]

	testCases := []struct {
		name        string
		malleate    func(contract *vm.Contract, delegator, grantee testkeyring.Key) []interface{}
		expErr      bool
		errContains string
		postCheck   func(delegator common.Address)
	}{
		{
			"success - the correct event is emitted",
			func(contract *vm.Contract, delegator, grantee testkeyring.Key) []interface{} {
				err := s.CreateAuthorization(ctx, delegator.AccAddr, grantee.AccAddr, staking.UndelegateAuthz, nil)
				s.Require().NoError(err)
				undelegateArgs := []interface{}{
					delegator.Addr,
					s.network.GetValidators()[0].OperatorAddress,
					big.NewInt(1000000000000000000),
				}
				_, err = s.precompile.Undelegate(ctx, delegator.Addr, contract, stDB, &methodUndelegate, undelegateArgs)
				s.Require().NoError(err)

				return []interface{}{
					delegator.Addr,
					s.network.GetValidators()[0].OperatorAddress,
					big.NewInt(1000000000000000000),
					big.NewInt(1),
				}
			},
			false,
			"",
			func(delegator common.Address) {
				log := stDB.Logs()[1]

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeCancelUnbondingDelegation]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight())) //nolint:gosec // G115

				optAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].OperatorAddress)
				s.Require().NoError(err)
				optHexAddr := common.BytesToAddress(optAddr)

				// Check event fields match the ones emitted
				var cancelUnbondEvent staking.EventCancelUnbonding
				err = cmn.UnpackLog(s.precompile.ABI, &cancelUnbondEvent, staking.EventTypeCancelUnbondingDelegation, *log)
				s.Require().NoError(err)
				s.Require().Equal(delegator, cancelUnbondEvent.DelegatorAddress)
				s.Require().Equal(optHexAddr, cancelUnbondEvent.ValidatorAddress)
				s.Require().Equal(big.NewInt(1000000000000000000), cancelUnbondEvent.Amount)
				s.Require().Equal(big.NewInt(1), cancelUnbondEvent.CreationHeight)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()
			stDB = s.network.GetStateDB()

			delegator := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			err := s.CreateAuthorization(ctx, delegator.AccAddr, grantee.AccAddr, staking.CancelUnbondingDelegationAuthz, nil)
			s.Require().NoError(err)

			contract := vm.NewContract(vm.AccountRef(delegator.Addr), s.precompile, big.NewInt(0), 20000)
			callArgs := tc.malleate(contract, delegator, grantee)
			_, err = s.precompile.CancelUnbondingDelegation(ctx, delegator.Addr, contract, stDB, &methodCancelUnbonding, callArgs)
			s.Require().NoError(err)

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				tc.postCheck(delegator.Addr)
			}
		})
	}
}
