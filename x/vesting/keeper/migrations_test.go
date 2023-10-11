package keeper_test

import (
	"time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	testutiltx "github.com/evmos/evmos/v14/testutil/tx"
	"github.com/evmos/evmos/v14/x/vesting/keeper"
	v1vestingtypes "github.com/evmos/evmos/v14/x/vesting/migrations/types"
	vestingtypes "github.com/evmos/evmos/v14/x/vesting/types"
)

func (suite *KeeperTestSuite) TestMigration() {
	suite.SetupTest()

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
