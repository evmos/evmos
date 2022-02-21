package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/tharsis/evmos/testutil"
	"github.com/tharsis/evmos/x/vesting/types"
)

func (suite *KeeperTestSuite) TestMsgCreateClawbackVestingAccount() {
	balances := sdk.NewCoins(sdk.NewInt64Coin("test", 1000))
	quarter := sdk.NewCoins(sdk.NewInt64Coin("test", 250))
	// addr1 := sdk.AccAddress(tests.GenerateAddress().Bytes())
	// addr2 := sdk.AccAddress(tests.GenerateAddress().Bytes())
	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	// addr3 := sdk.AccAddress([]byte("addr3_______________"))
	// addr4 := sdk.AccAddress([]byte("addr4_______________"))

	acc1 := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc1)

	lockupPeriods := []sdkvesting.Period{{Length: 5000, Amount: balances}}
	vestingPeriods := []sdkvesting.Period{
		{Length: 2000, Amount: quarter}, {Length: 2000, Amount: quarter},
		{Length: 2000, Amount: quarter}, {Length: 2000, Amount: quarter},
	}

	testCases := []struct {
		name       string
		from       sdk.AccAddress
		to         sdk.AccAddress
		startTime  int64
		lockup     []sdkvesting.Period
		vesting    []sdkvesting.Period
		merge      bool
		expectPass bool
	}{
		{"ok", addr1, addr2, 0, lockupPeriods, vestingPeriods, false, true},
		// {
		// 	name: "bad from addr",
		// 	msg: &types.MsgCreateClawbackVestingAccount{
		// 		FromAddress:    "foo",
		// 		ToAddress:      addr2.String(),
		// 		StartTime:      0,
		// 		LockupPeriods:  lockupPeriods,
		// 		VestingPeriods: vestingPeriods,
		// 	},
		// 	true,
		// },
		// {
		// 	name: "bad to addr",
		// 	msg: &types.MsgCreateClawbackVestingAccount{
		// 		FromAddress:    addr.String(),
		// 		ToAddress:      "foo",
		// 		StartTime:      0,
		// 		LockupPeriods:  lockupPeriods,
		// 		VestingPeriods: vestingPeriods,
		// 	},
		// 	true,
		// },
		// {
		// 	"default lockup",
		// 	types.NewMsgCreateClawbackVestingAccount(addr, addr3, 0, nil, vestingPeriods, false),
		// 	false,
		// },
		// {
		// 	"default vesting",
		// 	types.NewMsgCreateClawbackVestingAccount(addr, addr4, 0, lockupPeriods, nil, false),
		// 	false,
		// },
		// {
		// 	"different amounts",
		// 	types.NewMsgCreateClawbackVestingAccount(addr, addr2, 0, []sdkvesting.Period{{Length: 5000, Amount: quarter}}, vestingPeriods, false),
		// 	true,
		// },
		// {
		// 	"account exists",
		// 	types.NewMsgCreateClawbackVestingAccount(addr, addr, 0, lockupPeriods, vestingPeriods, false),
		// 	true,
		// },
		// {
		// 	"account exists no merge",
		// 	types.NewMsgCreateClawbackVestingAccount(addr, addr2, 0, lockupPeriods, vestingPeriods, false),
		// 	true,
		// },
		// {
		// 	"account exists merge not clawback",
		// 	types.NewMsgCreateClawbackVestingAccount(addr, addr, 0, lockupPeriods, vestingPeriods, true),
		// 	true,
		// },
		// {
		// 	"account exists merge",
		// 	types.NewMsgCreateClawbackVestingAccount(addr, addr2, 0, lockupPeriods, vestingPeriods, true),
		// 	false,
		// 	expectExtraBalance: 1000,
		// },
		// {
		// 	"merge wrong funder",
		// 	types.NewMsgCreateClawbackVestingAccount(addr2, addr2, 0, lockupPeriods, vestingPeriods, true),
		// 	true,
		// },
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			ctx := sdk.WrapSDKContext(suite.ctx)

			// now := tmtime.Now()
			// endTime := now.Add(24 * time.Hour)
			// addr := sdk.AccAddress(tests.GenerateAddress().Bytes())
			// bacc := authtypes.NewBaseAccountWithAddress(addr)
			// va := types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)
			// suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

			suite.Require().NoError(testutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, balances))

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
			// balanceDest := suite.app.BankKeeper.GetBalance(suite.ctx, tc.to, "test")

			if tc.expectPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expRes, res)

				accI := suite.app.AccountKeeper.GetAccount(suite.ctx, tc.to)
				suite.Require().NotNil(accI)
				suite.Require().IsType(&types.ClawbackVestingAccount{}, accI)
				suite.Require().Equal(balanceSource, sdk.NewInt64Coin("test", 0))
				// suite.Require().Equal(balanceDest, sdk.NewInt64Coin("test", 1000+tc.expectExtraBalance))
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

