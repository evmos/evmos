package erc20_test

import (
	"errors"
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/precompiles/authorization"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/precompiles/erc20"
	"github.com/evmos/evmos/v20/precompiles/testutil"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

//nolint:dupl // tests are not duplicate between the functions
func (s *PrecompileTestSuite) TestApprove() {
	method := s.precompile.Methods[authorization.ApproveMethod]
	amount := int64(100)

	testcases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func()
		expPass     bool
		errContains string
	}{
		{
			name:        "fail - empty args",
			malleate:    func() []interface{} { return nil },
			errContains: "invalid number of arguments",
		},
		{
			name: "fail - invalid number of arguments",
			malleate: func() []interface{} {
				return []interface{}{
					1, 2, 3,
				}
			},
			errContains: "invalid number of arguments",
		},
		{
			name: "fail - invalid address",
			malleate: func() []interface{} {
				return []interface{}{
					"invalid address", big.NewInt(2),
				}
			},
			errContains: "invalid address",
		},
		{
			name: "fail - invalid amount",
			malleate: func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(1), "invalid amount",
				}
			},
			errContains: "invalid amount",
		},
		{
			name: "fail - negative amount",
			malleate: func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(-1),
				}
			},
			errContains: erc20.ErrNegativeAmount.Error(),
		},
		{
			name: "fail - approve uint256 overflow",
			malleate: func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(1), new(big.Int).Add(abi.MaxUint256, common.Big1),
				}
			},
			errContains: "causes integer overflow",
		},
		{
			name: "fail - approve to zero with existing authorization only for other denominations",
			malleate: func() []interface{} {
				// NOTE: We are setting up a grant with a spend limit for a different denomination
				// and then trying to approve an amount of zero for the token denomination
				s.setupSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetPrivKey(0),
					sdk.NewCoins(
						sdk.NewInt64Coin(s.bondDenom, 1),
					),
				)

				return []interface{}{
					s.keyring.GetAddr(1), common.Big0,
				}
			},
			errContains: fmt.Sprintf(erc20.ErrNoAllowanceForToken, s.tokenDenom),
			postCheck: func() {
				// NOTE: Here we check that the authorization was not adjusted
				s.requireSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
					sdk.NewCoins(
						sdk.NewInt64Coin(s.bondDenom, 1),
					),
					[]string{},
				)
			},
		},
		{
			name: "pass - approve without existing authorization",
			malleate: func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(amount),
				}
			},
			expPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.tokenDenom, amount)),
					[]string{},
				)
			},
		},
		{
			name: "pass - approve with existing authorization",
			malleate: func() []interface{} {
				s.setupSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetPrivKey(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.tokenDenom, 1)),
				)

				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(amount),
				}
			},
			expPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.tokenDenom, amount)),
					[]string{},
				)
			},
		},
		{
			name: "pass - approve with existing authorization in different denomination",
			malleate: func() []interface{} {
				s.setupSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetPrivKey(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.bondDenom, 1)),
				)

				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(amount),
				}
			},
			expPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
					// Check that the approval is extended with the new denomination instead of overwritten
					sdk.NewCoins(sdk.NewInt64Coin(s.bondDenom, 1), sdk.NewInt64Coin(s.tokenDenom, amount)),
					[]string{},
				)
			},
		},
		{
			name: "pass - delete existing authorization",
			malleate: func() []interface{} {
				s.setupSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetPrivKey(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.tokenDenom, 1)),
				)

				return []interface{}{
					s.keyring.GetAddr(1), common.Big0,
				}
			},
			expPass: true,
			postCheck: func() {
				grants, err := s.grpcHandler.GetGrantsByGrantee(s.keyring.GetAccAddr(1).String())
				s.Require().NoError(err, "expected no error querying the grants")
				s.Require().Len(grants, 0, "expected grant to be deleted")
			},
		},
		{
			name: "pass - delete denomination from spend limit but leave other denoms",
			malleate: func() []interface{} {
				s.setupSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetPrivKey(0),
					sdk.NewCoins(
						sdk.NewInt64Coin(s.tokenDenom, 1),
						sdk.NewInt64Coin(s.bondDenom, 1),
					),
				)

				return []interface{}{
					s.keyring.GetAddr(1), common.Big0,
				}
			},
			expPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
					// Check that the approval does not have a spend limit for the deleted denomination
					// but still contains the other denom
					sdk.NewCoins(sdk.NewInt64Coin(s.bondDenom, 1)),
					[]string{},
				)
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()

			ctx := s.network.GetContext()

			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(
				s.T(),
				ctx,
				s.keyring.GetAddr(0),
				s.precompile,
				200_000,
			)

			var args []interface{}
			if tc.malleate != nil {
				args = tc.malleate()
			}

			bz, err := s.precompile.Approve(
				ctx,
				contract,
				s.network.GetStateDB(),
				&method,
				args,
			)

			if tc.expPass {
				s.Require().NoError(err, "expected no error")
				s.Require().NotNil(bz, "expected non-nil bytes")
			} else {
				s.Require().Error(err, "expected error")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
				s.Require().Empty(bz, "expected empty bytes")
			}

			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}

