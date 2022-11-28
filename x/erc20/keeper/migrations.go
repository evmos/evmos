package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	v3 "github.com/evmos/evmos/v10/x/erc20/migrations/v3"
	"github.com/evmos/evmos/v10/x/erc20/types"
)

var _ module.MigrationHandler = Migrator{}.Migrate2to3

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper         Keeper
	legacySubspace types.Subspace
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper, legacySubspace types.Subspace) Migrator {
	return Migrator{
		keeper:         keeper,
		legacySubspace: legacySubspace,
	}
}

// TODO: Possibly Delete
// Migrate1to2 migrates from consensus version 1 to 2.
// func (m Migrator) Migrate1to2(ctx sdk.Context) error {
//	return v2.UpdateParams(ctx, &m.keeper.paramstore)
// }

func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.MigrateStore(ctx, ctx.KVStore(m.keeper.storeKey), m.legacySubspace, m.keeper.cdc)
}
