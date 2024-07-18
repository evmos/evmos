package bech32_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/cmd/config"
	"github.com/evmos/evmos/v19/precompiles/bech32"
	cmn "github.com/evmos/evmos/v19/precompiles/common"
)

func (s *PrecompileTestSuite) TestHexToBech32() {
	// setup basic test suite
	s.SetupTest()

	method := s.precompile.Methods[bech32.HexToBech32Method]

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func(data []byte)
		expError    bool
		errContains string
	}{
		{
			"fail - invalid args length",
			func() []interface{} {
				return []interface{}{}
			},
			func([]byte) {},
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 2, 0),
		},
		{
			"fail - invalid hex address",
			func() []interface{} {
				return []interface{}{
					"",
					"",
				}
			},
			func([]byte) {},
			true,
			"invalid hex address",
		},
		{
			"fail - invalid bech32 HRP",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					"",
				}
			},
			func([]byte) {},
			true,
			"invalid bech32 human readable prefix (HRP)",
		},
		{
			"pass - valid hex address and valid bech32 HRP",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					config.Bech32Prefix,
				}
			},
			func(data []byte) {
				args, err := s.precompile.Unpack(bech32.HexToBech32Method, data)
				s.Require().NoError(err, "failed to unpack output")
				s.Require().Len(args, 1)
				addr, ok := args[0].(string)
				s.Require().True(ok)
				s.Require().Equal(s.keyring.GetAccAddr(0).String(), addr)
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			bz, err := s.precompile.HexToBech32(&method, tc.malleate())

			if tc.expError {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.errContains, err.Error())
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)
				tc.postCheck(bz)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestBech32ToHex() {
	// setup basic test suite
	s.SetupTest()

	method := s.precompile.Methods[bech32.Bech32ToHexMethod]

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func(data []byte)
		expError    bool
		errContains string
	}{
		{
			"fail - invalid args length",
			func() []interface{} {
				return []interface{}{}
			},
			func([]byte) {},
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 1, 0),
		},
		{
			"fail - empty bech32 address",
			func() []interface{} {
				return []interface{}{
					"",
				}
			},
			func([]byte) {},
			true,
			"invalid bech32 address",
		},
		{
			"fail - invalid bech32 address",
			func() []interface{} {
				return []interface{}{
					config.Bech32Prefix,
				}
			},
			func([]byte) {},
			true,
			fmt.Sprintf("invalid bech32 address: %s", config.Bech32Prefix),
		},
		{
			"fail - decoding bech32 failed",
			func() []interface{} {
				return []interface{}{
					config.Bech32Prefix + "1",
				}
			},
			func([]byte) {},
			true,
			"decoding bech32 failed",
		},
		{
			"fail - invalid address format",
			func() []interface{} {
				return []interface{}{
					sdk.AccAddress(make([]byte, 256)).String(),
				}
			},
			func([]byte) {},
			true,
			"address max length is 255",
		},
		{
			"success - valid bech32 address",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAccAddr(0).String(),
				}
			},
			func(data []byte) {
				args, err := s.precompile.Unpack(bech32.Bech32ToHexMethod, data)
				s.Require().NoError(err, "failed to unpack output")
				s.Require().Len(args, 1)
				addr, ok := args[0].(common.Address)
				s.Require().True(ok)
				s.Require().Equal(s.keyring.GetAddr(0), addr)
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			bz, err := s.precompile.Bech32ToHex(&method, tc.malleate())

			if tc.expError {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)
				tc.postCheck(bz)
			}
		})
	}
}
