package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/stretchr/testify/suite"

	utiltx "github.com/evmos/evmos/v12/testutil/tx"
	"github.com/evmos/evmos/v12/x/vesting/types"
)

type MsgsTestSuite struct {
	suite.Suite
}

func TestMsgsTestSuite(t *testing.T) {
	suite.Run(t, new(MsgsTestSuite))
}

func (suite *MsgsTestSuite) TestMsgCreateClawbackVestingAccountGetters() {
	msgInvalid := types.MsgCreateClawbackVestingAccount{}
	msg := types.NewMsgCreateClawbackVestingAccount(
		sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
		sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
		time.Unix(100200300, 0),
		sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
		sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
		true,
	)
	suite.Require().Equal(types.RouterKey, msg.Route())
	suite.Require().Equal(types.TypeMsgCreateClawbackVestingAccount, msg.Type())
	suite.Require().NotNil(msgInvalid.GetSignBytes())
	suite.Require().NotNil(msg.GetSigners())
}

func (suite *MsgsTestSuite) TestMsgCreateClawbackVestingAccountNew() {
	testCases := []struct {
		msg            string
		from           sdk.AccAddress
		to             sdk.AccAddress
		startTime      time.Time
		lockupPeriods  sdkvesting.Periods
		vestingPeriods sdkvesting.Periods
		merge          bool
		expectPass     bool
	}{
		{
			"msg create clawback vesting account - pass",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			true,
			true,
		},
	}

	for i, tc := range testCases {
		tx := types.NewMsgCreateClawbackVestingAccount(
			tc.from,
			tc.to,
			tc.startTime,
			tc.lockupPeriods,
			tc.vestingPeriods,
			tc.merge,
		)
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
		}
	}
}

func (suite *MsgsTestSuite) TestMsgCreateClawbackVestingAccount() {
	testCases := []struct {
		msg            string
		from           string
		to             string
		startTime      time.Time
		lockupPeriods  sdkvesting.Periods
		vestingPeriods sdkvesting.Periods
		merge          bool
		expectPass     bool
	}{
		{
			"msg create clawback vesting account - invalid from address",
			"foo",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			true,
			false,
		},
		{
			"msg create clawback vesting account - invalid to address",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			"foo",
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			true,
			false,
		},
		{
			"msg create clawback vesting account - invalid lockup period length",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 0, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			true,
			false,
		},
		{
			"msg create clawback vesting account - invalid lockup period amount",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 0)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			true,
			false,
		},
		{
			"msg create clawback vesting account - invalid vesting period length",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 0, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			true,
			false,
		},
		{
			"msg create clawback vesting account - invalid vesting period amount",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 0)}}},
			true,
			false,
		},
		{
			"msg create clawback vesting account - pass",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			true,
			true,
		},
	}

	for i, tc := range testCases {
		tx := types.MsgCreateClawbackVestingAccount{
			tc.from,
			tc.to,
			tc.startTime,
			tc.lockupPeriods,
			tc.vestingPeriods,
			tc.merge,
		}
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
		}
	}
}

func (suite *MsgsTestSuite) TestMsgClawbackGetters() {
	msgInvalid := types.MsgClawback{}
	msg := types.NewMsgClawback(
		sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
		sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
		sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
	)
	suite.Require().Equal(types.RouterKey, msg.Route())
	suite.Require().Equal(types.TypeMsgClawback, msg.Type())
	suite.Require().NotNil(msgInvalid.GetSignBytes())
	suite.Require().NotNil(msg.GetSigners())
}

func (suite *MsgsTestSuite) TestMsgClawbackNew() {
	testCases := []struct {
		msg        string
		funder     sdk.AccAddress
		addr       sdk.AccAddress
		dest       sdk.AccAddress
		expectPass bool
	}{
		{
			"msg clawback - pass",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			true,
		},
	}

	for i, tc := range testCases {
		tx := types.NewMsgClawback(
			tc.funder,
			tc.addr,
			tc.dest,
		)
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
		}
	}
}

