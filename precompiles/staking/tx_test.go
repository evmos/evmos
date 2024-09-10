package staking_test

import (
	"encoding/base64"
	"fmt"
	"math/big"

	testkeyring "github.com/evmos/evmos/v19/testutil/integration/evmos/keyring"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	cmn "github.com/evmos/evmos/v19/precompiles/common"
	"github.com/evmos/evmos/v19/precompiles/staking"
	"github.com/evmos/evmos/v19/precompiles/testutil"
	evmosutiltx "github.com/evmos/evmos/v19/testutil/tx"
	"github.com/evmos/evmos/v19/x/evm/core/vm"
	"github.com/evmos/evmos/v19/x/evm/statedb"
)

func (s *PrecompileTestSuite) TestCreateValidator() {
	var (
		stDB        *statedb.StateDB
		method      = s.precompile.Methods[staking.CreateValidatorMethod]
		description = staking.Description{
			Moniker:         "node0",
			Identity:        "",
			Website:         "",
			SecurityContact: "",
			Details:         "",
		}
		commission = staking.Commission{
			Rate:          math.LegacyOneDec().BigInt(),
			MaxRate:       math.LegacyOneDec().BigInt(),
			MaxChangeRate: math.LegacyOneDec().BigInt(),
		}
		minSelfDelegation = big.NewInt(1)
		pubkey            = "nfJ0axJC9dhta1MAE1EBFaVdxxkYzxYrBaHuJVjG//M="
		validatorAddress  common.Address
		value             = big.NewInt(1205000000000000000)
		diffAddr, _       = evmosutiltx.NewAddrKey()
	)

	testCases := []struct {
		name          string
		malleate      func() []interface{}
		gas           uint64
		callerAddress *common.Address
		postCheck     func(data []byte)
		expError      bool
		errContains   string
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			200000,
			nil,
			func([]byte) {},
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 6, 0),
		},
		{
			"fail - different origin than delegator",
			func() []interface{} {
				differentAddr := evmosutiltx.GenerateAddress()
				return []interface{}{
					description,
					commission,
					minSelfDelegation,
					differentAddr,
					pubkey,
					value,
				}
			},
			200000,
			nil,
			func([]byte) {},
			true,
			"is not the same as delegator address",
		},
		{
			"fail - invalid description",
			func() []interface{} {
				return []interface{}{
					"",
					commission,
					minSelfDelegation,
					validatorAddress,
					pubkey,
					value,
				}
			},
			200000,
			nil,
			func([]byte) {},
			true,
			"invalid description",
		},
		{
			"fail - invalid commission",
			func() []interface{} {
				return []interface{}{
					description,
					"",
					minSelfDelegation,
					validatorAddress,
					pubkey,
					value,
				}
			},
			200000,
			nil,
			func([]byte) {},
			true,
			"invalid commission",
		},
		{
			"fail - invalid min self delegation",
			func() []interface{} {
				return []interface{}{
					description,
					commission,
					"",
					validatorAddress,
					pubkey,
					value,
				}
			},
			200000,
			nil,
			func([]byte) {},
			true,
			"invalid amount",
		},
		{
			"fail - invalid validator address",
			func() []interface{} {
				return []interface{}{
					description,
					commission,
					minSelfDelegation,
					1205,
					pubkey,
					value,
				}
			},
			200000,
			nil,
			func([]byte) {},
			true,
			"invalid validator address",
		},
		{
			"fail - invalid pubkey",
			func() []interface{} {
				return []interface{}{
					description,
					commission,
					minSelfDelegation,
					validatorAddress,
					1205,
					value,
				}
			},
			200000,
			nil,
			func([]byte) {},
			true,
			"invalid type for",
		},
		{
			"fail - pubkey decoding error",
			func() []interface{} {
				return []interface{}{
					description,
					commission,
					minSelfDelegation,
					validatorAddress,
					"bHVrZQ=", // base64.StdEncoding.DecodeString error
					value,
				}
			},
			200000,
			nil,
			func([]byte) {},
			true,
			"illegal base64 data",
		},
		{
			"fail - consensus pubkey len is invalid",
			func() []interface{} {
				return []interface{}{
					description,
					commission,
					minSelfDelegation,
					validatorAddress,
					"bHVrZQ==",
					value,
				}
			},
			200000,
			nil,
			func([]byte) {},
			true,
			"consensus pubkey len is invalid",
		},
		{
			"fail - invalid value",
			func() []interface{} {
				return []interface{}{
					description,
					commission,
					minSelfDelegation,
					validatorAddress,
					pubkey,
					"",
				}
			},
			200000,
			nil,
			func([]byte) {},
			true,
			"invalid amount",
		},
		{
			"fail - cannot be called from address != than validator address",
			func() []interface{} {
				return []interface{}{
					description,
					commission,
					minSelfDelegation,
					validatorAddress,
					pubkey,
					value,
				}
			},
			200000,
			&diffAddr,
			func([]byte) {},
			true,
			"this method can only be called directly to the precompile",
		},
		{
			"success",
			func() []interface{} {
				return []interface{}{
					description,
					commission,
					minSelfDelegation,
					validatorAddress,
					pubkey,
					value,
				}
			},
			200000,
			nil,
			func(data []byte) {
				success, err := s.precompile.Unpack(staking.CreateValidatorMethod, data)
				s.Require().NoError(err)
				s.Require().Equal(success[0], true)

				log := stDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeCreateValidator]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.network.GetContext().BlockHeight())) //nolint:gosec // G115

				// Check the fully unpacked event matches the one emitted
				var createValidatorEvent staking.EventCreateValidator
				err = cmn.UnpackLog(s.precompile.ABI, &createValidatorEvent, staking.EventTypeCreateValidator, *log)
				s.Require().NoError(err)
				s.Require().Equal(validatorAddress, createValidatorEvent.ValidatorAddress)
				s.Require().Equal(value, createValidatorEvent.Value)
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			ctx := s.network.GetContext()
			stDB = s.network.GetStateDB()

			// reset sender
			validator := s.keyring.GetKey(0)
			validatorAddress = validator.Addr

			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(s.T(), ctx, validatorAddress, s.precompile, tc.gas)
			if tc.callerAddress != nil {
				contract.CallerAddress = *tc.callerAddress
			}

			bz, err := s.precompile.CreateValidator(ctx, validatorAddress, contract, stDB, &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				// query the validator in the staking keeper
				validator, err := s.network.App.StakingKeeper.Validator(ctx, validator.AccAddr.Bytes())
				s.Require().NoError(err)

				s.Require().NotNil(validator, "expected validator not to be nil")
				tc.postCheck(bz)

				isBonded := validator.IsBonded()
				s.Require().Equal(false, isBonded, "expected validator bonded to be %t; got %t", false, isBonded)

				consPubKey, err := validator.ConsPubKey()
				s.Require().NoError(err)
				consPubKeyBase64 := base64.StdEncoding.EncodeToString(consPubKey.Bytes())
				s.Require().Equal(pubkey, consPubKeyBase64, "expected validator pubkey to be %s; got %s", pubkey, consPubKeyBase64)

				operator := validator.GetOperator()
				s.Require().Equal(sdk.ValAddress(validatorAddress.Bytes()).String(), operator, "expected validator operator to be %s; got %s", validatorAddress, operator)

				commissionRate := validator.GetCommission()
				s.Require().Equal(commission.Rate.String(), commissionRate.BigInt().String(), "expected validator commission rate to be %s; got %s", commission.Rate.String(), commissionRate.String())

				valMinSelfDelegation := validator.GetMinSelfDelegation()
				s.Require().Equal(minSelfDelegation.String(), valMinSelfDelegation.String(), "expected validator min self delegation to be %s; got %s", minSelfDelegation.String(), valMinSelfDelegation.String())

				moniker := validator.GetMoniker()
				s.Require().Equal(description.Moniker, moniker, "expected validator moniker to be %s; got %s", description.Moniker, moniker)

				jailed := validator.IsJailed()
				s.Require().Equal(false, jailed, "expected validator jailed to be %t; got %t", false, jailed)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestEditValidator() {
	var (
		stDB              *statedb.StateDB
		ctx               sdk.Context
		validatorAddress  common.Address
		commissionRate    *big.Int
		minSelfDelegation *big.Int
		method            = s.precompile.Methods[staking.EditValidatorMethod]
		description       = staking.Description{
			Moniker:         "node0-edited",
			Identity:        "",
			Website:         "",
			SecurityContact: "",
			Details:         "",
		}
	)

	testCases := []struct {
		name          string
		malleate      func() []interface{}
		gas           uint64
		callerAddress *common.Address
		postCheck     func(data []byte)
		expError      bool
		errContains   string
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			200000,
			nil,
			func([]byte) {},
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 4, 0),
		},
		{
			"fail - different origin than delegator",
			func() []interface{} {
				differentAddr := evmosutiltx.GenerateAddress()
				return []interface{}{
					description,
					differentAddr,
					commissionRate,
					minSelfDelegation,
				}
			},
			200000,
			nil,
			func([]byte) {},
			true,
			"is not the same as validator operator address",
		},
		{
			"fail - invalid description",
			func() []interface{} {
				return []interface{}{
					"",
					validatorAddress,
					commissionRate,
					minSelfDelegation,
				}
			},
			200000,
			nil,
			func([]byte) {},
			true,
			"invalid description",
		},
		{
			"fail - invalid commission rate",
			func() []interface{} {
				return []interface{}{
					description,
					validatorAddress,
					"",
					minSelfDelegation,
				}
			},
			200000,
			nil,
			func([]byte) {},
			true,
			"invalid type for commissionRate",
		},
		{
			"fail - invalid min self delegation",
			func() []interface{} {
				return []interface{}{
					description,
					validatorAddress,
					commissionRate,
					"",
				}
			},
			200000,
			nil,
			func([]byte) {},
			true,
			"invalid type for minSelfDelegation",
		},
		{
			"fail - invalid validator address",
			func() []interface{} {
				return []interface{}{
					description,
					1205,
					commissionRate,
					minSelfDelegation,
				}
			},
			200000,
			nil,
			func([]byte) {},
			true,
			"invalid validator address",
		},
		{
			"fail - commission change rate too high",
			func() []interface{} {
				return []interface{}{
					description,
					validatorAddress,
					math.LegacyNewDecWithPrec(11, 2).BigInt(),
					minSelfDelegation,
				}
			},
			200000,
			nil,
			func([]byte) {},
			true,
			"commission cannot be changed more than max change rate",
		},
		{
			"fail - negative commission rate",
			func() []interface{} {
				return []interface{}{
					description,
					validatorAddress,
					math.LegacyNewDecWithPrec(-5, 2).BigInt(),
					minSelfDelegation,
				}
			},
			200000,
			nil,
			func([]byte) {},
			true,
			"commission rate must be between 0 and 1 (inclusive)",
		},
		{
			"fail - negative min self delegation",
			func() []interface{} {
				return []interface{}{
					description,
					validatorAddress,
					commissionRate,
					math.LegacyNewDecWithPrec(-5, 2).BigInt(),
				}
			},
			200000,
			nil,
			func([]byte) {},
			true,
			"minimum self delegation must be a positive integer",
		},
		{
			"fail - calling precompile from a different address than validator (smart contract call)",
			func() []interface{} {
				return []interface{}{
					description,
					validatorAddress,
					commissionRate,
					minSelfDelegation,
				}
			},
			200000,
			func() *common.Address {
				addr := s.keyring.GetAddr(0)
				return &addr
			}(),
			func([]byte) {},
			true,
			"this method can only be called directly to the precompile",
		},
		{
			"success",
			func() []interface{} {
				return []interface{}{
					description,
					validatorAddress,
					commissionRate,
					minSelfDelegation,
				}
			},
			200000,
			nil,
			func(data []byte) {
				success, err := s.precompile.Unpack(staking.EditValidatorMethod, data)
				s.Require().NoError(err)
				s.Require().Equal(success[0], true)

				log := stDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeEditValidator]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight())) //nolint:gosec // G115

				// Check the fully unpacked event matches the one emitted
				var editValidatorEvent staking.EventEditValidator
				err = cmn.UnpackLog(s.precompile.ABI, &editValidatorEvent, staking.EventTypeEditValidator, *log)
				s.Require().NoError(err)
				s.Require().Equal(validatorAddress, editValidatorEvent.ValidatorAddress)
				s.Require().Equal(commissionRate, editValidatorEvent.CommissionRate)
				s.Require().Equal(minSelfDelegation, editValidatorEvent.MinSelfDelegation)
			},
			false,
			"",
		},
		{
			"success - should not update commission rate",
			func() []interface{} {
				// expected commission rate is the previous one (5%)
				commissionRate = math.LegacyNewDecWithPrec(5, 2).BigInt()
				return []interface{}{
					description,
					validatorAddress,
					big.NewInt(-1),
					minSelfDelegation,
				}
			},
			200000,
			nil,
			func(data []byte) { //nolint:dupl
				success, err := s.precompile.Unpack(staking.EditValidatorMethod, data)
				s.Require().NoError(err)
				s.Require().Equal(success[0], true)

				log := stDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeEditValidator]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight())) //nolint:gosec // G115

				// Check the fully unpacked event matches the one emitted
				var editValidatorEvent staking.EventEditValidator
				err = cmn.UnpackLog(s.precompile.ABI, &editValidatorEvent, staking.EventTypeEditValidator, *log)
				s.Require().NoError(err)
				s.Require().Equal(validatorAddress, editValidatorEvent.ValidatorAddress)
			},
			false,
			"",
		},
		{
			"success - should not update minimum self delegation",
			func() []interface{} {
				// expected min self delegation is the previous one (0)
				minSelfDelegation = math.LegacyZeroDec().BigInt()
				return []interface{}{
					description,
					validatorAddress,
					commissionRate,
					big.NewInt(-1),
				}
			},
			200000,
			nil,
			func(data []byte) { //nolint:dupl
				success, err := s.precompile.Unpack(staking.EditValidatorMethod, data)
				s.Require().NoError(err)
				s.Require().Equal(success[0], true)

				log := stDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeEditValidator]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight())) //nolint:gosec // G115

				// Check the fully unpacked event matches the one emitted
				var editValidatorEvent staking.EventEditValidator
				err = cmn.UnpackLog(s.precompile.ABI, &editValidatorEvent, staking.EventTypeEditValidator, *log)
				s.Require().NoError(err)
				s.Require().Equal(validatorAddress, editValidatorEvent.ValidatorAddress)
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx = s.network.GetContext()
			stDB = s.network.GetStateDB()

			commissionRate = math.LegacyNewDecWithPrec(1, 1).BigInt()
			minSelfDelegation = big.NewInt(11)

			// reset sender
			valAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].OperatorAddress)
			s.Require().NoError(err)

			validatorAddress = common.BytesToAddress(valAddr.Bytes())

			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(s.T(), ctx, validatorAddress, s.precompile, tc.gas)
			if tc.callerAddress != nil {
				contract.CallerAddress = *tc.callerAddress
			}

			bz, err := s.precompile.EditValidator(ctx, validatorAddress, contract, stDB, &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)

				// query the validator in the staking keeper
				validator, err := s.network.App.StakingKeeper.Validator(ctx, valAddr.Bytes())
				s.Require().NoError(err)

				s.Require().NotNil(validator, "expected validator not to be nil")
				tc.postCheck(bz)

				isBonded := validator.IsBonded()
				s.Require().Equal(true, isBonded, "expected validator bonded to be %t; got %t", true, isBonded)

				operator := validator.GetOperator()
				s.Require().Equal(sdk.ValAddress(validatorAddress.Bytes()).String(), operator, "expected validator operator to be %s; got %s", validatorAddress, operator)

				updatedCommRate := validator.GetCommission()
				s.Require().Equal(commissionRate.String(), updatedCommRate.BigInt().String(), "expected validator commission rate to be %s; got %s", commissionRate.String(), commissionRate.String())

				valMinSelfDelegation := validator.GetMinSelfDelegation()
				s.Require().Equal(minSelfDelegation.String(), valMinSelfDelegation.String(), "expected validator min self delegation to be %s; got %s", minSelfDelegation.String(), valMinSelfDelegation.String())

				moniker := validator.GetMoniker()
				s.Require().Equal(description.Moniker, moniker, "expected validator moniker to be %s; got %s", description.Moniker, moniker)

				jailed := validator.IsJailed()
				s.Require().Equal(false, jailed, "expected validator jailed to be %t; got %t", false, jailed)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestDelegate() {
	var (
		ctx  sdk.Context
		stDB *statedb.StateDB
	)
	method := s.precompile.Methods[staking.DelegateMethod]

	testCases := []struct {
		name                string
		malleate            func(delegator, grantee testkeyring.Key, operatorAddress string) []interface{}
		gas                 uint64
		expDelegationShares *big.Int
		postCheck           func(data []byte)
		expError            bool
		errContains         string
	}{
		{
			"fail - empty input args",
			func(_, _ testkeyring.Key, _ string) []interface{} {
				return []interface{}{}
			},
			200000,
			big.NewInt(0),
			func([]byte) {},
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 3, 0),
		},
		// TODO: check case if authorization does not exist
		{
			name: "fail - different origin than delegator",
			malleate: func(_, _ testkeyring.Key, operatorAddress string) []interface{} {
				differentAddr := evmosutiltx.GenerateAddress()
				return []interface{}{
					differentAddr,
					operatorAddress,
					big.NewInt(1e18),
				}
			},
			gas:         200000,
			expError:    true,
			errContains: "is not the same as delegator address",
		},
		{
			"fail - invalid delegator address",
			func(_, _ testkeyring.Key, operatorAddress string) []interface{} {
				return []interface{}{
					"",
					operatorAddress,
					big.NewInt(1),
				}
			},
			200000,
			big.NewInt(1),
			func([]byte) {},
			true,
			fmt.Sprintf(cmn.ErrInvalidDelegator, ""),
		},
		{
			"fail - invalid amount",
			func(delegator, _ testkeyring.Key, operatorAddress string) []interface{} {
				return []interface{}{
					delegator.Addr,
					operatorAddress,
					nil,
				}
			},
			200000,
			big.NewInt(1),
			func([]byte) {},
			true,
			fmt.Sprintf(cmn.ErrInvalidAmount, nil),
		},
		{
			"fail - delegation failed because of insufficient funds",
			func(delegator, grantee testkeyring.Key, operatorAddress string) []interface{} {
				//  TODO: why is this necessary?
				err := s.CreateAuthorization(ctx, delegator.AccAddr, grantee.AccAddr, staking.DelegateAuthz, nil)
				s.Require().NoError(err)
				amt, ok := math.NewIntFromString("1000000000000000000000000000")
				s.Require().True(ok)
				return []interface{}{
					delegator.Addr,
					operatorAddress,
					amt.BigInt(),
				}
			},
			200000,
			big.NewInt(15),
			func([]byte) {},
			true,
			"insufficient funds",
		},
		// TODO: adjust tests to work with authorizations (currently does not work because origin == precompile caller which needs no authorization)
		// {
		//	"fail - delegation should not be possible to validators outside of the allow list",
		//	func(string) []interface{} {
		//		err := s.CreateAuthorization(validatorADdress, staking.DelegateAuthz, nil)
		//		s.Require().NoError(err)
		//
		//		// Create new validator --> this is not included in the authorized allow list
		//		testutil.CreateValidator(s.network.GetContext(), s.T(), s.privKey.PubKey(), s.network.App.StakingKeeper, math.NewInt(100))
		//		newValAddr := sdk.ValAddress(s.keyring.GetAccAddr(0))
		//
		//		return []interface{}{
		//			s.keyring.GetAddr(0),
		//			newValAddr.String(),
		//			big.NewInt(1e18),
		//		}
		//	},
		//	200000,
		//	big.NewInt(15),
		//	func( []byte) {},
		//	true,
		//	"cannot delegate/undelegate",
		// },
		{
			"success",
			func(delegator, grantee testkeyring.Key, operatorAddress string) []interface{} {
				// TODO: necessary?
				err := s.CreateAuthorization(ctx, delegator.AccAddr, grantee.AccAddr, staking.DelegateAuthz, nil)
				s.Require().NoError(err)
				return []interface{}{
					delegator.Addr,
					operatorAddress,
					big.NewInt(1e18),
				}
			},
			20000,
			big.NewInt(2),
			func(data []byte) {
				success, err := s.precompile.Unpack(staking.DelegateMethod, data)
				s.Require().NoError(err)
				s.Require().Equal(success[0], true)

				log := stDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeDelegate]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.network.GetContext().BlockHeight())) //nolint:gosec // G115
			},
			false,
			"",
		},
		// TODO: adjust tests to work with authorizations (currently does not work because origin == precompile caller which needs no authorization)
		// {
		//	"success - delegate and update the authorization for the delegator",
		//	func(operatorAddress string) []interface{} {
		//		err := s.CreateAuthorization(s.keyring.GetAddr(0), staking.DelegateAuthz, &sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewInt(2e18)})
		//		s.Require().NoError(err)
		//		return []interface{}{
		//			s.keyring.GetAddr(0),
		//			operatorAddress,
		//			big.NewInt(1e18),
		//		}
		//	},
		//	20000,
		//	big.NewInt(2),
		//	func(data []byte) {
		//		authorization, _ := s.network.App.AuthzKeeper.GetAuthorization(s.network.GetContext(), s.keyring.GetAccAddr(0), s.keyring.GetAccAddr(0), staking.DelegateMsg)
		//		s.Require().NotNil(authorization)
		//		stakeAuthorization := authorization.(*stakingtypes.StakeAuthorization)
		//		s.Require().Equal(math.NewInt(1e18), stakeAuthorization.MaxTokens.Amount)
		//	},
		//	false,
		//	"",
		// },
		// {
		//	"success - delegate and delete the authorization for the delegator",
		//	func(operatorAddress string) []interface{} {
		//		err := s.CreateAuthorization(s.keyring.GetAddr(0), staking.DelegateAuthz, &sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewInt(1e18)})
		//		s.Require().NoError(err)
		//		return []interface{}{
		//			s.keyring.GetAddr(0),
		//			operatorAddress,
		//			big.NewInt(1e18),
		//		}
		//	},
		//	20000,
		//	big.NewInt(2),
		//	func(data []byte) {
		//		authorization, _ := s.network.App.AuthzKeeper.GetAuthorization(s.network.GetContext(), s.keyring.GetAccAddr(0), s.keyring.GetAccAddr(0), staking.DelegateMsg)
		//		s.Require().Nil(authorization)
		//	},
		//	false,
		//	"",
		// },
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx = s.network.GetContext()
			stDB = s.network.GetStateDB()

			delegator := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			contract, ctx := testutil.NewPrecompileContract(s.T(), ctx, delegator.Addr, s.precompile, tc.gas)

			delegateArgs := tc.malleate(
				delegator,
				grantee,
				s.network.GetValidators()[0].OperatorAddress,
			)
			bz, err := s.precompile.Delegate(ctx, delegator.Addr, contract, stDB, &method, delegateArgs)

			// query the delegation in the staking keeper
			valAddr, valErr := sdk.ValAddressFromBech32(s.network.GetValidators()[0].OperatorAddress)
			s.Require().NoError(valErr)
			delegation, delErr := s.network.App.StakingKeeper.Delegation(ctx, delegator.AccAddr, valAddr)
			s.Require().NoError(delErr)
			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
				s.Require().Equal(s.network.GetValidators()[0].DelegatorShares, delegation.GetShares())
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(delegation, "expected delegation not to be nil")
				tc.postCheck(bz)

				expDelegationAmt := math.NewIntFromBigInt(tc.expDelegationShares)
				delegationAmt := delegation.GetShares().TruncateInt()

				s.Require().Equal(expDelegationAmt, delegationAmt, "expected delegation amount to be %d; got %d", expDelegationAmt, delegationAmt)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestUndelegate() {
	var (
		ctx  sdk.Context
		stDB *statedb.StateDB
	)
	method := s.precompile.Methods[staking.UndelegateMethod]

	testCases := []struct {
		name                  string
		malleate              func(delegator, grantee testkeyring.Key, operatorAddress string) []interface{}
		postCheck             func(data []byte)
		gas                   uint64
		expUndelegationShares *big.Int
		expError              bool
		errContains           string
	}{
		{
			"fail - empty input args",
			func(testkeyring.Key, testkeyring.Key, string) []interface{} {
				return []interface{}{}
			},
			func([]byte) {},
			200000,
			big.NewInt(0),
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 3, 0),
		},
		// TODO: check case if authorization does not exist
		{
			name: "fail - different origin than delegator",
			malleate: func(_, _ testkeyring.Key, operatorAddress string) []interface{} {
				differentAddr := evmosutiltx.GenerateAddress()
				return []interface{}{
					differentAddr,
					operatorAddress,
					big.NewInt(1000000000000000000),
				}
			},
			gas:         200000,
			expError:    true,
			errContains: "is not the same as delegator",
		},
		{
			"fail - invalid delegator address",
			func(_, _ testkeyring.Key, operatorAddress string) []interface{} {
				return []interface{}{
					"",
					operatorAddress,
					big.NewInt(1),
				}
			},
			func([]byte) {},
			200000,
			big.NewInt(1),
			true,
			fmt.Sprintf(cmn.ErrInvalidDelegator, ""),
		},
		{
			"fail - invalid amount",
			func(delegator, _ testkeyring.Key, operatorAddress string) []interface{} {
				return []interface{}{
					delegator.Addr,
					operatorAddress,
					nil,
				}
			},
			func([]byte) {},
			200000,
			big.NewInt(1),
			true,
			fmt.Sprintf(cmn.ErrInvalidAmount, nil),
		},
		{
			"success",
			func(delegator, grantee testkeyring.Key, operatorAddress string) []interface{} {
				err := s.CreateAuthorization(ctx, delegator.AccAddr, grantee.AccAddr, staking.UndelegateAuthz, nil)
				s.Require().NoError(err)
				return []interface{}{
					delegator.Addr,
					operatorAddress,
					big.NewInt(1000000000000000000),
				}
			},
			func(data []byte) {
				args, err := s.precompile.Unpack(staking.UndelegateMethod, data)
				s.Require().NoError(err, "failed to unpack output")
				s.Require().Len(args, 1)
				completionTime, ok := args[0].(int64)
				s.Require().True(ok, "completion time type %T", args[0])
				params, err := s.network.App.StakingKeeper.GetParams(ctx)
				s.Require().NoError(err)
				expCompletionTime := ctx.BlockTime().Add(params.UnbondingTime).UTC().Unix()
				s.Require().Equal(expCompletionTime, completionTime)
				// Check the event emitted
				log := stDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())
			},
			20000,
			big.NewInt(1000000000000000000),
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx = s.network.GetContext()
			stDB = s.network.GetStateDB()

			delegator := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(s.T(), ctx, delegator.Addr, s.precompile, tc.gas)

			undelegateArgs := tc.malleate(delegator, grantee, s.network.GetValidators()[0].OperatorAddress)
			bz, err := s.precompile.Undelegate(ctx, delegator.Addr, contract, stDB, &method, undelegateArgs)

			// query the unbonding delegations in the staking keeper
			undelegations, _ := s.network.App.StakingKeeper.GetAllUnbondingDelegations(ctx, delegator.AccAddr)

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)
				tc.postCheck(bz)

				s.Require().Equal(undelegations[0].DelegatorAddress, delegator.AccAddr.String())
				s.Require().Equal(undelegations[0].ValidatorAddress, s.network.GetValidators()[0].OperatorAddress)
				s.Require().Equal(undelegations[0].Entries[0].Balance, math.NewIntFromBigInt(tc.expUndelegationShares))
			}
		})
	}
}

