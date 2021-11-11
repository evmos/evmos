package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/tharsis/ethermint/tests"

	length "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"
)

type ProposalTestSuite struct {
	suite.Suite
}

func TestProposalTestSuite(t *testing.T) {
	suite.Run(t, new(ProposalTestSuite))
}

// func (suite *ProposalTestSuite) TestRegisterTokenPairProposal() {
// 	testCases := []struct {
// 		msg         string
// 		title       string
// 		description string
// 		pair        TokenPair
// 		expectPass  bool
// 	}{
// 		// Valid tests
// 		{msg: "Register token pair - valid pair enabled", title: "test", description: "test desc", pair: TokenPair{tests.GenerateAddress().String(), "test", true, MODULE_OWNER}, expectPass: true},
// 		{msg: "Register token pair - valid pair dissabled", title: "test", description: "test desc", pair: TokenPair{tests.GenerateAddress().String(), "test", false, MODULE_OWNER}, expectPass: true},
// 		// Missing params valid
// 		{msg: "Register token pair - invalid missing title ", title: "", description: "test desc", pair: TokenPair{tests.GenerateAddress().String(), "test", false, MODULE_OWNER}, expectPass: false},
// 		{msg: "Register token pair - invalid missing description ", title: "test", description: "", pair: TokenPair{tests.GenerateAddress().String(), "test", false, MODULE_OWNER}, expectPass: false},
// 		// Invalid address
// 		{msg: "Register token pair - invalid address (no hex)", title: "test", description: "test desc", pair: TokenPair{"0x5dCA2483280D9727c80b5518faC4556617fb19ZZ", "test", true, MODULE_OWNER}, expectPass: false},
// 		{msg: "Register token pair - invalid address (invalid length 1)", title: "test", description: "test desc", pair: TokenPair{"0x5dCA2483280D9727c80b5518faC4556617fb19", "test", true, MODULE_OWNER}, expectPass: false},
// 		{msg: "Register token pair - invalid address (invalid length 2)", title: "test", description: "test desc", pair: TokenPair{"0x5dCA2483280D9727c80b5518faC4556617fb194FFF", "test", true, MODULE_OWNER}, expectPass: false},
// 		{msg: "Register token pair - invalid address (invalid prefix)", title: "test", description: "test desc", pair: TokenPair{"1x5dCA2483280D9727c80b5518faC4556617fb19F", "test", true, MODULE_OWNER}, expectPass: false},
// 		// Invalid Regex (denom)
// 		{msg: "Register token pair - invalid starts with number", title: "test", description: "test desc", pair: TokenPair{tests.GenerateAddress().String(), "1test", true, MODULE_OWNER}, expectPass: false},
// 		{msg: "Register token pair - invalid char '('", title: "test", description: "test desc", pair: TokenPair{tests.GenerateAddress().String(), "(test", true, MODULE_OWNER}, expectPass: false},
// 		{msg: "Register token pair - invalid char '^'", title: "test", description: "test desc", pair: TokenPair{tests.GenerateAddress().String(), "^test", true, MODULE_OWNER}, expectPass: false},
// 		// Invalid length
// 		{msg: "Register token pair - invalid length token (0)", title: "test", description: "test desc", pair: TokenPair{tests.GenerateAddress().String(), "", true, MODULE_OWNER}, expectPass: false},
// 		{msg: "Register token pair - invalid length token (1)", title: "test", description: "test desc", pair: TokenPair{tests.GenerateAddress().String(), "a", true, MODULE_OWNER}, expectPass: false},
// 		{msg: "Register token pair - invalid length token (128)", title: "test", description: "test desc", pair: TokenPair{tests.GenerateAddress().String(), strings.Repeat("a", 129), true, MODULE_OWNER}, expectPass: false},
// 		{msg: "Register token pair - invalid length title (140)", title: strings.Repeat("a", length.MaxTitleLength+1), description: "test desc", pair: TokenPair{tests.GenerateAddress().String(), "test", true, MODULE_OWNER}, expectPass: false},
// 		{msg: "Register token pair - invalid length description (5000)", title: "title", description: strings.Repeat("a", length.MaxDescriptionLength+1), pair: TokenPair{tests.GenerateAddress().String(), "test", true, MODULE_OWNER}, expectPass: false},
// 	}

