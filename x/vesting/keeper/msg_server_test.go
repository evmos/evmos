package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/testutil"
	"github.com/tharsis/evmos/x/vesting/types"
)

var (
	balances       = sdk.NewCoins(sdk.NewInt64Coin("test", 1000))
	quarter        = sdk.NewCoins(sdk.NewInt64Coin("test", 250))
	addr           = sdk.AccAddress(tests.GenerateAddress().Bytes())
	addr2          = sdk.AccAddress(tests.GenerateAddress().Bytes())
	addr3          = sdk.AccAddress(tests.GenerateAddress().Bytes())
	addr4          = sdk.AccAddress(tests.GenerateAddress().Bytes())
	lockupPeriods  = []sdkvesting.Period{{Length: 5000, Amount: balances}}
	vestingPeriods = []sdkvesting.Period{
		{Length: 2000, Amount: quarter}, {Length: 2000, Amount: quarter},
		{Length: 2000, Amount: quarter}, {Length: 2000, Amount: quarter},
	}
)

func (suite *KeeperTestSuite) TestMsgCreateClawbackVestingAccount() {
	testCases := []struct {
		name               string
		malleate           func()
		from               sdk.AccAddress
		to                 sdk.AccAddress
		startTime          time.Time
		lockup             []sdkvesting.Period
		vesting            []sdkvesting.Period
		merge              bool
		expectExtraBalance int64
		expectPass         bool
	}{
		{
			"ok - new account",
			func() {},
			addr,
			addr2,
			time.Now(),
			lockupPeriods,
			vestingPeriods,
			false,
			0,
			true,
		},
		{
			"ok - new account - default lockup",
			func() {},
			addr,
			addr2,
			time.Now(),
			nil,
			vestingPeriods,
			false,
			0,
			true,
		},
		{
			"ok - new account - default vesting",
			func() {},
			addr,
			addr2,
			time.Now(),
			lockupPeriods,
			nil,
			false,
			0,
			true,
		},
		{
			"fail - different locking and vesting amounts",
			func() {},
			addr,
			addr2,
			time.Now(),
			[]sdkvesting.Period{
				{Length: 5000, Amount: quarter},
			},
			vestingPeriods,
			false,
			0,
			false,
		},
		{
			"fail - account exists - clawback but no merge",
			func() {
				// Existing clawback account
				vestingStart := s.ctx.BlockTime()
				baseAccount := authtypes.NewBaseAccountWithAddress(addr2)
				funder := sdk.AccAddress(types.ModuleName)
				clawbackAccount := types.NewClawbackVestingAccount(baseAccount, funder, balances, vestingStart, lockupPeriods, vestingPeriods)
				testutil.FundAccount(s.app.BankKeeper, s.ctx, addr2, balances)
				s.app.AccountKeeper.SetAccount(s.ctx, clawbackAccount)
			},
			addr,
			addr2,
			time.Now(),
			lockupPeriods,
			vestingPeriods,
			false,
			0,
			false,
		},
		{
			"fail - account exists - no clawback",
			func() {},
			addr,
			addr,
			time.Now(),
			lockupPeriods,
			vestingPeriods,
			false,
			0,
			false,
		},
		{
			"fail - account exists - merge but not clawback",
			func() {},
			addr,
			addr,
			time.Now(),
			lockupPeriods,
			vestingPeriods,
			true,
			0,
			false,
		},
		{
			"fail - account exists - wrong funder",
			func() {
				// Existing clawback account
				vestingStart := s.ctx.BlockTime()
				baseAccount := authtypes.NewBaseAccountWithAddress(addr2)
				funder := sdk.AccAddress(types.ModuleName)
				clawbackAccount := types.NewClawbackVestingAccount(baseAccount, funder, balances, vestingStart, lockupPeriods, vestingPeriods)
				testutil.FundAccount(s.app.BankKeeper, s.ctx, addr2, balances)
				s.app.AccountKeeper.SetAccount(s.ctx, clawbackAccount)
			},
			addr2,
			addr2,
			time.Now(),
			lockupPeriods,
			vestingPeriods,
			true,
			0,
			false,
		},
		{
			"ok - account exists - addGrant",
			func() {
				// Existing clawback account
				vestingStart := s.ctx.BlockTime()
				baseAccount := authtypes.NewBaseAccountWithAddress(addr2)
				funder := sdk.AccAddress(addr)
				clawbackAccount := types.NewClawbackVestingAccount(baseAccount, funder, balances, vestingStart, lockupPeriods, vestingPeriods)
				testutil.FundAccount(s.app.BankKeeper, s.ctx, addr2, balances)
				s.app.AccountKeeper.SetAccount(s.ctx, clawbackAccount)
			},
			addr,
			addr2,
			time.Now(),
			lockupPeriods,
			vestingPeriods,
			true,
			1000,
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // Reset
			ctx := sdk.WrapSDKContext(suite.ctx)

			tc.malleate()

			testutil.FundAccount(s.app.BankKeeper, s.ctx, addr, balances)

			msg := types.NewMsgCreateClawbackVestingAccount(
				tc.from,
				tc.to,
				tc.startTime,
				tc.lockup,
				tc.vesting,
				tc.merge,
			)
			res, err := suite.app.VestingKeeper.CreateClawbackVestingAccount(ctx, msg)

			expRes := &types.MsgCreateClawbackVestingAccountResponse{}
			balanceSource := suite.app.BankKeeper.GetBalance(suite.ctx, tc.from, "test")
			balanceDest := suite.app.BankKeeper.GetBalance(suite.ctx, tc.to, "test")

			if tc.expectPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expRes, res)

				accI := suite.app.AccountKeeper.GetAccount(suite.ctx, tc.to)
				suite.Require().NotNil(accI)
				suite.Require().IsType(&types.ClawbackVestingAccount{}, accI)
				suite.Require().Equal(sdk.NewInt64Coin("test", 0), balanceSource)
				suite.Require().Equal(sdk.NewInt64Coin("test", 1000+tc.expectExtraBalance), balanceDest)
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestMsgClawback() {
	testCases := []struct {
		name         string
		malleate     func()
		funder       sdk.AccAddress
		addr         sdk.AccAddress
		dest         sdk.AccAddress
		expectedPass bool
	}{
		{
			"no clawback account",
			func() {},
			addr,
			sdk.AccAddress(tests.GenerateAddress().Bytes()),
			addr3,
			false,
		},
		{
			"wrong account type",
			func() {
				baseAccount := authtypes.NewBaseAccountWithAddress(addr4)
				acc := sdkvesting.NewDelayedVestingAccount(baseAccount, balances, 500000)
				s.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			addr,
			addr4,
			addr3,
			false,
		},
		{
			"wrong funder",
			func() {},
			addr3,
			addr2,
			addr3,
			false,
		},
		{
			"pass",
			func() {
			},
			addr,
			addr2,
			addr3,
			true,
		},
		{
			"pass - without dest",
			func() {
			},
			addr,
			addr2,
			sdk.AccAddress([]byte{}),
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			ctx := sdk.WrapSDKContext(suite.ctx)

			// Set funder
			funder := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, tc.funder)
			suite.app.AccountKeeper.SetAccount(suite.ctx, funder)
			testutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr, balances)

			// Create Clawnback Vesting Account
			createMsg := types.NewMsgCreateClawbackVestingAccount(addr, addr2, suite.ctx.BlockTime(), lockupPeriods, vestingPeriods, false)
			createRes, err := suite.app.VestingKeeper.CreateClawbackVestingAccount(ctx, createMsg)
			suite.Require().NoError(err)
			suite.Require().NotNil(createRes)

			balanceDest := suite.app.BankKeeper.GetBalance(suite.ctx, addr2, "test")
			suite.Require().Equal(balanceDest, sdk.NewInt64Coin("test", 1000))

			tc.malleate()

			// Perform clawback
			msg := types.NewMsgClawback(tc.funder, tc.addr, tc.dest)
			res, err := suite.app.VestingKeeper.Clawback(ctx, msg)

			expRes := &types.MsgClawbackResponse{}
			balanceDest = suite.app.BankKeeper.GetBalance(suite.ctx, addr2, "test")
			balanceClaw := suite.app.BankKeeper.GetBalance(suite.ctx, tc.dest, "test")
			if len(tc.dest) == 0 {
				balanceClaw = suite.app.BankKeeper.GetBalance(suite.ctx, tc.funder, "test")
			}

			if tc.expectedPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
				suite.Require().Equal(sdk.NewInt64Coin("test", 0), balanceDest)
				suite.Require().Equal(balances[0], balanceClaw)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestClawbackVestingAccountStore() {
	suite.SetupTest()

	// Create and set clawback vesting account
	vestingStart := s.ctx.BlockTime()
	funder := sdk.AccAddress(types.ModuleName)
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())
	baseAccount := authtypes.NewBaseAccountWithAddress(addr)
	acc := types.NewClawbackVestingAccount(baseAccount, funder, balances, vestingStart, lockupPeriods, vestingPeriods)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	acc2 := suite.app.AccountKeeper.GetAccount(suite.ctx, acc.GetAddress())
	suite.Require().IsType(&types.ClawbackVestingAccount{}, acc2)
	suite.Require().Equal(acc.String(), acc2.String())
}

func (suite *KeeperTestSuite) TestClawbackVestingAccountMarshal() {
	suite.SetupTest()

	// Create and set clawback vesting account
	vestingStart := s.ctx.BlockTime()
	funder := sdk.AccAddress(types.ModuleName)
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())
	baseAccount := authtypes.NewBaseAccountWithAddress(addr)
	acc := types.NewClawbackVestingAccount(baseAccount, funder, balances, vestingStart, lockupPeriods, vestingPeriods)

	bz, err := suite.app.AccountKeeper.MarshalAccount(acc)
	suite.Require().NoError(err)

	acc2, err := suite.app.AccountKeeper.UnmarshalAccount(bz)
	suite.Require().NoError(err)
	suite.Require().IsType(&types.ClawbackVestingAccount{}, acc2)
	suite.Require().Equal(acc.String(), acc2.String())

	// error on bad bytes
	_, err = suite.app.AccountKeeper.UnmarshalAccount(bz[:len(bz)/2])
	suite.Require().Error(err)
}