func (s *PrecompileTestSuite) TestRedelegate() {
	var ctx sdk.Context
	method := s.precompile.Methods[staking.RedelegateMethod]

	testCases := []struct {
		name                  string
		malleate              func(delegator, grantee testkeyring.Key, srcOperatorAddr, dstOperatorAddr string) []interface{}
		postCheck             func(data []byte)
		gas                   uint64
		expRedelegationShares *big.Int
		expError              bool
		errContains           string
	}{
		{
			"fail - empty input args",
			func(_, _ testkeyring.Key, _, _ string) []interface{} {
				return []interface{}{}
			},
			func([]byte) {},
			200000,
			big.NewInt(0),
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 4, 0),
		},
		// TODO: check case if authorization does not exist
		{
			name: "fail - different origin than delegator",
			malleate: func(_, _ testkeyring.Key, srcOperatorAddr, dstOperatorAddr string) []interface{} {
				differentAddr := evmosutiltx.GenerateAddress()
				return []interface{}{
					differentAddr,
					srcOperatorAddr,
					dstOperatorAddr,
					big.NewInt(1000000000000000000),
				}
			},
			gas:         200000,
			expError:    true,
			errContains: "is not the same as delegator",
		},
		{
			"fail - invalid delegator address",
			func(_, _ testkeyring.Key, srcOperatorAddr, dstOperatorAddr string) []interface{} {
				return []interface{}{
					"",
					srcOperatorAddr,
					dstOperatorAddr,
					big.NewInt(1),
				}
			},
			func([]byte) {},
			200000,
			big.NewInt(1),
			true,
			fmt.Sprintf(cmn.ErrInvalidDelegator, ""),
		},
		{
			"fail - invalid amount",
			func(delegator, _ testkeyring.Key, srcOperatorAddr, dstOperatorAddr string) []interface{} {
				return []interface{}{
					delegator.Addr,
					srcOperatorAddr,
					dstOperatorAddr,
					nil,
				}
			},
			func([]byte) {},
			200000,
			big.NewInt(1),
			true,
			fmt.Sprintf(cmn.ErrInvalidAmount, nil),
		},
		{
			"fail - invalid shares amount",
			func(delegator, _ testkeyring.Key, srcOperatorAddr, dstOperatorAddr string) []interface{} {
				return []interface{}{
					delegator.Addr,
					srcOperatorAddr,
					dstOperatorAddr,
					big.NewInt(-1),
				}
			},
			func([]byte) {},
			200000,
			big.NewInt(1),
			true,
			"invalid shares amount",
		},
		{
			"success",
			func(delegator, grantee testkeyring.Key, srcOperatorAddr, dstOperatorAddr string) []interface{} {
				// TODO: necessary?
				err := s.CreateAuthorization(ctx, delegator.AccAddr, grantee.AccAddr, staking.RedelegateAuthz, nil)
				s.Require().NoError(err)
				return []interface{}{
					delegator.Addr,
					srcOperatorAddr,
					dstOperatorAddr,
					big.NewInt(1000000000000000000),
				}
			},
			func(data []byte) {
				args, err := s.precompile.Unpack(staking.RedelegateMethod, data)
				s.Require().NoError(err, "failed to unpack output")
				s.Require().Len(args, 1)
				completionTime, ok := args[0].(int64)
				s.Require().True(ok, "completion time type %T", args[0])
				params, err := s.network.App.StakingKeeper.GetParams(ctx)
				s.Require().NoError(err)
				expCompletionTime := ctx.BlockTime().Add(params.UnbondingTime).UTC().Unix()
				s.Require().Equal(expCompletionTime, completionTime)
			},
			200000,
			big.NewInt(1),
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx = s.network.GetContext()
			delegator := s.keyring.GetKey(0)
			// TODO: even necessary??
			grantee := s.keyring.GetKey(1)

			contract, ctx := testutil.NewPrecompileContract(s.T(), ctx, delegator.Addr, s.precompile, tc.gas)

			redelegateArgs := tc.malleate(
				delegator,
				grantee,
				s.network.GetValidators()[0].OperatorAddress,
				s.network.GetValidators()[1].OperatorAddress,
			)
			bz, err := s.precompile.Redelegate(ctx, delegator.Addr, contract, s.network.GetStateDB(), &method, redelegateArgs)

			// query the redelegations in the staking keeper
			redelegations, redelErr := s.network.App.StakingKeeper.GetRedelegations(ctx, delegator.AccAddr, 5)
			s.Require().NoError(redelErr)

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)

				s.Require().Equal(redelegations[0].DelegatorAddress, delegator.AccAddr.String())
				s.Require().Equal(redelegations[0].ValidatorSrcAddress, s.network.GetValidators()[0].OperatorAddress)
				s.Require().Equal(redelegations[0].ValidatorDstAddress, s.network.GetValidators()[1].OperatorAddress)
				s.Require().Equal(redelegations[0].Entries[0].SharesDst, math.LegacyNewDecFromBigInt(tc.expRedelegationShares))
			}
		})
	}
}

