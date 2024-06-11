package accesscontrol_test

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	accesscontrol "github.com/evmos/evmos/v18/precompiles/access_control"
	commonerr "github.com/evmos/evmos/v18/precompiles/common"
	"math/big"
)

func (s *PrecompileTestSuite) TestParseRoleArgs() {

	addr := s.keyring.GetAddr(0)

	testCases := []struct {
		name          string
		args          []interface{}
		expectedPass  bool
		errorContains string
	}{
		{
			name:          "fail - invalid number of arguments",
			args:          []interface{}{},
			expectedPass:  false,
			errorContains: fmt.Sprintf(commonerr.ErrInvalidNumberOfArgs, 2, 0),
		},
		{
			name: "fail - invalid role argument",
			args: []interface{}{
				"",
				common.Address{},
			},
			expectedPass:  false,
			errorContains: fmt.Sprintf(accesscontrol.ErrInvalidRoleArgument),
		},
		{
			name: "fail - invalid account argument",
			args: []interface{}{
				[32]uint8{},
				"",
			},
			expectedPass:  false,
			errorContains: fmt.Sprintf(accesscontrol.ErrInvalidAccountArgument),
		},
		{
			name: "pass - valid arguments",
			args: []interface{}{
				[32]uint8(accesscontrol.RoleDefaultAdmin.Bytes()),
				addr,
			},
			expectedPass:  true,
			errorContains: "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			role, account, err := accesscontrol.ParseRoleArgs(tc.args)

			if tc.expectedPass {
				s.NoError(err)
				s.Equal(accesscontrol.RoleDefaultAdmin, role)
				s.Equal(addr, account)
			} else {
				s.Error(err)
				s.Contains(err.Error(), tc.errorContains)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestParseBurnArgs() {

	testCases := []struct {
		name          string
		args          []interface{}
		expectedPass  bool
		errorContains string
	}{
		{
			name:          "fail - invalid number of arguments",
			args:          []interface{}{},
			expectedPass:  false,
			errorContains: fmt.Sprintf(commonerr.ErrInvalidNumberOfArgs, 1, 0),
		},
		{
			name: "fail - invalid amount argument",
			args: []interface{}{
				"",
			},
			expectedPass:  false,
			errorContains: fmt.Sprintf(commonerr.ErrInvalidAmount, ""),
		},
		{
			name: "pass - valid arguments",
			args: []interface{}{
				big.NewInt(1e18),
			},
			expectedPass:  true,
			errorContains: "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			amount, err := accesscontrol.ParseBurnArgs(tc.args)

			if tc.expectedPass {
				s.NoError(err)
				s.Equal(big.NewInt(1e18), amount)
			} else {
				s.Error(err)
				s.Contains(err.Error(), tc.errorContains)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestParseMintArgs() {
	toAddr := s.keyring.GetAddr(0)
	testCases := []struct {
		name          string
		args          []interface{}
		expectedPass  bool
		errorContains string
	}{
		{
			name:          "fail - invalid number of arguments",
			args:          []interface{}{},
			expectedPass:  false,
			errorContains: fmt.Sprintf(commonerr.ErrInvalidNumberOfArgs, 2, 0),
		},
		{
			name: "fail - invalid to argument",
			args: []interface{}{
				"", "",
			},
			expectedPass:  false,
			errorContains: fmt.Sprintf(accesscontrol.ErrInvalidMinterAddress),
		},
		{
			name: "fail - minting to zero address",
			args: []interface{}{
				common.Address{}, "",
			},
			expectedPass:  false,
			errorContains: fmt.Sprintf(accesscontrol.ErrMintToZeroAddress),
		},
		{
			name: "fail - invalid amount argument",
			args: []interface{}{
				toAddr, "",
			},
			expectedPass:  false,
			errorContains: fmt.Sprintf(commonerr.ErrInvalidAmount, ""),
		},
		{
			name: "fail - amount is negative",
			args: []interface{}{
				toAddr, big.NewInt(-1),
			},
			expectedPass:  false,
			errorContains: fmt.Sprintf(accesscontrol.ErrMintAmountNotGreaterThanZero),
		},
		{
			name: "pass - valid arguments",
			args: []interface{}{
				toAddr, big.NewInt(1e18),
			},
			expectedPass:  true,
			errorContains: "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			address, amount, err := accesscontrol.ParseMintArgs(tc.args)

			if tc.expectedPass {
				s.NoError(err)
				s.Equal(big.NewInt(1e18), amount)
				s.Equal(toAddr, address)
			} else {
				s.Error(err)
				s.Contains(err.Error(), tc.errorContains)
			}
		})
	}
}
