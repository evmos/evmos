package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	evm "github.com/tharsis/ethermint/x/evm/types"
	"gopkg.in/yaml.v2"
)

// Parameter store keys
var (
	KeyMintDenom              = []byte("KeyMintDenom")
	KeyExponentialCalculation = []byte("KeyExponentialCalculation")
	KeyInflationDistribution  = []byte("KeyInflationDistribution")
)

// ParamTable for inflation module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string,
	exponentialCalculation ExponentialCalculation,
	inflationDistribution InflationDistribution,
) Params {
	return Params{
		MintDenom:              mintDenom,
		ExponentialCalculation: exponentialCalculation,
		InflationDistribution:  inflationDistribution,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom: evm.DefaultEVMDenom,
		ExponentialCalculation: ExponentialCalculation{
			A:             sdk.NewDec(int64(300_000_000)),
			R:             sdk.NewDecWithPrec(50, 2), // 50%
			C:             sdk.NewDec(int64(9_375_000)),
			BondingTarget: sdk.NewDecWithPrec(66, 2), // 66%
			MaxVariance:   sdk.ZeroDec(),             // 0%
		},
		InflationDistribution: InflationDistribution{
			StakingRewards:  sdk.NewDecWithPrec(533333334, 9), // 0.53 = 40% / (1 - 25%)
			UsageIncentives: sdk.NewDecWithPrec(333333333, 9), // 0.33 = 25% / (1 - 25%)
			CommunityPool:   sdk.NewDecWithPrec(133333333, 9), // 0.13 = 10% / (1 - 25%)
		},
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
		paramtypes.NewParamSetPair(KeyExponentialCalculation, &p.ExponentialCalculation, validateExponentialCalculation),
		paramtypes.NewParamSetPair(KeyInflationDistribution, &p.InflationDistribution, validateInflationDistribution),
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

	// validate bonded target
	if v.BondingTarget.GT(sdk.NewDec(1)) {
		return fmt.Errorf("bonded target cannot be greater than 1")
	}

	if !v.BondingTarget.IsPositive() {
		return fmt.Errorf("bonded target cannot be zero or negative")
	}

	// validate max variance
	if v.MaxVariance.IsNegative() {
		return fmt.Errorf("max variance cannot be negative")
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

func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := validateExponentialCalculation(p.ExponentialCalculation); err != nil {
		return err
	}
	return validateInflationDistribution(p.InflationDistribution)
}
