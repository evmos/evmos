// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	evm "github.com/evmos/evmos/v12/x/evm/types"
	"github.com/evmos/evmos/v12/x/inflation/types"
)

var _ types.LegacyParams = &V2Params{}

var (
	ParamsKey                           = []byte("Params")
	ParamStoreKeyMintDenom              = []byte("ParamStoreKeyMintDenom")
	ParamStoreKeyExponentialCalculation = []byte("ParamStoreKeyExponentialCalculation")
	ParamStoreKeyInflationDistribution  = []byte("ParamStoreKeyInflationDistribution")
	ParamStoreKeyEnableInflation        = []byte("ParamStoreKeyEnableInflation")
)

var (
	DefaultInflationDenom         = evm.DefaultEVMDenom
	DefaultInflation              = true
	DefaultExponentialCalculation = V2ExponentialCalculation{
		A:             sdk.NewDec(int64(300_000_000)),
		R:             sdk.NewDecWithPrec(50, 2), // 50%
		C:             sdk.NewDec(int64(9_375_000)),
		BondingTarget: sdk.NewDecWithPrec(66, 2), // 66%
		MaxVariance:   sdk.ZeroDec(),             // 0%
	}
	DefaultInflationDistribution = V2InflationDistribution{
		StakingRewards:  sdk.NewDecWithPrec(533333334, 9), // 0.53 = 40% / (1 - 25%)
		UsageIncentives: sdk.NewDecWithPrec(333333333, 9), // 0.33 = 25% / (1 - 25%)
		CommunityPool:   sdk.NewDecWithPrec(133333333, 9), // 0.13 = 10% / (1 - 25%)
	}
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&V2Params{})
}

func NewParams(
	mintDenom string,
	exponentialCalculation V2ExponentialCalculation,
	inflationDistribution V2InflationDistribution,
	enableInflation bool,
) V2Params {
	return V2Params{
		MintDenom:              mintDenom,
		ExponentialCalculation: exponentialCalculation,
		InflationDistribution:  inflationDistribution,
		EnableInflation:        enableInflation,
	}
}

func DefaultParams() V2Params {
	return V2Params{
		MintDenom:              DefaultInflationDenom,
		ExponentialCalculation: DefaultExponentialCalculation,
		InflationDistribution:  DefaultInflationDistribution,
		EnableInflation:        DefaultInflation,
	}
}

// Implements params.ParamSet
func (p *V2Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyMintDenom, &p.MintDenom, validateMintDenom),
		paramtypes.NewParamSetPair(ParamStoreKeyExponentialCalculation, &p.ExponentialCalculation, validateExponentialCalculation),
		paramtypes.NewParamSetPair(ParamStoreKeyInflationDistribution, &p.InflationDistribution, validateInflationDistribution),
		paramtypes.NewParamSetPair(ParamStoreKeyEnableInflation, &p.EnableInflation, validateBool),
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
	v, ok := i.(V2ExponentialCalculation)
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
	v, ok := i.(V2InflationDistribution)
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

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func (p V2Params) Validate() error {
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