func (s *PrecompileTestSuite) TestCancelUnbondingDelegation() {
	var ctx sdk.Context
	method := s.precompile.Methods[staking.CancelUnbondingDelegationMethod]
	undelegateMethod := s.precompile.Methods[staking.UndelegateMethod]

	testCases := []struct {
		name               string
		malleate           func(delegator, grantee testkeyring.Key, operatorAddress string) []interface{}
		postCheck          func(data []byte)
		gas                uint64
		expDelegatedShares *big.Int
		expError           bool
		errContains        string
	}{
		{
			"fail - empty input args",
			func(_, _ testkeyring.Key, _ string) []interface{} {
				return []interface{}{}
			},
			func([]byte) {},
			200000,
			big.NewInt(0),
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 4, 0),
		},
		{
			"fail - invalid delegator address",
			func(_, _ testkeyring.Key, operatorAddress string) []interface{} {
				return []interface{}{
					"",
					operatorAddress,
					big.NewInt(1),
					big.NewInt(1),
				}
			},
			func([]byte) {},
			200000,
			big.NewInt(1),
			true,
			fmt.Sprintf(cmn.ErrInvalidDelegator, ""),
		},
		{
			"fail - creation height",
			func(delegator, _ testkeyring.Key, operatorAddress string) []interface{} {
				return []interface{}{
					delegator.Addr,
					operatorAddress,
					big.NewInt(1),
					nil,
				}
			},
			func([]byte) {},
			200000,
			big.NewInt(1),
			true,
			"invalid creation height",
		},
		{
			"fail - invalid amount",
			func(delegator, _ testkeyring.Key, operatorAddress string) []interface{} {
				return []interface{}{
					delegator.Addr,
					operatorAddress,
					nil,
					big.NewInt(1),
				}
			},
			func([]byte) {},
			200000,
			big.NewInt(1),
			true,
			fmt.Sprintf(cmn.ErrInvalidAmount, nil),
		},
		{
			"fail - invalid amount",
			func(delegator, _ testkeyring.Key, operatorAddress string) []interface{} {
				return []interface{}{
					delegator.Addr,
					operatorAddress,
					nil,
					big.NewInt(1),
				}
			},
			func([]byte) {},
			200000,
			big.NewInt(1),
			true,
			fmt.Sprintf(cmn.ErrInvalidAmount, nil),
		},
		{
			"fail - invalid shares amount",
			func(delegator, _ testkeyring.Key, operatorAddress string) []interface{} {
				return []interface{}{
					delegator.Addr,
					operatorAddress,
					big.NewInt(-1),
					big.NewInt(1),
				}
			},
			func([]byte) {},
			200000,
			big.NewInt(1),
			true,
			"invalid amount: invalid request",
		},
		{
			"success",
			func(delegator, grantee testkeyring.Key, operatorAddress string) []interface{} {
				// TODO: why is this necessary
				// TODO: remove also grantee from malleate
				err := s.CreateAuthorization(ctx, delegator.AccAddr, grantee.AccAddr, staking.DelegateAuthz, nil)
				s.Require().NoError(err)
				return []interface{}{
					delegator.Addr,
					operatorAddress,
					big.NewInt(1),
					big.NewInt(1),
				}
			},
			func(data []byte) {
				success, err := s.precompile.Unpack(staking.CancelUnbondingDelegationMethod, data)
				s.Require().NoError(err)
				s.Require().Equal(success[0], true)
			},
			200000,
			big.NewInt(1),
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx = s.network.GetContext()
			stDB := s.network.GetStateDB()

			delegator := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			contract, ctx := testutil.NewPrecompileContract(s.T(), ctx, delegator.Addr, s.precompile, tc.gas)
			cancelArgs := tc.malleate(delegator, grantee, s.network.GetValidators()[0].OperatorAddress)

			if tc.expError {
				bz, err := s.precompile.CancelUnbondingDelegation(ctx, delegator.Addr, contract, stDB, &method, cancelArgs)
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				undelegateArgs := []interface{}{
					delegator.Addr,
					s.network.GetValidators()[0].OperatorAddress,
					big.NewInt(1000000000000000000),
				}

				// TODO: is this even necessary?
				err := s.CreateAuthorization(ctx, delegator.AccAddr, grantee.AccAddr, staking.UndelegateAuthz, nil)
				s.Require().NoError(err)

				_, err = s.precompile.Undelegate(ctx, delegator.Addr, contract, stDB, &undelegateMethod, undelegateArgs)
				s.Require().NoError(err)

				valAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].GetOperator())
				s.Require().NoError(err)

				_, err = s.network.App.StakingKeeper.GetDelegation(ctx, delegator.AccAddr, valAddr)
				s.Require().Error(err)
				s.Require().Contains("no delegation for (address, validator) tuple", err.Error())

				err = s.CreateAuthorization(ctx, delegator.AccAddr, grantee.AccAddr, staking.CancelUnbondingDelegationAuthz, nil)
				s.Require().NoError(err)

				bz, err := s.precompile.CancelUnbondingDelegation(ctx, delegator.Addr, contract, stDB, &method, cancelArgs)
				s.Require().NoError(err)
				tc.postCheck(bz)

				delegation, err := s.network.App.StakingKeeper.GetDelegation(ctx, delegator.AccAddr, valAddr)
				s.Require().NoError(err)

				s.Require().Equal(delegation.DelegatorAddress, delegator.AccAddr.String())
				s.Require().Equal(delegation.ValidatorAddress, s.network.GetValidators()[0].OperatorAddress)
				s.Require().Equal(delegation.Shares, math.LegacyNewDecFromBigInt(tc.expDelegatedShares))

			}
		})
	}
}
