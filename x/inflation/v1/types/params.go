// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmostypes "github.com/evmos/evmos/v20/types"
)

// FoundationWallets contains the address of the foundation that held tokens to
// be distributed to the team. Tokens allocated to these wallets are excluded
// from the circulating supply.
var FoundationWallets = []string{
	"0x29fDcB7b64B84fD54D0fB0E04A8f6B062046fc6F",
	"0x83ef4C096F9A9daC61081121CCE30578fe437182",
	"0x7F98F5891780Dc73984da11b5499A70F31A6b2A8",
	"0xCdf843a5102cc1B3098940552faA9e6F39b3Bb61",
}

var ParamsKey = []byte("Params")

var (
	DefaultInflationDenom         = evmostypes.BaseDenom
	DefaultInflation              = true
	DefaultExponentialCalculation = ExponentialCalculation{
		A:             math.LegacyNewDec(int64(300_000_000)),
		R:             math.LegacyNewDecWithPrec(50, 2), // 50%
		C:             math.LegacyNewDec(int64(9_375_000)),
		BondingTarget: math.LegacyNewDecWithPrec(66, 2), // 66%
		MaxVariance:   math.LegacyZeroDec(),             // 0%
	}
	DefaultInflationDistribution = InflationDistribution{
		StakingRewards:  math.LegacyNewDecWithPrec(533333334, 9), // 0.53
		CommunityPool:   math.LegacyNewDecWithPrec(466666666, 9), // 0.47
		UsageIncentives: math.LegacyZeroDec(),                    // Deprecated
	}
)

func NewParams(
	mintDenom string,
	exponentialCalculation ExponentialCalculation,
	inflationDistribution InflationDistribution,
	enableInflation bool,
) Params {
	return Params{
		MintDenom:              mintDenom,
		ExponentialCalculation: exponentialCalculation,
		InflationDistribution:  inflationDistribution,
		EnableInflation:        enableInflation,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:              DefaultInflationDenom,
		ExponentialCalculation: DefaultExponentialCalculation,
		InflationDistribution:  DefaultInflationDistribution,
		EnableInflation:        DefaultInflation,
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

	return sdk.ValidateDenom(v)
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
	if v.R.GT(math.LegacyNewDec(1)) {
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
	if v.BondingTarget.GT(math.LegacyNewDec(1)) {
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

	if !v.UsageIncentives.IsZero() {
		return errors.New("incentives pool distribution is deprecated. UsageIncentives param should be zero")
	}

	if v.CommunityPool.IsNegative() {
		return errors.New("community pool distribution ratio must not be negative")
	}

	totalProportions := v.StakingRewards.Add(v.UsageIncentives).Add(v.CommunityPool)
	if !totalProportions.Equal(math.LegacyNewDec(1)) {
		return errors.New("total distributions ratio should be 1")
	}

	return nil
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
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
	if err := validateInflationDistribution(p.InflationDistribution); err != nil {
		return err
	}

	return validateBool(p.EnableInflation)
}
