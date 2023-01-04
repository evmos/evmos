package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	inflationkeeper "github.com/evmos/evmos/v10/x/inflation/keeper"
	v2types "github.com/evmos/evmos/v10/x/inflation/migrations/v2/types"
	"github.com/evmos/evmos/v10/x/inflation/types"
)

type mockSubspace struct {
	ps v2types.V2Params
}

func newMockSubspace(ps v2types.V2Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSetIfExists(ctx sdk.Context, ps types.LegacyParams) {
	*ps.(*v2types.V2Params) = ms.ps
}

func (suite *KeeperTestSuite) TestMigrations() {
	legacySubspace := newMockSubspace(v2types.DefaultParams())
	migrator := inflationkeeper.NewMigrator(suite.app.InflationKeeper, legacySubspace)

	testCases := []struct {
		name        string
		migrateFunc func(ctx sdk.Context) error
	}{
		{
			"Run Migrate1to2",
			migrator.Migrate1to2,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := tc.migrateFunc(suite.ctx)
			suite.Require().NoError(err)
		})
	}
}
