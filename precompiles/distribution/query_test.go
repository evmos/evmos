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
	"github.com/ethereum/go-ethereum/core/vm"

	"github.com/evmos/evmos/v17/testutil"
	testutiltx "github.com/evmos/evmos/v17/testutil/tx"

	cmn "github.com/evmos/evmos/v17/precompiles/common"
	"github.com/evmos/evmos/v17/precompiles/distribution"
)

var (
	expDelegationRewards int64 = 2000000000000000000
	expValAmount         int64 = 1
	rewards, _                 = math.NewIntFromString("1000000000000000000")
)

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
					s.validators[0].OperatorAddress,
				}
			},
			func([]byte) {},
			100000,
			true,
			"delegation does not exist",
		},
		{
			"success",
			func() []interface{} {
				addr := sdk.AccAddress(s.validators[0].GetOperator())
				// fund del account to make self-delegation
				err := testutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, addr, 10)
				s.Require().NoError(err)
				// make a self delegation
				_, err = s.app.StakingKeeper.Delegate(s.ctx, addr, math.NewInt(1), stakingtypes.Unspecified, s.validators[0], true)
				s.Require().NoError(err)
				return []interface{}{
					s.validators[0].OperatorAddress,
				}
			},
			func(bz []byte) {
				var out distribution.ValidatorDistributionInfoOutput
				err := s.precompile.UnpackIntoInterface(&out, distribution.ValidatorDistributionInfoMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)
				expAddr := sdk.AccAddress(s.validators[0].GetOperator())
				s.Require().Equal(expAddr.String(), out.DistributionInfo.OperatorAddress)
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
			contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), tc.gas)

			bz, err := s.precompile.ValidatorDistributionInfo(s.ctx, contract, &method, tc.malleate())

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

