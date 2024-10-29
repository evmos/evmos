package v5

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v5types "github.com/evmos/evmos/v19/x/erc20/migrations/v5/types"
	"github.com/evmos/evmos/v19/x/erc20/types"
	"github.com/stretchr/testify/require"
)

func TestMigrate(t *testing.T) {
	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := prefix.NewStore(ctx.KVStore(storeKey), types.KeyPrefixTokenPair)

	for _, pair := range v5types.DefaultTokenPairs {
		marshaledPair, err := pair.Marshal()
		require.NoError(t, err)
		store.Set(pair.GetID(), marshaledPair)
	}

	marshaledPair := store.Get(v5types.DefaultTokenPairs[0].GetID())
	fmt.Println("marshaledPair", marshaledPair)
	require.NotNil(t, marshaledPair)

	require.NoError(t, MigrateStore(ctx, storeKey))

	var tokenPairs []types.TokenPair
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var tokenPair types.TokenPair
		err := tokenPair.Unmarshal(iterator.Value())
		require.NoError(t, err)

		tokenPairs = append(tokenPairs, tokenPair)
	}

	require.Equal(t, types.DefaultTokenPairs, tokenPairs)
}