// 	for i, tc := range testCases {
// 		tx := NewRegisterCoinProposal(tc.title, tc.description, tc.pair)
// 		err := tx.ValidateBasic()

// 		if tc.expectPass {
// 			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
// 		} else {
// 			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
// 		}
// 	}
// }

func (suite *ProposalTestSuite) TestEnableTokenRelayProposal() {
	testCases := []struct {
		msg         string
		title       string
		description string
		token       string
		expectPass  bool
	}{
		{msg: "Enable token relay proposal - valid denom", title: "test", description: "test desc", token: "test", expectPass: true},
		{msg: "Enable token relay proposal - valid address", title: "test", description: "test desc", token: "0x5dCA2483280D9727c80b5518faC4556617fb194F", expectPass: true},
		{msg: "Enable token relay proposal - invalid address", title: "test", description: "test desc", token: "0x123", expectPass: false},

		// Invalid missing params
		{msg: "Enable token relay proposal - valid missing title", title: "", description: "test desc", token: "test", expectPass: false},
		{msg: "Enable token relay proposal - valid missing description", title: "test", description: "", token: "test", expectPass: false},
		{msg: "Enable token relay proposal - invalid missing token", title: "test", description: "test desc", token: "", expectPass: false},

		// Invalid regex
		{msg: "Enable token relay proposal - invalid denom", title: "test", description: "test desc", token: "^test", expectPass: false},
		// Invalid length
		{msg: "Enable token relay proposal - invalid length (1)", title: "test", description: "test desc", token: "a", expectPass: false},
		{msg: "Enable token relay proposal - invalid length (128)", title: "test", description: "test desc", token: strings.Repeat("a", 129), expectPass: false},

		{msg: "Enable token relay proposal - invalid length title (140)", title: strings.Repeat("a", length.MaxTitleLength+1), description: "test desc", token: "test", expectPass: false},
		{msg: "Enable token relay proposal - invalid length description (5000)", title: "title", description: strings.Repeat("a", length.MaxDescriptionLength+1), token: "test", expectPass: false},
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

func (suite *ProposalTestSuite) TestUpdateTokenPairERC20Proposal() {
	testCases := []struct {
		msg          string
		title        string
		description  string
		erc20Addr    common.Address
		newErc20Addr common.Address
		expectPass   bool
	}{
		{msg: "update token pair erc20 - pass", title: "test", description: "test desc", erc20Addr: tests.GenerateAddress(), newErc20Addr: tests.GenerateAddress(), expectPass: true},
		{msg: "update token pair erc20 - missing title", title: "", description: "test desc", erc20Addr: tests.GenerateAddress(), newErc20Addr: tests.GenerateAddress(), expectPass: false},
		{msg: "update token pair erc20 - missing description", title: "test", description: "", erc20Addr: tests.GenerateAddress(), newErc20Addr: tests.GenerateAddress(), expectPass: false},
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
func (suite *ProposalTestSuite) TestUpdateTokenPairERC20ProposalWithoutConstructor() {
	testCases := []struct {
		msg          string
		title        string
		description  string
		erc20Addr    string
		newErc20Addr string
		expectPass   bool
	}{
		{msg: "update token pair erc20 without constructor - valid", title: "test", description: "desc", erc20Addr: tests.GenerateAddress().String(), newErc20Addr: tests.GenerateAddress().String(), expectPass: true},
		{msg: "update token pair erc20 without constructor- invalid address 1", title: "test", description: "desc", erc20Addr: tests.GenerateAddress().String(), newErc20Addr: "1x5dCA2483280D9727c80b5518faC4556617fb19F", expectPass: false},
		{msg: "update token pair erc20 without constructor- invalid address 2", title: "test", description: "desc", erc20Addr: "1x5dCA2483280D9727c80b5518faC4556617fb19F", newErc20Addr: tests.GenerateAddress().String(), expectPass: false},
	}

	for i, tc := range testCases {
		tx := UpdateTokenPairERC20Proposal{
			Title:           tc.title,
			Description:     tc.description,
			Erc20Address:    tc.erc20Addr,
			NewErc20Address: tc.newErc20Addr,
		}
		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
		}
	}
}
