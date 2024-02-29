package distribution_test

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v16/precompiles/testutil"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	cmn "github.com/evmos/evmos/v16/precompiles/common"
	"github.com/evmos/evmos/v16/precompiles/distribution"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	utiltx "github.com/evmos/evmos/v16/testutil/tx"
	"github.com/evmos/evmos/v16/utils"
)

func (s *PrecompileTestSuite) TestSetWithdrawAddress() {
	var ctx sdk.Context
	method := s.precompile.Methods[distribution.SetWithdrawAddressMethod]
	newWithdrawerAddr := utiltx.GenerateAddress()

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
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 2, 0),
		},
		{
			"fail - invalid delegator address",
			func() []interface{} {
				return []interface{}{
					"",
					s.keyring.GetAddr(0).String(),
				}
			},
			func() {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidDelegator, ""),
		},
		{
			"fail - invalid withdrawer address",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					nil,
				}
			},
			func() {},
			200000,
			true,
			"invalid withdraw address: empty address string is not allowed: invalid address",
		},
		{
			"success - using the same address withdrawer address",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					s.keyring.GetAddr(0).String(),
				}
			},
			func() {
				withdrawerAddr, err := s.network.App.DistrKeeper.GetDelegatorWithdrawAddr(ctx, s.keyring.GetAccAddr(0))
				s.Require().NoError(err)
				s.Require().Equal(withdrawerAddr.String(), s.keyring.GetAccAddr(0).String())
			},
			20000,
			false,
			"",
		},
		{
			"success - using a different withdrawer address",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					newWithdrawerAddr.String(),
				}
			},
			func() {
				withdrawerAddr, err := s.network.App.DistrKeeper.GetDelegatorWithdrawAddr(ctx, s.keyring.GetAddr(0).Bytes())
				s.Require().NoError(err)
				s.Require().Equal(withdrawerAddr.Bytes(), newWithdrawerAddr.Bytes())
			},
			20000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx = s.network.GetContext()

			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(s.T(), ctx, s.keyring.GetAddr(0), s.precompile, tc.gas)

			_, err := s.precompile.SetWithdrawAddress(ctx, s.keyring.GetAddr(0), contract, s.network.GetStateDB(), &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}

