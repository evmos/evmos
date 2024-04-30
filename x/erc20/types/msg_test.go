package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	utiltx "github.com/evmos/evmos/v18/testutil/tx"
	"github.com/evmos/evmos/v18/x/erc20/types"

	"github.com/ethereum/go-ethereum/common"
)

type MsgsTestSuite struct {
	suite.Suite
}

func TestMsgsTestSuite(t *testing.T) {
	suite.Run(t, new(MsgsTestSuite))
}

func (suite *MsgsTestSuite) TestMsgConvertERC20Getters() {
	msgInvalid := types.MsgConvertERC20{}
	msg := types.NewMsgConvertERC20(
		math.NewInt(100),
		sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
		utiltx.GenerateAddress(),
		utiltx.GenerateAddress(),
	)
	suite.Require().Equal(types.RouterKey, msg.Route())
	suite.Require().Equal(types.TypeMsgConvertERC20, msg.Type())
	suite.Require().NotNil(msgInvalid.GetSignBytes())
	suite.Require().NotNil(msg.GetSigners())
}

func (suite *MsgsTestSuite) TestMsgConvertERC20New() {
	testCases := []struct {
		msg        string
		amount     math.Int
		receiver   sdk.AccAddress
		contract   common.Address
		sender     common.Address
		expectPass bool
	}{
		{
			"msg convert erc20 - pass",
			math.NewInt(100),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			utiltx.GenerateAddress(),
			utiltx.GenerateAddress(),
			true,
		},
	}

	for i, tc := range testCases {
		tx := types.NewMsgConvertERC20(tc.amount, tc.receiver, tc.contract, tc.sender)
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
		}
	}
}

func (suite *MsgsTestSuite) TestMsgConvertERC20() {
	testCases := []struct {
		msg        string
		amount     math.Int
		receiver   string
		contract   string
		sender     string
		expectPass bool
	}{
		{
			"invalid contract hex address",
			math.NewInt(100),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			sdk.AccAddress{}.String(),
			utiltx.GenerateAddress().String(),
			false,
		},
		{
			"negative coin amount",
			math.NewInt(-100),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			utiltx.GenerateAddress().String(),
			utiltx.GenerateAddress().String(),
			false,
		},
		{
			"invalid receiver address",
			math.NewInt(100),
			sdk.AccAddress{}.String(),
			utiltx.GenerateAddress().String(),
			utiltx.GenerateAddress().String(),
			false,
		},
		{
			"invalid sender address",
			math.NewInt(100),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			utiltx.GenerateAddress().String(),
			sdk.AccAddress{}.String(),
			false,
		},
		{
			"msg convert erc20 - pass",
			math.NewInt(100),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			utiltx.GenerateAddress().String(),
			utiltx.GenerateAddress().String(),
			true,
		},
	}

	for i, tc := range testCases {
		tx := types.MsgConvertERC20{tc.contract, tc.amount, tc.receiver, tc.sender}
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
		}
	}
}

func (suite *MsgsTestSuite) TestMsgUpdateValidateBasic() {
	testCases := []struct {
		name      string
		msgUpdate *types.MsgUpdateParams
		expPass   bool
	}{
		{
			"fail - invalid authority address",
			&types.MsgUpdateParams{
				Authority: "invalid",
				Params:    types.DefaultParams(),
			},
			false,
		},
		{
			"pass - valid msg",
			&types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params:    types.DefaultParams(),
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := tc.msgUpdate.ValidateBasic()
			if tc.expPass {
				suite.NoError(err)
			} else {
				suite.Error(err)
			}
		})
	}
}
