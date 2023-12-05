// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	v3 "github.com/evmos/evmos/v16/x/erc20/migrations/v3"
	"github.com/evmos/evmos/v16/x/erc20/types"
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

func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.MigrateStore(ctx, m.keeper.storeKey, m.legacySubspace)
}
