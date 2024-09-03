// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package inflation

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/x/inflation/v1/keeper"
	"github.com/evmos/evmos/v19/x/inflation/v1/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	ak types.AccountKeeper,
	_ types.StakingKeeper,
	genesisState types.GenesisState,
) {
	// Ensure inflation module account is set on genesis
	if acc := ak.GetModuleAccount(ctx, types.ModuleName); acc == nil {
		panic("the inflation module account has not been set")
	}

	// Set genesis state
	params := genesisState.Params
	err := k.SetParams(ctx, params)
	if err != nil {
		panic(errorsmod.Wrapf(err, "error setting params"))
	}

	period := genesisState.Period
	k.SetPeriod(ctx, period)

	epochIdentifier := genesisState.EpochIdentifier
	k.SetEpochIdentifier(ctx, epochIdentifier)

	epochsPerPeriod := genesisState.EpochsPerPeriod
	k.SetEpochsPerPeriod(ctx, epochsPerPeriod)

	skippedEpochs := genesisState.SkippedEpochs
	k.SetSkippedEpochs(ctx, skippedEpochs)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:          k.GetParams(ctx),
		Period:          k.GetPeriod(ctx),
		EpochIdentifier: k.GetEpochIdentifier(ctx),
		EpochsPerPeriod: k.GetEpochsPerPeriod(ctx),
		SkippedEpochs:   k.GetSkippedEpochs(ctx),
	}
}
