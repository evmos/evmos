package types

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	"github.com/tharsis/ethermint/tests"
)

type TokenPairTestSuite struct {
	suite.Suite
}

func TestTokenPairSuite(t *testing.T) {
	suite.Run(t, new(TokenPairTestSuite))
}

func (suite *TokenPairTestSuite) TestTokenPairNew() {
	testCases := []struct {
		msg          string
		erc20Address common.Address
		denom        string
		enabled      bool
		owner        Owner
		expectPass   bool
	}{
		{msg: "Register token pair - invalid starts with number", erc20Address: tests.GenerateAddress(), denom: "1test", enabled: true, owner: MODULE_OWNER, expectPass: false},
		{msg: "Register token pair - invalid char '('", erc20Address: tests.GenerateAddress(), denom: "(test", enabled: true, owner: MODULE_OWNER, expectPass: false},
		{msg: "Register token pair - invalid char '^'", erc20Address: tests.GenerateAddress(), denom: "^test", enabled: true, owner: MODULE_OWNER, expectPass: false},
		// TODO: (guille) should the "\" be allowed to support unicode names?
		{msg: "Register token pair - invalid char '\\'", erc20Address: tests.GenerateAddress(), denom: "-test", enabled: true, owner: MODULE_OWNER, expectPass: false},
		// Invalid length
		{msg: "Register token pair - invalid length token (0)", erc20Address: tests.GenerateAddress(), denom: "", enabled: true, owner: MODULE_OWNER, expectPass: false},
		{msg: "Register token pair - invalid length token (1)", erc20Address: tests.GenerateAddress(), denom: "a", enabled: true, owner: MODULE_OWNER, expectPass: false},
		{msg: "Register token pair - invalid length token (128)", erc20Address: tests.GenerateAddress(), denom: strings.Repeat("a", 129), enabled: true, owner: MODULE_OWNER, expectPass: false},
		{msg: "Register token pair - pass", erc20Address: tests.GenerateAddress(), denom: "test", enabled: true, owner: MODULE_OWNER, expectPass: true},
	}

	for i, tc := range testCases {
		tp := NewTokenPair(tc.erc20Address, tc.denom, tc.enabled, tc.owner)
		err := tp.Validate()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
		}
	}
}

func (suite *TokenPairTestSuite) TestTokenPair() {
	testCases := []struct {
		msg        string
		pair       TokenPair
		expectPass bool
	}{
		{msg: "Register token pair - invalid address (no hex)", pair: TokenPair{"0x5dCA2483280D9727c80b5518faC4556617fb19ZZ", "test", true, MODULE_OWNER}, expectPass: false},
		{msg: "Register token pair - invalid address (invalid length 1)", pair: TokenPair{"0x5dCA2483280D9727c80b5518faC4556617fb19", "test", true, MODULE_OWNER}, expectPass: false},
		{msg: "Register token pair - invalid address (invalid length 2)", pair: TokenPair{"0x5dCA2483280D9727c80b5518faC4556617fb194FFF", "test", true, MODULE_OWNER}, expectPass: false},
		{msg: "pass", pair: TokenPair{tests.GenerateAddress().String(), "test", true, MODULE_OWNER}, expectPass: true},
	}

	for i, tc := range testCases {
		err := tc.pair.Validate()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
		}
	}
}

func (suite *TokenPairTestSuite) TestIsNativeCoin() {
	testCases := []struct {
		name       string
		pair       TokenPair
		expectPass bool
	}{
		{
			"no owner",
			TokenPair{tests.GenerateAddress().String(), "test", true, INVALID_OWNER},
			false,
		},
		{
			"external ERC20 owner",
			TokenPair{tests.GenerateAddress().String(), "test", true, EXTERNAL_OWNER},
			false,
		},
		{
			"pass",
			TokenPair{tests.GenerateAddress().String(), "test", true, MODULE_OWNER},
			true,
		},
	}

	for _, tc := range testCases {
		res := tc.pair.IsNativeCoin()
		fmt.Println(res)

		if tc.expectPass {
			suite.Require().True(res, tc.name)
		} else {
			suite.Require().False(res, tc.name)
		}
	}
}
func (suite *TokenPairTestSuite) TestIsNativeERC20() {
	testCases := []struct {
		name       string
		pair       TokenPair
		expectPass bool
	}{
		{
			"no owner",
			TokenPair{tests.GenerateAddress().String(), "test", true, INVALID_OWNER},
			false,
		},
		{
			"module owner",
			TokenPair{tests.GenerateAddress().String(), "test", true, MODULE_OWNER},
			false,
		},
		{
			"pass",
			TokenPair{tests.GenerateAddress().String(), "test", true, EXTERNAL_OWNER},
			true,
		},
	}

	for _, tc := range testCases {
		res := tc.pair.IsNativeERC20()
		fmt.Println(res)

		if tc.expectPass {
			suite.Require().True(res, tc.name)
		} else {
			suite.Require().False(res, tc.name)
		}
	}
}
