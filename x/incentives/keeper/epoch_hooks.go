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

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/evmos/evmos/v12/x/epochs/types"
)

// BeforeEpochStart performs a no-op
func (k Keeper) BeforeEpochStart(_ sdk.Context, _ string, _ int64) {}

// AfterEpochEnd distributes the contract incentives at the end of each epoch
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, _ int64) {
	params := k.GetParams(ctx)

	// check if epochIdentifier signal equals the identifier in the params
	if epochIdentifier != params.IncentivesEpochIdentifier {
		return
	}

	// check if the Incentives are globally enabled
	if !params.EnableIncentives {
		return
	}

	if err := k.DistributeRewards(ctx); err != nil {
		panic(err)
	}
}

// ___________________________________________________________________________________________________

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// BeforeEpochStart implements EpochHooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

// AfterEpochEnd implements EpochHooks
func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
