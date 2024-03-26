package keeper_test

import (
	"testing"
	"time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	testutiltx "github.com/evmos/evmos/v16/testutil/tx"
	"github.com/evmos/evmos/v16/x/vesting/keeper"
	v1vestingtypes "github.com/evmos/evmos/v16/x/vesting/migrations/types"
	vestingtypes "github.com/evmos/evmos/v16/x/vesting/types"
	"github.com/stretchr/testify/require"
)

func TestMigration(t *testing.T) {
	nw := network.NewUnitTestNetwork()
	ctx := nw.GetContext()

	// Create account addresses for testing
	vestingAddr, _ := testutiltx.NewAccAddressAndKey()
	funder, _ := testutiltx.NewAccAddressAndKey()

	// create a base vesting account instead of a clawback vesting account at the vesting address
	baseAccount := authtypes.NewBaseAccountWithAddress(vestingAddr)
	baseAccount.AccountNumber = nw.App.AccountKeeper.NextAccountNumber(ctx)
	acc, err := sdkvesting.NewBaseVestingAccount(baseAccount, balances, 500000)
	require.NoError(t, err)

	oldAccount := &v1vestingtypes.ClawbackVestingAccount{
		BaseVestingAccount: acc,
		FunderAddress:      funder.String(),
		StartTime:          time.Now(),
		LockupPeriods:      lockupPeriods,
		VestingPeriods:     vestingPeriods,
	}
	nw.App.AccountKeeper.SetAccount(ctx, oldAccount)

	foundAcc := nw.App.AccountKeeper.GetAccount(ctx, vestingAddr)
	require.NotNil(t, foundAcc, "vesting account not found")
	require.IsType(t, &v1vestingtypes.ClawbackVestingAccount{}, foundAcc, "vesting account is not a v1 clawback vesting account")

	// migrate
	migrator := keeper.NewMigrator(nw.App.VestingKeeper)
	err = migrator.Migrate1to2(ctx)
	require.NoError(t, err, "migration failed")

	// check that the account is now a v2 base vesting account
	foundAcc = nw.App.AccountKeeper.GetAccount(ctx, vestingAddr)
	require.NotNil(t, foundAcc, "vesting account not found")
	require.IsType(t, &vestingtypes.ClawbackVestingAccount{}, foundAcc, "vesting account is not a v2 base vesting account")
}
