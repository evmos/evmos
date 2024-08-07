package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	feemarketkeeper "github.com/evmos/evmos/v19/x/feemarket/keeper"
	"github.com/evmos/evmos/v19/x/feemarket/types"
	"github.com/stretchr/testify/require"
)

type mockSubspace struct {
	ps types.Params
}

func newMockSubspace(ps types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSetIfExists(_ sdk.Context, ps types.LegacyParams) {
	*ps.(*types.Params) = ms.ps
}

func TestMigrations(t *testing.T) {
	nw := network.NewUnitTestNetwork()
	ctx := nw.GetContext()

	legacySubspace := newMockSubspace(types.DefaultParams())
	migrator := feemarketkeeper.NewMigrator(nw.App.FeeMarketKeeper, legacySubspace)

	testCases := []struct {
		name        string
		migrateFunc func(ctx sdk.Context) error
	}{
		{
			"Run Migrate3to4",
			migrator.Migrate3to4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.migrateFunc(ctx)
			require.NoError(t, err)
		})
	}
}
