package transfer

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/module"
	ibctransfer "github.com/cosmos/ibc-go/v5/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v5/modules/apps/transfer/keeper"
	"github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	"github.com/evmos/evmos/v10/x/ibc/transfer/keeper"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic embeds the IBC Transfer AppModuleBasic
type AppModuleBasic struct {
	*ibctransfer.AppModuleBasic
}

// AppModule represents the AppModule for this module
type AppModule struct {
	*ibctransfer.AppModule
	keeper keeper.Keeper
}

// NewAppModule creates a new 20-transfer module
func NewAppModule(k keeper.Keeper) AppModule {
	am := ibctransfer.NewAppModule(*k.Keeper)
	return AppModule{
		AppModule: &am,
		keeper:    k,
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// Override Transfer Msg Server
	types.RegisterMsgServer(cfg.MsgServer(), am.keeper)
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)

	ibcMigrator := ibctransferkeeper.NewMigrator(*am.keeper.Keeper)
	if err := cfg.RegisterMigration(types.ModuleName, 1, ibcMigrator.MigrateTraces); err != nil {
		panic(fmt.Sprintf("failed to migrate transfer app from version 1 to 2: %v", err))
	}

	m := keeper.NewMigrator(am.keeper)
	// register v2 -> v3 migration
	if err := cfg.RegisterMigration(types.ModuleName, 2, m.Migrate2to3); err != nil {
		panic(fmt.Errorf("failed to migrate %s to v3: %w", types.ModuleName, err))
	}
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 3 }
