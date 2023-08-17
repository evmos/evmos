package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutiltx "github.com/evmos/evmos/v14/testutil/tx"
)

func (s *KeeperTestSuite) TestGovClawbackStore() {
	s.SetupTest()

	addr := sdk.AccAddress(s.address.Bytes())

	// check that the address is not disabled by default
	disabled := s.app.VestingKeeper.HasGovClawbackDisabled(s.ctx, addr)
	s.Require().False(disabled, "expected address not to be found in store")

	// disable the address
	s.app.VestingKeeper.SetGovClawbackDisabled(s.ctx, addr)

	// check that the address is disabled
	disabled = s.app.VestingKeeper.HasGovClawbackDisabled(s.ctx, addr)
	s.Require().True(disabled, "expected address to be found in store")

	// delete the address
	s.app.VestingKeeper.DeleteGovClawbackDisabled(s.ctx, addr)

	// check that the address is not disabled
	disabled = s.app.VestingKeeper.HasGovClawbackDisabled(s.ctx, addr)
	s.Require().False(disabled, "expected address not to be found in store")
}

func (s *KeeperTestSuite) TestGovClawbackNoOps() {
	s.SetupTest()

	addr := sdk.AccAddress(s.address.Bytes())
	addr2 := sdk.AccAddress(testutiltx.GenerateAddress().Bytes())

	// disable the address
	s.app.VestingKeeper.SetGovClawbackDisabled(s.ctx, addr)

	// a duplicate entry should not panic but no-op
	s.app.VestingKeeper.SetGovClawbackDisabled(s.ctx, addr)

	// check that the address is disabled
	disabled := s.app.VestingKeeper.HasGovClawbackDisabled(s.ctx, addr)
	s.Require().True(disabled, "expected address to be found in store")

	// check that address 2 is not disabled
	disabled = s.app.VestingKeeper.HasGovClawbackDisabled(s.ctx, addr2)
	s.Require().False(disabled, "expected address not to be found in store")

	// deleting a non-existent entry should not panic but no-op
	s.app.VestingKeeper.DeleteGovClawbackDisabled(s.ctx, addr2)

	// check that the address is still not disabled
	disabled = s.app.VestingKeeper.HasGovClawbackDisabled(s.ctx, addr2)
	s.Require().False(disabled, "expected address not to be found in store")
}