// func (suite *KeeperTestSuite) TestMsgClawback() {
// 	ctx := suite.app.BaseApp.NewContext(false, tmproto.Header{Height: suite.app.LastBlockHeight() + 1})

// 	balances := sdk.NewCoins(sdk.NewInt64Coin("test", 1000))
// 	quarter := sdk.NewCoins(sdk.NewInt64Coin("test", 250))
// 	addr := sdk.AccAddress([]byte("addr_______________"))
// 	addr2 := sdk.AccAddress([]byte("addr2_______________"))
// 	addr3 := sdk.AccAddress([]byte("addr3_______________"))
// 	addr4 := sdk.AccAddress([]byte("addr4_______________"))

// 	funder := suite.app.AccountKeeper.NewAccountWithAddress(ctx, addr)
// 	suite.app.AccountKeeper.SetAccount(ctx, funder)
// 	suite.Require().NoError(simapp.FundAccount(suite.app.BankKeeper, ctx, addr, balances))

// 	lockupPeriods := []types.Period{{Length: 5000, Amount: balances}}
// 	vestingPeriods := []types.Period{
// 		{Length: 2000, Amount: quarter}, {Length: 2000, Amount: quarter},
// 		{Length: 2000, Amount: quarter}, {Length: 2000, Amount: quarter},
// 	}

// 	createMsg := types.NewMsgCreateClawbackVestingAccount(addr, addr2, 0, lockupPeriods, vestingPeriods, false)
// 	res, err := suite.handler(ctx, createMsg)
// 	suite.Require().NoError(err)
// 	suite.Require().NotNil(res)

// 	balanceDest := suite.app.BankKeeper.GetBalance(ctx, addr2, "test")
// 	suite.Require().Equal(balanceDest, sdk.NewInt64Coin("test", 1000))

// 	clawbackMsg := types.NewMsgClawback(addr, addr2, addr3)
// 	clawRes, err := suite.handler(ctx, clawbackMsg)
// 	suite.Require().NoError(err)
// 	suite.Require().NotNil(clawRes)

// 	balanceDest = suite.app.BankKeeper.GetBalance(ctx, addr2, "test")
// 	suite.Require().Equal(balanceDest, sdk.NewInt64Coin("test", 0))
// 	balanceClaw := suite.app.BankKeeper.GetBalance(ctx, addr3, "test")
// 	suite.Require().Equal(balanceClaw, sdk.NewInt64Coin("test", 1000))

// 	// test bad messages

// 	// bad funder
// 	clawbackMsg = &types.MsgClawback{
// 		FunderAddress: "foo",
// 		Address:       addr2.String(),
// 		DestAddress:   addr3.String(),
// 	}
// 	_, err = suite.handler(ctx, clawbackMsg)
// 	suite.Require().Error(err)

// 	// bad addr
// 	clawbackMsg = &types.MsgClawback{
// 		FunderAddress: addr.String(),
// 		Address:       "foo",
// 		DestAddress:   addr3.String(),
// 	}
// 	_, err = suite.handler(ctx, clawbackMsg)
// 	suite.Require().Error(err)

// 	// bad dest
// 	clawbackMsg = &types.MsgClawback{
// 		FunderAddress: addr.String(),
// 		Address:       addr2.String(),
// 		DestAddress:   "foo",
// 	}
// 	_, err = suite.handler(ctx, clawbackMsg)
// 	suite.Require().Error(err)

// 	// no account
// 	clawbackMsg = types.NewMsgClawback(addr, addr4, addr3)
// 	_, err = suite.handler(ctx, clawbackMsg)
// 	suite.Require().Error(err)

// 	// wrong account type
// 	clawbackMsg = types.NewMsgClawback(addr, addr3, addr4)
// 	_, err = suite.handler(ctx, clawbackMsg)
// 	suite.Require().Error(err)

// 	// wrong funder
// 	clawbackMsg = types.NewMsgClawback(addr4, addr2, addr3)
// 	_, err = suite.handler(ctx, clawbackMsg)
// 	suite.Require().Error(err)
// }
