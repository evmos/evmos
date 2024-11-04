package keeper_test

import (
	"testing"

	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	testutiltx "github.com/evmos/evmos/v20/testutil/tx"
	"github.com/stretchr/testify/require"
)

func TestGovClawbackStore(t *testing.T) {
	nw := network.NewUnitTestNetwork()
	ctx := nw.GetContext()

	addr, _ := testutiltx.NewAccAddressAndKey()

	// check that the address is not disabled by default
	disabled := nw.App.VestingKeeper.HasGovClawbackDisabled(ctx, addr)
	require.False(t, disabled, "expected address not to be found in store")

	// disable the address
	nw.App.VestingKeeper.SetGovClawbackDisabled(ctx, addr)

	// check that the address is disabled
	disabled = nw.App.VestingKeeper.HasGovClawbackDisabled(ctx, addr)
	require.True(t, disabled, "expected address to be found in store")

	// delete the address
	nw.App.VestingKeeper.DeleteGovClawbackDisabled(ctx, addr)

	// check that the address is not disabled
	disabled = nw.App.VestingKeeper.HasGovClawbackDisabled(ctx, addr)
	require.False(t, disabled, "expected address not to be found in store")
}

func TestGovClawbackNoOps(t *testing.T) {
	nw := network.NewUnitTestNetwork()
	ctx := nw.GetContext()

	addr, _ := testutiltx.NewAccAddressAndKey()
	addr2, _ := testutiltx.NewAccAddressAndKey()

	// disable the address
	nw.App.VestingKeeper.SetGovClawbackDisabled(ctx, addr)

	// a duplicate entry should not panic but no-op
	nw.App.VestingKeeper.SetGovClawbackDisabled(ctx, addr)

	// check that the address is disabled
	disabled := nw.App.VestingKeeper.HasGovClawbackDisabled(ctx, addr)
	require.True(t, disabled, "expected address to be found in store")

	// check that address 2 is not disabled
	disabled = nw.App.VestingKeeper.HasGovClawbackDisabled(ctx, addr2)
	require.False(t, disabled, "expected address not to be found in store")

	// deleting a non-existent entry should not panic but no-op
	nw.App.VestingKeeper.DeleteGovClawbackDisabled(ctx, addr2)

	// check that the address is still not disabled
	disabled = nw.App.VestingKeeper.HasGovClawbackDisabled(ctx, addr2)
	require.False(t, disabled, "expected address not to be found in store")
}
