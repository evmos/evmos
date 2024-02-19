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
	cmn "github.com/evmos/evmos/v16/precompiles/common"
	"github.com/evmos/evmos/v16/precompiles/distribution"
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
				withdrawerAddr, err := s.network.App.DistrKeeper.GetDelegatorWithdrawAddr(ctx, s.keyring.GetAddr(0).Bytes())
				s.Require().NoError(err)
				s.Require().Equal(withdrawerAddr.Bytes(), s.keyring.GetAddr(0).Bytes())
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
	var ctx sdk.Context
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
			func(operatorAddress string) []interface{} {
				return []interface{}{}
			},
			func(data []byte) {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 2, 0),
		},
		{
			"fail - invalid delegator address",
			func(operatorAddress string) []interface{} {
				return []interface{}{
					"",
					operatorAddress,
				}
			},
			func(data []byte) {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidDelegator, ""),
		},
		{
			"fail - invalid validator address",
			func(operatorAddress string) []interface{} {
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
			func(operatorAddress string) []interface{} {
				valAddr, err := sdk.ValAddressFromBech32(operatorAddress)
				s.Require().NoError(err)
				val, _ := s.network.App.StakingKeeper.GetValidator(ctx, valAddr)
				coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1e18)))
				s.network.App.DistrKeeper.AllocateTokensToValidator(ctx, val, sdk.NewDecCoinsFromCoins(coins...))
				return []interface{}{
					s.keyring.GetAddr(0),
					operatorAddress,
				}
			},
			func(data []byte) {
				var coins []cmn.Coin
				err := s.precompile.UnpackIntoInterface(&coins, distribution.WithdrawDelegatorRewardsMethod, data)
				s.Require().NoError(err, "failed to unpack output")
				s.Require().Equal(coins[0].Denom, utils.BaseDenom)
				s.Require().Equal(coins[0].Amount, big.NewInt(1000000000000000000))
				// Check bank balance after the withdrawal of rewards
				balance := s.network.App.BankKeeper.GetBalance(ctx, s.keyring.GetAddr(0).Bytes(), utils.BaseDenom)
				s.Require().Equal(balance.Amount.BigInt(), big.NewInt(6000000000000000000))
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

			// sanity check to make sure the starting balance is always 5 EVMOS
			balance := s.network.App.BankKeeper.GetBalance(ctx, s.keyring.GetAddr(0).Bytes(), utils.BaseDenom)
			s.Require().Equal(balance.Amount.BigInt(), big.NewInt(5000000000000000000))

			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(s.T(), ctx, s.keyring.GetAddr(0), s.precompile, tc.gas)

			bz, err := s.precompile.WithdrawDelegatorRewards(ctx, s.keyring.GetAddr(0), contract, s.network.GetStateDB(), &method, tc.malleate(s.network.GetValidators()[0].OperatorAddress))

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
	var ctx sdk.Context
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
			func(operatorAddress string) []interface{} {
				return []interface{}{}
			},
			func(data []byte) {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 1, 0),
		},
		{
			"fail - invalid validator address",
			func(operatorAddress string) []interface{} {
				return []interface{}{
					nil,
				}
			},
			func(data []byte) {},
			200000,
			true,
			"invalid validator address",
		},
		{
			"success - withdraw all commission from a single validator",
			func(operatorAddress string) []interface{} {
				valAddr, err := sdk.ValAddressFromBech32(operatorAddress)
				s.Require().NoError(err)
				valCommission := sdk.DecCoins{sdk.NewDecCoinFromDec(utils.BaseDenom, math.LegacyNewDecWithPrec(1000000000000000000, 1))}
				// set outstanding rewards
				s.network.App.DistrKeeper.SetValidatorOutstandingRewards(ctx, valAddr, types.ValidatorOutstandingRewards{Rewards: valCommission})
				// set commission
				s.network.App.DistrKeeper.SetValidatorAccumulatedCommission(ctx, valAddr, types.ValidatorAccumulatedCommission{Commission: valCommission})
				return []interface{}{
					operatorAddress,
				}
			},
			func(data []byte) {
				var coins []cmn.Coin
				err := s.precompile.UnpackIntoInterface(&coins, distribution.WithdrawValidatorCommissionMethod, data)
				s.Require().NoError(err, "failed to unpack output")
				s.Require().Equal(coins[0].Denom, utils.BaseDenom)
				s.Require().Equal(coins[0].Amount, big.NewInt(100000000000000000))
				// Check bank balance after the withdrawal of commission
				balance := s.network.App.BankKeeper.GetBalance(ctx, []byte(s.network.GetValidators()[0].GetOperator()), utils.BaseDenom)
				s.Require().Equal(balance.Amount.BigInt(), big.NewInt(100000000000000000))
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
			// Sanity check to make sure the starting balance is always 0
			balance := s.network.App.BankKeeper.GetBalance(ctx, []byte(s.network.GetValidators()[0].GetOperator()), utils.BaseDenom)
			s.Require().Equal(balance.Amount.BigInt(), big.NewInt(0))
			s.Require().Equal(balance.Denom, utils.BaseDenom)

			validatorAddress := common.BytesToAddress([]byte(s.network.GetValidators()[0].GetOperator()))
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
	var ctx sdk.Context
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
			func(data []byte) {
				balance := s.network.App.BankKeeper.GetBalance(ctx, s.keyring.GetAddr(0).Bytes(), utils.BaseDenom)
				s.Require().Equal(balance.Amount.BigInt(), big.NewInt(7e18))
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
				s.Require().Equal(balance.Amount.BigInt(), big.NewInt(6e18))
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

			// Sanity check to make sure the starting balance is always 5 EVMOS
			balance := s.network.App.BankKeeper.GetBalance(ctx, s.keyring.GetAddr(0).Bytes(), utils.BaseDenom)
			s.Require().Equal(balance.Amount.BigInt(), big.NewInt(5e18))

			// Distribute rewards to the 2 validators, 1 EVMOS each
			for _, val := range s.network.GetValidators() {
				coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1e18)))
				s.network.App.DistrKeeper.AllocateTokensToValidator(ctx, val, sdk.NewDecCoinsFromCoins(coins...))
			}

			bz, err := s.precompile.ClaimRewards(ctx, s.keyring.GetAddr(0), contract, s.network.GetStateDB(), &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck(bz)
			}
		})
	}
}
