package types

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/tests"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type MsgsTestSuite struct {
	suite.Suite
	contract      common.Address
	deployer      sdk.AccAddress
	deployerStr   string
	withdrawerStr string
}

func TestMsgsTestSuite(t *testing.T) {
	suite.Run(t, new(MsgsTestSuite))
}

func (suite *MsgsTestSuite) SetupTest() {
	deployer := tests.GenerateAddress()
	suite.contract = crypto.CreateAddress(deployer, 1)
	suite.deployer = sdk.AccAddress(deployer.Bytes())
	suite.deployerStr = suite.deployer.String()
	suite.withdrawerStr = sdk.AccAddress(tests.GenerateAddress().Bytes()).String()
}

func (suite *MsgsTestSuite) TestMsgRegisterFeeSplitGetters() {
	msgInvalid := MsgRegisterFeeSplit{}
	msg := NewMsgRegisterFeeSplit(
		suite.contract,
		suite.deployer,
		suite.deployer,
		[]uint64{1},
	)
	suite.Require().Equal(RouterKey, msg.Route())
	suite.Require().Equal(TypeMsgRegisterFeeSplit, msg.Type())
	suite.Require().NotNil(msgInvalid.GetSignBytes())
	suite.Require().NotNil(msg.GetSigners())
}

func (suite *MsgsTestSuite) TestMsgRegisterFeeSplitNew() {
	testCases := []struct {
		msg        string
		contract   string
		deployer   string
		withdraw   string
		nonces     []uint64
		expectPass bool
	}{
		{
			"pass",
			suite.contract.String(),
			suite.deployerStr,
			suite.withdrawerStr,
			[]uint64{1},
			true,
		},
		{
			"pass - empty withdrawer address",
			suite.contract.String(),
			suite.deployerStr,
			"",
			[]uint64{1},
			true,
		},
		{
			"pass - same withdrawer and deployer address",
			suite.contract.String(),
			suite.deployerStr,
			suite.deployerStr,
			[]uint64{1},
			true,
		},
		{
			"invalid contract address",
			"",
			suite.deployerStr,
			suite.withdrawerStr,
			[]uint64{1},
			false,
		},
		{
			"must not be zero: invalid address",
			"0x0000000000000000000000000000000000000000",
			suite.deployerStr,
			suite.withdrawerStr,
			[]uint64{1},
			false,
		},
		{
			"invalid deployer address",
			suite.contract.String(),
			"",
			suite.withdrawerStr,
			[]uint64{1},
			false,
		},
		{
			"invalid withdraw address",
			suite.contract.String(),
			suite.deployerStr,
			"withdraw",
			[]uint64{1},
			false,
		},
		{
			"invalid nonces",
			suite.contract.String(),
			suite.deployerStr,
			suite.withdrawerStr,
			[]uint64{},
			false,
		},
		{
			"invalid nonces - array length must be less than 20",
			suite.contract.String(),
			suite.deployerStr,
			suite.withdrawerStr,
			[]uint64{1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			false,
		},
	}

	for i, tc := range testCases {
		tx := MsgRegisterFeeSplit{
			ContractAddress:   tc.contract,
			DeployerAddress:   tc.deployer,
			WithdrawerAddress: tc.withdraw,
			Nonces:            tc.nonces,
		}
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s", i, tc.msg)
			suite.Require().Contains(err.Error(), tc.msg)
		}
	}
}

func (suite *MsgsTestSuite) TestMsgCancelFeeSplitGetters() {
	msgInvalid := MsgCancelFeeSplit{}
	msg := NewMsgCancelFeeSplit(
		suite.contract,
		sdk.AccAddress(suite.deployer.Bytes()),
	)
	suite.Require().Equal(RouterKey, msg.Route())
	suite.Require().Equal(TypeMsgCancelFeeSplit, msg.Type())
	suite.Require().NotNil(msgInvalid.GetSignBytes())
	suite.Require().NotNil(msg.GetSigners())
}

func (suite *MsgsTestSuite) TestMsgCancelFeeSplitNew() {
	testCases := []struct {
		msg        string
		contract   string
		deployer   string
		expectPass bool
	}{
		{
			"msg cancel contract fee - pass",
			suite.contract.String(),
			suite.deployerStr,
			true,
		},
		{
			"invalid contract address",
			"",
			suite.deployerStr,
			false,
		},
		{
			"must not be zero: invalid address",
			"0x0000000000000000000000000000000000000000",
			suite.deployerStr,
			false,
		},
		{
			"invalid deployer address",
			suite.contract.String(),
			"",
			false,
		},
	}

	for i, tc := range testCases {
		tx := MsgCancelFeeSplit{
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

func (suite *MsgsTestSuite) TestMsgUpdateFeeSplitGetters() {
	msgInvalid := MsgUpdateFeeSplit{}
	msg := NewMsgUpdateFeeSplit(
		suite.contract,
		sdk.AccAddress(suite.deployer.Bytes()),
		sdk.AccAddress(suite.deployer.Bytes()),
	)
	suite.Require().Equal(RouterKey, msg.Route())
	suite.Require().Equal(TypeMsgUpdateFeeSplit, msg.Type())
	suite.Require().NotNil(msgInvalid.GetSignBytes())
	suite.Require().NotNil(msg.GetSigners())
}

func (suite *MsgsTestSuite) TestMsgUpdateFeeSplitNew() {
	withdrawerStr := sdk.AccAddress(tests.GenerateAddress().Bytes()).String()
	testCases := []struct {
		msg        string
		contract   string
		deployer   string
		withdraw   string
		expectPass bool
	}{
		{
			"msg update fee - pass",
			suite.contract.String(),
			suite.deployerStr,
			withdrawerStr,
			true,
		},
		{
			"invalid contract address",
			"",
			suite.deployerStr,
			withdrawerStr,
			false,
		},
		{
			"must not be zero: invalid address",
			"0x0000000000000000000000000000000000000000",
			suite.deployerStr,
			withdrawerStr,
			false,
		},
		{
			"invalid deployer address",
			suite.contract.String(),
			"",
			suite.deployerStr,
			false,
		},
		{
			"invalid withdraw address",
			suite.contract.String(),
			suite.deployerStr,
			"withdraw",
			false,
		},
		{
			"withdraw address must be different that deployer",
			suite.contract.String(),
			suite.deployerStr,
			suite.deployerStr,
			false,
		},
	}

	for i, tc := range testCases {
		tx := MsgUpdateFeeSplit{
			ContractAddress:   tc.contract,
			DeployerAddress:   tc.deployer,
			WithdrawerAddress: tc.withdraw,
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
