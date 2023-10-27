// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v15rc2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/evmos/evmos/v15/utils"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v15.0.0-rc2.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ak authzkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		if utils.IsTestnet(ctx.ChainID()) {
			logger.Info("removing outdated distribution authorizations")
			RemoveDistributionAuthorizations(ctx, ak)
		}

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// RemoveDistributionAuthorizations removes any outdated distribution authorizations
// found in the authz keeper.
func RemoveDistributionAuthorizations(ctx sdk.Context, ak authzkeeper.Keeper) {
	logger := ctx.Logger().With("upgrade", UpgradeName, "module", "authz")

	ak.IterateGrants(ctx, func(granterAddr, granteeAddr sdk.AccAddress, grant authz.Grant) bool {
		authorization, err := grant.GetAuthorization()
		if err != nil {
			return false
		}

		distAuthz, ok := authorization.(*distributiontypes.DistributionAuthorization)
		if !ok {
			return false
		}

		if err = ak.DeleteGrant(ctx, granteeAddr, granterAddr, distAuthz.MsgTypeURL()); err != nil {
			logger.Error("failed to delete distribution authorization",
				"grantee", granteeAddr,
				"granter", granterAddr,
				"msg_type_url", distAuthz.MsgTypeURL(),
				"error", err,
			)
		}

		return false
	})
}
