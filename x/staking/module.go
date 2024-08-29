// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package staking

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v19/x/staking/keeper"
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

// InitGenesis delegates the InitGenesis call to the underlying x/staking module,
// however, it returns no validator updates as validators are tracked via the
// consumer chain's x/cvv/consumer module and so this module is not responsible
// for returning the initial validator set.
//
// Note: InitGenesis is not called during the soft upgrade of a module
// (as a part of a changeover from standalone -> consumer chain),
// so there is no special handling needed in this method for a consumer being in the pre-CCV state.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState

	cdc.MustUnmarshalJSON(data, &genesisState)
	_ = am.keeper.InitGenesis(ctx, &genesisState)

	return []abci.ValidatorUpdate{}
}

// EndBlock delegates the EndBlock call to the underlying x/staking module,
// however, it returns no validator updates as validators are tracked via the
// consumer chain's x/cvv/consumer module and so this module is not responsible
// for returning the initial validator set.
//
// Note: This method does not require any special handling for PreCCV being true
// (as a part of the changeover from standalone -> consumer chain).
// The ccv consumer Endblocker is ordered to run before the staking Endblocker,
// so if PreCCV is true during one block, the ccv consumer Enblocker will return the proper validator updates,
// the PreCCV flag will be toggled to false, and no validator updates should be returned by this method.
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	_ = am.keeper.BlockValidatorUpdates(ctx)
	return []abci.ValidatorUpdate{}
}
