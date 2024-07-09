package distribution_test

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	"github.com/evmos/evmos/v18/precompiles/testutil"
	"github.com/evmos/evmos/v18/x/evm/core/vm"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	"github.com/evmos/evmos/v18/precompiles/distribution"
	utiltx "github.com/evmos/evmos/v18/testutil/tx"
	"github.com/evmos/evmos/v18/utils"
)

func (s *PrecompileTestSuite) TestSetWithdrawAddress() {
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
					s.address.String(),
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
					s.address,
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
					s.address,
					s.address.String(),
				}
			},
			func() {
				withdrawerAddr := s.app.DistrKeeper.GetDelegatorWithdrawAddr(s.ctx, s.address.Bytes())
				s.Require().Equal(withdrawerAddr.Bytes(), s.address.Bytes())
			},
			20000,
			false,
			"",
		},
		{
			"success - using a different withdrawer address",
			func() []interface{} {
				return []interface{}{
					s.address,
					newWithdrawerAddr.String(),
				}
			},
			func() {
				withdrawerAddr := s.app.DistrKeeper.GetDelegatorWithdrawAddr(s.ctx, s.address.Bytes())
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

			var contract *vm.Contract
			contract, s.ctx = testutil.NewPrecompileContract(s.T(), s.ctx, s.address, s.precompile, tc.gas)

			_, err := s.precompile.SetWithdrawAddress(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate())

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
			func([]byte) {},
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
			func([]byte) {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidDelegator, ""),
		},
		{
			"fail - invalid validator address",
			func(string) []interface{} {
				return []interface{}{
					s.address,
					nil,
				}
			},
			func([]byte) {},
			200000,
			true,
			"invalid validator address",
		},
		{
			"success - withdraw rewards from a single validator without commission",
			func(operatorAddress string) []interface{} {
				valAddr, err := sdk.ValAddressFromBech32(operatorAddress)
				s.Require().NoError(err)
				val, _ := s.app.StakingKeeper.GetValidator(s.ctx, valAddr)
				coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1e18)))
				s.app.DistrKeeper.AllocateTokensToValidator(s.ctx, val, sdk.NewDecCoinsFromCoins(coins...))
				return []interface{}{
					s.address,
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
				balance := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), utils.BaseDenom)
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

			// sanity check to make sure the starting balance is always 5 EVMOS
			balance := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), utils.BaseDenom)
			s.Require().Equal(balance.Amount.BigInt(), big.NewInt(5000000000000000000))

			var contract *vm.Contract
			contract, s.ctx = testutil.NewPrecompileContract(s.T(), s.ctx, s.address, s.precompile, tc.gas)

			bz, err := s.precompile.WithdrawDelegatorRewards(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate(s.validators[0].OperatorAddress))

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
			func([]byte) {},
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
			func([]byte) {},
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
				s.app.DistrKeeper.SetValidatorOutstandingRewards(s.ctx, valAddr, types.ValidatorOutstandingRewards{Rewards: valCommission})
				// set commission
				s.app.DistrKeeper.SetValidatorAccumulatedCommission(s.ctx, valAddr, types.ValidatorAccumulatedCommission{Commission: valCommission})
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
				balance := s.app.BankKeeper.GetBalance(s.ctx, s.validators[0].GetOperator().Bytes(), utils.BaseDenom)
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

			// Sanity check to make sure the starting balance is always 0
			balance := s.app.BankKeeper.GetBalance(s.ctx, s.validators[0].GetOperator().Bytes(), utils.BaseDenom)
			s.Require().Equal(balance.Amount.BigInt(), big.NewInt(0))
			s.Require().Equal(balance.Denom, utils.BaseDenom)

			validatorAddress := common.BytesToAddress(s.validators[0].GetOperator().Bytes())
			var contract *vm.Contract
			contract, s.ctx = testutil.NewPrecompileContract(s.T(), s.ctx, validatorAddress, s.precompile, tc.gas)

			bz, err := s.precompile.WithdrawValidatorCommission(s.ctx, validatorAddress, contract, s.stateDB, &method, tc.malleate(s.validators[0].OperatorAddress))

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
			func([]byte) {},
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
			func([]byte) {},
			200000,
			true,
			"invalid delegator address",
		},
		{
			"fail - invalid type for maxRetrieve: expected uint32",
			func() []interface{} {
				return []interface{}{
					s.address,
					big.NewInt(100000000000000000),
				}
			},
			func([]byte) {},
			200000,
			true,
			"invalid type for maxRetrieve: expected uint32",
		},
		{
			"pass - withdraw from validators with maxRetrieve higher than number of validators",
			func() []interface{} {
				return []interface{}{
					s.address,
					uint32(10),
				}
			},
			func([]byte) {
				balance := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), utils.BaseDenom)
				s.Require().Equal(balance.Amount.BigInt(), big.NewInt(7e18))
			},
			20000,
			false,
			"",
		},
		{
			"fail - too many retrieved results",
			func() []interface{} {
				return []interface{}{
					s.address,
					uint32(32_000_000),
				}
			},
			func([]byte) {},
			200000,
			true,
			"maxRetrieve (32000000) parameter exceeds the maximum number of validators (100)",
		},
		{
			"success - withdraw from all validators - 2",
			func() []interface{} {
				return []interface{}{
					s.address,
					uint32(2),
				}
			},
			func([]byte) {
				balance := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), utils.BaseDenom)
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
					s.address,
					uint32(1),
				}
			},
			func([]byte) {
				balance := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), utils.BaseDenom)
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

			var contract *vm.Contract
			contract, s.ctx = testutil.NewPrecompileContract(s.T(), s.ctx, s.address, s.precompile, tc.gas)

			// Sanity check to make sure the starting balance is always 5 EVMOS
			balance := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), utils.BaseDenom)
			s.Require().Equal(balance.Amount.BigInt(), big.NewInt(5e18))

			// Distribute rewards to the 2 validators, 1 EVMOS each
			for _, val := range s.validators {
				coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1e18)))
				s.app.DistrKeeper.AllocateTokensToValidator(s.ctx, val, sdk.NewDecCoinsFromCoins(coins...))
			}

			bz, err := s.precompile.ClaimRewards(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck(bz)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestFundCommunityPool() {
	method := s.precompile.Methods[distribution.FundCommunityPoolMethod]

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
			func([]byte) {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 2, 0),
		},
		{
			"fail - invalid depositor address",
			func() []interface{} {
				return []interface{}{
					nil,
					big.NewInt(1e18),
				}
			},
			func([]byte) {},
			200000,
			true,
			"invalid hex address address",
		},
		{
			"success - fund the community pool 1 EVMOS",
			func() []interface{} {
				return []interface{}{
					s.address,
					big.NewInt(1e18),
				}
			},
			func([]byte) {
				coins := s.app.DistrKeeper.GetFeePoolCommunityCoins(s.ctx)
				expectedAmount := new(big.Int).Mul(big.NewInt(1e18), new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(sdk.Precision)), nil))
				s.Require().Equal(expectedAmount, coins.AmountOf(utils.BaseDenom).BigInt())
				userBalance := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), utils.BaseDenom)
				s.Require().Equal(big.NewInt(4e18), userBalance.Amount.BigInt())
			},
			20000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			var contract *vm.Contract
			contract, s.ctx = testutil.NewPrecompileContract(s.T(), s.ctx, s.address, s.precompile, tc.gas)

			// Sanity check to make sure the starting balance is always 5 EVMOS
			balance := s.app.BankKeeper.GetBalance(s.ctx, s.address.Bytes(), utils.BaseDenom)
			s.Require().Equal(balance.Amount.BigInt(), big.NewInt(5e18))

			bz, err := s.precompile.FundCommunityPool(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck(bz)
			}
		})
	}
}
