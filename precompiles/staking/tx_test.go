package staking_test

import (
	"encoding/base64"
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	geth "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	"github.com/evmos/evmos/v18/precompiles/staking"
	"github.com/evmos/evmos/v18/precompiles/testutil"
	evmosutiltx "github.com/evmos/evmos/v18/testutil/tx"
)

func (s *PrecompileTestSuite) TestCreateValidator() {
	var (
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
		validatorAddress  = s.address
		pubkey            = "nfJ0axJC9dhta1MAE1EBFaVdxxkYzxYrBaHuJVjG//M="
		value             = big.NewInt(1205000000000000000)
	)

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		gas         uint64
		postCheck   func(data []byte)
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			200000,
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
			func([]byte) {},
			true,
			"invalid amount",
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
			func(data []byte) {
				success, err := s.precompile.Unpack(staking.CreateValidatorMethod, data)
				s.Require().NoError(err)
				s.Require().Equal(success[0], true)

				log := s.stateDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeCreateValidator]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), geth.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.ctx.BlockHeight()))

				// Check the fully unpacked event matches the one emitted
				var createValidatorEvent staking.EventCreateValidator
				err = cmn.UnpackLog(s.precompile.ABI, &createValidatorEvent, staking.EventTypeCreateValidator, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.address, createValidatorEvent.ValidatorAddress)
				s.Require().Equal(value, createValidatorEvent.Value)
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// reset sender
			validatorAddress = s.address

			var contract *vm.Contract
			contract, s.ctx = testutil.NewPrecompileContract(s.T(), s.ctx, s.address, s.precompile, tc.gas)

			bz, err := s.precompile.CreateValidator(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate())

			// query the validator in the staking keeper
			validator := s.app.StakingKeeper.Validator(s.ctx, s.address.Bytes())
			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
				s.Require().Nil(validator)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(validator, "expected validator not to be nil")
				tc.postCheck(bz)

				isBonded := validator.IsBonded()
				s.Require().Equal(false, isBonded, "expected validator bonded to be %t; got %t", false, isBonded)

				consPubKey, err := validator.ConsPubKey()
				s.Require().NoError(err)
				consPubKeyBase64 := base64.StdEncoding.EncodeToString(consPubKey.Bytes())
				s.Require().Equal(pubkey, consPubKeyBase64, "expected validator pubkey to be %s; got %s", pubkey, consPubKeyBase64)

				operator := validator.GetOperator().String()
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

