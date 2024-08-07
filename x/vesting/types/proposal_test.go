package types_test

import (
	"testing"

	"github.com/evmos/evmos/v19/x/vesting/types"
	"github.com/stretchr/testify/suite"
)

type ProposalTestSuite struct {
	suite.Suite
}

func TestProposalTestSuite(t *testing.T) {
	suite.Run(t, new(ProposalTestSuite))
}

func (suite *ProposalTestSuite) TestKeysTypes() {
	suite.Require().Equal("vesting", (&types.ClawbackProposal{}).ProposalRoute())
	suite.Require().Equal("Clawback", (&types.ClawbackProposal{}).ProposalType())
}

func (suite *ProposalTestSuite) TestClawbackProposal() {
	testCases := []struct {
		msg                string
		title              string
		description        string
		address            string
		destinationAddress string
		expectPass         bool
	}{
		// Valid tests
		{
			msg:         "Clawback proposal - valid address",
			title:       "test",
			description: "test desc",
			address:     "evmos19y7d5jz7q0v86zw5m0300mhprpvu0ccc4x6xgg",
			expectPass:  true,
		},
		// Invalid - Missing params
		{
			msg:         "Clawback proposal - invalid missing title ",
			title:       "",
			description: "test desc",
			address:     "evmos19y7d5jz7q0v86zw5m0300mhprpvu0ccc4x6xgg",
			expectPass:  false,
		},
		{
			msg:         "Clawback proposal - invalid missing description ",
			title:       "test",
			description: "",
			address:     "evmos19y7d5jz7q0v86zw5m0300mhprpvu0ccc4x6xgg",
			expectPass:  false,
		},
		// Invalid address
		{
			msg:         "Clawback proposal - invalid address (no hex)",
			title:       "test",
			description: "test desc",
			address:     "evmos19y7d5jz7q0v86zw5m0300mhprpvu0ccc4x6ggg",
			expectPass:  false,
		},
		{
			msg:                "Clawback proposal - invalid destination addr",
			title:              "test",
			description:        "test desc",
			address:            "evmos19y7d5jz7q0v86zw5m0300mhprpvu0ccc4x6xgg",
			destinationAddress: "125182ujaisch8hsgs",
			expectPass:         false,
		},
	}

	for i, tc := range testCases {
		tx := types.NewClawbackProposal(tc.title, tc.description, tc.address, tc.destinationAddress)
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
		}
	}
}