//nolint:dupl // tests are not duplicate between the functions
func (s *PrecompileTestSuite) TestIncreaseAllowance() {
	method := s.precompile.Methods[authorization.IncreaseAllowanceMethod]
	amount := int64(100)
	increaseAmount := int64(200)

	testcases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func()
		expPass     bool
		errContains string
	}{
		{
			name:        "fail - empty args",
			malleate:    func() []interface{} { return nil },
			errContains: "invalid number of arguments",
		},
		{
			name: "fail - invalid number of arguments",
			malleate: func() []interface{} {
				return []interface{}{
					1, 2, 3,
				}
			},
			errContains: "invalid number of arguments",
		},
		{
			name: "fail - invalid address",
			malleate: func() []interface{} {
				return []interface{}{
					"invalid address", big.NewInt(2),
				}
			},
			errContains: "invalid address",
		},
		{
			name: "fail - invalid amount",
			malleate: func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(1), "invalid amount",
				}
			},
			errContains: "invalid amount",
		},
		{
			name: "fail - negative amount",
			malleate: func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(-1),
				}
			},
			errContains: erc20.ErrIncreaseNonPositiveValue.Error(),
		},
		{
			name: "pass - increase allowance without existing authorization",
			malleate: func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(increaseAmount),
				}
			},
			expPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.tokenDenom, increaseAmount)),
					[]string{},
				)
			},
		},
		{
			name: "pass - increase allowance with existing authorization",
			malleate: func() []interface{} {
				s.setupSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetPrivKey(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.tokenDenom, amount)),
				)

				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(increaseAmount),
				}
			},
			expPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.tokenDenom, amount+increaseAmount)),
					[]string{},
				)
			},
		},
		{
			name: "fail - uint256 overflow when increasing allowance",
			malleate: func() []interface{} {
				// NOTE: We are setting up a grant with a spend limit of the maximum uint256 value
				// and then trying to approve an amount that would overflow the uint256 value
				s.setupSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetPrivKey(0),
					sdk.NewCoins(
						sdk.NewInt64Coin(s.bondDenom, 1),
						sdk.NewCoin(s.tokenDenom, math.NewIntFromBigInt(abi.MaxUint256)),
					),
				)

				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(amount),
				}
			},
			errContains: erc20.ConvertErrToERC20Error(errors.New(cmn.ErrIntegerOverflow)).Error(),
			postCheck: func() {
				s.requireSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
					// NOTE: The amounts should not have been adjusted after failing the overflow check.
					sdk.NewCoins(
						sdk.NewInt64Coin(s.bondDenom, 1),
						sdk.NewCoin(s.tokenDenom, math.NewIntFromBigInt(abi.MaxUint256)),
					),
					[]string{},
				)
			},
		},
		{
			name: "pass - increase allowance with existing authorization in different denomination",
			malleate: func() []interface{} {
				s.setupSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetPrivKey(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.bondDenom, amount)),
				)

				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(increaseAmount),
				}
			},
			expPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
					// NOTE: The approval in the precompile denomination is added to the existing
					// approval for the different denomination.
					sdk.NewCoins(sdk.NewInt64Coin(s.bondDenom, amount), sdk.NewInt64Coin(s.tokenDenom, increaseAmount)),
					[]string{},
				)
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()

			ctx := s.network.GetContext()

			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(
				s.T(),
				ctx,
				s.keyring.GetAddr(0),
				s.precompile,
				200_000,
			)

			var args []interface{}
			if tc.malleate != nil {
				args = tc.malleate()
			}

			bz, err := s.precompile.IncreaseAllowance(
				ctx,
				contract,
				s.network.GetStateDB(),
				&method,
				args,
			)

			if tc.expPass {
				s.Require().NoError(err, "expected no error")
				s.Require().NotNil(bz, "expected non-nil bytes")
			} else {
				s.Require().Error(err, "expected error")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
				s.Require().Empty(bz, "expected empty bytes")
			}

			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}

