package keeper_test

// import (
// 	"fmt"
// 	"time"
//
// 	"github.com/cosmos/cosmos-sdk/types/query"
// 	"github.com/evmos/evmos/v16/x/epochs/types"
// )
//
// func (suite *KeeperTestSuite) TestEpochInfo() {
// 	var (
// 		req    *types.QueryEpochsInfoRequest
// 		expRes *types.QueryEpochsInfoResponse
// 	)
//
// 	testCases := []struct {
// 		name     string
// 		malleate func()
// 		expPass  bool
// 	}{
// 		{
// 			"pass - default EpochInfos",
// 			func() {
// 				req = &types.QueryEpochsInfoRequest{}
//
// 				day := types.EpochInfo{
// 					Identifier:              types.DayEpochID,
// 					StartTime:               time.Time{},
// 					Duration:                time.Hour * 24,
// 					CurrentEpoch:            1,
// 					CurrentEpochStartHeight: 1,
// 					CurrentEpochStartTime:   time.Time{},
// 					EpochCountingStarted:    false,
// 				}
// 				day.StartTime = suite.network.GetContext().BlockTime()
// 				day.CurrentEpochStartHeight = suite.network.GetContext().BlockHeight()
//
// 				week := types.EpochInfo{
// 					Identifier:              types.WeekEpochID,
// 					StartTime:               time.Time{},
// 					Duration:                time.Hour * 24 * 7,
// 					CurrentEpoch:            1,
// 					CurrentEpochStartHeight: 1,
// 					CurrentEpochStartTime:   time.Time{},
// 					EpochCountingStarted:    false,
// 				}
// 				week.StartTime = suite.network.GetContext().BlockTime()
// 				week.CurrentEpochStartHeight = suite.network.GetContext().BlockHeight()
//
// 				expRes = &types.QueryEpochsInfoResponse{
// 					Epochs: []types.EpochInfo{day, week},
// 					Pagination: &query.PageResponse{
// 						NextKey: nil,
// 						Total:   uint64(2),
// 					},
// 				}
// 			},
// 			true,
// 		},
// 		// {
// 		// 	"set epoch info",
// 		// 	func() {
// 		// 		day := types.EpochInfo{
// 		// 			Identifier:              types.DayEpochID,
// 		// 			StartTime:               time.Time{},
// 		// 			Duration:                time.Hour * 24,
// 		// 			CurrentEpoch:            0,
// 		// 			CurrentEpochStartHeight: 1,
// 		// 			CurrentEpochStartTime:   time.Time{},
// 		// 			EpochCountingStarted:    false,
// 		// 		}
// 		// 		day.StartTime = suite.network.GetContext().BlockTime()
// 		// 		day.CurrentEpochStartHeight = suite.network.GetContext().BlockHeight()
// 		//
// 		// 		week := types.EpochInfo{
// 		// 			Identifier:              types.WeekEpochID,
// 		// 			StartTime:               time.Time{},
// 		// 			Duration:                time.Hour * 24 * 7,
// 		// 			CurrentEpoch:            0,
// 		// 			CurrentEpochStartHeight: 1,
// 		// 			CurrentEpochStartTime:   time.Time{},
// 		// 			EpochCountingStarted:    false,
// 		// 		}
// 		// 		week.StartTime = suite.network.GetContext().BlockTime()
// 		// 		week.CurrentEpochStartHeight = suite.network.GetContext().BlockHeight()
// 		//
// 		// 		quarter := types.EpochInfo{
// 		// 			Identifier:              "quarter",
// 		// 			StartTime:               time.Time{},
// 		// 			Duration:                time.Hour * 24 * 7 * 13,
// 		// 			CurrentEpoch:            0,
// 		// 			CurrentEpochStartHeight: 1,
// 		// 			CurrentEpochStartTime:   time.Time{},
// 		// 			EpochCountingStarted:    false,
// 		// 		}
// 		//               ctx := suite.network.GetContext()
// 		// 		quarter.StartTime = ctx.BlockTime()
// 		// 		quarter.CurrentEpochStartHeight = ctx.BlockHeight()
// 		// 		suite.network.App.EpochsKeeper.SetEpochInfo(ctx, quarter)
// 		// 		suite.network.App.Commit()
// 		//
// 		// 		req = &types.QueryEpochsInfoRequest{}
// 		// 		expRes = &types.QueryEpochsInfoResponse{
// 		// 			Epochs: []types.EpochInfo{day, quarter, week},
// 		// 			Pagination: &query.PageResponse{
// 		// 				NextKey: nil,
// 		// 				Total:   uint64(3),
// 		// 			},
// 		// 		}
// 		// 	},
// 		// 	true,
// 		// },
// 	}
// 	for _, tc := range testCases {
// 		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
// 			suite.SetupTest([]types.EpochInfo{})
//
// 			ctx := suite.network.GetContext()
// 			tc.malleate()
//
// 			res, err := suite.network.GetEpochsClient().EpochInfos(ctx, req)
// 			if tc.expPass {
// 				suite.Require().NoError(err)
// 				suite.Require().Equal(expRes, res)
// 			} else {
// 				suite.Require().Error(err)
// 			}
// 		})
// 	}
// }
//
// func (suite *KeeperTestSuite) TestCurrentEpoch() {
// 	var (
// 		req    *types.QueryCurrentEpochRequest
// 		expRes *types.QueryCurrentEpochResponse
// 	)
//
// 	testCases := []struct {
// 		name     string
// 		malleate func()
// 		expPass  bool
// 	}{
// 		{
// 			"fail - unknown identifier",
// 			func() {
// 				req = &types.QueryCurrentEpochRequest{Identifier: "second"}
// 			},
// 			false,
// 		},
// 		{
// 			"pass - week identifier",
// 			func() {
// 				currentEpoch := int64(1)
// 				req = &types.QueryCurrentEpochRequest{Identifier: types.WeekEpochID}
// 				expRes = &types.QueryCurrentEpochResponse{
// 					CurrentEpoch: currentEpoch,
// 				}
// 			},
// 			true,
// 		},
// 		{
// 			"pass - day identifier",
// 			func() {
// 				currentEpoch := int64(1)
// 				req = &types.QueryCurrentEpochRequest{Identifier: types.DayEpochID}
// 				expRes = &types.QueryCurrentEpochResponse{
// 					CurrentEpoch: currentEpoch,
// 				}
// 			},
// 			true,
// 		},
// 	}
// 	for _, tc := range testCases {
// 		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
// 			suite.SetupTest([]types.EpochInfo{})
//
// 			ctx := suite.network.GetContext()
// 			tc.malleate()
//
// 			res, err := suite.network.GetEpochsClient().CurrentEpoch(ctx, req)
// 			if tc.expPass {
// 				suite.Require().NoError(err)
// 				suite.Require().Equal(expRes, res)
// 			} else {
// 				suite.Require().Error(err)
// 			}
// 		})
// 	}
// }
