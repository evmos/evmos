package staking_test

//import (
//	"fmt"
//	"math/big"
//
//	"cosmossdk.io/math"
//	sdk "github.com/cosmos/cosmos-sdk/types"
//	"github.com/cosmos/cosmos-sdk/types/query"
//	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
//	"github.com/ethereum/go-ethereum/common"
//	"github.com/ethereum/go-ethereum/core/vm"
//	"github.com/evmos/evmos/v16/precompiles/authorization"
//	cmn "github.com/evmos/evmos/v16/precompiles/common"
//	"github.com/evmos/evmos/v16/precompiles/staking"
//	testutiltx "github.com/evmos/evmos/v16/testutil/tx"
//)
//
//func (s *PrecompileTestSuite) TestDelegation() {
//	method := s.precompile.Methods[staking.DelegationMethod]
//
//	testCases := []struct {
//		name        string
//		malleate    func(operatorAddress string) []interface{}
//		postCheck   func(bz []byte)
//		gas         uint64
//		expErr      bool
//		errContains string
//	}{
//		{
//			"fail - empty input args",
//			func(operatorAddress string) []interface{} {
//				return []interface{}{}
//			},
//			func(bz []byte) {},
//			100000,
//			true,
//			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 2, 0),
//		},
//		{
//			"fail - invalid delegator address",
//			func(operatorAddress string) []interface{} {
//				return []interface{}{
//					"invalid",
//					operatorAddress,
//				}
//			},
//			func(bz []byte) {},
//			100000,
//			true,
//			fmt.Sprintf(cmn.ErrInvalidDelegator, "invalid"),
//		},
//		{
//			"fail - invalid operator address",
//			func(operatorAddress string) []interface{} {
//				return []interface{}{
//					s.keyring.GetAddr(0),
//					"invalid",
//				}
//			},
//			func(bz []byte) {},
//			100000,
//			true,
//			"decoding bech32 failed: invalid bech32 string",
//		},
//		{
//			"success - empty delegation",
//			func(operatorAddress string) []interface{} {
//				addr, _ := testutiltx.NewAddrKey()
//				return []interface{}{
//					addr,
//					operatorAddress,
//				}
//			},
//			func(bz []byte) {
//				var delOut staking.DelegationOutput
//				err := s.precompile.UnpackIntoInterface(&delOut, staking.DelegationMethod, bz)
//				s.Require().NoError(err, "failed to unpack output")
//				s.Require().Equal(delOut.Shares.Int64(), big.NewInt(0).Int64())
//			},
//			100000,
//			false,
//			"",
//		},
//		{
//			"success",
//			func(operatorAddress string) []interface{} {
//				return []interface{}{
//					s.keyring.GetAddr(0),
//					operatorAddress,
//				}
//			},
//			func(bz []byte) {
//				var delOut staking.DelegationOutput
//				err := s.precompile.UnpackIntoInterface(&delOut, staking.DelegationMethod, bz)
//				s.Require().NoError(err, "failed to unpack output")
//				s.Require().Equal(delOut.Shares, big.NewInt(1e18))
//			},
//			100000,
//			false,
//			"",
//		},
//	}
//
//	for _, tc := range testCases {
//		s.Run(tc.name, func() {
//			s.SetupTest() // reset
//			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)
//
//			bz, err := s.precompile.Delegation(s.network.GetContext(), contract, &method, tc.malleate(s.network.GetValidators()[0].OperatorAddress))
//
//			if tc.expErr {
//				s.Require().Error(err)
//				s.Require().Contains(err.Error(), tc.errContains)
//			} else {
//				s.Require().NoError(err)
//				s.Require().NotEmpty(bz)
//				tc.postCheck(bz)
//			}
//		})
//	}
//}
//
//func (s *PrecompileTestSuite) TestUnbondingDelegation() {
//	method := s.precompile.Methods[staking.UnbondingDelegationMethod]
//
//	testCases := []struct {
//		name        string
//		malleate    func(operatorAddress string) []interface{}
//		postCheck   func(bz []byte)
//		gas         uint64
//		expErr      bool
//		errContains string
//	}{
//		{
//			"fail - empty input args",
//			func(operatorAddress string) []interface{} {
//				return []interface{}{}
//			},
//			func(bz []byte) {},
//			100000,
//			true,
//			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 2, 0),
//		},
//		{
//			"fail - invalid delegator address",
//			func(operatorAddress string) []interface{} {
//				return []interface{}{
//					"invalid",
//					operatorAddress,
//				}
//			},
//			func(bz []byte) {},
//			100000,
//			true,
//			fmt.Sprintf(cmn.ErrInvalidDelegator, "invalid"),
//		},
//		{
//			"success - no unbonding delegation found",
//			func(operatorAddress string) []interface{} {
//				addr, _ := testutiltx.NewAddrKey()
//				return []interface{}{
//					addr,
//					operatorAddress,
//				}
//			},
//			func(data []byte) {
//				var ubdOut staking.UnbondingDelegationOutput
//				err := s.precompile.UnpackIntoInterface(&ubdOut, staking.UnbondingDelegationMethod, data)
//				s.Require().NoError(err, "failed to unpack output")
//				s.Require().Len(ubdOut.UnbondingDelegation.Entries, 0)
//			},
//			100000,
//			false,
//			"",
//		},
//		{
//			"success",
//			func(operatorAddress string) []interface{} {
//				return []interface{}{
//					s.keyring.GetAddr(0),
//					operatorAddress,
//				}
//			},
//			func(data []byte) {
//				var ubdOut staking.UnbondingDelegationOutput
//				err := s.precompile.UnpackIntoInterface(&ubdOut, staking.UnbondingDelegationMethod, data)
//				s.Require().NoError(err, "failed to unpack output")
//				s.Require().Len(ubdOut.UnbondingDelegation.Entries, 1)
//				s.Require().Equal(ubdOut.UnbondingDelegation.Entries[0].CreationHeight, s.network.GetContext().BlockHeight())
//				s.Require().Equal(ubdOut.UnbondingDelegation.Entries[0].Balance, big.NewInt(1e18))
//			},
//			100000,
//			false,
//			"",
//		},
//	}
//
//	for _, tc := range testCases {
//		s.Run(tc.name, func() {
//			s.SetupTest() // reset
//			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)
//
//			valAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].GetOperator())
//			s.Require().NoError(err)
//
//			_, _, err = s.network.App.StakingKeeper.Undelegate(s.network.GetContext(), s.keyring.GetAccAddr(0), valAddr, math.LegacyNewDec(1))
//			s.Require().NoError(err)
//
//			bz, err := s.precompile.UnbondingDelegation(s.network.GetContext(), contract, &method, tc.malleate(s.network.GetValidators()[0].OperatorAddress))
//
//			if tc.expErr {
//				s.Require().Error(err)
//				s.Require().Contains(err.Error(), tc.errContains)
//			} else {
//				s.Require().NoError(err)
//				s.Require().NotNil(bz)
//				tc.postCheck(bz)
//			}
//		})
//	}
//}
//
//func (s *PrecompileTestSuite) TestValidator() {
//	method := s.precompile.Methods[staking.ValidatorMethod]
//
//	testCases := []struct {
//		name        string
//		malleate    func(operatorAddress common.Address) []interface{}
//		postCheck   func(bz []byte)
//		gas         uint64
//		expErr      bool
//		errContains string
//	}{
//		{
//			"fail - empty input args",
//			func(operatorAddress common.Address) []interface{} {
//				return []interface{}{}
//			},
//			func(_ []byte) {},
//			100000,
//			true,
//			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 1, 0),
//		},
//		{
//			"success",
//			func(operatorAddress common.Address) []interface{} {
//				return []interface{}{
//					operatorAddress,
//				}
//			},
//			func(data []byte) {
//				var valOut staking.ValidatorOutput
//				err := s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorMethod, data)
//				s.Require().NoError(err, "failed to unpack output")
//
//				operatorAddress, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].OperatorAddress)
//				s.Require().NoError(err)
//
//				s.Require().Equal(common.HexToAddress(valOut.Validator.OperatorAddress), common.BytesToAddress(operatorAddress.Bytes()))
//			},
//			100000,
//			false,
//			"",
//		},
//		{
//			name: "success - empty validator",
//			malleate: func(_ common.Address) []interface{} {
//				newAddr, _ := testutiltx.NewAccAddressAndKey()
//				newValAddr := sdk.ValAddress(newAddr)
//				return []interface{}{
//					common.BytesToAddress(newValAddr.Bytes()),
//				}
//			},
//			postCheck: func(data []byte) {
//				var valOut staking.ValidatorOutput
//				err := s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorMethod, data)
//				s.Require().NoError(err, "failed to unpack output")
//				s.Require().Equal(valOut.Validator.OperatorAddress, "")
//				s.Require().Equal(valOut.Validator.Status, uint8(0))
//			},
//			gas: 100000,
//		},
//	}
//
//	for _, tc := range testCases {
//		s.Run(tc.name, func() {
//			s.SetupTest() // reset
//			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)
//
//			operatorAddress, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].OperatorAddress)
//			s.Require().NoError(err)
//
//			bz, err := s.precompile.Validator(s.network.GetContext(), &method, contract, tc.malleate(common.BytesToAddress(operatorAddress.Bytes())))
//
//			if tc.expErr {
//				s.Require().Error(err)
//				s.Require().Contains(err.Error(), tc.errContains)
//			} else {
//				s.Require().NoError(err)
//				s.Require().NotNil(bz)
//				tc.postCheck(bz)
//			}
//		})
//	}
//}
//
//func (s *PrecompileTestSuite) TestValidators() {
//	method := s.precompile.Methods[staking.ValidatorsMethod]
//
//	testCases := []struct {
//		name        string
//		malleate    func() []interface{}
//		postCheck   func(bz []byte)
//		gas         uint64
//		expErr      bool
//		errContains string
//	}{
//		{
//			"fail - empty input args",
//			func() []interface{} {
//				return []interface{}{}
//			},
//			func(_ []byte) {},
//			100000,
//			true,
//			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 2, 0),
//		},
//		{
//			"fail - invalid number of arguments",
//			func() []interface{} {
//				return []interface{}{
//					stakingtypes.Bonded.String(),
//				}
//			},
//			func(_ []byte) {},
//			100000,
//			true,
//			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 2, 1),
//		},
//		{
//			"success - bonded status & pagination w/countTotal",
//			func() []interface{} {
//				return []interface{}{
//					stakingtypes.Bonded.String(),
//					query.PageRequest{
//						Limit:      1,
//						CountTotal: true,
//					},
//				}
//			},
//			func(data []byte) {
//				const expLen = 1
//				var valOut staking.ValidatorsOutput
//				err := s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorsMethod, data)
//				s.Require().NoError(err, "failed to unpack output")
//
//				s.Require().Len(valOut.Validators, expLen)
//				// passed CountTotal = true
//				s.Require().Equal(len(s.network.GetValidators()), int(valOut.PageResponse.Total))
//				s.Require().NotEmpty(valOut.PageResponse.NextKey)
//				s.assertValidatorsResponse(valOut.Validators, expLen)
//			},
//			100000,
//			false,
//			"",
//		},
//		{
//			"success - bonded status & pagination w/countTotal & key is []byte{0}",
//			func() []interface{} {
//				return []interface{}{
//					stakingtypes.Bonded.String(),
//					query.PageRequest{
//						Key:        []byte{0},
//						Limit:      1,
//						CountTotal: true,
//					},
//				}
//			},
//			func(data []byte) {
//				const expLen = 1
//				var valOut staking.ValidatorsOutput
//				err := s.precompile.UnpackIntoInterface(&valOut, staking.ValidatorsMethod, data)
//				s.Require().NoError(err, "failed to unpack output")
//
//				s.Require().Len(valOut.Validators, expLen)
//				// passed CountTotal = true
//				s.Require().Equal(len(s.network.GetValidators()), int(valOut.PageResponse.Total))
//				s.Require().NotEmpty(valOut.PageResponse.NextKey)
//				s.assertValidatorsResponse(valOut.Validators, expLen)
//			},
//			100000,
//			false,
//			"",
//		},
//	}
//
//	for _, tc := range testCases {
//		s.Run(tc.name, func() {
//			s.SetupTest() // reset
//			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)
//
//			bz, err := s.precompile.Validators(s.network.GetContext(), &method, contract, tc.malleate())
//
//			if tc.expErr {
//				s.Require().Error(err)
//				s.Require().Contains(err.Error(), tc.errContains)
//			} else {
//				s.Require().NoError(err)
//				s.Require().NotNil(bz)
//				tc.postCheck(bz)
//			}
//		})
//	}
//}
//
//func (s *PrecompileTestSuite) TestRedelegation() {
//	method := s.precompile.Methods[staking.RedelegationMethod]
//	redelegateMethod := s.precompile.Methods[staking.RedelegateMethod]
//
//	testCases := []struct {
//		name        string
//		malleate    func(srcOperatorAddr, destOperatorAddr string) []interface{}
//		postCheck   func(bz []byte)
//		gas         uint64
//		expErr      bool
//		errContains string
//	}{
//		{
//			"fail - empty input args",
//			func(srcOperatorAddr, destOperatorAddr string) []interface{} {
//				return []interface{}{}
//			},
//			func(bz []byte) {},
//			100000,
//			true,
//			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 3, 0),
//		},
//		{
//			"fail - invalid delegator address",
//			func(srcOperatorAddr, destOperatorAddr string) []interface{} {
//				return []interface{}{
//					"invalid",
//					srcOperatorAddr,
//					destOperatorAddr,
//				}
//			},
//			func(bz []byte) {},
//			100000,
//			true,
//			fmt.Sprintf(cmn.ErrInvalidDelegator, "invalid"),
//		},
//		{
//			"fail - empty src validator addr",
//			func(srcOperatorAddr, destOperatorAddr string) []interface{} {
//				return []interface{}{
//					s.keyring.GetAddr(0),
//					"",
//					destOperatorAddr,
//				}
//			},
//			func(bz []byte) {},
//			100000,
//			true,
//			"empty address string is not allowed",
//		},
//		{
//			"fail - empty destination addr",
//			func(srcOperatorAddr, destOperatorAddr string) []interface{} {
//				return []interface{}{
//					s.keyring.GetAddr(0),
//					srcOperatorAddr,
//					"",
//				}
//			},
//			func(bz []byte) {},
//			100000,
//			true,
//			"empty address string is not allowed",
//		},
//		{
//			"success",
//			func(srcOperatorAddr, destOperatorAddr string) []interface{} {
//				return []interface{}{
//					s.keyring.GetAddr(0),
//					srcOperatorAddr,
//					destOperatorAddr,
//				}
//			},
//			func(data []byte) {
//				var redOut staking.RedelegationOutput
//				err := s.precompile.UnpackIntoInterface(&redOut, staking.RedelegationMethod, data)
//				s.Require().NoError(err, "failed to unpack output")
//				s.Require().Len(redOut.Redelegation.Entries, 1)
//				s.Require().Equal(redOut.Redelegation.Entries[0].CreationHeight, s.network.GetContext().BlockHeight())
//				s.Require().Equal(redOut.Redelegation.Entries[0].SharesDst, big.NewInt(1e18))
//			},
//			100000,
//			false,
//			"",
//		},
//		{
//			name: "success - no redelegation found",
//			malleate: func(srcOperatorAddr, _ string) []interface{} {
//				nonExistentOperator := sdk.ValAddress([]byte("non-existent-operator"))
//				return []interface{}{
//					s.keyring.GetAddr(0),
//					srcOperatorAddr,
//					nonExistentOperator.String(),
//				}
//			},
//			postCheck: func(data []byte) {
//				var redOut staking.RedelegationOutput
//				err := s.precompile.UnpackIntoInterface(&redOut, staking.RedelegationMethod, data)
//				s.Require().NoError(err, "failed to unpack output")
//				s.Require().Len(redOut.Redelegation.Entries, 0)
//			},
//			gas: 100000,
//		},
//	}
//
//	for _, tc := range testCases {
//		s.Run(tc.name, func() {
//			s.SetupTest() // reset
//			ctx := s.network.GetContext()
//
//			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)
//
//			delegationArgs := []interface{}{
//				s.keyring.GetAddr(0),
//				s.network.GetValidators()[0].OperatorAddress,
//				s.network.GetValidators()[1].OperatorAddress,
//				big.NewInt(1e18),
//			}
//
//			err := s.CreateAuthorization(ctx, s.keyring.GetAddr(0), staking.RedelegateAuthz, nil)
//			s.Require().NoError(err)
//
//			_, err = s.precompile.Redelegate(ctx, s.keyring.GetAddr(0), contract, s.network.GetStateDB(), &redelegateMethod, delegationArgs)
//			s.Require().NoError(err)
//
//			bz, err := s.precompile.Redelegation(ctx, &method, contract, tc.malleate(s.network.GetValidators()[0].OperatorAddress, s.network.GetValidators()[1].OperatorAddress))
//
//			if tc.expErr {
//				s.Require().Error(err)
//				s.Require().Contains(err.Error(), tc.errContains)
//			} else {
//				s.Require().NoError(err)
//				s.Require().NotNil(bz)
//				tc.postCheck(bz)
//			}
//		})
//	}
//}
//
//func (s *PrecompileTestSuite) TestRedelegations() {
//	var (
//		ctx             sdk.Context
//		delAmt                 = big.NewInt(3e17)
//		redelTotalCount uint64 = 2
//		method                 = s.precompile.Methods[staking.RedelegationsMethod]
//	)
//
//	testCases := []struct {
//		name        string
//		malleate    func() []interface{}
//		postCheck   func(bz []byte)
//		gas         uint64
//		expErr      bool
//		errContains string
//	}{
//		{
//			"fail - empty input args",
//			func() []interface{} {
//				return []interface{}{}
//			},
//			func(bz []byte) {},
//			100000,
//			true,
//			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 4, 0),
//		},
//		{
//			"fail - invalid delegator address",
//			func() []interface{} {
//				return []interface{}{
//					common.BytesToAddress([]byte("invalid")),
//					s.network.GetValidators()[0].OperatorAddress,
//					s.network.GetValidators()[1].OperatorAddress,
//					query.PageRequest{},
//				}
//			},
//			func(bz []byte) {},
//			100000,
//			true,
//			"redelegation not found",
//		},
//		{
//			"fail - invalid query | all empty args ",
//			func() []interface{} {
//				return []interface{}{
//					common.Address{},
//					"",
//					"",
//					query.PageRequest{},
//				}
//			},
//			func(data []byte) {},
//			100000,
//			true,
//			"invalid query. Need to specify at least a source validator address or delegator address",
//		},
//		{
//			"fail - invalid query | only destination validator address",
//			func() []interface{} {
//				return []interface{}{
//					common.Address{},
//					"",
//					s.network.GetValidators()[1].OperatorAddress,
//					query.PageRequest{},
//				}
//			},
//			func(data []byte) {},
//			100000,
//			true,
//			"invalid query. Need to specify at least a source validator address or delegator address",
//		},
//		{
//			"success - specified delegator, source & destination",
//			func() []interface{} {
//				return []interface{}{
//					s.keyring.GetAddr(0),
//					s.network.GetValidators()[0].OperatorAddress,
//					s.network.GetValidators()[1].OperatorAddress,
//					query.PageRequest{},
//				}
//			},
//			func(data []byte) {
//				s.assertRedelegationsOutput(data, 0, delAmt, ctx.BlockHeight(), false)
//			},
//			100000,
//			false,
//			"",
//		},
//		{
//			"success - specifying only source w/pagination",
//			func() []interface{} {
//				return []interface{}{
//					common.Address{},
//					s.network.GetValidators()[0].OperatorAddress,
//					"",
//					query.PageRequest{
//						Limit:      1,
//						CountTotal: true,
//					},
//				}
//			},
//			func(data []byte) {
//				s.assertRedelegationsOutput(data, redelTotalCount, delAmt, ctx.BlockHeight(), true)
//			},
//			100000,
//			false,
//			"",
//		},
//		{
//			"success - get all existing redelegations for a delegator w/pagination",
//			func() []interface{} {
//				return []interface{}{
//					s.keyring.GetAddr(0),
//					"",
//					"",
//					query.PageRequest{
//						Limit:      1,
//						CountTotal: true,
//					},
//				}
//			},
//			func(data []byte) {
//				s.assertRedelegationsOutput(data, redelTotalCount, delAmt, ctx.BlockHeight(), true)
//			},
//			100000,
//			false,
//			"",
//		},
//	}
//
//	for _, tc := range testCases {
//		s.Run(tc.name, func() {
//			s.SetupTest() // reset
//			ctx = s.network.GetContext()
//			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)
//
//			err := s.setupRedelegations(ctx, delAmt)
//			s.Require().NoError(err)
//
//			// query redelegations
//			bz, err := s.precompile.Redelegations(ctx, &method, contract, tc.malleate())
//
//			if tc.expErr {
//				s.Require().Error(err)
//				s.Require().Contains(err.Error(), tc.errContains)
//			} else {
//				s.Require().NoError(err)
//				s.Require().NotNil(bz)
//				tc.postCheck(bz)
//			}
//		})
//	}
//}
//
//func (s *PrecompileTestSuite) TestAllowance() {
//	var ctx sdk.Context
//	approvedCoin := sdk.Coin{Denom: s.bondDenom, Amount: math.NewInt(1e18)}
//	granteeAddr := testutiltx.GenerateAddress()
//	method := s.precompile.Methods[authorization.AllowanceMethod]
//
//	testCases := []struct {
//		name        string
//		malleate    func() []interface{}
//		postCheck   func(bz []byte)
//		gas         uint64
//		expErr      bool
//		errContains string
//	}{
//		{
//			"fail - empty input args",
//			func() []interface{} {
//				return []interface{}{}
//			},
//			func(bz []byte) {},
//			100000,
//			true,
//			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 3, 0),
//		},
//		{
//			"success - query delegate method allowance",
//			func() []interface{} {
//				err := s.CreateAuthorization(ctx, granteeAddr, staking.DelegateAuthz, &approvedCoin)
//				s.Require().NoError(err)
//
//				return []interface{}{
//					granteeAddr,
//					s.keyring.GetAddr(0),
//					staking.DelegateMsg,
//				}
//			},
//			func(bz []byte) {
//				var amountsOut *big.Int
//				err := s.precompile.UnpackIntoInterface(&amountsOut, authorization.AllowanceMethod, bz)
//				s.Require().NoError(err, "failed to unpack output")
//				s.Require().Equal(big.NewInt(1e18), amountsOut, "expected different allowed amount")
//			},
//			100000,
//			false,
//			"",
//		},
//		{
//			"success - return empty allowance if authorization is not found",
//			func() []interface{} {
//				return []interface{}{
//					granteeAddr,
//					s.keyring.GetAddr(0),
//					staking.UndelegateMsg,
//				}
//			},
//			func(bz []byte) {
//				var amountsOut *big.Int
//				err := s.precompile.UnpackIntoInterface(&amountsOut, authorization.AllowanceMethod, bz)
//				s.Require().NoError(err, "failed to unpack output")
//				s.Require().Equal(int64(0), amountsOut.Int64(), "expected no allowance")
//			},
//			100000,
//			false,
//			"",
//		},
//	}
//
//	for _, tc := range testCases {
//		s.Run(tc.name, func() {
//			s.SetupTest() // reset
//			ctx = s.network.GetContext()
//
//			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)
//
//			args := tc.malleate()
//			bz, err := s.precompile.Allowance(ctx, &method, contract, args)
//
//			if tc.expErr {
//				s.Require().Error(err)
//				s.Require().Contains(err.Error(), tc.errContains)
//			} else {
//				s.Require().NoError(err)
//				s.Require().NotNil(bz)
//				tc.postCheck(bz)
//			}
//		})
//	}
//}
