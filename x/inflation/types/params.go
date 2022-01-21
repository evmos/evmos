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
var (
	KeyMintDenom                          = []byte("MintDenom")
	KeyGenesisEpochProvisions             = []byte("GenesisEpochProvisions")
	KeyEpochIdentifier                    = []byte("EpochIdentifier")
	KeyReductionPeriodInEpochs            = []byte("ReductionPeriodInEpochs")
	KeyReductionFactor                    = []byte("ReductionFactor")
	KeyPoolAllocationRatio                = []byte("PoolAllocationRatio")
	KeyTeamVestingProvision               = []byte("TeamVestingProvision")
	KeyTeamVestingReceiver                = []byte("TeamVestingReceiver")
	KeyMintingRewardsAllocationStartEpoch = []byte("MintingRewardsAllocationStartEpoch")
)

// ParamTable for minting module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string,
	genesisEpochProvisions sdk.Dec,
	epochIdentifier string,
	reductionFactor sdk.Dec,
	reductionPeriodInEpochs int64,
	allocationProportions AllocationProportions,
	teamVestingProvision sdk.Coin,
	teamVestingReceiver string,
	mintingRewardsAllocationStartEpoch int64,
) Params {
	return Params{
		MintDenom:                          mintDenom,
		GenesisEpochProvisions:             genesisEpochProvisions,
		EpochIdentifier:                    epochIdentifier,
		ReductionPeriodInEpochs:            reductionPeriodInEpochs,
		ReductionFactor:                    reductionFactor,
		AllocationProportions:              allocationProportions,
		TeamVestingProvision:               teamVestingProvision,
		TeamVestingReceiver:                teamVestingReceiver,
		MintingRewardsAllocationStartEpoch: mintingRewardsAllocationStartEpoch,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:               sdk.DefaultBondDenom,
		GenesisEpochProvisions:  sdk.NewDec(5000000),
		EpochIdentifier:         "day",                    // 1 day
		ReductionPeriodInEpochs: 365,                      // 1 year
		ReductionFactor:         sdk.NewDecWithPrec(5, 1), // 0.5
		AllocationProportions: AllocationProportions{
			StakingRewards:  sdk.NewDecWithPrec(4, 1),  // 0.4
			TeamVesting:     sdk.NewDecWithPrec(25, 2), // 0.25
			UsageIncentives: sdk.NewDecWithPrec(25, 2), // 0.25
			CommunityPool:   sdk.NewDecWithPrec(1, 1),  // 0.1
		},
		TeamVestingProvision: sdk.NewCoin(
			sdk.DefaultBondDenom,
			sdk.NewInt(136986), // 200000000/(4*365)
		),
		TeamVestingReceiver:                "",
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
		paramtypes.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
		paramtypes.NewParamSetPair(KeyGenesisEpochProvisions, &p.GenesisEpochProvisions, validateGenesisEpochProvisions),
		paramtypes.NewParamSetPair(KeyEpochIdentifier, &p.EpochIdentifier, epochtypes.ValidateEpochIdentifierInterface),
		paramtypes.NewParamSetPair(KeyReductionPeriodInEpochs, &p.ReductionPeriodInEpochs, validateReductionPeriodInEpochs),
		paramtypes.NewParamSetPair(KeyReductionFactor, &p.ReductionFactor, validateReductionFactor),
		paramtypes.NewParamSetPair(KeyPoolAllocationRatio, &p.AllocationProportions, validateAllocationProportions),
		paramtypes.NewParamSetPair(KeyTeamVestingProvision, &p.TeamVestingProvision, validateTeamVestingProvision),
		paramtypes.NewParamSetPair(KeyTeamVestingReceiver, &p.TeamVestingReceiver, validateTeamVestingReceiver),
		paramtypes.NewParamSetPair(KeyMintingRewardsAllocationStartEpoch, &p.MintingRewardsAllocationStartEpoch, validateMintingRewardsAllocationStartEpoch),
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

func validateReductionPeriodInEpochs(i interface{}) error {
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

func validateAllocationProportions(i interface{}) error {
	v, ok := i.(AllocationProportions)
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
func validateTeamVestingReceiver(i interface{}) error {
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
	if err := validateGenesisEpochProvisions(p.GenesisEpochProvisions); err != nil {
		return err
	}
	if err := epochtypes.ValidateEpochIdentifierInterface(p.EpochIdentifier); err != nil {
		return err
	}
	if err := validateReductionPeriodInEpochs(p.ReductionPeriodInEpochs); err != nil {
		return err
	}
	if err := validateReductionFactor(p.ReductionFactor); err != nil {
		return err
	}
	if err := validateAllocationProportions(p.AllocationProportions); err != nil {
		return err
	}
	if err := validateTeamVestingProvision(p.TeamVestingProvision); err != nil {
		return err
	}
	if err := validateTeamVestingReceiver(p.TeamVestingReceiver); err != nil {
		return err
	}

	return validateMintingRewardsAllocationStartEpoch(p.MintingRewardsAllocationStartEpoch)
}