func (s *PrecompileTestSuite) TestDelegate() {
	method := s.precompile.Methods[staking.DelegateMethod]

	testCases := []struct {
		name                string
		malleate            func(operatorAddress string) []interface{}
		gas                 uint64
		expDelegationShares *big.Int
		postCheck           func(data []byte)
		expError            bool
		errContains         string
	}{
		{
			"fail - empty input args",
			func(string) []interface{} {
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
			malleate: func(operatorAddress string) []interface{} {
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
			func(operatorAddress string) []interface{} {
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
			func(operatorAddress string) []interface{} {
				return []interface{}{
					s.address,
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
			func(operatorAddress string) []interface{} {
				err := s.CreateAuthorization(s.address, staking.DelegateAuthz, nil)
				s.Require().NoError(err)
				return []interface{}{
					s.address,
					operatorAddress,
					big.NewInt(9e18),
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
		//		err := s.CreateAuthorization(s.address, staking.DelegateAuthz, nil)
		//		s.Require().NoError(err)
		//
		//		// Create new validator --> this is not included in the authorized allow list
		//		testutil.CreateValidator(s.ctx, s.T(), s.privKey.PubKey(), s.app.StakingKeeper, math.NewInt(100))
		//		newValAddr := sdk.ValAddress(s.address.Bytes())
		//
		//		return []interface{}{
		//			s.address,
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
			func(operatorAddress string) []interface{} {
				err := s.CreateAuthorization(s.address, staking.DelegateAuthz, nil)
				s.Require().NoError(err)
				return []interface{}{
					s.address,
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

				log := s.stateDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[staking.EventTypeDelegate]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), geth.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.ctx.BlockHeight()))
			},
			false,
			"",
		},
		// TODO: adjust tests to work with authorizations (currently does not work because origin == precompile caller which needs no authorization)
		// {
		//	"success - delegate and update the authorization for the delegator",
		//	func(operatorAddress string) []interface{} {
		//		err := s.CreateAuthorization(s.address, staking.DelegateAuthz, &sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewInt(2e18)})
		//		s.Require().NoError(err)
		//		return []interface{}{
		//			s.address,
		//			operatorAddress,
		//			big.NewInt(1e18),
		//		}
		//	},
		//	20000,
		//	big.NewInt(2),
		//	func(data []byte) {
		//		authorization, _ := s.app.AuthzKeeper.GetAuthorization(s.ctx, s.address.Bytes(), s.address.Bytes(), staking.DelegateMsg)
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
		//		err := s.CreateAuthorization(s.address, staking.DelegateAuthz, &sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewInt(1e18)})
		//		s.Require().NoError(err)
		//		return []interface{}{
		//			s.address,
		//			operatorAddress,
		//			big.NewInt(1e18),
		//		}
		//	},
		//	20000,
		//	big.NewInt(2),
		//	func(data []byte) {
		//		authorization, _ := s.app.AuthzKeeper.GetAuthorization(s.ctx, s.address.Bytes(), s.address.Bytes(), staking.DelegateMsg)
		//		s.Require().Nil(authorization)
		//	},
		//	false,
		//	"",
		// },
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			var contract *vm.Contract
			contract, s.ctx = testutil.NewPrecompileContract(s.T(), s.ctx, s.address, s.precompile, tc.gas)

			bz, err := s.precompile.Delegate(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate(s.validators[0].OperatorAddress))

			// query the delegation in the staking keeper
			delegation := s.app.StakingKeeper.Delegation(s.ctx, s.address.Bytes(), s.validators[0].GetOperator())
			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
				s.Require().Equal(s.validators[0].DelegatorShares, delegation.GetShares())
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
	method := s.precompile.Methods[staking.UndelegateMethod]

	testCases := []struct {
		name                  string
		malleate              func(operatorAddress string) []interface{}
		postCheck             func(data []byte)
		gas                   uint64
		expUndelegationShares *big.Int
		expError              bool
		errContains           string
	}{
		{
			"fail - empty input args",
			func(string) []interface{} {
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
			malleate: func(operatorAddress string) []interface{} {
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
			func(operatorAddress string) []interface{} {
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
			func(operatorAddress string) []interface{} {
				return []interface{}{
					s.address,
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
			func(operatorAddress string) []interface{} {
				err := s.CreateAuthorization(s.address, staking.UndelegateAuthz, nil)
				s.Require().NoError(err)
				return []interface{}{
					s.address,
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
				params := s.app.StakingKeeper.GetParams(s.ctx)
				expCompletionTime := s.ctx.BlockTime().Add(params.UnbondingTime).UTC().Unix()
				s.Require().Equal(expCompletionTime, completionTime)
				// Check the event emitted
				log := s.stateDB.Logs()[0]
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

			var contract *vm.Contract
			contract, s.ctx = testutil.NewPrecompileContract(s.T(), s.ctx, s.address, s.precompile, tc.gas)

			bz, err := s.precompile.Undelegate(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate(s.validators[0].OperatorAddress))

			// query the unbonding delegations in the staking keeper
			undelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, s.address.Bytes())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)
				tc.postCheck(bz)

				bech32Addr, err := sdk.Bech32ifyAddressBytes("evmos", s.address.Bytes())
				s.Require().NoError(err)
				s.Require().Equal(undelegations[0].DelegatorAddress, bech32Addr)
				s.Require().Equal(undelegations[0].ValidatorAddress, s.validators[0].OperatorAddress)
				s.Require().Equal(undelegations[0].Entries[0].Balance, math.NewIntFromBigInt(tc.expUndelegationShares))
			}
		})
	}
}

func (s *PrecompileTestSuite) TestRedelegate() {
	method := s.precompile.Methods[staking.RedelegateMethod]

	testCases := []struct {
		name                  string
		malleate              func(srcOperatorAddr, dstOperatorAddr string) []interface{}
		postCheck             func(data []byte)
		gas                   uint64
		expRedelegationShares *big.Int
		expError              bool
		errContains           string
	}{
		{
			"fail - empty input args",
			func(string, string) []interface{} {
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
			malleate: func(srcOperatorAddr, dstOperatorAddr string) []interface{} {
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
			func(srcOperatorAddr, dstOperatorAddr string) []interface{} {
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
			func(srcOperatorAddr, dstOperatorAddr string) []interface{} {
				return []interface{}{
					s.address,
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
			func(srcOperatorAddr, dstOperatorAddr string) []interface{} {
				return []interface{}{
					s.address,
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
			func(srcOperatorAddr, dstOperatorAddr string) []interface{} {
				err := s.CreateAuthorization(s.address, staking.RedelegateAuthz, nil)
				s.Require().NoError(err)
				return []interface{}{
					s.address,
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
				params := s.app.StakingKeeper.GetParams(s.ctx)
				expCompletionTime := s.ctx.BlockTime().Add(params.UnbondingTime).UTC().Unix()
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

			var contract *vm.Contract
			contract, s.ctx = testutil.NewPrecompileContract(s.T(), s.ctx, s.address, s.precompile, tc.gas)

			bz, err := s.precompile.Redelegate(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate(s.validators[0].OperatorAddress, s.validators[1].OperatorAddress))

			// query the redelegations in the staking keeper
			redelegations := s.app.StakingKeeper.GetRedelegations(s.ctx, s.address.Bytes(), 5)

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)

				bech32Addr, err := sdk.Bech32ifyAddressBytes("evmos", s.address.Bytes())
				s.Require().NoError(err)
				s.Require().Equal(redelegations[0].DelegatorAddress, bech32Addr)
				s.Require().Equal(redelegations[0].ValidatorSrcAddress, s.validators[0].OperatorAddress)
				s.Require().Equal(redelegations[0].ValidatorDstAddress, s.validators[1].OperatorAddress)
				s.Require().Equal(redelegations[0].Entries[0].SharesDst, math.LegacyNewDecFromBigInt(tc.expRedelegationShares))
			}
		})
	}
}

func (s *PrecompileTestSuite) TestCancelUnbondingDelegation() {
	method := s.precompile.Methods[staking.CancelUnbondingDelegationMethod]
	undelegateMethod := s.precompile.Methods[staking.UndelegateMethod]

	testCases := []struct {
		name               string
		malleate           func(operatorAddress string) []interface{}
		postCheck          func(data []byte)
		gas                uint64
		expDelegatedShares *big.Int
		expError           bool
		errContains        string
	}{
		{
			"fail - empty input args",
			func(string) []interface{} {
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
			func(operatorAddress string) []interface{} {
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
			func(operatorAddress string) []interface{} {
				return []interface{}{
					s.address,
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
			func(operatorAddress string) []interface{} {
				return []interface{}{
					s.address,
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
			func(operatorAddress string) []interface{} {
				return []interface{}{
					s.address,
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
			func(operatorAddress string) []interface{} {
				return []interface{}{
					s.address,
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
			func(operatorAddress string) []interface{} {
				err := s.CreateAuthorization(s.address, staking.DelegateAuthz, nil)
				s.Require().NoError(err)
				return []interface{}{
					s.address,
					operatorAddress,
					big.NewInt(1),
					big.NewInt(2),
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

			var contract *vm.Contract
			contract, s.ctx = testutil.NewPrecompileContract(s.T(), s.ctx, s.address, s.precompile, tc.gas)

			if tc.expError {
				bz, err := s.precompile.CancelUnbondingDelegation(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate(s.validators[0].OperatorAddress))
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				undelegateArgs := []interface{}{
					s.address,
					s.validators[0].OperatorAddress,
					big.NewInt(1000000000000000000),
				}

				err := s.CreateAuthorization(s.address, staking.UndelegateAuthz, nil)
				s.Require().NoError(err)

				_, err = s.precompile.Undelegate(s.ctx, s.address, contract, s.stateDB, &undelegateMethod, undelegateArgs)
				s.Require().NoError(err)

				_, found := s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), s.validators[0].GetOperator())
				s.Require().False(found)

				err = s.CreateAuthorization(s.address, staking.CancelUnbondingDelegationAuthz, nil)
				s.Require().NoError(err)

				bz, err := s.precompile.CancelUnbondingDelegation(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate(s.validators[0].OperatorAddress))
				s.Require().NoError(err)
				tc.postCheck(bz)

				delegation, found := s.app.StakingKeeper.GetDelegation(s.ctx, s.address.Bytes(), s.validators[0].GetOperator())
				s.Require().True(found)

				bech32Addr, err := sdk.Bech32ifyAddressBytes("evmos", s.address.Bytes())
				s.Require().NoError(err)
				s.Require().Equal(delegation.DelegatorAddress, bech32Addr)
				s.Require().Equal(delegation.ValidatorAddress, s.validators[0].OperatorAddress)
				s.Require().Equal(delegation.Shares, math.LegacyNewDecFromBigInt(tc.expDelegatedShares))

			}
		})
	}
}
