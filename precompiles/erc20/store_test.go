package erc20_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/precompiles/erc20"
)

func (s *PrecompileTestSuite) TestSetContractOwnerAddress() {

	testcases := []struct {
		name string
		malleate func(ctx sdk.Context, precompile *erc20.Precompile)
		expPass bool
		errContains string
	}{
		{
			name: "success",
			malleate: nil,
			expPass: true,
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			ctx := s.network.GetContext()
			precompile := s.setupERC20Precompile(s.tokenDenom, s.keyring.GetAccAddr(0).String())

			if tc.malleate != nil {
				tc.malleate(ctx, precompile)
			}

			err := precompile.SetContractOwnerAddress(ctx, s.keyring.GetAccAddr(1))
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestGetContractOwnerAddress() {
	testcases := []struct {
		name string
		malleate func(ctx sdk.Context, precompile *erc20.Precompile) *erc20.Precompile
		expOwner sdk.AccAddress
		expPass bool
		errContains string
	}{
		{
			name: "success",
			malleate: nil,
			expOwner: s.keyring.GetAccAddr(1),
			expPass: true,
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			ctx := s.network.GetContext()
			precompile := s.setupERC20Precompile(s.tokenDenom, s.keyring.GetAccAddr(1).String())

			if tc.malleate != nil {
				precompile = tc.malleate(ctx, precompile)
			}

			owner, err := precompile.GetContractOwnerAddress(ctx)
			if tc.expPass {
				s.Require().NoError(err)
				s.Require().Equal(tc.expOwner, owner)
			} else {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			}
		})
	}	
}