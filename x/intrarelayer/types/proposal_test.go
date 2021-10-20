package types

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/tharsis/ethermint/tests"

	"github.com/ethereum/go-ethereum/common"
)

type ProposalTestSuite struct {
	suite.Suite
}

func TestProposalTestSuite(t *testing.T) {
	suite.Run(t, new(MsgsTestSuite))
}

func (suite *MsgsTestSuite) TestRegisterTokenPairProposal() {
	testCases := []struct {
		msg         string
		title       string
		description string
		pair        TokenPair
		expectPass  bool
	}{
		{msg: "Register token pair - pass", title: "test", description: "test desc", pair: TokenPair{tests.GenerateAddress().String(), "test", true}, expectPass: true},
	}

	for i, tc := range testCases {
		tx := NewRegisterTokenPairProposal(tc.title, tc.description, tc.pair)
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
		}
	}
}

func (suite *MsgsTestSuite) TestEnableTokenRelayProposal() {
	testCases := []struct {
		msg         string
		title       string
		description string
		token       string
		expectPass  bool
	}{
		{msg: "Enable token relay proposal - pass", title: "test", description: "test desc", token: "test", expectPass: true},
	}

	for i, tc := range testCases {
		tx := NewEnableTokenRelayProposal(tc.title, tc.description, tc.token)
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
		}
	}
}

func (suite *MsgsTestSuite) TestUpdateTokenPairERC20Proposal() {
	testCases := []struct {
		msg          string
		title        string
		description  string
		erc20Addr    common.Address
		newErc20Addr common.Address
		expectPass   bool
	}{
		{msg: "update token pair erc20 - pass", title: "test", description: "test desc", erc20Addr: tests.GenerateAddress(), newErc20Addr: tests.GenerateAddress(), expectPass: true},
	}

	for i, tc := range testCases {
		tx := NewUpdateTokenPairERC20Proposal(tc.title, tc.description, tc.erc20Addr, tc.newErc20Addr)
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
		}
	}
}
