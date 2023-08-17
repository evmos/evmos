package types_test

import (
	"strings"
	"testing"

	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/ethereum/go-ethereum/common"
	utiltx "github.com/evmos/evmos/v14/testutil/tx"
	"github.com/evmos/evmos/v14/x/erc20/types"
	"github.com/stretchr/testify/suite"
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
		owner        types.Owner
		expectPass   bool
	}{
		{msg: "Register token pair - invalid starts with number", erc20Address: utiltx.GenerateAddress(), denom: "1test", owner: types.OWNER_MODULE, expectPass: false},
		{msg: "Register token pair - invalid char '('", erc20Address: utiltx.GenerateAddress(), denom: "(test", owner: types.OWNER_MODULE, expectPass: false},
		{msg: "Register token pair - invalid char '^'", erc20Address: utiltx.GenerateAddress(), denom: "^test", owner: types.OWNER_MODULE, expectPass: false},
		// TODO: (guille) should the "\" be allowed to support unicode names?
		{msg: "Register token pair - invalid char '\\'", erc20Address: utiltx.GenerateAddress(), denom: "-test", owner: types.OWNER_MODULE, expectPass: false},
		// Invalid length
		{msg: "Register token pair - invalid length token (0)", erc20Address: utiltx.GenerateAddress(), denom: "", owner: types.OWNER_MODULE, expectPass: false},
		{msg: "Register token pair - invalid length token (1)", erc20Address: utiltx.GenerateAddress(), denom: "a", owner: types.OWNER_MODULE, expectPass: false},
		{msg: "Register token pair - invalid length token (128)", erc20Address: utiltx.GenerateAddress(), denom: strings.Repeat("a", 129), owner: types.OWNER_MODULE, expectPass: false},
		{msg: "Register token pair - pass", erc20Address: utiltx.GenerateAddress(), denom: "test", owner: types.OWNER_MODULE, expectPass: true},
	}

	for i, tc := range testCases {
		tp := types.NewTokenPair(tc.erc20Address, tc.denom, tc.owner)
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
		pair       types.TokenPair
		expectPass bool
	}{
		{msg: "Register token pair - invalid address (no hex)", pair: types.TokenPair{"0x5dCA2483280D9727c80b5518faC4556617fb19ZZ", "test", true, types.OWNER_MODULE}, expectPass: false},
		{msg: "Register token pair - invalid address (invalid length 1)", pair: types.TokenPair{"0x5dCA2483280D9727c80b5518faC4556617fb19", "test", true, types.OWNER_MODULE}, expectPass: false},
		{msg: "Register token pair - invalid address (invalid length 2)", pair: types.TokenPair{"0x5dCA2483280D9727c80b5518faC4556617fb194FFF", "test", true, types.OWNER_MODULE}, expectPass: false},
		{msg: "pass", pair: types.TokenPair{utiltx.GenerateAddress().String(), "test", true, types.OWNER_MODULE}, expectPass: true},
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

func (suite *TokenPairTestSuite) TestGetID() {
	addr := utiltx.GenerateAddress()
	denom := "test"
	pair := types.NewTokenPair(addr, denom, types.OWNER_MODULE)
	id := pair.GetID()
	expID := tmhash.Sum([]byte(addr.String() + "|" + denom))
	suite.Require().Equal(expID, id)
}

func (suite *TokenPairTestSuite) TestGetERC20Contract() {
	expAddr := utiltx.GenerateAddress()
	denom := "test"
	pair := types.NewTokenPair(expAddr, denom, types.OWNER_MODULE)
	addr := pair.GetERC20Contract()
	suite.Require().Equal(expAddr, addr)
}

func (suite *TokenPairTestSuite) TestIsNativeCoin() {
	testCases := []struct {
		name       string
		pair       types.TokenPair
		expectPass bool
	}{
		{
			"no owner",
			types.TokenPair{utiltx.GenerateAddress().String(), "test", true, types.OWNER_UNSPECIFIED},
			false,
		},
		{
			"external ERC20 owner",
			types.TokenPair{utiltx.GenerateAddress().String(), "test", true, types.OWNER_EXTERNAL},
			false,
		},
		{
			"pass",
			types.TokenPair{utiltx.GenerateAddress().String(), "test", true, types.OWNER_MODULE},
			true,
		},
	}

	for _, tc := range testCases {
		res := tc.pair.IsNativeCoin()
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
		pair       types.TokenPair
		expectPass bool
	}{
		{
			"no owner",
			types.TokenPair{utiltx.GenerateAddress().String(), "test", true, types.OWNER_UNSPECIFIED},
			false,
		},
		{
			"module owner",
			types.TokenPair{utiltx.GenerateAddress().String(), "test", true, types.OWNER_MODULE},
			false,
		},
		{
			"pass",
			types.TokenPair{utiltx.GenerateAddress().String(), "test", true, types.OWNER_EXTERNAL},
			true,
		},
	}

	for _, tc := range testCases {
		res := tc.pair.IsNativeERC20()
		if tc.expectPass {
			suite.Require().True(res, tc.name)
		} else {
			suite.Require().False(res, tc.name)
		}
	}
}
