package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestGetAllSTRV2Address() {
	address1 := suite.address.Bytes()
	address2 := suite.consAddress.Bytes()

	testCases := []struct {
		name     string
		malleate func()
		expected []sdk.AccAddress
	}{
		{
			"space is empty",
			func() {
			},
			[]sdk.AccAddress{},
		},
		{
			"one address",
			func() {
				suite.app.Erc20Keeper.SetSTRv2Address(suite.ctx, address1)
			},
			[]sdk.AccAddress{address1},
		},
		{
			"two addresses",
			func() {
				suite.app.Erc20Keeper.SetSTRv2Address(suite.ctx, address1)
				suite.app.Erc20Keeper.SetSTRv2Address(suite.ctx, address2)
			},
			[]sdk.AccAddress{address1, address2},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			addresses := suite.app.Erc20Keeper.GetAllSTRV2Address(suite.ctx)
			suite.Require().ElementsMatch(tc.expected, addresses)
		})
	}
}
