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

package inflation

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/evmos/evmos/v12/x/inflation/client/cli"
	"github.com/evmos/evmos/v12/x/inflation/keeper"
	"github.com/evmos/evmos/v12/x/inflation/types"
)

// type check to ensure the interface is properly implemented
var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

// app module Basics object
type AppModuleBasic struct{}

// Name returns the inflation module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the inflation module's types on the given LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// ConsensusVersion returns the consensus state-breaking version for the module.
func (AppModuleBasic) ConsensusVersion() uint64 {
	return 3
}

// RegisterInterfaces registers interfaces and implementations of the incentives
// module.
func (AppModuleBasic) RegisterInterfaces(interfaceRegistry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(interfaceRegistry)
}

// DefaultGenesis returns default genesis state as raw bytes for the incentives
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the inflation module.
func (b AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var genesisState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genesisState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return genesisState.Validate()
}

// RegisterRESTRoutes performs a no-op as the inflation module doesn't expose REST
// endpoints
func (AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the inflation module.
func (b AppModuleBasic) RegisterGRPCGatewayRoutes(c client.Context, serveMux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), serveMux, types.NewQueryClient(c)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the root tx command for the inflation module.
func (AppModuleBasic) GetTxCmd() *cobra.Command { return nil }

// GetQueryCmd returns no root query command for the inflation module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// ___________________________________________________________________________

// AppModule implements an application module for the inflation module.
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
	ak     authkeeper.AccountKeeper
	sk     stakingkeeper.Keeper
	// legacySubspace is used solely for migration of x/params managed parameters
	legacySubspace types.Subspace
}

// NewAppModule creates a new AppModule Object
func NewAppModule(
	k keeper.Keeper,
	ak authkeeper.AccountKeeper,
	sk stakingkeeper.Keeper,
	ss types.Subspace,
) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
		ak:             ak,
		sk:             sk,
		legacySubspace: ss,
	}
}

// Name returns the inflation module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants registers the inflation module invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route returns the message routing key for the inflation module.
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, NewHandler(am.keeper))
}

// QuerierRoute returns the inflation module's querier route name.
func (am AppModule) QuerierRoute() string {
	return types.RouterKey
}

// LegacyQuerierHandler returns the inflation module sdk.Querier.
func (am AppModule) LegacyQuerierHandler(_ *codec.LegacyAmino) sdk.Querier {
	return nil
}

// RegisterServices registers a gRPC query service to respond to the
// module-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), am.keeper)
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)

	m := keeper.NewMigrator(am.keeper, am.legacySubspace)

	// Migrate to version 2 of store
	err := cfg.RegisterMigration(types.ModuleName, 1, m.Migrate1to2)
	if err != nil {
		panic(err)
	}

	// Migrate to version 3 of store
	err = cfg.RegisterMigration(types.ModuleName, 2, m.Migrate2to3)
	if err != nil {
		panic(err)
	}
}

// BeginBlock returns the begin blocker for the inflation module.
func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {
}

// EndBlock returns the end blocker for the inflation module. It returns no validator
// updates.
func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// InitGenesis performs genesis initialization for the inflation module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState

	cdc.MustUnmarshalJSON(data, &genesisState)
	InitGenesis(ctx, am.keeper, am.ak, am.sk, genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the inflation
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(gs)
}

// ___________________________________________________________________________

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the inflation module.
func (am AppModule) GenerateGenesisState(_ *module.SimulationState) {
}

// ProposalContents doesn't return any content functions for governance proposals.
func (am AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{}
}

// RandomizedParams creates randomized inflation param changes for the simulator.
func (am AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{}
}

// RegisterStoreDecoder registers a decoder for inflation module's types.
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {
}

// WeightedOperations doesn't return any inflation module operation.
func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	return []simtypes.WeightedOperation{}
}
