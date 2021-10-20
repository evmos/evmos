package types

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/ethermint/tests"

	"github.com/ethereum/go-ethereum/common"
)

type MsgsTestSuite struct {
	suite.Suite
}

func TestMsgsTestSuite(t *testing.T) {
	suite.Run(t, new(MsgsTestSuite))
}

func (suite *MsgsTestSuite) TestMsgConvertCoin() {
	testCases := []struct {
		msg        string
		coin       sdk.Coin
		receiver   common.Address
		sender     sdk.AccAddress
		expectPass bool
	}{
		{msg: "msg convert coin - pass", coin: sdk.NewCoin("test", sdk.NewInt(100)), receiver: tests.GenerateAddress(), sender: sdk.AccAddress(tests.GenerateAddress().Bytes()), expectPass: true},
	}

	for i, tc := range testCases {
		tx := NewMsgConvertCoin(tc.coin, tc.receiver, tc.sender)
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
		}
	}
}

// TODO: Uncomment when validate basic is fixed
// func (suite *MsgsTestSuite) TestMsgConvertERC20() {
// 	testCases := []struct {
// 		msg        string
// 		amount     sdk.Int
// 		receiver   sdk.AccAddress
// 		contract   common.Address
// 		sender     common.Address
// 		expectPass bool
// 	}{
// 		{msg: "msg convert erc20 - pass", amount: sdk.NewInt(100), receiver: sdk.AccAddress(tests.GenerateAddress().Bytes()), contract: tests.GenerateAddress(), sender: tests.GenerateAddress(), expectPass: true},
// 	}

// 	for i, tc := range testCases {
// 		tx := NewMsgConvertERC20(tc.amount, tc.receiver, tc.contract, tc.sender)
// 		err := tx.ValidateBasic()

// 		if tc.expectPass {
// 			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
// 		} else {
// 			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
// 		}
// 	}
// }
