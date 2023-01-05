package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v2 "github.com/evmos/evmos/v10/x/inflation/migrations/v2"
	v3 "github.com/evmos/evmos/v10/x/inflation/migrations/v3"
	"github.com/evmos/evmos/v10/x/inflation/types"
)

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

// Migrate1to2 migrates the store from consensus version 1 to 2
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.MigrateStore(ctx, ctx.KVStore(m.keeper.storeKey), m.legacySubspace, m.keeper.cdc)
}

// Migrate2to3 migrates the store from consensus version 2 to 3
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.MigrateStore(ctx.KVStore(m.keeper.storeKey))
}
