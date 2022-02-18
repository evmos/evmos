package keeper_test

// import (
// 	"testing"

// 	"github.com/stretchr/testify/suite"
// 	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

// 	"github.com/cosmos/cosmos-sdk/simapp"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
// 	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
// )

// type HandlerTestSuite struct {
// 	suite.Suite

// 	handler sdk.Handler
// 	app     *simapp.SimApp
// }

// func (suite *HandlerTestSuite) SetupTest() {
// 	checkTx := false
// 	app := simapp.Setup(checkTx)

// 	suite.handler = vesting.NewHandler(app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
// 	suite.app = app
// }

// func (suite *HandlerTestSuite) TestMsgCreateClawbackVestingAccount() {
// 	ctx := suite.app.BaseApp.NewContext(false, tmproto.Header{Height: suite.app.LastBlockHeight() + 1})

// 	balances := sdk.NewCoins(sdk.NewInt64Coin("test", 1000))
// 	quarter := sdk.NewCoins(sdk.NewInt64Coin("test", 250))
// 	addr1 := sdk.AccAddress([]byte("addr1_______________"))
// 	addr2 := sdk.AccAddress([]byte("addr2_______________"))
// 	addr3 := sdk.AccAddress([]byte("addr3_______________"))
// 	addr4 := sdk.AccAddress([]byte("addr4_______________"))

// 	acc1 := suite.app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
// 	suite.app.AccountKeeper.SetAccount(ctx, acc1)

// 	lockupPeriods := []types.Period{{Length: 5000, Amount: balances}}
// 	vestingPeriods := []types.Period{
// 		{Length: 2000, Amount: quarter}, {Length: 2000, Amount: quarter},
// 		{Length: 2000, Amount: quarter}, {Length: 2000, Amount: quarter},
// 	}

// 	testCases := []struct {
// 		name               string
// 		msg                *types.MsgCreateClawbackVestingAccount
// 		expectErr          bool
// 		expectExtraBalance int64
// 	}{
// 		{
// 			name:      "create clawback vesting account",
// 			msg:       types.NewMsgCreateClawbackVestingAccount(addr1, addr2, 0, lockupPeriods, vestingPeriods, false),
// 			expectErr: false,
// 		},
// 		{
// 			name: "bad from addr",
// 			msg: &types.MsgCreateClawbackVestingAccount{
// 				FromAddress:    "foo",
// 				ToAddress:      addr2.String(),
// 				StartTime:      0,
// 				LockupPeriods:  lockupPeriods,
// 				VestingPeriods: vestingPeriods,
// 			},
// 			expectErr: true,
// 		},
// 		{
// 			name: "bad to addr",
// 			msg: &types.MsgCreateClawbackVestingAccount{
// 				FromAddress:    addr1.String(),
// 				ToAddress:      "foo",
// 				StartTime:      0,
// 				LockupPeriods:  lockupPeriods,
// 				VestingPeriods: vestingPeriods,
// 			},
// 			expectErr: true,
// 		},
// 		{
// 			name:      "default lockup",
// 			msg:       types.NewMsgCreateClawbackVestingAccount(addr1, addr3, 0, nil, vestingPeriods, false),
// 			expectErr: false,
// 		},
// 		{
// 			name:      "default vesting",
// 			msg:       types.NewMsgCreateClawbackVestingAccount(addr1, addr4, 0, lockupPeriods, nil, false),
// 			expectErr: false,
// 		},
// 		{
// 			name:      "different amounts",
// 			msg:       types.NewMsgCreateClawbackVestingAccount(addr1, addr2, 0, []types.Period{{Length: 5000, Amount: quarter}}, vestingPeriods, false),
// 			expectErr: true,
// 		},
// 		{
// 			name:      "account exists",
// 			msg:       types.NewMsgCreateClawbackVestingAccount(addr1, addr1, 0, lockupPeriods, vestingPeriods, false),
// 			expectErr: true,
// 		},
// 		{
// 			name:      "account exists no merge",
// 			msg:       types.NewMsgCreateClawbackVestingAccount(addr1, addr2, 0, lockupPeriods, vestingPeriods, false),
// 			expectErr: true,
// 		},
// 		{
// 			name:      "account exists merge not clawback",
// 			msg:       types.NewMsgCreateClawbackVestingAccount(addr1, addr1, 0, lockupPeriods, vestingPeriods, true),
// 			expectErr: true,
// 		},
// 		{
// 			name:               "account exists merge",
// 			msg:                types.NewMsgCreateClawbackVestingAccount(addr1, addr2, 0, lockupPeriods, vestingPeriods, true),
// 			expectErr:          false,
// 			expectExtraBalance: 1000,
// 		},
// 		{
// 			name:      "merge wrong funder",
// 			msg:       types.NewMsgCreateClawbackVestingAccount(addr2, addr2, 0, lockupPeriods, vestingPeriods, true),
// 			expectErr: true,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		suite.Run(tc.name, func() {
// 			if !tc.expectErr {
// 				suite.Require().NoError(simapp.FundAccount(suite.app.BankKeeper, ctx, addr1, balances))
// 			}
// 			res, err := suite.handler(ctx, tc.msg)
// 			if tc.expectErr {
// 				suite.Require().Error(err)
// 			} else {
// 				suite.Require().NoError(err)
// 				suite.Require().NotNil(res)

