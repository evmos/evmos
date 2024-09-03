// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v20

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/v4/x/ccv/consumer/keeper"
	consumertypes "github.com/cosmos/interchain-security/v4/x/ccv/consumer/types"
	evmkeeper "github.com/evmos/evmos/v19/x/evm/keeper"
	"github.com/spf13/cast"

	stakingkeeper "github.com/evmos/evmos/v19/x/staking/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v20
// ICS Consumer chain upgrade handler
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	appOpts servertypes.AppOptions,
	ibcKeeper ibckeeper.Keeper,
	consumerKeeper *ccvconsumerkeeper.Keeper,
	stakingKeeper stakingkeeper.Keeper,
	evmKeeper evmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)
		logger.Info("Starting upgrade handler for v20")

		fromVM := make(map[string]uint64)

		ibcKeeper.ConnectionKeeper.SetParams(ctx, ibcKeeper.ConnectionKeeper.GetParams(ctx))

		// Evm params change
		evmParams := evmKeeper.GetParams(ctx)
		// TODO: Change accordingly to testnet and mainnet ATOM voucher
		evmParams.EvmDenom = "" // The EVM denom will now be the IBC voucher for ATOM

		nodeHome := cast.ToString(appOpts.Get(flags.FlagHome))
		consumerUpgradeGenFile := nodeHome + "/config/ccv.json"
		appState, _, err := genutiltypes.GenesisStateFromGenFile(consumerUpgradeGenFile)
		if err != nil {
			return fromVM, fmt.Errorf("failed to unmarshal genesis state: %w", err)
		}

		consumerGenesis := consumertypes.GenesisState{}
		cdc.MustUnmarshalJSON(appState[consumertypes.ModuleName], &consumerGenesis)

		consumerGenesis.PreCCV = true

		// TODO: Replace uatom with the IBC voucher for ATOM
		consumerGenesis.Params.RewardDenoms = []string{"uatom"} // Allow Evmos and ATOM rewards
		consumerKeeper.InitGenesis(ctx, &consumerGenesis)
		return fromVM, nil
	}
}
