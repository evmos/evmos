package v2_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tharsis/ethermint/encoding"
	"github.com/tharsis/evmos/v4/app"
	v1 "github.com/tharsis/evmos/v4/x/epochs/migrations/v1"
	v2 "github.com/tharsis/evmos/v4/x/epochs/migrations/v2"
	types "github.com/tharsis/evmos/v4/x/epochs/types"
)

func TestStoreMigration(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	epochsKey := sdk.NewKVStoreKey(types.StoreKey)
	tEpochsKey := sdk.NewTransientStoreKey(fmt.Sprintf("%s_test", types.StoreKey))
	ctx := testutil.DefaultContext(epochsKey, tEpochsKey)
	store := ctx.KVStore(epochsKey)
	oldstore := prefix.NewStore(store, types.KeyPrefixEpoch)
	newstore := oldstore

	// store pre-migration epochs
	oldGenesis := v1.DefaultGenesisState()
	epochWeek := oldGenesis.Epochs[0]
	epochDay := oldGenesis.Epochs[1]

	keyEpochWeek := []byte(epochWeek.Identifier)
	bzEpochWeek := encCfg.Marshaler.MustMarshal(&epochWeek)
	oldstore.Set(keyEpochWeek, bzEpochWeek)

	keyEpochDay := []byte(epochDay.Identifier)
	bzEpochDay := encCfg.Marshaler.MustMarshal(&epochDay)
	oldstore.Set(keyEpochDay, bzEpochDay)

	// check pre-migration state is intact
	require.True(t, oldstore.Has(keyEpochWeek))
	require.True(t, oldstore.Has(keyEpochDay))
	require.Equal(t, oldstore.Get(keyEpochWeek), bzEpochWeek)
	require.Equal(t, oldstore.Get(keyEpochDay), bzEpochDay)

	// Run migrations
	err := v2.MigrateStore(ctx, epochsKey, encCfg.Marshaler)
	require.NoError(t, err)

	durationWeek := types.DurationToBz(epochWeek.Duration)
	durationDay := types.DurationToBz(epochDay.Duration)

	// Make sure epoch info values have been moved with duration as key
	require.True(t, newstore.Has(durationWeek))
	require.True(t, newstore.Has(durationDay))
	require.Equal(t, bzEpochWeek, newstore.Get(durationWeek))
	require.Equal(t, bzEpochDay, newstore.Get(durationDay))

	// Old keys have been removed
	require.False(t, newstore.Has(keyEpochWeek))
	require.False(t, newstore.Has(keyEpochDay))
}

func TestMigrateJSON(t *testing.T) {
	oldGenesis := v1.DefaultGenesisState()

	// Check identifiers exist in old genesis state
	require.Equal(t, oldGenesis.Epochs[0].Identifier, types.WeekEpochID)
	require.Equal(t, oldGenesis.Epochs[1].Identifier, types.DayEpochID)

	newGenesis := v2.MigrateJSON(*oldGenesis)

	for i, epoch := range oldGenesis.Epochs {
		newepoch := newGenesis.Epochs[i]

		// Check all other field values aside from identifiers are correct
		require.Equal(t, epoch.StartTime, newepoch.StartTime)
		require.Equal(t, epoch.Duration, newepoch.Duration)
		require.Equal(t, epoch.CurrentEpoch, newepoch.CurrentEpoch)
		require.Equal(t, epoch.CurrentEpochStartHeight, newepoch.CurrentEpochStartHeight)
		require.Equal(t, epoch.CurrentEpochStartTime, newepoch.CurrentEpochStartTime)
		require.Equal(t, epoch.EpochCountingStarted, newepoch.EpochCountingStarted)
	}
}
