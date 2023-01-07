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

package v11

import (
	"fmt"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ica "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts"
	icacontrollertypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

var (
	// create ICS27 Controller submodule params, with the controller module NOT enabled
	controllerParams = icacontrollertypes.Params{}

	// create ICS27 Host submodule params with a list of allowed messages
	hostParams = icahosttypes.Params{
		HostEnabled: true,
		AllowMessages: []string{
			sdk.MsgTypeURL(&banktypes.MsgSend{}),
			sdk.MsgTypeURL(&banktypes.MsgMultiSend{}),
			sdk.MsgTypeURL(&distrtypes.MsgSetWithdrawAddress{}),
			sdk.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}),
			sdk.MsgTypeURL(&govtypes.MsgVote{}),
			sdk.MsgTypeURL(&govtypes.MsgVoteWeighted{}),
			sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
			sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
			sdk.MsgTypeURL(&stakingtypes.MsgCancelUnbondingDelegation{}),
			sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
			sdk.MsgTypeURL(&transfertypes.MsgTransfer{}),
		},
	}
)

// CreateUpgradeHandler creates an SDK upgrade handler for v11
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ak authkeeper.AccountKeeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		MigrateEscrowAccounts(ctx, ak)

		// Register the consensus version in the version map to avoid the SDK from triggering the default
		// InitGenesis function.
		vm[icatypes.ModuleName] = ica.AppModule{}.ConsensusVersion()

		// cast ica module (stored as AppModule type in module manager) to ica.AppModule type in order
		// to use the InitModule method. This is an alternative to InitGenesis, and has the advantage,
		// that the used parameters for the ica controller and host can directly be passed in.
		icaModule, correctTypecast := mm.Modules[icatypes.ModuleName].(ica.AppModule)
		if !correctTypecast {
			panic("mm.Modules[icatypes.ModuleName] is not of type ica.AppModule")
		}

		// Run InitModule because we are adjusting the ICA host submodule parameters to only allow
		// selected message types
		icaModule.InitModule(ctx, controllerParams, hostParams)

		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// MigrateEscrowAccounts updates the IBC transfer escrow accounts type to ModuleAccount
func MigrateEscrowAccounts(ctx sdk.Context, ak authkeeper.AccountKeeper) {
	for i := 0; i <= openChannels; i++ {
		channelID := fmt.Sprintf("channel-%d", i)
		address := transfertypes.GetEscrowAddress(transfertypes.PortID, channelID)

		// check if account exists
		existingAcc := ak.GetAccount(ctx, address)

		// account does NOT exist, so don't create it
		if existingAcc == nil {
			continue
		}

		// if existing account is ModuleAccount, no-op
		if _, isModuleAccount := existingAcc.(authtypes.ModuleAccountI); isModuleAccount {
			continue
		}

		// account name based on the address derived by the transfertypes.GetEscrowAddress
		// this function appends the current IBC transfer module version to the provided port and channel IDs
		// To pass account validation, need to have address derived from account name
		accountName := fmt.Sprintf("%s\x00%s/%s", transfertypes.Version, transfertypes.PortID, channelID)
		baseAcc := authtypes.NewBaseAccountWithAddress(address)

		// no special permissions defined for the module account
		acc := authtypes.NewModuleAccount(baseAcc, accountName)
		ak.SetModuleAccount(ctx, acc)
	}
}
