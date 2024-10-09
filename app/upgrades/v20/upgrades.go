// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v20

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	evmkeeper "github.com/evmos/evmos/v20/x/evm/keeper"
	"github.com/evmos/evmos/v20/x/evm/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for Evmos v20
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ek *evmkeeper.Keeper,
	gk govkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		logger := ctx.Logger().With("upgrade", UpgradeName)

		logger.Debug("Enabling gov precompile...")
		if err := EnableGovPrecompile(ctx, ek); err != nil {
			logger.Error("error while enabling gov precompile", "error", err.Error())
		}

		// run the sdk v0.50 migrations
		logger.Debug("Running module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		logger.Debug("Updating expedited prop params...")
		if err := UpdateExpeditedPropsParams(ctx, gk); err != nil {
			logger.Error("error while updating gov params", "error", err.Error())
		}

		return vm, nil
	}
}

func EnableGovPrecompile(ctx sdk.Context, ek *evmkeeper.Keeper) error {
	// Enable gov precompile
	params := ek.GetParams(ctx)
	params.ActiveStaticPrecompiles = append(params.ActiveStaticPrecompiles, types.GovPrecompileAddress)
	if err := params.Validate(); err != nil {
		return err
	}
	return ek.SetParams(ctx, params)
}

func UpdateExpeditedPropsParams(ctx sdk.Context, gv govkeeper.Keeper) error {
	params, err := gv.Params.Get(ctx)
	if err != nil {
		return err
	}

	// use the same denom as the min deposit denom
	// also amount must be greater than MinDeposit amount
	denom := params.MinDeposit[0].Denom
	expDepAmt := params.ExpeditedMinDeposit[0].Amount
	if expDepAmt.LTE(params.MinDeposit[0].Amount) {
		expDepAmt = params.MinDeposit[0].Amount.MulRaw(govv1.DefaultMinExpeditedDepositTokensRatio)
	}
	params.ExpeditedMinDeposit = sdk.NewCoins(sdk.NewCoin(denom, expDepAmt))

	// if expedited voting period > voting period
	// set expedited voting period to be half the voting period
	if params.ExpeditedVotingPeriod != nil && params.VotingPeriod != nil && *params.ExpeditedVotingPeriod > *params.VotingPeriod {
		expPeriod := *params.VotingPeriod / 2
		params.ExpeditedVotingPeriod = &expPeriod
	}

	if err := params.ValidateBasic(); err != nil {
		return err
	}
	return gv.Params.Set(ctx, params)
}
