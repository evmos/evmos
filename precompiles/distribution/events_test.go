package distribution_test

import (
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	cmn "github.com/evmos/evmos/v16/precompiles/common"
	"github.com/evmos/evmos/v16/precompiles/distribution"
	"github.com/evmos/evmos/v16/utils"
)

func (s *PrecompileTestSuite) TestSetWithdrawAddressEvent() {
	method := s.precompile.Methods[distribution.SetWithdrawAddressMethod]
	testCases := []struct {
		name        string
		malleate    func(operatorAddress string) []interface{}
		postCheck   func()
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"success - the correct event is emitted",
			func(operatorAddress string) []interface{} {
				return []interface{}{
					s.address,
					s.address.String(),
				}
			},
			func() {
				log := s.stateDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[distribution.EventTypeSetWithdrawAddress]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.ctx.BlockHeight()))

				// Check the fully unpacked event matches the one emitted
				var setWithdrawerAddrEvent distribution.EventSetWithdrawAddress
				err := cmn.UnpackLog(s.precompile.ABI, &setWithdrawerAddrEvent, distribution.EventTypeSetWithdrawAddress, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.address, setWithdrawerAddrEvent.Caller)
				s.Require().Equal(sdk.MustBech32ifyAddressBytes("evmos", s.address.Bytes()), setWithdrawerAddrEvent.WithdrawerAddress)
			},
			20000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.SetupTest()

		contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), tc.gas)
		s.ctx = s.ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
		initialGas := s.ctx.GasMeter().GasConsumed()
		s.Require().Zero(initialGas)

		_, err := s.precompile.SetWithdrawAddress(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate(s.validators[0].OperatorAddress))

		if tc.expError {
			s.Require().Error(err)
			s.Require().Contains(err.Error(), tc.errContains)
		} else {
			s.Require().NoError(err)
			tc.postCheck()
		}
	}
}

func (s *PrecompileTestSuite) TestWithdrawDelegatorRewardsEvent() {
	method := s.precompile.Methods[distribution.WithdrawDelegatorRewardsMethod]
	testCases := []struct {
		name        string
		malleate    func(operatorAddress string) []interface{}
		postCheck   func()
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"success - the correct event is emitted",
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
			func() {
				log := s.stateDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[distribution.EventTypeWithdrawDelegatorRewards]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.ctx.BlockHeight()))

				optAddr, err := sdk.ValAddressFromBech32(s.validators[0].OperatorAddress)
				s.Require().NoError(err)
				optHexAddr := common.BytesToAddress(optAddr)

				// Check the fully unpacked event matches the one emitted
				var delegatorRewards distribution.EventWithdrawDelegatorRewards
				err = cmn.UnpackLog(s.precompile.ABI, &delegatorRewards, distribution.EventTypeWithdrawDelegatorRewards, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.address, delegatorRewards.DelegatorAddress)
				s.Require().Equal(optHexAddr, delegatorRewards.ValidatorAddress)
				s.Require().Equal(big.NewInt(1000000000000000000), delegatorRewards.Amount)
			},
			20000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.SetupTest()

		contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), tc.gas)
		s.ctx = s.ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
		initialGas := s.ctx.GasMeter().GasConsumed()
		s.Require().Zero(initialGas)

		_, err := s.precompile.WithdrawDelegatorRewards(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate(s.validators[0].OperatorAddress))

		if tc.expError {
			s.Require().Error(err)
			s.Require().Contains(err.Error(), tc.errContains)
		} else {
			s.Require().NoError(err)
			tc.postCheck()
		}
	}
}

func (s *PrecompileTestSuite) TestWithdrawValidatorCommissionEvent() {
	method := s.precompile.Methods[distribution.WithdrawValidatorCommissionMethod]
	testCases := []struct {
		name        string
		malleate    func(operatorAddress string) []interface{}
		postCheck   func()
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"success - the correct event is emitted",
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
			func() {
				log := s.stateDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[distribution.EventTypeWithdrawValidatorCommission]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.ctx.BlockHeight()))

				// Check the fully unpacked event matches the one emitted
				var validatorRewards distribution.EventWithdrawValidatorRewards
				err := cmn.UnpackLog(s.precompile.ABI, &validatorRewards, distribution.EventTypeWithdrawValidatorCommission, *log)
				s.Require().NoError(err)
				s.Require().Equal(crypto.Keccak256Hash([]byte(s.validators[0].OperatorAddress)), validatorRewards.ValidatorAddress)
				s.Require().Equal(big.NewInt(100000000000000000), validatorRewards.Commission)
			},
			20000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.SetupTest()

		validatorAddress := common.BytesToAddress(s.validators[0].GetOperator().Bytes())
		contract := vm.NewContract(vm.AccountRef(validatorAddress), s.precompile, big.NewInt(0), tc.gas)
		s.ctx = s.ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
		initialGas := s.ctx.GasMeter().GasConsumed()
		s.Require().Zero(initialGas)

		_, err := s.precompile.WithdrawValidatorCommission(s.ctx, validatorAddress, contract, s.stateDB, &method, tc.malleate(s.validators[0].OperatorAddress))

		if tc.expError {
			s.Require().Error(err)
			s.Require().Contains(err.Error(), tc.errContains)
		} else {
			s.Require().NoError(err)
			tc.postCheck()
		}
	}
}

func (s *PrecompileTestSuite) TestClaimRewardsEvent() {
	testCases := []struct {
		name      string
		coins     sdk.Coins
		postCheck func()
	}{
		{
			"success",
			sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1e18))),
			func() {
				log := s.stateDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[distribution.EventTypeClaimRewards]
				s.Require().Equal(event.ID, common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.ctx.BlockHeight()))

				var claimRewardsEvent distribution.EventClaimRewards
				err := cmn.UnpackLog(s.precompile.ABI, &claimRewardsEvent, distribution.EventTypeClaimRewards, *log)
				s.Require().NoError(err)
				s.Require().Equal(common.BytesToAddress(s.address.Bytes()), claimRewardsEvent.DelegatorAddress)
				s.Require().Equal(big.NewInt(1e18), claimRewardsEvent.Amount)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			err := s.precompile.EmitClaimRewardsEvent(s.ctx, s.stateDB, s.address, tc.coins)
			s.Require().NoError(err)
			tc.postCheck()
		})
	}
}
