package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutiltx "github.com/evmos/evmos/v15/testutil/tx"
)

func (suite *KeeperTestSuite) TestGovClawbackStore() {
	suite.SetupTest()

	addr := sdk.AccAddress(suite.address.Bytes())

	// check that the address is not disabled by default
	disabled := suite.app.VestingKeeper.HasGovClawbackDisabled(suite.ctx, addr)
	suite.Require().False(disabled, "expected address not to be found in store")

	// disable the address
	suite.app.VestingKeeper.SetGovClawbackDisabled(suite.ctx, addr)

	// check that the address is disabled
	disabled = suite.app.VestingKeeper.HasGovClawbackDisabled(suite.ctx, addr)
	suite.Require().True(disabled, "expected address to be found in store")

	// delete the address
	suite.app.VestingKeeper.DeleteGovClawbackDisabled(suite.ctx, addr)

	// check that the address is not disabled
	disabled = suite.app.VestingKeeper.HasGovClawbackDisabled(suite.ctx, addr)
	suite.Require().False(disabled, "expected address not to be found in store")
}

func (suite *KeeperTestSuite) TestGovClawbackNoOps() {
	suite.SetupTest()

	addr := sdk.AccAddress(suite.address.Bytes())
	addr2 := sdk.AccAddress(testutiltx.GenerateAddress().Bytes())

	// disable the address
	suite.app.VestingKeeper.SetGovClawbackDisabled(suite.ctx, addr)

	// a duplicate entry should not panic but no-op
	suite.app.VestingKeeper.SetGovClawbackDisabled(suite.ctx, addr)

	// check that the address is disabled
	disabled := suite.app.VestingKeeper.HasGovClawbackDisabled(suite.ctx, addr)
	suite.Require().True(disabled, "expected address to be found in store")

	// check that address 2 is not disabled
	disabled = suite.app.VestingKeeper.HasGovClawbackDisabled(suite.ctx, addr2)
	suite.Require().False(disabled, "expected address not to be found in store")

	// deleting a non-existent entry should not panic but no-op
	suite.app.VestingKeeper.DeleteGovClawbackDisabled(suite.ctx, addr2)

	// check that the address is still not disabled
	disabled = suite.app.VestingKeeper.HasGovClawbackDisabled(suite.ctx, addr2)
	suite.Require().False(disabled, "expected address not to be found in store")
}
