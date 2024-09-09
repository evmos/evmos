package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/evmos/evmos/v20/x/epochs/types"
)

func TestEpochInfo(t *testing.T) {
	var (
		// suite is defined here so it is available inside the malleate function.
		suite  *KeeperTestSuite
		req    *types.QueryEpochsInfoRequest
		expRes *types.QueryEpochsInfoResponse
	)

	testCases := []struct {
		name     string
		malleate func() sdktypes.Context
		expPass  bool
	}{
		{
			"pass - default EpochInfos",
			func() sdktypes.Context {
				req = &types.QueryEpochsInfoRequest{}

				ctx := suite.network.GetContext()
				currentBlockHeight := ctx.BlockHeight()
				currentBlockTime := ctx.BlockTime()

				dayEpoch := types.EpochInfo{
					Identifier:              types.DayEpochID,
					Duration:                day,
					CurrentEpoch:            1,
					CurrentEpochStartHeight: 1,
					EpochCountingStarted:    true,
				}
				dayEpoch.StartTime = currentBlockTime
				dayEpoch.CurrentEpochStartTime = currentBlockTime
				dayEpoch.CurrentEpochStartHeight = currentBlockHeight

				weekEpoch := types.EpochInfo{
					Identifier:              types.WeekEpochID,
					Duration:                week,
					CurrentEpoch:            1,
					CurrentEpochStartHeight: 1,
					EpochCountingStarted:    true,
				}
				weekEpoch.StartTime = currentBlockTime
				weekEpoch.CurrentEpochStartTime = currentBlockTime
				weekEpoch.CurrentEpochStartHeight = currentBlockHeight

				expRes = &types.QueryEpochsInfoResponse{
					Epochs: []types.EpochInfo{dayEpoch, weekEpoch},
					Pagination: &query.PageResponse{
						NextKey: nil,
						Total:   uint64(2),
					},
				}

				return ctx
			},
			true,
		},
		{
			"set epoch info",
			func() sdktypes.Context {
				ctx := suite.network.GetContext()
				currentBlockHeight := ctx.BlockHeight()
				currentBlockTime := ctx.BlockTime()

				dayEpoch := types.EpochInfo{
					Identifier:              types.DayEpochID,
					Duration:                time.Hour * 24,
					CurrentEpoch:            1,
					CurrentEpochStartHeight: 1,
					EpochCountingStarted:    true,
				}
				dayEpoch.StartTime = currentBlockTime
				dayEpoch.CurrentEpochStartTime = currentBlockTime
				dayEpoch.CurrentEpochStartHeight = currentBlockHeight

				weekEpoch := types.EpochInfo{
					Identifier:              types.WeekEpochID,
					Duration:                time.Hour * 24 * 7,
					CurrentEpoch:            1,
					CurrentEpochStartHeight: 1,
					EpochCountingStarted:    true,
				}
				weekEpoch.StartTime = currentBlockTime
				weekEpoch.CurrentEpochStartTime = currentBlockTime
				weekEpoch.CurrentEpochStartHeight = currentBlockHeight

				quarterEpoch := types.EpochInfo{
					Identifier:              "quarter",
					Duration:                time.Hour * 24 * 7 * 13,
					CurrentEpoch:            0,
					CurrentEpochStartHeight: 1,
					EpochCountingStarted:    false,
				}

				quarterEpoch.StartTime = currentBlockTime
				quarterEpoch.CurrentEpochStartTime = currentBlockTime
				quarterEpoch.CurrentEpochStartHeight = currentBlockHeight
				suite.network.App.EpochsKeeper.SetEpochInfo(ctx, quarterEpoch)

				req = &types.QueryEpochsInfoRequest{}
				expRes = &types.QueryEpochsInfoResponse{
					Epochs: []types.EpochInfo{dayEpoch, quarterEpoch, weekEpoch},
					Pagination: &query.PageResponse{
						NextKey: nil,
						Total:   uint64(3),
					},
				}

				return ctx
			},
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			// Default epoch infos at genesis with day and week.
			suite = SetupTest([]types.EpochInfo{})
			ctx := tc.malleate()

			res, err := suite.network.GetEpochsClient().EpochInfos(ctx, req)
			if tc.expPass {
				require.NoError(t, err)
				require.Equal(t, expRes, res)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestCurrentEpoch(t *testing.T) {
	var (
		suite  *KeeperTestSuite
		req    *types.QueryCurrentEpochRequest
		expRes *types.QueryCurrentEpochResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"fail - unknown identifier",
			func() {
				req = &types.QueryCurrentEpochRequest{Identifier: "second"}
			},
			false,
		},
		{
			"pass - week identifier",
			func() {
				currentEpoch := int64(1)
				req = &types.QueryCurrentEpochRequest{Identifier: types.WeekEpochID}
				expRes = &types.QueryCurrentEpochResponse{
					CurrentEpoch: currentEpoch,
				}
			},
			true,
		},
		{
			"pass - day identifier",
			func() {
				currentEpoch := int64(1)
				req = &types.QueryCurrentEpochRequest{Identifier: types.DayEpochID}
				expRes = &types.QueryCurrentEpochResponse{
					CurrentEpoch: currentEpoch,
				}
			},
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			suite = SetupTest([]types.EpochInfo{})

			tc.malleate()

			res, err := suite.network.GetEpochsClient().CurrentEpoch(suite.network.GetContext(), req)
			if tc.expPass {
				require.NoError(t, err)
				require.Equal(t, expRes, res)
			} else {
				require.Error(t, err)
			}
		})
	}
}
