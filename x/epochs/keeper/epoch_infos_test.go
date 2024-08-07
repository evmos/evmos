package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/evmos/evmos/v19/x/epochs/types"
)

func TestEpochLifeCycle(t *testing.T) {
	// The default genesis includes day and week epochs.
	suite := SetupTest([]types.EpochInfo{})

	epochInfo := types.EpochInfo{
		Identifier:            monthIdentifier,
		StartTime:             time.Time{},
		Duration:              time.Hour * 24 * 30,
		CurrentEpoch:          0,
		CurrentEpochStartTime: time.Time{},
		EpochCountingStarted:  false,
	}
	ctx := suite.network.GetContext()
	suite.network.App.EpochsKeeper.SetEpochInfo(ctx, epochInfo)
	epochInfoSaved, found := suite.network.App.EpochsKeeper.GetEpochInfo(ctx, monthIdentifier)
	require.True(t, found)
	require.Equal(t, epochInfo, epochInfoSaved)

	allEpochs := suite.network.App.EpochsKeeper.AllEpochInfos(ctx)
	require.Len(t, allEpochs, 3)
	require.Equal(t, allEpochs[0].Identifier, types.DayEpochID) // alphabetical order
	require.Equal(t, allEpochs[1].Identifier, monthIdentifier)
	require.Equal(t, allEpochs[2].Identifier, types.WeekEpochID)
}
