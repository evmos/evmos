package distribution_test

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v20/x/evm/core/vm"

	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/precompiles/distribution"
	testutiltx "github.com/evmos/evmos/v20/testutil/tx"
)

var expValAmount int64 = 1

type distrTestCases struct {
	name        string
	malleate    func() []interface{}
	postCheck   func(bz []byte)
	gas         uint64
	expErr      bool
	errContains string
}

var baseTestCases = []distrTestCases{
	{
		"fail - empty input args",
		func() []interface{} {
			return []interface{}{}
		},
		func([]byte) {},
		100000,
		true,
		"invalid number of arguments",
	},
	{
		"fail - invalid validator address",
		func() []interface{} {
			return []interface{}{
				"invalid",
			}
		},
		func([]byte) {},
		100000,
		true,
		"invalid bech32 string",
	},
}

func (s *PrecompileTestSuite) TestValidatorDistributionInfo() {
	var ctx sdk.Context
	method := s.precompile.Methods[distribution.ValidatorDistributionInfoMethod]

	testCases := []distrTestCases{
		{
			"fail - nonexistent validator address",
			func() []interface{} {
				pv := mock.NewPV()
				pk, err := pv.GetPubKey()
				s.Require().NoError(err)
				return []interface{}{
					sdk.ValAddress(pk.Address().Bytes()).String(),
				}
			},
			func([]byte) {},
			100000,
			true,
			"validator does not exist",
		},
		{
			"fail - existent validator but without self delegation",
			func() []interface{} {
				return []interface{}{
					s.network.GetValidators()[0].OperatorAddress,
				}
			},
			func([]byte) {},
			100000,
			true,
			"no delegation for (address, validator) tuple",
		},
		{
			"success",
			func() []interface{} {
				valAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].GetOperator())
				s.Require().NoError(err)
				s.Require().NoError(err)

				// fund account for self delegation
				amt := math.NewInt(1)
				err = s.fundAccountWithBaseDenom(ctx, valAddr.Bytes(), amt)
				s.Require().NoError(err)

				// make a self delegation
				_, err = s.network.App.StakingKeeper.Delegate(ctx, valAddr.Bytes(), amt, stakingtypes.Unspecified, s.network.GetValidators()[0], true)
				s.Require().NoError(err)
				return []interface{}{
					s.network.GetValidators()[0].OperatorAddress,
				}
			},
			func(bz []byte) {
				var out distribution.ValidatorDistributionInfoOutput
				err := s.precompile.UnpackIntoInterface(&out, distribution.ValidatorDistributionInfoMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)

				valAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].GetOperator())
				s.Require().NoError(err)

				s.Require().Equal(sdk.AccAddress(valAddr.Bytes()).String(), out.DistributionInfo.OperatorAddress)
				s.Require().Equal(0, len(out.DistributionInfo.Commission))
				s.Require().Equal(0, len(out.DistributionInfo.SelfBondRewards))
			},
			100000,
			false,
			"",
		},
	}
	testCases = append(testCases, baseTestCases...)

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()
			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)

			bz, err := s.precompile.ValidatorDistributionInfo(ctx, contract, &method, tc.malleate())

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)
				tc.postCheck(bz)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestValidatorOutstandingRewards() {
	var ctx sdk.Context
	method := s.precompile.Methods[distribution.ValidatorOutstandingRewardsMethod]

	testCases := []distrTestCases{
		{
			"fail - nonexistent validator address",
			func() []interface{} {
				pv := mock.NewPV()
				pk, err := pv.GetPubKey()
				s.Require().NoError(err)
				return []interface{}{
					sdk.ValAddress(pk.Address().Bytes()).String(),
				}
			},
			func(bz []byte) {
				var out []sdk.DecCoin
				err := s.precompile.UnpackIntoInterface(&out, distribution.ValidatorOutstandingRewardsMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(0, len(out))
			},
			100000,
			true,
			"validator does not exist",
		},
		{
			"success - existent validator, no outstanding rewards",
			func() []interface{} {
				return []interface{}{
					s.network.GetValidators()[0].OperatorAddress,
				}
			},
			func(bz []byte) {
				var out []sdk.DecCoin
				err := s.precompile.UnpackIntoInterface(&out, distribution.ValidatorOutstandingRewardsMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(0, len(out))
			},
			100000,
			false,
			"",
		},
		{
			"success - with outstanding rewards",
			func() []interface{} {
				valRewards := sdk.DecCoins{sdk.NewDecCoinFromDec(s.bondDenom, math.LegacyNewDec(1))}
				// set outstanding rewards
				valAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].GetOperator())
				s.Require().NoError(err)

				err = s.network.App.DistrKeeper.SetValidatorOutstandingRewards(ctx, valAddr, types.ValidatorOutstandingRewards{Rewards: valRewards})
				s.Require().NoError(err)

				return []interface{}{
					s.network.GetValidators()[0].OperatorAddress,
				}
			},
			func(bz []byte) {
				var out []cmn.DecCoin
				err := s.precompile.UnpackIntoInterface(&out, distribution.ValidatorOutstandingRewardsMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(1, len(out))
				s.Require().Equal(uint8(18), out[0].Precision)
				s.Require().Equal(s.bondDenom, out[0].Denom)
				s.Require().Equal(expValAmount, out[0].Amount.Int64())
			},
			100000,
			false,
			"",
		},
	}
	testCases = append(testCases, baseTestCases...)

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()
			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)

			bz, err := s.precompile.ValidatorOutstandingRewards(ctx, contract, &method, tc.malleate())

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)
				tc.postCheck(bz)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestValidatorCommission() {
	var ctx sdk.Context
	method := s.precompile.Methods[distribution.ValidatorCommissionMethod]

	testCases := []distrTestCases{
		{
			"fail - nonexistent validator address",
			func() []interface{} {
				pv := mock.NewPV()
				pk, err := pv.GetPubKey()
				s.Require().NoError(err)
				return []interface{}{
					sdk.ValAddress(pk.Address().Bytes()).String(),
				}
			},
			func(bz []byte) {
				var out []sdk.DecCoin
				err := s.precompile.UnpackIntoInterface(&out, distribution.ValidatorCommissionMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(0, len(out))
			},
			100000,
			true,
			"validator does not exist",
		},
		{
			"success - existent validator, no accumulated commission",
			func() []interface{} {
				return []interface{}{
					s.network.GetValidators()[0].OperatorAddress,
				}
			},
			func(bz []byte) {
				var out []sdk.DecCoin
				err := s.precompile.UnpackIntoInterface(&out, distribution.ValidatorCommissionMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(0, len(out))
			},
			100000,
			false,
			"",
		},
		{
			"success - with accumulated commission",
			func() []interface{} {
				commAmt := math.LegacyNewDec(1)
				validator := s.network.GetValidators()[0]
				valAddr, err := sdk.ValAddressFromBech32(validator.GetOperator())
				s.Require().NoError(err)
				valCommission := sdk.DecCoins{sdk.NewDecCoinFromDec(s.bondDenom, commAmt)}
				err = s.network.App.DistrKeeper.SetValidatorAccumulatedCommission(ctx, valAddr, types.ValidatorAccumulatedCommission{Commission: valCommission})
				s.Require().NoError(err)

				// set distribution module account balance which pays out the commission
				coins := sdk.NewCoins(sdk.NewCoin(s.bondDenom, commAmt.RoundInt()))
				err = s.mintCoinsForDistrMod(ctx, coins)
				s.Require().NoError(err)

				return []interface{}{
					validator.OperatorAddress,
				}
			},
			func(bz []byte) {
				var out []cmn.DecCoin
				err := s.precompile.UnpackIntoInterface(&out, distribution.ValidatorCommissionMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(1, len(out))
				s.Require().Equal(uint8(18), out[0].Precision)
				s.Require().Equal(s.bondDenom, out[0].Denom)
				s.Require().Equal(expValAmount, out[0].Amount.Int64())
			},
			100000,
			false,
			"",
		},
	}
	testCases = append(testCases, baseTestCases...)

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()
			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)

			bz, err := s.precompile.ValidatorCommission(ctx, contract, &method, tc.malleate())

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)
				tc.postCheck(bz)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestValidatorSlashes() {
	var ctx sdk.Context
	method := s.precompile.Methods[distribution.ValidatorSlashesMethod]

	testCases := []distrTestCases{
		{
			"fail - invalid validator address",
			func() []interface{} {
				return []interface{}{
					"invalid", uint64(1), uint64(5), query.PageRequest{},
				}
			},
			func([]byte) {
			},
			100000,
			true,
			"invalid validator address",
		},
		{
			"fail - invalid starting height type",
			func() []interface{} {
				return []interface{}{
					s.network.GetValidators()[0].OperatorAddress,
					int64(1), uint64(5),
					query.PageRequest{},
				}
			},
			func([]byte) {
			},
			100000,
			true,
			"invalid type for startingHeight: expected uint64, received int64",
		},
		{
			"fail - starting height greater than ending height",
			func() []interface{} {
				return []interface{}{
					s.network.GetValidators()[0].OperatorAddress,
					uint64(6), uint64(5),
					query.PageRequest{},
				}
			},
			func([]byte) {
			},
			100000,
			true,
			"starting height greater than ending height",
		},
		{
			"success - nonexistent validator address",
			func() []interface{} {
				pv := mock.NewPV()
				pk, err := pv.GetPubKey()
				s.Require().NoError(err)
				return []interface{}{
					sdk.ValAddress(pk.Address().Bytes()).String(),
					uint64(1),
					uint64(5),
					query.PageRequest{},
				}
			},
			func(bz []byte) {
				var out distribution.ValidatorSlashesOutput
				err := s.precompile.UnpackIntoInterface(&out, distribution.ValidatorSlashesMethod, bz)
				s.Require().NoError(err, "failed to unpack output")
				s.Require().Equal(0, len(out.Slashes))
				s.Require().Equal(uint64(0), out.PageResponse.Total)
			},
			100000,
			false,
			"",
		},
		{
			"success - existent validator, no slashes",
			func() []interface{} {
				return []interface{}{
					s.network.GetValidators()[0].OperatorAddress,
					uint64(1),
					uint64(5),
					query.PageRequest{},
				}
			},
			func(bz []byte) {
				var out distribution.ValidatorSlashesOutput
				err := s.precompile.UnpackIntoInterface(&out, distribution.ValidatorSlashesMethod, bz)
				s.Require().NoError(err, "failed to unpack output")
				s.Require().Equal(0, len(out.Slashes))
				s.Require().Equal(uint64(0), out.PageResponse.Total)
			},
			100000,
			false,
			"",
		},
		{
			"success - with slashes",
			func() []interface{} {
				valAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].GetOperator())
				s.Require().NoError(err)
				err = s.network.App.DistrKeeper.SetValidatorSlashEvent(ctx, valAddr, 2, 1, types.ValidatorSlashEvent{ValidatorPeriod: 1, Fraction: math.LegacyNewDec(5)})
				s.Require().NoError(err)
				return []interface{}{
					s.network.GetValidators()[0].OperatorAddress,
					uint64(1), uint64(5),
					query.PageRequest{},
				}
			},
			func(bz []byte) {
				var out distribution.ValidatorSlashesOutput
				err := s.precompile.UnpackIntoInterface(&out, distribution.ValidatorSlashesMethod, bz)
				s.Require().NoError(err, "failed to unpack output")
				s.Require().Equal(1, len(out.Slashes))
				s.Require().Equal(math.LegacyNewDec(5).BigInt(), out.Slashes[0].Fraction.Value)
				s.Require().Equal(uint64(1), out.Slashes[0].ValidatorPeriod)
				s.Require().Equal(uint64(1), out.PageResponse.Total)
			},
			100000,
			false,
			"",
		},
		{
			"success - with slashes w/pagination",
			func() []interface{} {
				valAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].GetOperator())
				s.Require().NoError(err)
				err = s.network.App.DistrKeeper.SetValidatorSlashEvent(ctx, valAddr, 2, 1, types.ValidatorSlashEvent{ValidatorPeriod: 1, Fraction: math.LegacyNewDec(5)})
				s.Require().NoError(err)
				return []interface{}{
					s.network.GetValidators()[0].OperatorAddress,
					uint64(1),
					uint64(5),
					query.PageRequest{Limit: 1, CountTotal: true},
				}
			},
			func(bz []byte) {
				var out distribution.ValidatorSlashesOutput
				err := s.precompile.UnpackIntoInterface(&out, distribution.ValidatorSlashesMethod, bz)
				s.Require().NoError(err, "failed to unpack output")
				s.Require().Equal(1, len(out.Slashes))
				s.Require().Equal(math.LegacyNewDec(5).BigInt(), out.Slashes[0].Fraction.Value)
				s.Require().Equal(uint64(1), out.Slashes[0].ValidatorPeriod)
				s.Require().Equal(uint64(1), out.PageResponse.Total)
			},
			100000,
			false,
			"",
		},
	}
	testCases = append(testCases, baseTestCases[0])

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()
			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)

			bz, err := s.precompile.ValidatorSlashes(ctx, contract, &method, tc.malleate())

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)
				tc.postCheck(bz)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestDelegationRewards() {
	var (
		ctx sdk.Context
		err error
	)
	method := s.precompile.Methods[distribution.DelegationRewardsMethod]

	testCases := []distrTestCases{
		{
			"fail - invalid validator address",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					"invalid",
				}
			},
			func([]byte) {},
			100000,
			true,
			"invalid bech32 string",
		},
		{
			"fail - nonexistent validator address",
			func() []interface{} {
				pv := mock.NewPV()
				pk, err := pv.GetPubKey()
				s.Require().NoError(err)
				return []interface{}{
					s.keyring.GetAddr(0),
					sdk.ValAddress(pk.Address().Bytes()).String(),
				}
			},
			func([]byte) {},
			100000,
			true,
			"validator does not exist",
		},
		{
			"fail - existent validator, no delegation",
			func() []interface{} {
				newAddr, _ := testutiltx.NewAddrKey()
				return []interface{}{
					newAddr,
					s.network.GetValidators()[0].OperatorAddress,
				}
			},
			func([]byte) {},
			100000,
			true,
			"no delegation for (address, validator) tuple",
		},
		{
			"success - existent validator & delegation, but no rewards",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					s.network.GetValidators()[0].OperatorAddress,
				}
			},
			func(bz []byte) {
				var out []cmn.DecCoin
				err := s.precompile.UnpackIntoInterface(&out, distribution.DelegationRewardsMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(0, len(out))
			},
			100000,
			false,
			"",
		},
		{
			"success - with rewards",
			func() []interface{} {
				ctx, err = s.prepareStakingRewards(ctx, stakingRewards{s.keyring.GetAddr(0).Bytes(), s.network.GetValidators()[0], testRewardsAmt})
				s.Require().NoError(err, "failed to prepare staking rewards", err)
				return []interface{}{
					s.keyring.GetAddr(0),
					s.network.GetValidators()[0].OperatorAddress,
				}
			},
			func(bz []byte) {
				var out []cmn.DecCoin
				err := s.precompile.UnpackIntoInterface(&out, distribution.DelegationRewardsMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(1, len(out))
				s.Require().Equal(uint8(18), out[0].Precision)
				s.Require().Equal(s.bondDenom, out[0].Denom)
				s.Require().Equal(expRewardsAmt.Int64(), out[0].Amount.Int64())
			},
			100000,
			false,
			"",
		},
	}
	testCases = append(testCases, baseTestCases[0])

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()
			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)

			args := tc.malleate()
			bz, err := s.precompile.DelegationRewards(ctx, contract, &method, args)

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)
				tc.postCheck(bz)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestDelegationTotalRewards() {
	var (
		ctx sdk.Context
		err error
	)
	method := s.precompile.Methods[distribution.DelegationTotalRewardsMethod]

	testCases := []distrTestCases{
		{
			"fail - invalid delegator address",
			func() []interface{} {
				return []interface{}{
					"invalid",
				}
			},
			func([]byte) {},
			100000,
			true,
			fmt.Sprintf(cmn.ErrInvalidDelegator, "invalid"),
		},
		{
			"success - no delegations",
			func() []interface{} {
				newAddr, _ := testutiltx.NewAddrKey()
				return []interface{}{
					newAddr,
				}
			},
			func(bz []byte) {
				var out distribution.DelegationTotalRewardsOutput
				err := s.precompile.UnpackIntoInterface(&out, distribution.DelegationTotalRewardsMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(0, len(out.Rewards))
				s.Require().Equal(0, len(out.Total))
			},
			100000,
			false,
			"",
		},
		{
			"success - existent validator & delegation, but no rewards",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
				}
			},
			func(bz []byte) {
				var out distribution.DelegationTotalRewardsOutput
				err := s.precompile.UnpackIntoInterface(&out, distribution.DelegationTotalRewardsMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)

				validatorsCount := len(s.network.GetValidators())
				s.Require().Equal(validatorsCount, len(out.Rewards))

				// no rewards
				s.Require().Equal(0, len(out.Rewards[0].Reward))
				s.Require().Equal(0, len(out.Rewards[1].Reward))
				s.Require().Equal(0, len(out.Rewards[2].Reward))
				s.Require().Equal(0, len(out.Total))
			},
			100000,
			false,
			"",
		},
		{
			"success - with rewards",
			func() []interface{} {
				ctx, err = s.prepareStakingRewards(ctx, stakingRewards{s.keyring.GetAccAddr(0), s.network.GetValidators()[0], testRewardsAmt})
				s.Require().NoError(err, "failed to prepare staking rewards", err)

				return []interface{}{
					s.keyring.GetAddr(0),
				}
			},
			func(bz []byte) {
				var (
					out distribution.DelegationTotalRewardsOutput
					i   int
				)
				err := s.precompile.UnpackIntoInterface(&out, distribution.DelegationTotalRewardsMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)

				validators := s.network.GetValidators()
				valWithRewards := validators[0]
				validatorsCount := len(s.network.GetValidators())
				s.Require().Equal(validatorsCount, len(out.Rewards))

				// the response order may change
				for index, or := range out.Rewards {
					if or.ValidatorAddress == valWithRewards.OperatorAddress {
						i = index
					} else {
						s.Require().Equal(0, len(out.Rewards[index].Reward))
					}
				}

				// only validator[i] has rewards
				s.Require().Equal(1, len(out.Rewards[i].Reward))
				s.Require().Equal(s.bondDenom, out.Rewards[i].Reward[0].Denom)
				s.Require().Equal(uint8(math.LegacyPrecision), out.Rewards[i].Reward[0].Precision)
				s.Require().Equal(expRewardsAmt.Int64(), out.Rewards[i].Reward[0].Amount.Int64())

				s.Require().Equal(1, len(out.Total))
				s.Require().Equal(expRewardsAmt.Int64(), out.Total[0].Amount.Int64())
			},
			100000,
			false,
			"",
		},
	}
	testCases = append(testCases, baseTestCases[0])

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()

			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)

			args := tc.malleate()
			bz, err := s.precompile.DelegationTotalRewards(ctx, contract, &method, args)

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)
				tc.postCheck(bz)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestDelegatorValidators() {
	var ctx sdk.Context
	method := s.precompile.Methods[distribution.DelegatorValidatorsMethod]

	testCases := []distrTestCases{
		{
			"fail - invalid delegator address",
			func() []interface{} {
				return []interface{}{
					"invalid",
				}
			},
			func([]byte) {},
			100000,
			true,
			fmt.Sprintf(cmn.ErrInvalidDelegator, "invalid"),
		},
		{
			"success - no delegations",
			func() []interface{} {
				newAddr, _ := testutiltx.NewAddrKey()
				return []interface{}{
					newAddr,
				}
			},
			func(bz []byte) {
				var out []string
				err := s.precompile.UnpackIntoInterface(&out, distribution.DelegatorValidatorsMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(0, len(out))
			},
			100000,
			false,
			"",
		},
		{
			"success - existent delegations",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
				}
			},
			func(bz []byte) {
				var out []string
				err := s.precompile.UnpackIntoInterface(&out, distribution.DelegatorValidatorsMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(3, len(out))
				for _, val := range s.network.GetValidators() {
					s.Require().Contains(
						out,
						val.OperatorAddress,
						"expected operator address %q to be in output",
						val.OperatorAddress,
					)
				}
			},
			100000,
			false,
			"",
		},
	}
	testCases = append(testCases, baseTestCases[0])

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()
			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)

			bz, err := s.precompile.DelegatorValidators(ctx, contract, &method, tc.malleate())

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)
				tc.postCheck(bz)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestDelegatorWithdrawAddress() {
	var ctx sdk.Context
	method := s.precompile.Methods[distribution.DelegatorWithdrawAddressMethod]

	testCases := []distrTestCases{
		{
			"fail - invalid delegator address",
			func() []interface{} {
				return []interface{}{
					"invalid",
				}
			},
			func([]byte) {},
			100000,
			true,
			fmt.Sprintf(cmn.ErrInvalidDelegator, "invalid"),
		},
		{
			"success - withdraw address same as delegator address",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
				}
			},
			func(bz []byte) {
				var out string
				err := s.precompile.UnpackIntoInterface(&out, distribution.DelegatorWithdrawAddressMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(sdk.AccAddress(s.keyring.GetAddr(0).Bytes()).String(), out)
			},
			100000,
			false,
			"",
		},
	}
	testCases = append(testCases, baseTestCases[0])

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()
			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)

			bz, err := s.precompile.DelegatorWithdrawAddress(ctx, contract, &method, tc.malleate())

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)
				tc.postCheck(bz)
			}
		})
	}
}
