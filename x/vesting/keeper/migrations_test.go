package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	testutiltx "github.com/evmos/evmos/v19/testutil/tx"
	"github.com/evmos/evmos/v19/x/vesting/keeper"
	v1vestingtypes "github.com/evmos/evmos/v19/x/vesting/migrations/types"
	vestingtypes "github.com/evmos/evmos/v19/x/vesting/types"
)

func (suite *KeeperTestSuite) TestMigrate1to2() {
	if err := suite.SetupTest(); err != nil {
		panic(err)
	}

	// Create account addresses for testing
	vestingAddr, _ := testutiltx.NewAccAddressAndKey()
	funder, _ := testutiltx.NewAccAddressAndKey()

	// create a base vesting account instead of a clawback vesting account at the vesting address
	baseAccount := authtypes.NewBaseAccountWithAddress(vestingAddr)
	acc := sdkvesting.NewBaseVestingAccount(baseAccount, balances, 500000)

	oldAccount := &v1vestingtypes.ClawbackVestingAccount{
		BaseVestingAccount: acc,
		FunderAddress:      funder.String(),
		StartTime:          time.Now(),
		LockupPeriods:      lockupPeriods,
		VestingPeriods:     vestingPeriods,
	}
	suite.app.AccountKeeper.SetAccount(suite.ctx, oldAccount)

	foundAcc := suite.app.AccountKeeper.GetAccount(suite.ctx, vestingAddr)
	suite.Require().NotNil(foundAcc, "vesting account not found")
	suite.Require().IsType(&v1vestingtypes.ClawbackVestingAccount{}, foundAcc, "vesting account is not a v1 clawback vesting account")

	// migrate
	migrator := keeper.NewMigrator(suite.app.VestingKeeper)
	err = migrator.Migrate1to2(suite.ctx)
	suite.Require().NoError(err, "migration failed")

	// check that the account is now a v2 base vesting account
	foundAcc = suite.app.AccountKeeper.GetAccount(suite.ctx, vestingAddr)
	suite.Require().NotNil(foundAcc, "vesting account not found")
	suite.Require().IsType(&vestingtypes.ClawbackVestingAccount{}, foundAcc, "vesting account is not a v2 base vesting account")
}

func (suite *KeeperTestSuite) TestMigrate2to3() {
	var emtpyCoins types.Coins
	if err := suite.SetupTest(); err != nil {
		panic(err)
	}

	// Create account addresses for testing
	vestingAddr, _ := testutiltx.NewAccAddressAndKey()
	funder, _ := testutiltx.NewAccAddressAndKey()

	// create a base vesting account instead of a clawback vesting account at the vesting address
	baseAccount := authtypes.NewBaseAccountWithAddress(vestingAddr)
	acc := sdkvesting.NewBaseVestingAccount(baseAccount, balances, 500000)

	testCases := []struct {
		name                    string
		initialDelegatedVesting types.Coins
		initialDelegatedFree    types.Coins
		expectedDelegatedFree   types.Coins
	}{
		{
			name:                    "delegated vesting > 0 and delegated free == 0",
			initialDelegatedVesting: quarter,
			initialDelegatedFree:    emtpyCoins,
			expectedDelegatedFree:   quarter,
		},
		{
			name:                    "delegated vesting > 0 and delegated free > 0",
			initialDelegatedVesting: quarter,
			initialDelegatedFree:    quarter,
			expectedDelegatedFree:   types.NewCoins(types.NewInt64Coin("test", 500)),
		},
		{
			name:                    "delegated vesting == 0 and delegated free > 0",
			initialDelegatedVesting: emtpyCoins,
			initialDelegatedFree:    quarter,
			expectedDelegatedFree:   quarter,
		},
		{
			name:                    "delegated vesting == 0 and delegated free == 0",
			initialDelegatedVesting: emtpyCoins,
			initialDelegatedFree:    emtpyCoins,
			expectedDelegatedFree:   emtpyCoins,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			acc.DelegatedVesting = tc.initialDelegatedVesting
			acc.DelegatedFree = tc.initialDelegatedFree

			vestAcc := &vestingtypes.ClawbackVestingAccount{
				BaseVestingAccount: acc,
				FunderAddress:      funder.String(),
				StartTime:          time.Now(),
				LockupPeriods:      lockupPeriods,
				VestingPeriods:     vestingPeriods,
			}
			suite.app.AccountKeeper.SetAccount(suite.ctx, vestAcc)

			// check account was created successfully
			foundAcc := suite.app.AccountKeeper.GetAccount(suite.ctx, vestingAddr)
			suite.Require().NotNil(foundAcc, "vesting account not found")
			vestAcc, ok := foundAcc.(*vestingtypes.ClawbackVestingAccount)
			suite.Require().True(ok)
			suite.Require().Equal(tc.initialDelegatedVesting, vestAcc.DelegatedVesting)
			suite.Require().Equal(tc.initialDelegatedFree, vestAcc.DelegatedFree)

			// migrate
			migrator := keeper.NewMigrator(suite.app.VestingKeeper)
			err = migrator.Migrate2to3(suite.ctx)
			suite.Require().NoError(err, "migration failed")

			// check that the account delegated vesting coins were migrated
			// to the delegated free coins
			foundAcc = suite.app.AccountKeeper.GetAccount(suite.ctx, vestingAddr)
			suite.Require().NotNil(foundAcc, "vesting account not found")
			vestAcc, ok = foundAcc.(*vestingtypes.ClawbackVestingAccount)
			suite.Require().True(ok)
			suite.Require().Equal(emtpyCoins, vestAcc.DelegatedVesting)
			suite.Require().Equal(tc.expectedDelegatedFree, vestAcc.DelegatedFree)
		})
	}
}
