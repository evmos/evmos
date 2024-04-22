package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	length "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	utiltx "github.com/evmos/evmos/v18/testutil/tx"
	"github.com/evmos/evmos/v18/x/erc20/types"
)

type ProposalTestSuite struct {
	suite.Suite
}

func TestProposalTestSuite(t *testing.T) {
	suite.Run(t, new(ProposalTestSuite))
}

func (suite *ProposalTestSuite) TestKeysTypes() {
	suite.Require().Equal("erc20", (&types.RegisterERC20Proposal{}).ProposalRoute())
	suite.Require().Equal("RegisterERC20", (&types.RegisterERC20Proposal{}).ProposalType())
	suite.Require().Equal("erc20", (&types.ToggleTokenConversionProposal{}).ProposalRoute())
	suite.Require().Equal("ToggleTokenConversion", (&types.ToggleTokenConversionProposal{}).ProposalType())
}

func (suite *ProposalTestSuite) TestCreateDenomDescription() {
	testCases := []struct {
		name      string
		denom     string
		expString string
	}{
		{
			"with valid address",
			"0xdac17f958d2ee523a2206206994597c13d831ec7",
			"Cosmos coin token representation of 0xdac17f958d2ee523a2206206994597c13d831ec7",
		},
		{
			"with empty string",
			"",
			"Cosmos coin token representation of ",
		},
	}
	for _, tc := range testCases {
		desc := types.CreateDenomDescription(tc.denom)
		suite.Require().Equal(desc, tc.expString)
	}
}

func (suite *ProposalTestSuite) TestCreateDenom() {
	testCases := []struct {
		name      string
		denom     string
		expString string
	}{
		{
			"with valid address",
			"0xdac17f958d2ee523a2206206994597c13d831ec7",
			"erc20/0xdac17f958d2ee523a2206206994597c13d831ec7",
		},
		{
			"with empty string",
			"",
			"erc20/",
		},
	}
	for _, tc := range testCases {
		desc := types.CreateDenom(tc.denom)
		suite.Require().Equal(desc, tc.expString)
	}
}

func (suite *ProposalTestSuite) TestValidateErc20Denom() {
	testCases := []struct {
		name    string
		denom   string
		expPass bool
	}{
		{
			"- instead of /",
			"erc20-0xdac17f958d2ee523a2206206994597c13d831ec7",
			false,
		},
		{
			"without /",
			"conversionCoin",
			false,
		},
		{
			"// instead of /",
			"erc20//0xdac17f958d2ee523a2206206994597c13d831ec7",
			false,
		},
		{
			"multiple /",
			"erc20/0xdac17f958d2ee523a2206206994597c13d831ec7/test",
			false,
		},
		{
			"pass",
			"erc20/0xdac17f958d2ee523a2206206994597c13d831ec7",
			true,
		},
	}
	for _, tc := range testCases {
		err := types.ValidateErc20Denom(tc.denom)

		if tc.expPass {
			suite.Require().Nil(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

func (suite *ProposalTestSuite) TestRegisterERC20Proposal() {
	testCases := []struct {
		msg         string
		title       string
		description string
		pair        types.TokenPair
		expectPass  bool
	}{
		// Valid tests
		{msg: "Register token pair - valid pair enabled", title: "test", description: "test desc", pair: types.TokenPair{utiltx.GenerateAddress().String(), "test", true, types.OWNER_MODULE}, expectPass: true},
		{msg: "Register token pair - valid pair dissabled", title: "test", description: "test desc", pair: types.TokenPair{utiltx.GenerateAddress().String(), "test", false, types.OWNER_MODULE}, expectPass: true},
		// Missing params valid
		{msg: "Register token pair - invalid missing title ", title: "", description: "test desc", pair: types.TokenPair{utiltx.GenerateAddress().String(), "test", false, types.OWNER_MODULE}, expectPass: false},
		{msg: "Register token pair - invalid missing description ", title: "test", description: "", pair: types.TokenPair{utiltx.GenerateAddress().String(), "test", false, types.OWNER_MODULE}, expectPass: false},
		// Invalid address
		{msg: "Register token pair - invalid address (no hex)", title: "test", description: "test desc", pair: types.TokenPair{"0x5dCA2483280D9727c80b5518faC4556617fb19ZZ", "test", true, types.OWNER_MODULE}, expectPass: false},
		{msg: "Register token pair - invalid address (invalid length 1)", title: "test", description: "test desc", pair: types.TokenPair{"0x5dCA2483280D9727c80b5518faC4556617fb19", "test", true, types.OWNER_MODULE}, expectPass: false},
		{msg: "Register token pair - invalid address (invalid length 2)", title: "test", description: "test desc", pair: types.TokenPair{"0x5dCA2483280D9727c80b5518faC4556617fb194FFF", "test", true, types.OWNER_MODULE}, expectPass: false},
		{msg: "Register token pair - invalid address (invalid prefix)", title: "test", description: "test desc", pair: types.TokenPair{"1x5dCA2483280D9727c80b5518faC4556617fb19F", "test", true, types.OWNER_MODULE}, expectPass: false},
	}

	for i, tc := range testCases {
		tx := types.NewRegisterERC20Proposal(tc.title, tc.description, tc.pair.Erc20Address)
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
		}
	}
}

func (suite *ProposalTestSuite) TestToggleTokenConversionProposal() {
	testCases := []struct {
		msg         string
		title       string
		description string
		token       string
		expectPass  bool
	}{
		{msg: "Enable token conversion proposal - valid denom", title: "test", description: "test desc", token: "test", expectPass: true},
		{msg: "Enable token conversion proposal - valid address", title: "test", description: "test desc", token: "0x5dCA2483280D9727c80b5518faC4556617fb194F", expectPass: true}, //gitleaks:allow
		{msg: "Enable token conversion proposal - invalid address", title: "test", description: "test desc", token: "0x123", expectPass: false},

		// Invalid missing params
		{msg: "Enable token conversion proposal - valid missing title", title: "", description: "test desc", token: "test", expectPass: false},
		{msg: "Enable token conversion proposal - valid missing description", title: "test", description: "", token: "test", expectPass: false},
		{msg: "Enable token conversion proposal - invalid missing token", title: "test", description: "test desc", token: "", expectPass: false},

		// Invalid regex
		{msg: "Enable token conversion proposal - invalid denom", title: "test", description: "test desc", token: "^test", expectPass: false},
		// Invalid length
		{msg: "Enable token conversion proposal - invalid length (1)", title: "test", description: "test desc", token: "a", expectPass: false},
		{msg: "Enable token conversion proposal - invalid length (128)", title: "test", description: "test desc", token: strings.Repeat("a", 129), expectPass: false},

		{msg: "Enable token conversion proposal - invalid length title (140)", title: strings.Repeat("a", length.MaxTitleLength+1), description: "test desc", token: "test", expectPass: false},
		{msg: "Enable token conversion proposal - invalid length description (5000)", title: "title", description: strings.Repeat("a", length.MaxDescriptionLength+1), token: "test", expectPass: false},
	}

	for i, tc := range testCases {
		tx := types.NewToggleTokenConversionProposal(tc.title, tc.description, tc.token)
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
		}
	}
}
