package keeper_test

import (
	"fmt"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"

	"github.com/evmos/evmos/v11/x/claims/types"
)

func (suite *KeeperTestSuite) TestUpdateParams() {
	// Add channels to the channel keeper
	channel := channeltypes.Channel{}
	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, transfertypes.PortID, channeltypes.ChannelPrefix+"0", channel)
	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, transfertypes.PortID, channeltypes.ChannelPrefix+"3", channel)

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
			name: "fail - valid Update msg with invalid channel name",
			request: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params: types.Params{
					AuthorizedChannels: []string{"channel-0", "abc"},
				},
			},
			expectErr:   true,
			errContains: "invalid authorized channel",
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
