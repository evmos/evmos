package types

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/ethermint/tests"

	"github.com/ethereum/go-ethereum/crypto"
)

type MsgsTestSuite struct {
	suite.Suite
}

func TestMsgsTestSuite(t *testing.T) {
	suite.Run(t, new(MsgsTestSuite))
}

func (suite *MsgsTestSuite) TestMsgRegisterDevFeeInfoGetters() {
	msgInvalid := MsgRegisterDevFeeInfo{}
	deployer := tests.GenerateAddress()
	contract := crypto.CreateAddress(deployer, 1)
	msg := NewMsgRegisterDevFeeInfo(
		contract,
		sdk.AccAddress(deployer.Bytes()),
		sdk.AccAddress(deployer.Bytes()),
		[]uint64{1},
	)
	suite.Require().Equal(RouterKey, msg.Route())
	suite.Require().Equal(TypeMsgRegisterDevFeeInfo, msg.Type())
	suite.Require().NotNil(msgInvalid.GetSignBytes())
	suite.Require().Nil(msgInvalid.GetSigners())
	suite.Require().NotNil(msg.GetSigners())
}

func (suite *MsgsTestSuite) TestMsgRegisterDevFeeInfoNew() {
	deployer := tests.GenerateAddress()
	deployerStr := sdk.AccAddress(deployer.Bytes()).String()
	contract := crypto.CreateAddress(deployer, 1)
	testCases := []struct {
		msg        string
		contract   string
		deployer   string
		withdraw   string
		nonces     []uint64
		expectPass bool
	}{
		{
			"msg register contract - pass",
			contract.String(),
			deployerStr,
			deployerStr,
			[]uint64{1},
			true,
		},
		{
			"msg register contract empty withdraw - pass",
			contract.String(),
			deployerStr,
			deployerStr,
			[]uint64{1},
			true,
		},
		{
			"invalid contract address",
			"",
			deployerStr,
			deployerStr,
			[]uint64{1},
			false,
		},
		{
			"address must not be empty",
			"0x0000000000000000000000000000000000000000",
			deployerStr,
			deployerStr,
			[]uint64{1},
			false,
		},
		{
			"invalid deployer address",
			contract.String(),
			"",
			deployerStr,
			[]uint64{1},
			false,
		},
		{
			"invalid withdraw address",
			contract.String(),
			deployerStr,
			"withdraw",
			[]uint64{1},
			false,
		},
		{
			"invalid nonces",
			contract.String(),
			deployerStr,
			deployerStr,
			[]uint64{},
			false,
		},
	}

	for i, tc := range testCases {
		tx := MsgRegisterDevFeeInfo{
			ContractAddress: tc.contract,
			DeployerAddress: tc.deployer,
			WithdrawAddress: tc.withdraw,
			Nonces:          tc.nonces,
		}
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
			suite.Require().Contains(err.Error(), tc.msg)
		}
	}
}

func (suite *MsgsTestSuite) TestMsgCancelDevFeeInfoGetters() {
	msgInvalid := MsgCancelDevFeeInfo{}
	deployer := tests.GenerateAddress()
	contract := crypto.CreateAddress(deployer, 1)
	msg := NewMsgCancelDevFeeInfo(
		contract,
		sdk.AccAddress(deployer.Bytes()),
	)
	suite.Require().Equal(RouterKey, msg.Route())
	suite.Require().Equal(TypeMsgCancelDevFeeInfo, msg.Type())
	suite.Require().NotNil(msgInvalid.GetSignBytes())
	suite.Require().Nil(msgInvalid.GetSigners())
	suite.Require().NotNil(msg.GetSigners())
}

func (suite *MsgsTestSuite) TestMsgCancelDevFeeInfoNew() {
	deployer := tests.GenerateAddress()
	deployerStr := sdk.AccAddress(deployer.Bytes()).String()
	contract := crypto.CreateAddress(deployer, 1)
	testCases := []struct {
		msg        string
		contract   string
		deployer   string
		expectPass bool
	}{
		{
			"msg cancel contract fee - pass",
			contract.String(),
			deployerStr,
			true,
		},
		{
			"invalid contract address",
			"",
			deployerStr,
			false,
		},
		{
			"address must not be empty",
			"0x0000000000000000000000000000000000000000",
			deployerStr,
			false,
		},
		{
			"invalid deployer address",
			contract.String(),
			"",
			false,
		},
	}

	for i, tc := range testCases {
		tx := MsgCancelDevFeeInfo{
			ContractAddress: tc.contract,
			DeployerAddress: tc.deployer,
		}
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
			suite.Require().Contains(err.Error(), tc.msg)
		}
	}
}

func (suite *MsgsTestSuite) TestMsgUpdateDevFeeInfoGetters() {
	msgInvalid := MsgUpdateDevFeeInfo{}
	deployer := tests.GenerateAddress()
	contract := crypto.CreateAddress(deployer, 1)
	msg := NewMsgUpdateDevFeeInfo(
		contract,
		sdk.AccAddress(deployer.Bytes()),
		sdk.AccAddress(deployer.Bytes()),
	)
	suite.Require().Equal(RouterKey, msg.Route())
	suite.Require().Equal(TypeMsgUpdateDevFeeInfo, msg.Type())
	suite.Require().NotNil(msgInvalid.GetSignBytes())
	suite.Require().Nil(msgInvalid.GetSigners())
	suite.Require().NotNil(msg.GetSigners())
}

func (suite *MsgsTestSuite) TestMsgUpdateDevFeeInfoNew() {
	deployer := tests.GenerateAddress()
	deployerStr := sdk.AccAddress(deployer.Bytes()).String()
	withdrawStr := sdk.AccAddress(tests.GenerateAddress().Bytes()).String()
	contract := crypto.CreateAddress(deployer, 1)
	testCases := []struct {
		msg        string
		contract   string
		deployer   string
		withdraw   string
		expectPass bool
	}{
		{
			"msg update fee info - pass",
			contract.String(),
			deployerStr,
			withdrawStr,
			true,
		},
		{
			"invalid contract address",
			"",
			deployerStr,
			withdrawStr,
			false,
		},
		{
			"address must not be empty",
			"0x0000000000000000000000000000000000000000",
			deployerStr,
			withdrawStr,
			false,
		},
		{
			"invalid deployer address",
			contract.String(),
			"",
			deployerStr,
			false,
		},
		{
			"invalid withdraw address",
			contract.String(),
			deployerStr,
			"withdraw",
			false,
		},
		{
			"withdraw address must be different that deployer",
			contract.String(),
			deployerStr,
			deployerStr,
			false,
		},
	}

	for i, tc := range testCases {
		tx := MsgUpdateDevFeeInfo{
			ContractAddress: tc.contract,
			DeployerAddress: tc.deployer,
			WithdrawAddress: tc.withdraw,
		}
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
			suite.Require().Contains(err.Error(), tc.msg)
		}
	}
}