func (s *PrecompileTestSuite) TestValidatorOutstandingRewards() { //nolint:dupl
	method := s.precompile.Methods[distribution.ValidatorOutstandingRewardsMethod]

	testCases := []distrTestCases{
		{
			"success - nonexistent validator address",
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
			false,
			"",
		},
		{
			"success - existent validator, no outstanding rewards",
			func() []interface{} {
				return []interface{}{
					s.validators[0].OperatorAddress,
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
				s.app.DistrKeeper.SetValidatorOutstandingRewards(s.ctx, s.validators[0].GetOperator(), types.ValidatorOutstandingRewards{Rewards: valRewards})
				return []interface{}{
					s.validators[0].OperatorAddress,
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
			contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), tc.gas)

			bz, err := s.precompile.ValidatorOutstandingRewards(s.ctx, contract, &method, tc.malleate())

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

func (s *PrecompileTestSuite) TestValidatorCommission() { //nolint:dupl
	method := s.precompile.Methods[distribution.ValidatorCommissionMethod]

	testCases := []distrTestCases{
		{
			"success - nonexistent validator address",
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
			false,
			"",
		},
		{
			"success - existent validator, no accumulated commission",
			func() []interface{} {
				return []interface{}{
					s.validators[0].OperatorAddress,
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
				valCommission := sdk.DecCoins{sdk.NewDecCoinFromDec(s.bondDenom, math.LegacyNewDec(1))}
				s.app.DistrKeeper.SetValidatorAccumulatedCommission(s.ctx, s.validators[0].GetOperator(), types.ValidatorAccumulatedCommission{Commission: valCommission})
				return []interface{}{
					s.validators[0].OperatorAddress,
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
			contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), tc.gas)

			bz, err := s.precompile.ValidatorCommission(s.ctx, contract, &method, tc.malleate())

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
					s.validators[0].OperatorAddress,
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
					s.validators[0].OperatorAddress,
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
					s.validators[0].OperatorAddress,
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
				s.app.DistrKeeper.SetValidatorSlashEvent(s.ctx, s.validators[0].GetOperator(), 2, 1, types.ValidatorSlashEvent{ValidatorPeriod: 1, Fraction: math.LegacyNewDec(5)})
				return []interface{}{
					s.validators[0].OperatorAddress,
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
				s.app.DistrKeeper.SetValidatorSlashEvent(s.ctx, s.validators[0].GetOperator(), 2, 1, types.ValidatorSlashEvent{ValidatorPeriod: 1, Fraction: math.LegacyNewDec(5)})
				return []interface{}{
					s.validators[0].OperatorAddress,
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
			contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), tc.gas)

			bz, err := s.precompile.ValidatorSlashes(s.ctx, contract, &method, tc.malleate())

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
	method := s.precompile.Methods[distribution.DelegationRewardsMethod]

	testCases := []distrTestCases{
		{
			"fail - invalid validator address",
			func() []interface{} {
				return []interface{}{
					s.address,
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
					s.address,
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
					s.validators[0].OperatorAddress,
				}
			},
			func([]byte) {},
			100000,
			true,
			"delegation does not exist",
		},
		{
			"success - existent validator & delegation, but no rewards",
			func() []interface{} {
				return []interface{}{
					s.address,
					s.validators[0].OperatorAddress,
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
				s.prepareStakingRewards(stakingRewards{s.address.Bytes(), s.validators[0], rewards})
				return []interface{}{
					s.address,
					s.validators[0].OperatorAddress,
				}
			},
			func(bz []byte) {
				var out []cmn.DecCoin
				err := s.precompile.UnpackIntoInterface(&out, distribution.DelegationRewardsMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(1, len(out))
				s.Require().Equal(uint8(18), out[0].Precision)
				s.Require().Equal(s.bondDenom, out[0].Denom)
				s.Require().Equal(expDelegationRewards, out[0].Amount.Int64())
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
			contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), tc.gas)

			bz, err := s.precompile.DelegationRewards(s.ctx, contract, &method, tc.malleate())

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
					s.address,
				}
			},
			func(bz []byte) {
				var out distribution.DelegationTotalRewardsOutput
				err := s.precompile.UnpackIntoInterface(&out, distribution.DelegationTotalRewardsMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(2, len(out.Rewards))
				// the response order may change
				if out.Rewards[0].ValidatorAddress == s.validators[0].OperatorAddress {
					s.Require().Equal(s.validators[0].OperatorAddress, out.Rewards[0].ValidatorAddress)
					s.Require().Equal(s.validators[1].OperatorAddress, out.Rewards[1].ValidatorAddress)
				} else {
					s.Require().Equal(s.validators[1].OperatorAddress, out.Rewards[0].ValidatorAddress)
					s.Require().Equal(s.validators[0].OperatorAddress, out.Rewards[1].ValidatorAddress)
				}
				// no rewards
				s.Require().Equal(0, len(out.Rewards[0].Reward))
				s.Require().Equal(0, len(out.Rewards[1].Reward))
				s.Require().Equal(0, len(out.Total))
			},
			100000,
			false,
			"",
		},
		{
			"success - with rewards",
			func() []interface{} {
				s.prepareStakingRewards(stakingRewards{s.address.Bytes(), s.validators[0], rewards})
				return []interface{}{
					s.address,
				}
			},
			func(bz []byte) {
				var (
					out distribution.DelegationTotalRewardsOutput
					i   int
				)
				err := s.precompile.UnpackIntoInterface(&out, distribution.DelegationTotalRewardsMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(2, len(out.Rewards))

				// the response order may change
				if out.Rewards[0].ValidatorAddress == s.validators[0].OperatorAddress {
					s.Require().Equal(s.validators[0].OperatorAddress, out.Rewards[0].ValidatorAddress)
					s.Require().Equal(s.validators[1].OperatorAddress, out.Rewards[1].ValidatorAddress)
					s.Require().Equal(0, len(out.Rewards[1].Reward))
				} else {
					i = 1
					s.Require().Equal(s.validators[0].OperatorAddress, out.Rewards[1].ValidatorAddress)
					s.Require().Equal(s.validators[1].OperatorAddress, out.Rewards[0].ValidatorAddress)
					s.Require().Equal(0, len(out.Rewards[0].Reward))
				}

				// only validator[i] has rewards
				s.Require().Equal(1, len(out.Rewards[i].Reward))
				s.Require().Equal(s.bondDenom, out.Rewards[i].Reward[0].Denom)
				s.Require().Equal(uint8(math.LegacyPrecision), out.Rewards[i].Reward[0].Precision)
				s.Require().Equal(expDelegationRewards, out.Rewards[i].Reward[0].Amount.Int64())

				s.Require().Equal(1, len(out.Total))
				s.Require().Equal(expDelegationRewards, out.Total[0].Amount.Int64())
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
			contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), tc.gas)

			bz, err := s.precompile.DelegationTotalRewards(s.ctx, contract, &method, tc.malleate())

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
					s.address,
				}
			},
			func(bz []byte) {
				var out []string
				err := s.precompile.UnpackIntoInterface(&out, distribution.DelegatorValidatorsMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(2, len(out))
				// the order may change
				if out[0] == s.validators[0].OperatorAddress {
					s.Require().Equal(s.validators[0].OperatorAddress, out[0])
					s.Require().Equal(s.validators[1].OperatorAddress, out[1])
				} else {
					s.Require().Equal(s.validators[1].OperatorAddress, out[0])
					s.Require().Equal(s.validators[0].OperatorAddress, out[1])
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
			contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), tc.gas)

			bz, err := s.precompile.DelegatorValidators(s.ctx, contract, &method, tc.malleate())

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
					s.address,
				}
			},
			func(bz []byte) {
				var out string
				err := s.precompile.UnpackIntoInterface(&out, distribution.DelegatorWithdrawAddressMethod, bz)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(sdk.AccAddress(s.address.Bytes()).String(), out)
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
			contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), tc.gas)

			bz, err := s.precompile.DelegatorWithdrawAddress(s.ctx, contract, &method, tc.malleate())

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
