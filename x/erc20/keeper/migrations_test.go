package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v10/x/erc20/exported"
	erc20keeper "github.com/evmos/evmos/v10/x/erc20/keeper"
	"github.com/evmos/evmos/v10/x/erc20/types"
)

type mockSubspace struct {
	ps types.Params
}

func newMockSubspace(ps types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSet(ctx sdk.Context, ps exported.Params) {
	*ps.(*types.Params) = ms.ps
}

func (suite *KeeperTestSuite) TestMigrations() {
	legacySubspace := newMockSubspace(types.DefaultParams())
	migrator := erc20keeper.NewMigrator(suite.app.Erc20Keeper, legacySubspace)

	testCases := []struct {
		name        string
		migrateFunc func(ctx sdk.Context) error
	}{
		// TODO: Figure out if these will be deleted
		{
			"Run Migrate2to3",
			migrator.Migrate2to3,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := tc.migrateFunc(suite.ctx)
			suite.Require().NoError(err)
		})
	}
}
