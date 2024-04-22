// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package staking

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/evmos/evmos/v18/x/staking/keeper"
)

var (
	_ module.BeginBlockAppModule = AppModule{}
	_ module.EndBlockAppModule   = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic defines the basic application module used by the staking module.
type AppModuleBasic struct {
	*staking.AppModuleBasic
}

// AppModule represents a wrapper around the Cosmos SDK staking module AppModule and
// the Evmos custom staking module keeper.
type AppModule struct {
	*staking.AppModule
	keeper *keeper.Keeper
}

// NewAppModule creates a wrapper for the staking module.
func NewAppModule(
	cdc codec.Codec,
	k *keeper.Keeper,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	ls exported.Subspace,
) AppModule {
	am := staking.NewAppModule(cdc, k.Keeper, ak, bk, ls)
	return AppModule{
		AppModule: &am,
		keeper:    k,
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// Override Staking Msg Server
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	querier := stakingkeeper.Querier{Keeper: am.keeper.Keeper}
	types.RegisterQueryServer(cfg.QueryServer(), querier)

	// !! NOTE: when upgrading to a new cosmos-sdk version
	// !! Check if there're store migrations for the staking module
	// !! if so, you'll need to add them here
}
