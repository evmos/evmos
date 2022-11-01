package v2_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/evoblockchain/ethermint/encoding"
	"github.com/evoblockchain/evoblock/v8/app"
	v2 "github.com/evoblockchain/evoblock/v8/x/claims/migrations/v2"
	claims "github.com/evoblockchain/evoblock/v8/x/claims/types"
	"github.com/stretchr/testify/require"
)

func TestStoreMigration(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	claimsKey := sdk.NewKVStoreKey(claims.StoreKey)
	tClaimsKey := sdk.NewTransientStoreKey(fmt.Sprintf("%s_test", claims.StoreKey))
	ctx := testutil.DefaultContext(claimsKey, tClaimsKey)
	paramstore := paramtypes.NewSubspace(
		encCfg.Marshaler, encCfg.Amino, claimsKey, tClaimsKey, "claims",
	)
	paramstore = paramstore.WithKeyTable(claims.ParamKeyTable())
	require.True(t, paramstore.HasKeyTable())

	// setup to pre-migration state
	defParam := claims.DefaultParams()
	paramstore.Set(ctx, claims.ParamStoreKeyEnableClaims, defParam.EnableClaims)
	paramstore.Set(ctx, claims.ParamStoreKeyAirdropStartTime, defParam.AirdropStartTime)
	paramstore.Set(ctx, claims.ParamStoreKeyDurationUntilDecay, defParam.DurationUntilDecay)
	paramstore.Set(ctx, claims.ParamStoreKeyDurationOfDecay, defParam.DurationOfDecay)
	paramstore.Set(ctx, claims.ParamStoreKeyClaimsDenom, defParam.ClaimsDenom)

	// check pre-migration state are intact
	require.True(t, paramstore.Has(ctx, claims.ParamStoreKeyEnableClaims))
	require.True(t, paramstore.Has(ctx, claims.ParamStoreKeyAirdropStartTime))
	require.True(t, paramstore.Has(ctx, claims.ParamStoreKeyDurationUntilDecay))
	require.True(t, paramstore.Has(ctx, claims.ParamStoreKeyDurationOfDecay))
	require.True(t, paramstore.Has(ctx, claims.ParamStoreKeyClaimsDenom))

	// check no new params
	require.False(t, paramstore.Has(ctx, claims.ParamStoreKeyEVMChannels))
	require.False(t, paramstore.Has(ctx, claims.ParamStoreKeyAuthorizedChannels))

	// Run migrations
	err := v2.MigrateStore(ctx, &paramstore)
	require.NoError(t, err)

	// Make sure the new params are set
	require.True(t, paramstore.Has(ctx, claims.ParamStoreKeyAuthorizedChannels))
	require.True(t, paramstore.Has(ctx, claims.ParamStoreKeyEVMChannels))

	// Make sure the old params are there too
	require.True(t, paramstore.Has(ctx, claims.ParamStoreKeyEnableClaims))
	require.True(t, paramstore.Has(ctx, claims.ParamStoreKeyAirdropStartTime))
	require.True(t, paramstore.Has(ctx, claims.ParamStoreKeyDurationUntilDecay))
	require.True(t, paramstore.Has(ctx, claims.ParamStoreKeyDurationOfDecay))
	require.True(t, paramstore.Has(ctx, claims.ParamStoreKeyClaimsDenom))

	var authorizedChannels, evmChannels []string

	require.NotPanics(t, func() {
		paramstore.Get(ctx, claims.ParamStoreKeyAuthorizedChannels, &authorizedChannels)
		paramstore.Get(ctx, claims.ParamStoreKeyEVMChannels, &evmChannels)
	})

	// check that the values are the expected ones
	require.Equal(t, claims.DefaultAuthorizedChannels, authorizedChannels)
	require.Equal(t, claims.DefaultEVMChannels, evmChannels)
}
