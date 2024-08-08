package types_test

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/stretchr/testify/suite"

	utiltx "github.com/evmos/evmos/v19/testutil/tx"
	"github.com/evmos/evmos/v19/x/vesting/types"
)

type MsgsTestSuite struct {
	suite.Suite
}

func TestMsgsTestSuite(t *testing.T) {
	suite.Run(t, new(MsgsTestSuite))
}

var zeroAddress = common.Address{}.Bytes()

func (suite *MsgsTestSuite) TestMsgCreateClawbackVestingAccountGetters() {
	msgInvalid := types.MsgCreateClawbackVestingAccount{}
	msg := types.NewMsgCreateClawbackVestingAccount(
		sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
		sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
		false,
	)
	suite.Require().Equal(types.RouterKey, msg.Route())
	suite.Require().Equal(types.TypeMsgCreateClawbackVestingAccount, msg.Type())
	suite.Require().NotNil(msgInvalid.GetSignBytes())
}

// TestMsgCreateClawbackVestingAccountNew checks if creating a clawback vesting account message
// is possible with the NewMsgCreateClawbackVestingAccount constructor
//
// NOTE: Other functionality-related tests are in TestMsgCreateClawbackVestingAccount
func (suite *MsgsTestSuite) TestMsgCreateClawbackVestingAccountNew() {
	funder := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	vestingAddr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	msg := types.NewMsgCreateClawbackVestingAccount(funder, vestingAddr, false)
	suite.Require().Equal(funder.String(), msg.FunderAddress)
	suite.Require().Equal(vestingAddr.String(), msg.VestingAddress)
}

func (suite *MsgsTestSuite) TestMsgCreateClawbackVestingAccount() {
	testCases := []struct {
		name        string
		funderAddr  string
		vestingAddr string
		expPass     bool
	}{
		{
			name:        "msg create clawback vesting account - invalid from address",
			funderAddr:  "foo",
			vestingAddr: sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			expPass:     false,
		},
		{
			name:        "msg create clawback vesting account - invalid to address",
			funderAddr:  sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			vestingAddr: "foo",
			expPass:     false,
		},
		{
			name:        "msg create clawback vesting account - valid",
			funderAddr:  sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			vestingAddr: sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			expPass:     true,
		},
		{
			name:        "msg create clawback vesting account - fail if zero address for funder",
			funderAddr:  sdk.AccAddress(common.Address{}.Bytes()).String(),
			vestingAddr: sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			expPass:     false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tx := types.MsgCreateClawbackVestingAccount{
				FunderAddress:  tc.funderAddr,
				VestingAddress: tc.vestingAddr,
			}
			err := tx.ValidateBasic()
			if tc.expPass {
				suite.Require().NoError(err, "failed to validate message")
			} else {
				suite.Require().Error(err, "expected message validation to fail")
			}
		})
	}
}

func (suite *MsgsTestSuite) TestMsgFundVestingAccountGetters() {
	msgInvalid := types.MsgFundVestingAccount{}
	msg := types.NewMsgFundVestingAccount(
		sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
		sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
		time.Unix(100200300, 0),
		sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
		sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
	)
	suite.Require().Equal(types.RouterKey, msg.Route())
	suite.Require().Equal(types.TypeMsgFundVestingAccount, msg.Type())
	suite.Require().NotNil(msgInvalid.GetSignBytes())
}

// TestMsgFundVestingAccountNew checks if creating a fund vesting account message
// is possible with the NewMsgFundVestingAccount constructor
//
// NOTE: Other functionality-related tests are in TestMsgFundVestingAccount
func (suite *MsgsTestSuite) TestMsgFundVestingAccountNew() {
	funder := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	vestingAddr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	startTime := time.Now()
	expLockupPeriods := sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}}
	expVestingPeriods := sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}}
	tx := types.NewMsgFundVestingAccount(
		funder,
		vestingAddr,
		startTime,
		expLockupPeriods,
		expVestingPeriods,
	)
	err := tx.ValidateBasic()
	suite.Require().NoError(err, "failed to validate message")
	suite.Require().Equal(funder.String(), tx.FunderAddress, "expect different funder address")
	suite.Require().Equal(vestingAddr.String(), tx.VestingAddress, "expect different vesting address")
	suite.Require().Equal(startTime, tx.StartTime, "expect different start time")
	suite.Require().Equal(expLockupPeriods, tx.LockupPeriods, "expect different lockup periods")
	suite.Require().Equal(expVestingPeriods, tx.VestingPeriods, "expect different vesting periods")
}

func (suite *MsgsTestSuite) TestMsgFundVestingAccount() {
	testCases := []struct {
		msg            string
		funderAddr     string
		vestingAddr    string
		startTime      time.Time
		lockupPeriods  sdkvesting.Periods
		vestingPeriods sdkvesting.Periods
		expPass        bool
	}{
		{
			"msg fund vesting account - invalid from address",
			"foo",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			false,
		},
		{
			"msg fund vesting account - invalid to address",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			"foo",
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			false,
		},
		{
			"msg fund vesting account - invalid lockup period length",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 0, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			false,
		},
		{
			"msg fund vesting account - invalid lockup period amount",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 0)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			false,
		},
		{
			"msg fund vesting account - invalid vesting period length",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 0, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			false,
		},
		{
			"msg fund vesting account - invalid vesting period amount",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 0)}}},
			false,
		},
		{
			"msg fund vesting account - vesting address is zero address",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(zeroAddress).String(),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 0)}}},
			false,
		},
		{
			"msg fund vesting account - pass",
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			time.Unix(100200300, 0),
			sdkvesting.Periods{{Length: 200000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			sdkvesting.Periods{{Length: 300000, Amount: sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}}},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.msg, func() {
			tx := types.MsgFundVestingAccount{
				tc.funderAddr,
				tc.vestingAddr,
				tc.startTime,
				tc.lockupPeriods,
				tc.vestingPeriods,
			}
			err := tx.ValidateBasic()

			if tc.expPass {
				suite.Require().NoError(err, "failed to validate message")
			} else {
				suite.Require().Error(err, "expected message validation to fail")
			}
		})
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
			name: "msg update vesting funder - zero address for new funder",
			msg: &types.MsgUpdateVestingFunder{
				funder.String(),
				sdk.AccAddress(common.Address{}.Bytes()).String(),
				funder.String(),
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

func (suite *MsgsTestSuite) TestMsgConvertVestingAccountGetters() {
	msgInvalid := types.MsgConvertVestingAccount{}
	msg := types.NewMsgConvertVestingAccount(
		sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
	)
	suite.Require().Equal(types.RouterKey, msg.Route())
	suite.Require().Equal(types.TypeMsgConvertVestingAccount, msg.Type())
	suite.Require().NotNil(msgInvalid.GetSignBytes())
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
