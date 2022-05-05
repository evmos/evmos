package v2_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tharsis/ethermint/encoding"
	"github.com/tharsis/evmos/v4/app"
	v2 "github.com/tharsis/evmos/v4/x/epochs/migrations/v2"
	types "github.com/tharsis/evmos/v4/x/epochs/types"
)

func TestStoreMigration(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	epochsKey := sdk.NewKVStoreKey(types.StoreKey)
	tEpochsKey := sdk.NewTransientStoreKey(fmt.Sprintf("%s_test", types.StoreKey))
	ctx := testutil.DefaultContext(epochsKey, tEpochsKey)
	store := ctx.KVStore(epochsKey)
	oldstore := prefix.NewStore(store, v2.KeyPrefixEpoch)
	durationStore := prefix.NewStore(store, v2.KeyPrefixEpochDuration)
	epochStore := oldstore

	// Store pre-migration epochs
	epochWeek := types.EpochInfo{
		Identifier:              types.WeekEpochID,
		StartTime:               time.Time{},
		Duration:                time.Hour * 24 * 7,
		CurrentEpoch:            0,
		CurrentEpochStartHeight: 0,
		CurrentEpochStartTime:   time.Time{},
		EpochCountingStarted:    false,
	}
	epochDay := types.EpochInfo{
		Identifier:              types.DayEpochID,
		StartTime:               time.Time{},
		Duration:                time.Hour * 24,
		CurrentEpoch:            0,
		CurrentEpochStartHeight: 0,
		CurrentEpochStartTime:   time.Time{},
		EpochCountingStarted:    false,
	}

	keyEpochWeek := []byte(epochWeek.Identifier)
	bzEpochWeek := encCfg.Marshaler.MustMarshal(&epochWeek)
	oldstore.Set(keyEpochWeek, bzEpochWeek)

	keyEpochDay := []byte(epochDay.Identifier)
	bzEpochDay := encCfg.Marshaler.MustMarshal(&epochDay)
	oldstore.Set(keyEpochDay, bzEpochDay)

	// Check pre-migration state is intact
	require.True(t, oldstore.Has(keyEpochWeek))
	require.True(t, oldstore.Has(keyEpochDay))
	require.Equal(t, oldstore.Get(keyEpochWeek), bzEpochWeek)
	require.Equal(t, oldstore.Get(keyEpochDay), bzEpochDay)

	// Run migrations
	err := v2.MigrateStore(ctx, epochsKey, encCfg.Marshaler)
	require.NoError(t, err)

	durationWeek := v2.DurationToBz(epochWeek.Duration)
	durationDay := v2.DurationToBz(epochDay.Duration)

	// Make sure epoch info values have been moved with duration as key
	require.True(t, epochStore.Has(durationWeek))
	require.True(t, epochStore.Has(durationDay))
	require.Equal(t, bzEpochWeek, epochStore.Get(durationWeek))
	require.Equal(t, bzEpochDay, epochStore.Get(durationDay))

	// Make sure the new identifier => duration store has correct values
	require.True(t, durationStore.Has(keyEpochWeek))
	require.True(t, durationStore.Has(keyEpochDay))
	require.Equal(t, durationWeek, durationStore.Get(keyEpochWeek))
	require.Equal(t, durationDay, durationStore.Get(keyEpochDay))
}

func TestJSONMigration(t *testing.T) {
	// Pre-migration epochs
	oldEpochs := []types.EpochInfo{
		{
			Identifier:              types.WeekEpochID,
			StartTime:               time.Time{},
			Duration:                time.Hour * 24 * 7,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
		{
			Identifier:              types.DayEpochID,
			StartTime:               time.Time{},
			Duration:                time.Hour * 24,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
	}

	// Run genesis state migration
	oldGenesis := types.GenesisState{Epochs: oldEpochs}
	newGenesis := v2.MigrateJSON(oldGenesis)

	require.Equal(t, 4, len(newGenesis.Epochs))

	for i, epoch := range oldEpochs {
		require.Equal(t, epoch.Identifier, newGenesis.Epochs[i].Identifier)
		require.Equal(t, epoch.StartTime, newGenesis.Epochs[i].StartTime)
		require.Equal(t, epoch.Duration, newGenesis.Epochs[i].Duration)
		require.Equal(t, epoch.CurrentEpoch, newGenesis.Epochs[i].CurrentEpoch)
		require.Equal(t, epoch.CurrentEpochStartHeight, newGenesis.Epochs[i].CurrentEpochStartHeight)
		require.Equal(t, epoch.CurrentEpochStartTime, newGenesis.Epochs[i].CurrentEpochStartTime)
		require.Equal(t, epoch.EpochCountingStarted, newGenesis.Epochs[i].EpochCountingStarted)
	}

	for _i, epoch := range v2.NewEpochs {
		i := _i + 2
		require.Equal(t, epoch.Identifier, newGenesis.Epochs[i].Identifier)
		require.Equal(t, epoch.StartTime, newGenesis.Epochs[i].StartTime)
		require.Equal(t, epoch.Duration, newGenesis.Epochs[i].Duration)
		require.Equal(t, epoch.CurrentEpoch, newGenesis.Epochs[i].CurrentEpoch)
		require.Equal(t, epoch.CurrentEpochStartHeight, newGenesis.Epochs[i].CurrentEpochStartHeight)
		require.Equal(t, epoch.CurrentEpochStartTime, newGenesis.Epochs[i].CurrentEpochStartTime)
		require.Equal(t, epoch.EpochCountingStarted, newGenesis.Epochs[i].EpochCountingStarted)
	}
}
