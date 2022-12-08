package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	v3 "github.com/evmos/evmos/v10/x/ibc/transfer/migrations/v3"
)

var _ module.MigrationHandler = Migrator{}.Migrate2to3

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{
		keeper: keeper,
	}
}

// Migrate2to3 migrates from consensus version 1 to 2.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.MigrateEscrowAccounts(ctx, m.keeper.accountKeeper)
}
