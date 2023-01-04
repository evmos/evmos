package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	erc20keeper "github.com/evmos/evmos/v10/x/erc20/keeper"
	v3types "github.com/evmos/evmos/v10/x/erc20/migrations/v3/types"
	"github.com/evmos/evmos/v10/x/erc20/types"
)

type mockSubspace struct {
	ps v3types.V3Params
}

func newMockSubspace(ps v3types.V3Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSet(ctx sdk.Context, ps types.LegacyParams) {
	*ps.(*v3types.V3Params) = ms.ps
}

func (suite *KeeperTestSuite) TestMigrations() {
	legacySubspace := newMockSubspace(v3types.DefaultParams())
	migrator := erc20keeper.NewMigrator(suite.app.Erc20Keeper, legacySubspace)

	testCases := []struct {
		name        string
		migrateFunc func(ctx sdk.Context) error
	}{
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
