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
	"github.com/evmos/evmos/v18/utils"
	"github.com/spf13/cast"

	stakingkeeper "github.com/evmos/evmos/v18/x/staking/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v20
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	appOpts servertypes.AppOptions,
	ibcKeeper ibckeeper.Keeper,
	consumerKeeper *ccvconsumerkeeper.Keeper,
	stakingKeeper stakingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)
		logger.Info("Starting upgrade handler for v20")

		fromVM := make(map[string]uint64)

		ibcKeeper.ConnectionKeeper.SetParams(ctx, ibcKeeper.ConnectionKeeper.GetParams(ctx))

		nodeHome := cast.ToString(appOpts.Get(flags.FlagHome))
		consumerUpgradeGenFile := nodeHome + "/config/ccv.json"
		appState, _, err := genutiltypes.GenesisStateFromGenFile(consumerUpgradeGenFile)
		if err != nil {
			return fromVM, fmt.Errorf("failed to unmarshal genesis state: %w", err)
		}

		var consumerGenesis = consumertypes.GenesisState{}
		cdc.MustUnmarshalJSON(appState[consumertypes.ModuleName], &consumerGenesis)

		consumerGenesis.PreCCV = true
		consumerGenesis.Params.ConsumerRedistributionFraction = "0.75" // 25% of the rewards go towards the Hub
		//consumerGenesis.Params.SoftOptOutThreshold = "0.05"
		consumerGenesis.Params.RewardDenoms = []string{utils.BaseDenom, "uatom"} // Allow Evmos and ATOM rewards
		consumerKeeper.InitGenesis(ctx, &consumerGenesis)
		consumerKeeper.SetDistributionTransmissionChannel(ctx, "channel-3") // The Cosmos hub channel

		return fromVM, nil
	}
}
