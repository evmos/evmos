package keeper_test

import (
	"fmt"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/evmos/evmos/v16/x/epochs/types"
)

func (suite *KeeperTestSuite) TestEpochInfo() {
	var (
		req    *types.QueryEpochsInfoRequest
		expRes *types.QueryEpochsInfoResponse
	)

	day := time.Hour * 24
	week := time.Hour * 24 * 7

	testCases := []struct {
		name     string
		malleate func(ctx sdktypes.Context) sdktypes.Context
		expPass  bool
	}{
		{
			"pass - default EpochInfos",
			func(ctx sdktypes.Context) sdktypes.Context {
				req = &types.QueryEpochsInfoRequest{}

				currentBlockHeight := ctx.BlockHeight()
				currentBlockTime := ctx.BlockTime()

				dayEpoch := types.EpochInfo{
					Identifier:              types.DayEpochID,
					StartTime:               time.Time{},
					Duration:                day,
					CurrentEpoch:            1,
					CurrentEpochStartHeight: 1,
					CurrentEpochStartTime:   time.Time{},
					EpochCountingStarted:    true,
				}
				dayEpoch.StartTime = currentBlockTime
				dayEpoch.CurrentEpochStartHeight = currentBlockHeight

				weekEpoch := types.EpochInfo{
					Identifier:              types.WeekEpochID,
					StartTime:               time.Time{},
					Duration:                week,
					CurrentEpoch:            1,
					CurrentEpochStartHeight: 1,
					CurrentEpochStartTime:   time.Time{},
					EpochCountingStarted:    true,
				}
				weekEpoch.StartTime = currentBlockTime
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
			func(ctx sdktypes.Context) sdktypes.Context {
				currentBlockHeight := ctx.BlockHeight()
				currentBlockTime := ctx.BlockTime()

				dayEpoch := types.EpochInfo{
					Identifier:              types.DayEpochID,
					StartTime:               time.Time{},
					Duration:                time.Hour * 24,
					CurrentEpoch:            1,
					CurrentEpochStartHeight: 1,
					CurrentEpochStartTime:   time.Time{},
					EpochCountingStarted:    true,
				}
				dayEpoch.StartTime = currentBlockTime
				dayEpoch.CurrentEpochStartHeight = currentBlockHeight

				weekEpoch := types.EpochInfo{
					Identifier:              types.WeekEpochID,
					StartTime:               time.Time{},
					Duration:                time.Hour * 24 * 7,
					CurrentEpoch:            1,
					CurrentEpochStartHeight: 1,
					CurrentEpochStartTime:   time.Time{},
					EpochCountingStarted:    true,
				}
				weekEpoch.StartTime = currentBlockTime
				weekEpoch.CurrentEpochStartHeight = currentBlockHeight

				quarterEpoch := types.EpochInfo{
					Identifier:              "quarter",
					StartTime:               time.Time{},
					Duration:                time.Hour * 24 * 7 * 13,
					CurrentEpoch:            0,
					CurrentEpochStartHeight: 1,
					CurrentEpochStartTime:   time.Time{},
					EpochCountingStarted:    false,
				}

				quarterEpoch.StartTime = currentBlockTime
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
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			// Default epoch infos at genesis with day and week.
			suite.SetupTest([]types.EpochInfo{})

			ctx := suite.network.GetContext()
			ctx = tc.malleate(ctx)

			res, err := suite.network.GetEpochsClient().EpochInfos(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCurrentEpoch() {
	var (
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
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest([]types.EpochInfo{})

			ctx := suite.network.GetContext()
			tc.malleate()

			res, err := suite.network.GetEpochsClient().CurrentEpoch(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