func (s *PrecompileTestSuite) TestWithdrawDelegatorRewards() {
	var (
		ctx sdk.Context
		err error
	)
	method := s.precompile.Methods[distribution.WithdrawDelegatorRewardsMethod]

	testCases := []struct {
		name        string
		malleate    func(val stakingtypes.Validator) []interface{}
		postCheck   func(data []byte)
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func(stakingtypes.Validator) []interface{} {
				return []interface{}{}
			},
			func(data []byte) {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 2, 0),
		},
		{
			"fail - invalid delegator address",
			func(val stakingtypes.Validator) []interface{} {
				return []interface{}{
					"",
					val.OperatorAddress,
				}
			},
			func(data []byte) {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidDelegator, ""),
		},
		{
			"fail - invalid validator address",
			func(stakingtypes.Validator) []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					nil,
				}
			},
			func(data []byte) {},
			200000,
			true,
			"invalid validator address",
		},
		{
			"success - withdraw rewards from a single validator without commission",
			func(val stakingtypes.Validator) []interface{} {
				ctx, err = s.prepareStakingRewards(
					ctx,
					stakingRewards{
						Validator: val,
						Delegator: s.keyring.GetAccAddr(0),
						RewardAmt: testRewardsAmt,
					},
				)
				s.Require().NoError(err, "failed to unpack output")
				return []interface{}{
					s.keyring.GetAddr(0),
					val.OperatorAddress,
				}
			},
			func(data []byte) {
				var coins []cmn.Coin
				err := s.precompile.UnpackIntoInterface(&coins, distribution.WithdrawDelegatorRewardsMethod, data)
				s.Require().NoError(err, "failed to unpack output")
				s.Require().Equal(coins[0].Denom, utils.BaseDenom)
				s.Require().Equal(coins[0].Amount.Int64(), expRewardsAmt.Int64())
				// Check bank balance after the withdrawal of rewards
				balance := s.network.App.BankKeeper.GetBalance(ctx, s.keyring.GetAddr(0).Bytes(), utils.BaseDenom)
				s.Require().True(balance.Amount.GT(network.PrefundedAccountInitialBalance))
			},
			20000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx = s.network.GetContext()

			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(s.T(), ctx, s.keyring.GetAddr(0), s.precompile, tc.gas)

			args := tc.malleate(s.network.GetValidators()[0])
			bz, err := s.precompile.WithdrawDelegatorRewards(ctx, s.keyring.GetAddr(0), contract, s.network.GetStateDB(), &method, args)

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck(bz)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestWithdrawValidatorCommission() {
	var (
		ctx         sdk.Context
		prevBalance sdk.Coin
	)
	method := s.precompile.Methods[distribution.WithdrawDelegatorRewardsMethod]

	testCases := []struct {
		name        string
		malleate    func(operatorAddress string) []interface{}
		postCheck   func(data []byte)
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func(string) []interface{} {
				return []interface{}{}
			},
			func(data []byte) {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 1, 0),
		},
		{
			"fail - invalid validator address",
			func(string) []interface{} {
				return []interface{}{
					nil,
				}
			},
			func(data []byte) {},
			200000,
			true,
			"empty address string is not allowed",
		},
		{
			"success - withdraw all commission from a single validator",
			func(operatorAddress string) []interface{} {
				valAddr, err := sdk.ValAddressFromBech32(operatorAddress)
				s.Require().NoError(err)
				amt := math.LegacyNewDecWithPrec(1000000000000000000, 1)
				valCommission := sdk.DecCoins{sdk.NewDecCoinFromDec(utils.BaseDenom, amt)}
				// set outstanding rewards
				s.network.App.DistrKeeper.SetValidatorOutstandingRewards(ctx, valAddr, types.ValidatorOutstandingRewards{Rewards: valCommission})
				// set commission
				s.network.App.DistrKeeper.SetValidatorAccumulatedCommission(ctx, valAddr, types.ValidatorAccumulatedCommission{Commission: valCommission})

				// fund distr mod to pay for rewards + commission
				coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, amt.Mul(math.LegacyNewDec(2)).RoundInt()))
				err = s.mintCoinsForDistrMod(ctx, coins)
				s.Require().NoError(err)
				return []interface{}{
					operatorAddress,
				}
			},
			func(data []byte) {
				var coins []cmn.Coin
				amt := math.NewInt(100000000000000000)
				err := s.precompile.UnpackIntoInterface(&coins, distribution.WithdrawValidatorCommissionMethod, data)
				s.Require().NoError(err, "failed to unpack output")
				s.Require().Equal(coins[0].Denom, utils.BaseDenom)
				s.Require().Equal(coins[0].Amount, amt.BigInt())

				// Check bank balance after the withdrawal of commission
				valAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].GetOperator())
				s.Require().NoError(err)
				balance := s.network.App.BankKeeper.GetBalance(ctx, valAddr.Bytes(), utils.BaseDenom)
				s.Require().Equal(balance.Amount, prevBalance.Amount.Add(amt))
				s.Require().Equal(balance.Denom, utils.BaseDenom)
			},
			20000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx = s.network.GetContext()

			valAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].GetOperator())
			s.Require().NoError(err)

			prevBalance = s.network.App.BankKeeper.GetBalance(ctx, valAddr.Bytes(), utils.BaseDenom)

			validatorAddress := common.BytesToAddress(valAddr.Bytes())
			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(s.T(), ctx, validatorAddress, s.precompile, tc.gas)

			bz, err := s.precompile.WithdrawValidatorCommission(ctx, validatorAddress, contract, s.network.GetStateDB(), &method, tc.malleate(s.network.GetValidators()[0].OperatorAddress))

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck(bz)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestClaimRewards() {
	var (
		ctx         sdk.Context
		prevBalance sdk.Coin
	)
	method := s.precompile.Methods[distribution.ClaimRewardsMethod]

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func(data []byte)
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			func(data []byte) {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 2, 0),
		},
		{
			"fail - invalid delegator address",
			func() []interface{} {
				return []interface{}{
					nil,
					10,
				}
			},
			func(data []byte) {},
			200000,
			true,
			"invalid delegator address",
		},
		{
			"fail - invalid type for maxRetrieve: expected uint32",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					big.NewInt(100000000000000000),
				}
			},
			func(data []byte) {},
			200000,
			true,
			"invalid type for maxRetrieve: expected uint32",
		},
		{
			"success - withdraw from all validators - 2",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					uint32(2),
				}
			},
			func(_ []byte) {
				balance := s.network.App.BankKeeper.GetBalance(ctx, s.keyring.GetAddr(0).Bytes(), utils.BaseDenom)
				// twice the rewards amount (rewards from 2 validators) - 5% commission
				expRewards := expRewardsAmt.Mul(math.NewInt(2))
				s.Require().Equal(balance.Amount, prevBalance.Amount.Add(expRewards))
			},
			20000,
			false,
			"",
		},
		{
			"success - withdraw from only 1 validator",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					uint32(1),
				}
			},
			func(data []byte) {
				balance := s.network.App.BankKeeper.GetBalance(ctx, s.keyring.GetAddr(0).Bytes(), utils.BaseDenom)
				s.Require().Equal(balance.Amount, prevBalance.Amount.Add(expRewardsAmt))
			},
			20000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx = s.network.GetContext()

			var (
				contract *vm.Contract
				err      error
			)
			addr := s.keyring.GetAddr(0)
			contract, ctx = testutil.NewPrecompileContract(s.T(), ctx, addr, s.precompile, tc.gas)

			validators := s.network.GetValidators()
			srs := make([]stakingRewards, len(validators))
			for i, val := range validators {
				srs[i] = stakingRewards{
					Delegator: addr.Bytes(),
					Validator: val,
					RewardAmt: testRewardsAmt,
				}
			}

			ctx, err = s.prepareStakingRewards(ctx, srs...)
			s.Require().NoError(err)

			// get previous balance to compare final balance in the postCheck func
			prevBalance = s.network.App.BankKeeper.GetBalance(ctx, addr.Bytes(), utils.BaseDenom)

			bz, err := s.precompile.ClaimRewards(ctx, addr, contract, s.network.GetStateDB(), &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck(bz)
			}
		})
	}
}
