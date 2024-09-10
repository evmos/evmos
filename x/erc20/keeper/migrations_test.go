package keeper_test

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/evmos/evmos/v20/encoding"
	erc20keeper "github.com/evmos/evmos/v20/x/erc20/keeper"
	v3types "github.com/evmos/evmos/v20/x/erc20/migrations/v3/types"
	"github.com/evmos/evmos/v20/x/erc20/types"
)

type mockSubspace struct {
	ps           v3types.V3Params
	storeKey     storetypes.StoreKey
	transientKey storetypes.StoreKey
}

func newMockSubspace(ps v3types.V3Params, storeKey, transientKey storetypes.StoreKey) mockSubspace {
	return mockSubspace{ps: ps, storeKey: storeKey, transientKey: transientKey}
}

func (ms mockSubspace) GetParamSet(_ sdk.Context, ps types.LegacyParams) {
	*ps.(*v3types.V3Params) = ms.ps
}

func (ms mockSubspace) WithKeyTable(keyTable paramtypes.KeyTable) paramtypes.Subspace {
	encCfg := encoding.MakeConfig()
	cdc := encCfg.Codec
	return paramtypes.NewSubspace(cdc, encCfg.Amino, ms.storeKey, ms.transientKey, "test").WithKeyTable(keyTable)
}

func (suite *KeeperTestSuite) TestMigrations() {
	storeKey := storetypes.NewKVStoreKey(types.ModuleName)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)

	var outputParams v3types.V3Params
	inputParams := v3types.DefaultParams()
	legacySubspace := newMockSubspace(v3types.DefaultParams(), storeKey, tKey).WithKeyTable(v3types.ParamKeyTable())
	legacySubspace.SetParamSet(ctx, &inputParams)
	legacySubspace.GetParamSetIfExists(ctx, &outputParams)

	// Added dummy keeper in order to use the test store and store key
	mockKeeper := erc20keeper.NewKeeper(storeKey, nil, authtypes.NewModuleAddress(govtypes.ModuleName), nil, nil, nil, nil, suite.network.App.AuthzKeeper, nil)
	mockSubspace := newMockSubspace(v3types.DefaultParams(), storeKey, tKey)
	migrator := erc20keeper.NewMigrator(mockKeeper, mockSubspace)

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
			err := tc.migrateFunc(ctx)
			suite.Require().NoError(err)
		})
	}
}
