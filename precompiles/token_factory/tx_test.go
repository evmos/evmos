package tokenfactory_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core/vm"
	access "github.com/evmos/evmos/v18/precompiles/access_control"
	"github.com/evmos/evmos/v18/precompiles/testutil"
	tokenfactory "github.com/evmos/evmos/v18/precompiles/token_factory"
)

const (
	// TokenName is the name of the test token
	TokenName = "TEST"
	// TokenDenom is the symbol of the test token
	TokenDenom = "TST"
)

func (s *PrecompileTestSuite) TestCreateERC20() {
	method := s.precompile.Methods[tokenfactory.MethodCreateERC20]

	fromAddr := s.keyring.GetKey(0).Addr

	testCases := []struct {
		name          string
		args          []interface{}
		expectedError bool
		postCheck     func()
	}{
		{
			"pass - creates an ERC20 token factory token with access control",
			[]interface{}{TokenName, TokenDenom, uint8(18), big.NewInt(1e18)},
			false,
			func() {
				// Check token pair exists
				tokenPairs := s.network.App.Erc20Keeper.GetTokenPairs(s.network.GetContext())
				s.Require().Len(tokenPairs, 1)

				// Check the bank keeper for the token
				balance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), fromAddr.Bytes(), tokenPairs[0].Denom)
				s.Require().Equal(sdk.NewInt(1e18), balance.Amount)

				// Check denom metadata
				denomMetadata, found := s.network.App.BankKeeper.GetDenomMetaData(s.network.GetContext(), tokenPairs[0].Denom)
				s.Require().True(found)
				s.Require().Equal(denomMetadata.Name, TokenName)
				s.Require().Equal(denomMetadata.Symbol, TokenDenom)
				s.Require().Equal(denomMetadata.Base, tokenPairs[0].Denom)
				s.Require().Equal(denomMetadata.DenomUnits[1].Exponent, uint32(18))

				// Check access control set correctly
				owner, found := s.network.App.AccessControlKeeper.GetOwner(s.network.GetContext(), tokenPairs[0].GetERC20Contract())
				s.Require().True(found)
				s.Require().Equal(owner, fromAddr)

				// Check roles set correctly
				hasAdminRole := s.network.App.AccessControlKeeper.HasRole(s.network.GetContext(), tokenPairs[0].GetERC20Contract(), access.RoleDefaultAdmin, fromAddr)
				s.Require().True(hasAdminRole)
				hasMinter := s.network.App.AccessControlKeeper.HasRole(s.network.GetContext(), tokenPairs[0].GetERC20Contract(), access.RoleMinter, fromAddr)
				s.Require().True(hasMinter)
				hasBurner := s.network.App.AccessControlKeeper.HasRole(s.network.GetContext(), tokenPairs[0].GetERC20Contract(), access.RoleBurner, fromAddr)
				s.Require().True(hasBurner)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()
			stateDB := s.network.GetStateDB()

			var contract *vm.Contract
			contract, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), fromAddr, s.precompile, 0)

			_, err := s.precompile.CreateERC20(ctx, contract, stateDB, &method, tc.args)

			if tc.expectedError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}
