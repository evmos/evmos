package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/stretchr/testify/suite"

	"github.com/evoblockchain/ethermint/tests"
)

type MsgsTestSuite struct {
	suite.Suite
}

func TestMsgsTestSuite(t *testing.T) {
	suite.Run(t, new(MsgsTestSuite))
}

func (suite *MsgsTestSuite) TestMsgCreateClawbackVestingAccountGetters() {
	msgInvalid := MsgCreateClawbackVestingAccount{}
	msg := NewMsgCreateClawbackVestingAccount(
		sdk.AccAddress(tests.GenerateAddress().Bytes()),
		sdk.AccAddress(tests.GenerateAddress().Bytes()),
		time.Unix(100200300, 0),
		sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
		sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
		true,
	)
	suite.Require().Equal(RouterKey, msg.Route())
	suite.Require().Equal(TypeMsgCreateClawbackVestingAccount, msg.Type())
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
			sdk.AccAddress(tests.GenerateAddress().Bytes()),
			sdk.AccAddress(tests.GenerateAddress().Bytes()),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			true,
			true,
		},
	}

	for i, tc := range testCases {
		tx := NewMsgCreateClawbackVestingAccount(
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
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			true,
			false,
		},
		{
			"msg create clawback vesting account - invalid to address",
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			"foo",
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			true,
			false,
		},
		{
			"msg create clawback vesting account - invalid lockup period length",
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 0, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			true,
			false,
		},
		{
			"msg create clawback vesting account - invalid lockup period amount",
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 0)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			true,
			false,
		},
		{
			"msg create clawback vesting account - invalid vesting period length",
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 0, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			true,
			false,
		},
		{
			"msg create clawback vesting account - invalid vesting period amount",
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 0)}}},
			true,
			false,
		},
		{
			"msg create clawback vesting account - pass",
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			true,
			true,
		},
	}

	for i, tc := range testCases {
		tx := MsgCreateClawbackVestingAccount{
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
	msgInvalid := MsgClawback{}
	msg := NewMsgClawback(
		sdk.AccAddress(tests.GenerateAddress().Bytes()),
		sdk.AccAddress(tests.GenerateAddress().Bytes()),
		sdk.AccAddress(tests.GenerateAddress().Bytes()),
	)
	suite.Require().Equal(RouterKey, msg.Route())
	suite.Require().Equal(TypeMsgClawback, msg.Type())
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
			sdk.AccAddress(tests.GenerateAddress().Bytes()),
			sdk.AccAddress(tests.GenerateAddress().Bytes()),
			sdk.AccAddress(tests.GenerateAddress().Bytes()),
			true,
		},
	}

	for i, tc := range testCases {
		tx := NewMsgClawback(
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
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			false,
		},
		{
			"msg create clawback vesting account - invalid addr address",
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			"foo",
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			false,
		},
		{
			"msg create clawback vesting account - invalid dest address",
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			"foo",
			false,
		},
		{
			"msg create clawback vesting account - pass empty dest address",
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			"",
			true,
		},
		{
			"msg create clawback vesting account - pass",
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			true,
		},
	}

	for i, tc := range testCases {
		tx := MsgClawback{
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
