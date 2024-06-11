package accesscontrol_test

import (
	"cosmossdk.io/math"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core/vm"
	accesscontrol "github.com/evmos/evmos/v18/precompiles/access_control"
	"github.com/evmos/evmos/v18/precompiles/testutil"
	erc20types "github.com/evmos/evmos/v18/x/erc20/types"
	"math/big"
)

func (s *PrecompileTestSuite) TestMint() {
	toAddr := s.keyring.GetKey(0).Addr
	testCases := []struct {
		name        string
		malleate    func() []interface{}
		expErr      bool
		errContains string
	}{
		{
			"fail - minter doesn't have mint role",
			func() []interface{} {
				return []interface{}{toAddr, big.NewInt(1e18)}
			},
			true,
			fmt.Sprintf(accesscontrol.ErrSenderNoRole),
		},
		{
			"pass - mints Bank coins to the recipient address",
			func() []interface{} {
				// Set the minter role
				s.network.App.AccessControlKeeper.SetRole(s.network.GetContext(), s.tokenPair.GetERC20Contract(), accesscontrol.RoleMinter, toAddr)
				return []interface{}{toAddr, big.NewInt(1e18)}
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			stateDB := s.network.GetStateDB()

			var contract *vm.Contract
			contract, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), toAddr, s.precompile, 0)

			balanceBefore := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), toAddr.Bytes(), s.tokenPair.Denom)
			s.Require().Equal(sdk.NewInt(0), balanceBefore.Amount)

			_, err := s.precompile.Mint(ctx, contract, stateDB, tc.malleate())

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				balanceAfter := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), toAddr.Bytes(), s.tokenPair.Denom)
				s.Require().Equal(sdk.NewInt(1e18), balanceAfter.Amount)
				s.Require().NoError(err)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestBurn() {
	toAddr := s.keyring.GetKey(0).Addr
	amount := big.NewInt(1e18)
	testCases := []struct {
		name        string
		malleate    func() []interface{}
		expErr      bool
		errContains string
	}{
		{
			"fail - burner doesn't have burn role",
			func() []interface{} {
				return []interface{}{amount}
			},
			true,
			fmt.Sprintf(accesscontrol.ErrSenderNoRole),
		},
		{
			"pass - burns Bank coins from the sender address",
			func() []interface{} {
				// Set the minter role
				s.network.App.AccessControlKeeper.SetRole(s.network.GetContext(), s.tokenPair.GetERC20Contract(), accesscontrol.RoleBurner, toAddr)

				balanceBefore := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), toAddr.Bytes(), s.tokenPair.Denom)
				s.Require().Equal(math.NewInt(0), balanceBefore.Amount)

				// Mint new Coins and send them to the recipient address
				err := s.network.App.BankKeeper.MintCoins(s.network.GetContext(), erc20types.ModuleName, sdk.Coins{{Denom: s.tokenPair.Denom, Amount: math.NewIntFromBigInt(amount)}})
				s.Require().NoError(err)
				err = s.network.App.BankKeeper.SendCoinsFromModuleToAccount(s.network.GetContext(), erc20types.ModuleName, toAddr.Bytes(), sdk.Coins{{Denom: s.tokenPair.Denom, Amount: math.NewIntFromBigInt(amount)}})
				s.Require().NoError(err)

				balanceAfterMint := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), toAddr.Bytes(), s.tokenPair.Denom)
				s.Require().Equal(math.NewInt(amount.Int64()), balanceAfterMint.Amount)

				return []interface{}{amount}
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			stateDB := s.network.GetStateDB()

			var contract *vm.Contract
			contract, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), toAddr, s.precompile, 0)

			_, err := s.precompile.Burn(ctx, contract, stateDB, tc.malleate())

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				balanceAfter := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), toAddr.Bytes(), s.tokenPair.Denom)
				s.Require().Equal(math.NewInt(0), balanceAfter.Amount)
				s.Require().NoError(err)
			}
		})
	}
}
