package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v10/x/inflation/exported"
	v2 "github.com/evmos/evmos/v10/x/inflation/migrations/v2"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper         Keeper
	legacySubspace exported.Subspace
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper, legacySubspace exported.Subspace) Migrator {
	return Migrator{
		keeper:         keeper,
		legacySubspace: legacySubspace,
	}
}

// Migrate1to2 migrates the store from consensus version v1 to v2
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.MigrateStore(ctx, ctx.KVStore(m.keeper.storeKey), m.legacySubspace, m.keeper.cdc)
}
