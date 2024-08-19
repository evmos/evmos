package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	epochstypes "github.com/evmos/evmos/v19/x/epochs/types"
	"github.com/evmos/evmos/v19/x/inflation/v1/types"
	"github.com/stretchr/testify/require"
)

func TestSetGetEpochIdentifier(t *testing.T) {
	var (
		ctx sdk.Context
		nw  *network.UnitTestNetwork
	)
	defaultEpochIdentifier := types.DefaultGenesisState().EpochIdentifier
	expEpochIdentifier := epochstypes.WeekEpochID

	testCases := []struct {
		name     string
		malleate func()
		ok       bool
	}{
		{
			"default epochIdentifier",
			func() {},
			false,
		},
		{
			"epochIdentifier set",
			func() {
				nw.App.InflationKeeper.SetEpochIdentifier(ctx, expEpochIdentifier)
			},
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			// reset
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()

			tc.malleate()

			epochIdentifier := nw.App.InflationKeeper.GetEpochIdentifier(ctx)
			if tc.ok {
				require.Equal(t, expEpochIdentifier, epochIdentifier, tc.name)
			} else {
				require.Equal(t, defaultEpochIdentifier, epochIdentifier, tc.name)
			}
		})
	}
}

func TestSetGetEpochsPerPeriod(t *testing.T) { //nolint:dupl
	var (
		ctx sdk.Context
		nw  *network.UnitTestNetwork
	)
	defaultEpochsPerPeriod := types.DefaultGenesisState().EpochsPerPeriod
	expEpochsPerPeriod := int64(180)

	testCases := []struct {
		name     string
		malleate func()
		ok       bool
	}{
		{
			"default period",
			func() {},
			false,
		},
		{
			"period set",
			func() {
				nw.App.InflationKeeper.SetEpochsPerPeriod(ctx, expEpochsPerPeriod)
			},
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()

			tc.malleate()

			epochsPerPeriod := nw.App.InflationKeeper.GetEpochsPerPeriod(ctx)
			if tc.ok {
				require.Equal(t, expEpochsPerPeriod, epochsPerPeriod, tc.name)
			} else {
				require.Equal(t, defaultEpochsPerPeriod, epochsPerPeriod, tc.name)
			}
		})
	}
}

func TestSetGetSkippedEpochs(t *testing.T) { //nolint:dupl
	var (
		ctx sdk.Context
		nw  *network.UnitTestNetwork
	)
	defaultSkippedEpochs := types.DefaultGenesisState().SkippedEpochs
	expSkippedepochs := uint64(20)

	testCases := []struct {
		name     string
		malleate func()
		ok       bool
	}{
		{
			"default skipped epoch",
			func() {},
			false,
		},
		{
			"skipped epoch set",
			func() {
				nw.App.InflationKeeper.SetSkippedEpochs(ctx, expSkippedepochs)
			},
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()

			tc.malleate()

			epochsPerPeriod := nw.App.InflationKeeper.GetSkippedEpochs(ctx)
			if tc.ok {
				require.Equal(t, expSkippedepochs, epochsPerPeriod, tc.name)
			} else {
				require.Equal(t, defaultSkippedEpochs, epochsPerPeriod, tc.name)
			}
		})
	}
}
