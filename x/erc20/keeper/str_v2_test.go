package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestSetDeleteSTRV2Address() {
	address1 := suite.address.Bytes()
	address2 := suite.consAddress.Bytes()

	suite.SetupTest()

	// Set the same address twice, and it shouldnt fail
	suite.app.Erc20Keeper.SetSTRv2Address(suite.ctx, address1)
	suite.Require().Equal(true, suite.app.Erc20Keeper.HasSTRv2Address(suite.ctx, address1))
	suite.app.Erc20Keeper.SetSTRv2Address(suite.ctx, address1)
	suite.Require().Equal(true, suite.app.Erc20Keeper.HasSTRv2Address(suite.ctx, address1))

	// Set a different address and it shouldnt affect the first address
	suite.app.Erc20Keeper.SetSTRv2Address(suite.ctx, address2)
	suite.Require().Equal(true, suite.app.Erc20Keeper.HasSTRv2Address(suite.ctx, address2))
	suite.app.Erc20Keeper.SetSTRv2Address(suite.ctx, address1)
	suite.Require().Equal(true, suite.app.Erc20Keeper.HasSTRv2Address(suite.ctx, address1))

	// Delete the first address.
	// - it should delete the first address
	// - it shouldnt affect the second one
	suite.app.Erc20Keeper.DeleteSTRv2Address(suite.ctx, address1)
	suite.Require().Equal(false, suite.app.Erc20Keeper.HasSTRv2Address(suite.ctx, address1))
	suite.app.Erc20Keeper.DeleteSTRv2Address(suite.ctx, address1)
	suite.Require().Equal(false, suite.app.Erc20Keeper.HasSTRv2Address(suite.ctx, address1))
	suite.Require().Equal(true, suite.app.Erc20Keeper.HasSTRv2Address(suite.ctx, address2))

	// Set the deleted address again
	suite.app.Erc20Keeper.SetSTRv2Address(suite.ctx, address1)
	suite.Require().Equal(true, suite.app.Erc20Keeper.HasSTRv2Address(suite.ctx, address1))
	suite.Require().Equal(true, suite.app.Erc20Keeper.HasSTRv2Address(suite.ctx, address2))
}

func (suite *KeeperTestSuite) TestHasSTRV2Address() {
	address1 := suite.address.Bytes()
	address2 := suite.consAddress.Bytes()

	testCases := []struct {
		name     string
		malleate func()
		expected bool
	}{
		{
			"space is empty",
			func() {
			},
			false,
		},
		{
			"one address",
			func() {
				suite.app.Erc20Keeper.SetSTRv2Address(suite.ctx, address1)
			},
			true,
		},
		{
			"two addresses",
			func() {
				suite.app.Erc20Keeper.SetSTRv2Address(suite.ctx, address1)
				suite.app.Erc20Keeper.SetSTRv2Address(suite.ctx, address2)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			actual := suite.app.Erc20Keeper.HasSTRv2Address(suite.ctx, address1)
			suite.Require().Equal(tc.expected, actual)
		})
	}
}

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