//nolint:dupl // tests are not duplicate between the functions
func (s *PrecompileTestSuite) TestDecreaseAllowance() {
	method := s.precompile.Methods[authorization.DecreaseAllowanceMethod]
	amount := int64(100)
	decreaseAmount := int64(50)

	testcases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func()
		expPass     bool
		errContains string
	}{
		{
			name:        "fail - empty args",
			malleate:    func() []interface{} { return nil },
			errContains: "invalid number of arguments",
		},
		{
			name: "fail - invalid number of arguments",
			malleate: func() []interface{} {
				return []interface{}{
					1, 2, 3,
				}
			},
			errContains: "invalid number of arguments",
		},
		{
			name: "fail - invalid address",
			malleate: func() []interface{} {
				return []interface{}{
					"invalid address", big.NewInt(2),
				}
			},
			errContains: "invalid address",
		},
		{
			name: "fail - invalid amount",
			malleate: func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(1), "invalid amount",
				}
			},
			errContains: "invalid amount",
		},
		{
			name: "fail - negative amount",
			malleate: func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(-1),
				}
			},
			errContains: erc20.ErrDecreaseNonPositiveValue.Error(),
		},
		{
			name: "fail - decrease allowance without existing authorization",
			malleate: func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(decreaseAmount),
				}
			},
			errContains: "does not exist or is expired",
		},
		{
			name: "fail - decrease allowance with existing authorization only for other denominations",
			malleate: func() []interface{} {
				// NOTE: We are setting up a grant with a spend limit for a different denomination
				// and then trying to decrease the allowance for the token denomination
				s.setupSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetPrivKey(0),
					sdk.NewCoins(
						sdk.NewInt64Coin(s.bondDenom, 1),
					),
				)

				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(decreaseAmount),
				}
			},
			errContains: fmt.Sprintf("allowance for token %s does not exist", s.tokenDenom),
			postCheck: func() {
				// NOTE: Here we check that the authorization was not adjusted
				s.requireSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
					sdk.NewCoins(
						sdk.NewInt64Coin(s.bondDenom, 1),
					),
					[]string{},
				)
			},
		},
		{
			name: "pass - decrease allowance with existing authorization",
			malleate: func() []interface{} {
				s.setupSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetPrivKey(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.tokenDenom, amount)),
				)

				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(decreaseAmount),
				}
			},
			expPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.tokenDenom, decreaseAmount)),
					[]string{},
				)
			},
		},
		{
			name: "pass - decrease to zero and delete existing authorization",
			malleate: func() []interface{} {
				s.setupSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetPrivKey(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.tokenDenom, amount)),
				)

				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(amount),
				}
			},
			expPass: true,
			postCheck: func() {
				// Check that the authorization was deleted
				grants, err := s.grpcHandler.GetGrantsByGrantee(s.keyring.GetAccAddr(1).String())
				s.Require().NoError(err, "expected no error querying the grants")
				s.Require().Len(grants, 0, "expected grant to be deleted")
			},
		},
		{
			name: "pass - decrease allowance with existing authorization in different denomination",
			malleate: func() []interface{} {
				s.setupSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetPrivKey(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.bondDenom, amount), sdk.NewInt64Coin(s.tokenDenom, amount)),
				)

				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(decreaseAmount),
				}
			},
			expPass: true,
			postCheck: func() {
				// NOTE: Here we check that the authorization for the other denom was not deleted and the spend limit
				// for token denom was adjusted as expected
				s.requireSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.bondDenom, amount), sdk.NewInt64Coin(s.tokenDenom, amount-decreaseAmount)),
					[]string{},
				)
			},
		},
		{
			name: "pass - decrease allowance to zero for denom with existing authorization in other denominations",
			malleate: func() []interface{} {
				s.setupSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetPrivKey(0),
					sdk.NewCoins(
						sdk.NewInt64Coin(s.bondDenom, amount),
						sdk.NewInt64Coin(s.tokenDenom, amount),
					),
				)

				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(amount),
				}
			},
			expPass: true,
			postCheck: func() {
				// NOTE: Here we check that the authorization for the other denom was not deleted
				s.requireSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.bondDenom, amount)),
					[]string{},
				)
			},
		},
		{
			name: "fail - decrease allowance with existing authorization but decreased amount too high",
			malleate: func() []interface{} {
				s.setupSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetPrivKey(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.tokenDenom, amount)),
				)

				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(amount + 1),
				}
			},
			errContains: erc20.ConvertErrToERC20Error(errors.New("subtracted value cannot be greater than existing allowance")).Error(),
		},
		{
			name: "fail - decrease allowance with existing authorization in different denomination",
			malleate: func() []interface{} {
				s.setupSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetPrivKey(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.bondDenom, amount)),
				)

				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(decreaseAmount),
				}
			},
			errContains: fmt.Sprintf(erc20.ErrNoAllowanceForToken, s.tokenDenom),
			postCheck: func() {
				// NOTE: Here we check that the authorization for the other denom was not deleted
				s.requireSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.bondDenom, amount)),
					[]string{},
				)
			},
		},
		{
			name: "fail - decrease allowance with existing authorization in different denomination but decreased amount too high",
			malleate: func() []interface{} {
				s.setupSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetPrivKey(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.bondDenom, amount), sdk.NewInt64Coin(s.tokenDenom, 1)),
				)

				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(decreaseAmount),
				}
			},
			errContains: erc20.ConvertErrToERC20Error(errors.New("subtracted value cannot be greater than existing allowance")).Error(),
			postCheck: func() {
				// NOTE: Here we check that the authorization was not adjusted
				s.requireSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.bondDenom, amount), sdk.NewInt64Coin(s.tokenDenom, 1)),
					[]string{},
				)
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()

			ctx := s.network.GetContext()

			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(
				s.T(),
				ctx,
				s.keyring.GetAddr(0),
				s.precompile,
				200_000,
			)

			var args []interface{}
			if tc.malleate != nil {
				args = tc.malleate()
			}

			bz, err := s.precompile.DecreaseAllowance(
				ctx,
				contract,
				s.network.GetStateDB(),
				&method,
				args,
			)

			if tc.expPass {
				s.Require().NoError(err, "expected no error")
				s.Require().NotNil(bz, "expected non-nil bytes")
			} else {
				s.Require().Error(err, "expected error")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
				s.Require().Empty(bz, "expected empty bytes")
			}

			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}
