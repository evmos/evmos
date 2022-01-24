package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"

	epochtypes "github.com/tharsis/evmos/x/epochs/types"
)

// Parameter store keys
const (
	KeyMintDenom = iota + 1
	KeyEpochIdentifier
	KeyEpochsPerPeriod
	KeyExponentialCalculation
	KeyInflationDistribution
	KeyTeamAddress
	KeyTeamVestingProvision
	KeyGenesisEpochProvisions
	KeyMintingRewardsAllocationStartEpoch
)

// ParamTable for minting module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string,
	epochIdentifier string,
	epochsPerPeriod int64,
	exponentialCalculation ExponentialCalculation,
	inflationDistribution InflationDistribution,
	teamAddress string,
	teamVestingProvision sdk.Dec,
	genesisEpochProvisions sdk.Dec,
	mintingRewardsAllocationStartEpoch int64,
) Params {
	return Params{
		MintDenom:                          mintDenom,
		EpochIdentifier:                    epochIdentifier,
		EpochsPerPeriod:                    epochsPerPeriod,
		ExponentialCalculation:             exponentialCalculation,
		InflationDistribution:              inflationDistribution,
		TeamAddress:                        teamAddress,
		TeamVestingProvision:               teamVestingProvision,
		GenesisEpochProvisions:             genesisEpochProvisions,
		MintingRewardsAllocationStartEpoch: mintingRewardsAllocationStartEpoch,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:       sdk.DefaultBondDenom,
		EpochIdentifier: "day", // 1 day
		EpochsPerPeriod: 365,   // 1 year
		ExponentialCalculation: ExponentialCalculation{
			A: sdk.NewDec(int64(300000000)),
			R: sdk.NewDecWithPrec(5, 1), // 0.5
			C: sdk.NewDec(int64(9375000)),
			B: sdk.ZeroDec(),
		},
		InflationDistribution: InflationDistribution{
			StakingRewards:  sdk.NewDecWithPrec(4, 1),  // 0.4
			TeamVesting:     sdk.NewDecWithPrec(25, 2), // 0.25
			UsageIncentives: sdk.NewDecWithPrec(25, 2), // 0.25
			CommunityPool:   sdk.NewDecWithPrec(1, 1),  // 0.1
		},
		TeamAddress:                        ModuleAddress.Hex(),
		TeamVestingProvision:               sdk.NewDec(int64(136986)), // 200000000/(4*365)
		GenesisEpochProvisions:             sdk.NewDec(5000000),
		MintingRewardsAllocationStartEpoch: 0,
	}
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair([]byte{KeyMintDenom}, &p.MintDenom, validateMintDenom),
		paramtypes.NewParamSetPair([]byte{KeyEpochIdentifier}, &p.EpochIdentifier, epochtypes.ValidateEpochIdentifierInterface),
		paramtypes.NewParamSetPair([]byte{KeyEpochsPerPeriod}, &p.EpochsPerPeriod, validateEpochsPerPeriod),
		paramtypes.NewParamSetPair([]byte{KeyExponentialCalculation}, &p.ExponentialCalculation, validateExponentialCalculation),
		paramtypes.NewParamSetPair([]byte{KeyInflationDistribution}, &p.InflationDistribution, validateInflationDistribution),
		paramtypes.NewParamSetPair([]byte{KeyTeamAddress}, &p.TeamAddress, validateTeamAddress),
		paramtypes.NewParamSetPair([]byte{KeyTeamVestingProvision}, &p.TeamVestingProvision, validateTeamVestingProvision),
		paramtypes.NewParamSetPair([]byte{KeyGenesisEpochProvisions}, &p.GenesisEpochProvisions, validateGenesisEpochProvisions),
		paramtypes.NewParamSetPair([]byte{KeyMintingRewardsAllocationStartEpoch}, &p.MintingRewardsAllocationStartEpoch, validateMintingRewardsAllocationStartEpoch),
	}
}

func validateMintDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("mint denom cannot be blank")
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func validateGenesisEpochProvisions(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.LT(sdk.ZeroDec()) {
		return fmt.Errorf("genesis epoch provision must be non-negative")
	}

	return nil
}

func validateEpochsPerPeriod(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("max validators must be positive: %d", v)
	}

	return nil
}

func validateReductionFactor(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.GT(sdk.NewDec(1)) {
		return fmt.Errorf("reduction factor cannot be greater than 1")
	}

	if v.IsNegative() {
		return fmt.Errorf("reduction factor cannot be negative")
	}

	return nil
}
func validateExponentialCalculation(i interface{}) error {
	// TODO
	return nil
}

func validateInflationDistribution(i interface{}) error {
	v, ok := i.(InflationDistribution)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.StakingRewards.IsNegative() {
		return errors.New("staking allocation ratio must not be negative")
	}

	if v.UsageIncentives.IsNegative() {
		return errors.New("pool incentives allocation ratio must not be negative")
	}

	if v.TeamVesting.IsNegative() {
		return errors.New("developer rewards allocation ratio must not be negative")
	}

	// TODO: Maybe we should allow this :joy:, lets you burn osmo from community pool
	// for new chains
	if v.CommunityPool.IsNegative() {
		return errors.New("community pool allocation ratio must not be negative")
	}

	totalProportions := v.StakingRewards.Add(v.UsageIncentives).Add(v.TeamVesting).Add(v.CommunityPool)

	if !totalProportions.Equal(sdk.NewDec(1)) {
		return errors.New("total distributions ratio should be 1")
	}

	return nil
}

// TODO
func validateTeamVestingProvision(i interface{}) error {
	return nil
}

// TODO
func validateTeamAddress(i interface{}) error {
	return nil
}

func validateMintingRewardsAllocationStartEpoch(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v < 0 {
		return fmt.Errorf("start epoch must be non-negative")
	}

	return nil
}

func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := epochtypes.ValidateEpochIdentifierInterface(p.EpochIdentifier); err != nil {
		return err
	}
	if err := validateEpochsPerPeriod(p.EpochsPerPeriod); err != nil {
		return err
	}
	if err := validateExponentialCalculation(p.ExponentialCalculation); err != nil {
		return err
	}
	if err := validateInflationDistribution(p.InflationDistribution); err != nil {
		return err
	}
	if err := validateTeamAddress(p.TeamAddress); err != nil {
		return err
	}
	if err := validateTeamVestingProvision(p.TeamVestingProvision); err != nil {
		return err
	}

	if err := validateGenesisEpochProvisions(p.GenesisEpochProvisions); err != nil {
		return err
	}
	return validateMintingRewardsAllocationStartEpoch(p.MintingRewardsAllocationStartEpoch)
}