func (suite *MsgsTestSuite) TestMsgClawback() {
	testCases := []struct {
		msg        string
		funder     string
		addr       string
		dest       string
		expectPass bool
	}{
		{
			"msg create clawback vesting account - invalid fund address",
			"foo",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			false,
		},
		{
			"msg create clawback vesting account - invalid addr address",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			"foo",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			false,
		},
		{
			"msg create clawback vesting account - invalid dest address",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			"foo",
			false,
		},
		{
			"msg create clawback vesting account - pass empty dest address",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			"",
			true,
		},
		{
			"msg create clawback vesting account - pass",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			true,
		},
	}

	for i, tc := range testCases {
		tx := types.MsgClawback{
			tc.funder,
			tc.addr,
			tc.dest,
		}
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
		}
	}
}

func (suite *MsgsTestSuite) TestMsgUpdateVestingFunderGetters() {
	msgInvalid := types.MsgUpdateVestingFunder{}
	msg := types.NewMsgUpdateVestingFunder(
		sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
		sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
		sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
	)
	suite.Require().Equal(types.RouterKey, msg.Route())
	suite.Require().Equal(types.TypeMsgUpdateVestingFunder, msg.Type())
	suite.Require().NotNil(msgInvalid.GetSignBytes())
	suite.Require().NotNil(msg.GetSigners())
}

func (suite *MsgsTestSuite) TestMsgUpdateVestingFunder() {
	var (
		funder     = sdk.AccAddress(utiltx.GenerateAddress().Bytes())
		newFunder  = sdk.AccAddress(utiltx.GenerateAddress().Bytes())
		vestingAcc = sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	)

	testCases := []struct {
		name       string
		msg        *types.MsgUpdateVestingFunder
		expectPass bool
	}{
		{
			name: "msg update vesting funder - valid addresses",
			msg: types.NewMsgUpdateVestingFunder(
				funder,
				vestingAcc,
				newFunder,
			),
			expectPass: true,
		},
		{
			name: "msg update vesting funder - invalid funder address",
			msg: &types.MsgUpdateVestingFunder{
				"invalid_address",
				vestingAcc.String(),
				newFunder.String(),
			},
			expectPass: false,
		},
		{
			name: "msg update vesting funder - invalid new funder address",
			msg: &types.MsgUpdateVestingFunder{
				funder.String(),
				"invalid_address",
				newFunder.String(),
			},
			expectPass: false,
		},
		{
			name: "msg update vesting funder - invalid vesting address",
			msg: &types.MsgUpdateVestingFunder{
				funder.String(),
				vestingAcc.String(),
				"invalid_address",
			},
			expectPass: false,
		},
		{
			name: "msg update vesting funder - empty address",
			msg: &types.MsgUpdateVestingFunder{
				funder.String(),
				vestingAcc.String(),
				"",
			},
			expectPass: false,
		},
		{
			name: "msg update vesting funder - new funder address is equal to current funder address",
			msg: &types.MsgUpdateVestingFunder{
				funder.String(),
				utiltx.GenerateAddress().String(),
				funder.String(),
			},
			expectPass: false,
		},
	}

	for i, tc := range testCases {
		err := tc.msg.ValidateBasic()
		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.name)
		}
	}
}

func (suite *MsgsTestSuite) TestMsgConvertVestingAccount() {
	testCases := []struct {
		name    string
		msg     *types.MsgConvertVestingAccount
		expPass bool
	}{
		{
			"fail - not a valid vesting address",
			&types.MsgConvertVestingAccount{
				"invalid_address",
			},
			false,
		},

		{
			"pass - valid vesting address",
			&types.MsgConvertVestingAccount{
				sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			},
			true,
		},
	}

	for i, tc := range testCases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.name)
		}
	}
}
