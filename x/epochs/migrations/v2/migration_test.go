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

	// store pre-migration epochs
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

	// check pre-migration state is intact
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
