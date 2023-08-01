package keeper_test

import (
	"fmt"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"

	"github.com/evmos/evmos/v14/x/claims/types"
)

func (suite *KeeperTestSuite) TestUpdateParams() {
	// Add open channels to the channel keeper (0 & 3 are the default channels, 2 is the default evm channel)
	channel := channeltypes.Channel{State: channeltypes.OPEN}
	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, transfertypes.PortID, channeltypes.ChannelPrefix+"0", channel)
	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, transfertypes.PortID, channeltypes.ChannelPrefix+"3", channel)
	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, transfertypes.PortID, channeltypes.ChannelPrefix+"2", channel)
	// Add closed channel to the channel keeper
	closedChannel := channeltypes.Channel{State: channeltypes.CLOSED}
	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, transfertypes.PortID, channeltypes.ChannelPrefix+"4", closedChannel)

	testCases := []struct {
		name        string
		request     *types.MsgUpdateParams
		expectErr   bool
		errContains string
	}{
		{
			name:        "fail - invalid authority",
			request:     &types.MsgUpdateParams{Authority: "foobar"},
			expectErr:   true,
			errContains: "invalid authority",
		},
		{
			name: "pass - valid Update msg",
			request: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params:    types.DefaultParams(),
			},
			expectErr:   false,
			errContains: "",
		},
		{
			name: "fail - valid Update msg with unknown channel",
			request: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params: types.Params{
					AuthorizedChannels: []string{"channel-0", "channel-1"},
				},
			},
			expectErr:   true,
			errContains: "it is not found in the app's IBCKeeper.ChannelKeeper: channel-1",
		},
		{
			name: "fail - valid Update msg with closed channel",
			request: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params: types.Params{
					AuthorizedChannels: []string{"channel-0", "channel-4"},
				},
			},
			expectErr:   true,
			errContains: "it is not in the OPEN state: channel-4",
		},
		{
			name: "pass - valid Update msg with valid EVM channels",
			request: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params: types.Params{
					EVMChannels: []string{"channel-0", "channel-2"},
				},
			},
			expectErr:   false,
			errContains: "",
		},
		{
			name: "fail - valid Update msg with unknown EVM channel",
			request: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params: types.Params{
					EVMChannels: []string{"channel-6"},
				},
			},
			expectErr:   true,
			errContains: "it is not found in the app's IBCKeeper.ChannelKeeper: channel-6",
		},
		{
			name: "fail - valid Update msg with closed EVM channel",
			request: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params: types.Params{
					EVMChannels: []string{"channel-4"},
				},
			},
			expectErr:   true,
			errContains: "it is not in the OPEN state: channel-4",
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("MsgUpdateParams - %s", tc.name), func() {
			_, err := suite.app.ClaimsKeeper.UpdateParams(suite.ctx, tc.request)
			if tc.expectErr {
				suite.Require().ErrorContains(err, tc.errContains)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
