package keeper_test

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/evmos/evmos/v19/app"
	"github.com/evmos/evmos/v19/encoding"
	inflationkeeper "github.com/evmos/evmos/v19/x/inflation/v1/keeper"
	v2types "github.com/evmos/evmos/v19/x/inflation/v1/migrations/v2/types"
	"github.com/evmos/evmos/v19/x/inflation/v1/types"
)

type mockSubspace struct {
	ps           v2types.V2Params
	storeKey     storetypes.StoreKey
	transientKey storetypes.StoreKey
}

func newMockSubspace(ps v2types.V2Params, storeKey, transientKey storetypes.StoreKey) mockSubspace {
	return mockSubspace{ps: ps, storeKey: storeKey, transientKey: transientKey}
}

func (ms mockSubspace) GetParamSetIfExists(_ sdk.Context, ps types.LegacyParams) {
	*ps.(*v2types.V2Params) = ms.ps
}

func (ms mockSubspace) WithKeyTable(keyTable paramtypes.KeyTable) paramtypes.Subspace {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	cdc := encCfg.Codec
	return paramtypes.NewSubspace(cdc, encCfg.Amino, ms.storeKey, ms.transientKey, "test").WithKeyTable(keyTable)
}

func (suite *KeeperTestSuite) TestMigrations() {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)

	var outputParams v2types.V2Params
	inputParams := v2types.DefaultParams()
	legacySubspace := newMockSubspace(v2types.DefaultParams(), storeKey, tKey).WithKeyTable(v2types.ParamKeyTable())
	legacySubspace.SetParamSet(ctx, &inputParams)
	legacySubspace.GetParamSetIfExists(ctx, &outputParams)

	// Added dummy keeper in order to use the test store and store key
	mockKeeper := inflationkeeper.NewKeeper(storeKey, encCfg.Codec, authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper, nil, nil, nil, "")
	mockSubspace := newMockSubspace(v2types.DefaultParams(), storeKey, tKey)
	migrator := inflationkeeper.NewMigrator(mockKeeper, mockSubspace)

	testCases := []struct {
		name        string
		migrateFunc func(ctx sdk.Context) error
	}{
		{
			"Run Migrate1to2",
			migrator.Migrate1to2,
		},
		{
			"Run Migrate2to3",
			migrator.Migrate2to3,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := tc.migrateFunc(ctx)
			suite.Require().NoError(err)
		})
	}
}
