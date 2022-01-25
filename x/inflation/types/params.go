package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ethermint "github.com/tharsis/ethermint/types"
	"gopkg.in/yaml.v2"

	epochtypes "github.com/tharsis/evmos/x/epochs/types"
)

// Parameter store keys
var (
	KeyMintDenom              = []byte("KeyMintDenom")
	KeyEpochIdentifier        = []byte("KeyEpochIdentifier")
	KeyEpochsPerPeriod        = []byte("KeyEpochsPerPeriod")
	KeyExponentialCalculation = []byte("KeyExponentialCalculation")
	KeyInflationDistribution  = []byte("KeyInflationDistribution")
	KeyTeamAddress            = []byte("KeyTeamAddress")
	KeyTeamVestingProvision   = []byte("KeyTeamVestingProvision")
)

// ParamTable for inflation module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string,
	epochIdentifier string,
	epochsPerPeriod int64,
	exponentialCalculation ExponentialCalculation,
	inflationDistribution InflationDistribution,
	teamAddress sdk.AccAddress,
	teamVestingProvision sdk.Dec,
) Params {
	return Params{
		MintDenom:              mintDenom,
		EpochIdentifier:        epochIdentifier,
		EpochsPerPeriod:        epochsPerPeriod,
		ExponentialCalculation: exponentialCalculation,
		InflationDistribution:  inflationDistribution,
		TeamAddress:            teamAddress.String(),
		TeamVestingProvision:   teamVestingProvision,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:       sdk.DefaultBondDenom,
		EpochIdentifier: "day", // 1 day
		EpochsPerPeriod: 365,   // 1 year
		ExponentialCalculation: ExponentialCalculation{
			A: sdk.NewDec(int64(300_000_000)),
			R: sdk.NewDecWithPrec(5, 1), // 0.5
			C: sdk.NewDec(int64(9_375_000)),
			B: sdk.OneDec(),
		},
		InflationDistribution: InflationDistribution{
			StakingRewards:  sdk.NewDecWithPrec(533334, 6), // 0.53 = 40% / (1 - 25%)
			UsageIncentives: sdk.NewDecWithPrec(333333, 6), // 0.33 = 25% / (1 - 25%)
			CommunityPool:   sdk.NewDecWithPrec(133333, 6), // 0.13 = 10% / (1 - 25%)
		},
		TeamAddress:          TharsisAccountAddress.Hex(),
		TeamVestingProvision: sdk.NewDec(int64(136986)), // 200000000/(4*365)
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
		paramtypes.NewParamSetPair(KeyEpochIdentifier, &p.EpochIdentifier, epochtypes.ValidateEpochIdentifierInterface),
		paramtypes.NewParamSetPair(KeyEpochsPerPeriod, &p.EpochsPerPeriod, validateEpochsPerPeriod),
		paramtypes.NewParamSetPair(KeyExponentialCalculation, &p.ExponentialCalculation, validateExponentialCalculation),
		paramtypes.NewParamSetPair(KeyInflationDistribution, &p.InflationDistribution, validateInflationDistribution),
		paramtypes.NewParamSetPair(KeyTeamAddress, &p.TeamAddress, validateTeamAddress),
		paramtypes.NewParamSetPair(KeyTeamVestingProvision, &p.TeamVestingProvision, validateTeamVestingProvision),
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

func validateExponentialCalculation(i interface{}) error {
	v, ok := i.(ExponentialCalculation)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// validate initial value
	if v.A.IsNegative() {
		return fmt.Errorf("initial value cannot be negative")
	}

	// validate reduction factor
	if v.R.GT(sdk.NewDec(1)) {
		return fmt.Errorf("reduction factor cannot be greater than 1")
	}

	if v.R.IsNegative() {
		return fmt.Errorf("reduction factor cannot be negative")
	}

	// validate long term inflation
	if v.C.IsNegative() {
		return fmt.Errorf("long term inflation cannot be negative")
	}

	// validate bonding factor
	if v.B.GT(sdk.NewDec(1)) {
		return fmt.Errorf("bonding factor cannot be greater than 1")
	}

	if v.B.IsNegative() {
		return fmt.Errorf("bonding factor cannot be negative")
	}

	return nil
}

func validateInflationDistribution(i interface{}) error {
	v, ok := i.(InflationDistribution)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.StakingRewards.IsNegative() {
		return errors.New("staking distribution ratio must not be negative")
	}

	if v.UsageIncentives.IsNegative() {
		return errors.New("pool incentives distribution ratio must not be negative")
	}

	if v.CommunityPool.IsNegative() {
		return errors.New("community pool distribution ratio must not be negative")
	}

	totalProportions := v.StakingRewards.Add(v.UsageIncentives).Add(v.CommunityPool)
	if !totalProportions.Equal(sdk.NewDec(1)) {
		return errors.New("total distributions ratio should be 1")
	}

	return nil
}

func validateTeamAddress(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if err := ethermint.ValidateAddress(v); err != nil {
		return fmt.Errorf("invalid receiver hex address %w", err)
	}

	return nil
}

func validateTeamVestingProvision(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.LT(sdk.ZeroDec()) {
		return errors.New("team vesting provision must not be negative")
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
	return validateTeamVestingProvision(p.TeamVestingProvision)
}
