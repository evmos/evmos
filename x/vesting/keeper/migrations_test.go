package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	testutiltx "github.com/evmos/evmos/v19/testutil/tx"
	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/vesting/keeper"
	v1vestingtypes "github.com/evmos/evmos/v19/x/vesting/migrations/types"
	vestingtypes "github.com/evmos/evmos/v19/x/vesting/types"
	"github.com/stretchr/testify/require"
)

func TestMigrate1to2(t *testing.T) {
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

func TestMigrate2to3(t *testing.T) {
	var emptyCoins types.Coins
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

	testCases := []struct {
		name                    string
		initialDelegatedVesting types.Coins
		initialDelegatedFree    types.Coins
		expectedDelegatedFree   types.Coins
	}{
		{
			name:                    "delegated vesting > 0 and delegated free == 0",
			initialDelegatedVesting: quarter,
			initialDelegatedFree:    emptyCoins,
			expectedDelegatedFree:   quarter,
		},
		{
			name:                    "delegated vesting > 0 and delegated free > 0",
			initialDelegatedVesting: quarter,
			initialDelegatedFree:    quarter,
			expectedDelegatedFree:   types.NewCoins(types.NewInt64Coin(utils.BaseDenom, 500)),
		},
		{
			name:                    "delegated vesting == 0 and delegated free > 0",
			initialDelegatedVesting: emptyCoins,
			initialDelegatedFree:    quarter,
			expectedDelegatedFree:   quarter,
		},
		{
			name:                    "delegated vesting == 0 and delegated free == 0",
			initialDelegatedVesting: emptyCoins,
			initialDelegatedFree:    emptyCoins,
			expectedDelegatedFree:   emptyCoins,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			acc.DelegatedVesting = tc.initialDelegatedVesting
			acc.DelegatedFree = tc.initialDelegatedFree

			vestAcc := &vestingtypes.ClawbackVestingAccount{
				BaseVestingAccount: acc,
				FunderAddress:      funder.String(),
				StartTime:          time.Now(),
				LockupPeriods:      lockupPeriods,
				VestingPeriods:     vestingPeriods,
			}
			nw.App.AccountKeeper.SetAccount(ctx, vestAcc)

			// check account was created successfully
			foundAcc := nw.App.AccountKeeper.GetAccount(ctx, vestingAddr)
			require.NotNil(t, foundAcc, "vesting account not found")
			vestAcc, ok := foundAcc.(*vestingtypes.ClawbackVestingAccount)
			require.True(t, ok)
			require.Equal(t, tc.initialDelegatedVesting, vestAcc.DelegatedVesting)
			require.Equal(t, tc.initialDelegatedFree, vestAcc.DelegatedFree)

			// migrate
			migrator := keeper.NewMigrator(nw.App.VestingKeeper)
			err = migrator.Migrate2to3(ctx)
			require.NoError(t, err, "migration failed")

			// check that the account delegated vesting coins were migrated
			// to the delegated free coins
			foundAcc = nw.App.AccountKeeper.GetAccount(ctx, vestingAddr)
			require.NotNil(t, foundAcc, "vesting account not found")
			vestAcc, ok = foundAcc.(*vestingtypes.ClawbackVestingAccount)
			require.True(t, ok)
			require.Equal(t, emptyCoins, vestAcc.DelegatedVesting)
			require.Equal(t, tc.expectedDelegatedFree, vestAcc.DelegatedFree)
		})
	}
}
