package keeper_test

import (
	"time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	testutiltx "github.com/evmos/evmos/v13/testutil/tx"
	"github.com/evmos/evmos/v13/x/vesting/keeper"
	v1vestingtypes "github.com/evmos/evmos/v13/x/vesting/migrations/types"
	vestingtypes "github.com/evmos/evmos/v13/x/vesting/types"
)

func (s *KeeperTestSuite) TestMigration() {
	s.SetupTest()

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
	s.app.AccountKeeper.SetAccount(s.ctx, oldAccount)

	foundAcc := s.app.AccountKeeper.GetAccount(s.ctx, vestingAddr)
	s.Require().NotNil(foundAcc, "vesting account not found")
	s.Require().IsType(&v1vestingtypes.ClawbackVestingAccount{}, foundAcc, "vesting account is not a v1 clawback vesting account")

	// migrate
	migrator := keeper.NewMigrator(s.app.VestingKeeper)
	err = migrator.Migrate1to2(s.ctx)
	s.Require().NoError(err, "migration failed")

	// check that the account is now a v2 base vesting account
	foundAcc = s.app.AccountKeeper.GetAccount(s.ctx, vestingAddr)
	s.Require().NotNil(foundAcc, "vesting account not found")
	s.Require().IsType(&vestingtypes.ClawbackVestingAccount{}, foundAcc, "vesting account is not a v2 base vesting account")
}
