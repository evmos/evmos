package v2_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"
	"github.com/tharsis/ethermint/encoding"
	"github.com/tharsis/evmos/v2/app"
	v2 "github.com/tharsis/evmos/v2/x/claims/migrations/v2"
	claimstypes "github.com/tharsis/evmos/v2/x/claims/types"
)

func TestStoreMigration(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	claimsKey := sdk.NewKVStoreKey(claimstypes.StoreKey)
	tClaimsKey := sdk.NewTransientStoreKey(fmt.Sprintf("%s_test", claimstypes.StoreKey))
	ctx := testutil.DefaultContext(claimsKey, tClaimsKey)
	paramstore := paramtypes.NewSubspace(
		encCfg.Marshaler, encCfg.Amino, claimsKey, tClaimsKey, "claims",
	)

	// check no params
	require.False(t, paramstore.Has(ctx, claimstypes.ParamStoreKeyEVMChannels))
	require.False(t, paramstore.Has(ctx, claimstypes.ParamStoreKeyAuthorizedChannels))

	// Run migrations
	err := v2.MigrateStore(ctx, paramstore)
	require.NoError(t, err)

	// Make sure the new params are set
	require.True(t, paramstore.Has(ctx, claimstypes.ParamStoreKeyAuthorizedChannels))
	require.True(t, paramstore.Has(ctx, claimstypes.ParamStoreKeyEVMChannels))
}