// 				toAddr, err := sdk.AccAddressFromBech32(tc.msg.ToAddress)

// 				suite.Require().NoError(err)
// 				fromAddr, err := sdk.AccAddressFromBech32(tc.msg.FromAddress)
// 				suite.Require().NoError(err)

// 				accI := suite.app.AccountKeeper.GetAccount(ctx, toAddr)
// 				suite.Require().NotNil(accI)
// 				suite.Require().IsType(&types.ClawbackVestingAccount{}, accI)
// 				balanceSource := suite.app.BankKeeper.GetBalance(ctx, fromAddr, "test")
// 				suite.Require().Equal(balanceSource, sdk.NewInt64Coin("test", 0))
// 				balanceDest := suite.app.BankKeeper.GetBalance(ctx, toAddr, "test")
// 				suite.Require().Equal(balanceDest, sdk.NewInt64Coin("test", 1000+tc.expectExtraBalance))

// 			}
// 		})
// 	}
// }

// func (suite *HandlerTestSuite) TestMsgClawback() {
// 	ctx := suite.app.BaseApp.NewContext(false, tmproto.Header{Height: suite.app.LastBlockHeight() + 1})

// 	balances := sdk.NewCoins(sdk.NewInt64Coin("test", 1000))
// 	quarter := sdk.NewCoins(sdk.NewInt64Coin("test", 250))
// 	addr1 := sdk.AccAddress([]byte("addr1_______________"))
// 	addr2 := sdk.AccAddress([]byte("addr2_______________"))
// 	addr3 := sdk.AccAddress([]byte("addr3_______________"))
// 	addr4 := sdk.AccAddress([]byte("addr4_______________"))

// 	funder := suite.app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
// 	suite.app.AccountKeeper.SetAccount(ctx, funder)
// 	suite.Require().NoError(simapp.FundAccount(suite.app.BankKeeper, ctx, addr1, balances))

// 	lockupPeriods := []types.Period{{Length: 5000, Amount: balances}}
// 	vestingPeriods := []types.Period{
// 		{Length: 2000, Amount: quarter}, {Length: 2000, Amount: quarter},
// 		{Length: 2000, Amount: quarter}, {Length: 2000, Amount: quarter},
// 	}

// 	createMsg := types.NewMsgCreateClawbackVestingAccount(addr1, addr2, 0, lockupPeriods, vestingPeriods, false)
// 	res, err := suite.handler(ctx, createMsg)
// 	suite.Require().NoError(err)
// 	suite.Require().NotNil(res)

// 	balanceDest := suite.app.BankKeeper.GetBalance(ctx, addr2, "test")
// 	suite.Require().Equal(balanceDest, sdk.NewInt64Coin("test", 1000))

// 	clawbackMsg := types.NewMsgClawback(addr1, addr2, addr3)
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
// 		FunderAddress: addr1.String(),
// 		Address:       "foo",
// 		DestAddress:   addr3.String(),
// 	}
// 	_, err = suite.handler(ctx, clawbackMsg)
// 	suite.Require().Error(err)

// 	// bad dest
// 	clawbackMsg = &types.MsgClawback{
// 		FunderAddress: addr1.String(),
// 		Address:       addr2.String(),
// 		DestAddress:   "foo",
// 	}
// 	_, err = suite.handler(ctx, clawbackMsg)
// 	suite.Require().Error(err)

// 	// no account
// 	clawbackMsg = types.NewMsgClawback(addr1, addr4, addr3)
// 	_, err = suite.handler(ctx, clawbackMsg)
// 	suite.Require().Error(err)

// 	// wrong account type
// 	clawbackMsg = types.NewMsgClawback(addr1, addr3, addr4)
// 	_, err = suite.handler(ctx, clawbackMsg)
// 	suite.Require().Error(err)

// 	// wrong funder
// 	clawbackMsg = types.NewMsgClawback(addr4, addr2, addr3)
// 	_, err = suite.handler(ctx, clawbackMsg)
// 	suite.Require().Error(err)

// }

// func TestHandlerTestSuite(t *testing.T) {
// 	suite.Run(t, new(HandlerTestSuite))
// }
