package keeper_test

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	utiltx "github.com/evmos/evmos/v19/testutil/tx"
	"github.com/evmos/evmos/v19/x/erc20/types"
)

func (suite *KeeperTestSuite) TestMintingEnabled() {
	sender := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	receiver := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	expPair := types.NewTokenPair(utiltx.GenerateAddress(), "coin", types.OWNER_MODULE)
	id := expPair.GetID()

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"conversion is disabled globally",
			func() {
				params := types.DefaultParams()
				params.EnableErc20 = false
				suite.app.Erc20Keeper.SetParams(suite.ctx, params) //nolint:errcheck
			},
			false,
		},
		{
			"token pair not found",
			func() {},
			false,
		},
		{
			"conversion is disabled for the given pair",
			func() {
				expPair.Enabled = false
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, expPair)
				suite.app.Erc20Keeper.SetDenomMap(suite.ctx, expPair.Denom, id)
				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, expPair.GetERC20Contract(), id)
			},
			false,
		},
		{
			"token transfers are disabled",
			func() {
				expPair.Enabled = true
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, expPair)
				suite.app.Erc20Keeper.SetDenomMap(suite.ctx, expPair.Denom, id)
				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, expPair.GetERC20Contract(), id)

				params := banktypes.DefaultParams()
				params.SendEnabled = []*banktypes.SendEnabled{ //nolint:staticcheck
					{Denom: expPair.Denom, Enabled: false},
				}
				err := suite.app.BankKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"token not registered",
			func() {
				suite.app.Erc20Keeper.SetDenomMap(suite.ctx, expPair.Denom, id)
				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, expPair.GetERC20Contract(), id)
			},
			false,
		},
		{
			"receiver address is blocked (module account)",
			func() {
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, expPair)
				suite.app.Erc20Keeper.SetDenomMap(suite.ctx, expPair.Denom, id)
				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, expPair.GetERC20Contract(), id)

				acc := suite.app.AccountKeeper.GetModuleAccount(suite.ctx, types.ModuleName)
				receiver = acc.GetAddress()
			},
			false,
		},
		{
			"ok",
			func() {
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, expPair)
				suite.app.Erc20Keeper.SetDenomMap(suite.ctx, expPair.Denom, id)
				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, expPair.GetERC20Contract(), id)

				receiver = sdk.AccAddress(utiltx.GenerateAddress().Bytes())
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			pair, err := suite.app.Erc20Keeper.MintingEnabled(suite.ctx, sender, receiver, expPair.Erc20Address)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expPair, pair)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMintCoins() {
	sender := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	to := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	expPair := types.NewTokenPair(utiltx.GenerateAddress(), "coin", types.OWNER_MODULE)
	expPair.SetOwnerAddress(sender.String())
	amount := big.NewInt(1000000)
	id := expPair.GetID()

	params := types.DefaultParams()
	params.EnableErc20 = true
	s.app.Erc20Keeper.SetParams(s.ctx, params) //nolint:errcheck

	testcases := []struct {
		name        string
		malleate    func() 
		postCheck   func()
		expErr      bool
		errContains string
	}{
		{
			"fail - conversion is disabled globally",
			func() {
				params := types.DefaultParams()
				params.EnableErc20 = false
				s.app.Erc20Keeper.SetParams(s.ctx, params) //nolint:errcheck
			},
			func() {},
			true,
			"",
		},
		{
			"fail - token pair not found",
			func() {},
			func() {},
			true,
			"",
		},
		{
			"fail - conversion is disabled for the given pair",
			func() {
				expPair.Enabled = false
				s.app.Erc20Keeper.SetTokenPair(s.ctx, expPair)
				s.app.Erc20Keeper.SetDenomMap(s.ctx, expPair.Denom, id)
				s.app.Erc20Keeper.SetERC20Map(s.ctx, expPair.GetERC20Contract(), id)
			},
			func() {},
			true,
			"",
		},
		{
			"fail - token transfers are disabled",
			func() {
				expPair.Enabled = true
				s.app.Erc20Keeper.SetTokenPair(s.ctx, expPair)
				s.app.Erc20Keeper.SetDenomMap(s.ctx, expPair.Denom, id)
				s.app.Erc20Keeper.SetERC20Map(s.ctx, expPair.GetERC20Contract(), id)

				params := banktypes.DefaultParams()
				params.SendEnabled = []*banktypes.SendEnabled{ //nolint:staticcheck
					{Denom: expPair.Denom, Enabled: false},
				}
				err := s.app.BankKeeper.SetParams(s.ctx, params)
				s.Require().NoError(err)
			},
			func() {},
			true,
			"",
		},
		{
			"fail - token not registered",
			func() {
				s.app.Erc20Keeper.SetDenomMap(s.ctx, expPair.Denom, id)
				s.app.Erc20Keeper.SetERC20Map(s.ctx, expPair.GetERC20Contract(), id)
			},
			func() {},
			true,
			"",
		},
		{
			"fail - receiver address is blocked (module account)",
			func() {
				s.app.Erc20Keeper.SetTokenPair(s.ctx, expPair)
				s.app.Erc20Keeper.SetDenomMap(s.ctx, expPair.Denom, id)
				s.app.Erc20Keeper.SetERC20Map(s.ctx, expPair.GetERC20Contract(), id)

				acc := s.app.AccountKeeper.GetModuleAccount(s.ctx, types.ModuleName)
				to = acc.GetAddress()
			},
			func() {},
			true,
			"",
		},
		{
			"fail - pair is not native coin",
			func() {
				expPair.ContractOwner = types.OWNER_EXTERNAL
				s.app.Erc20Keeper.SetTokenPair(s.ctx, expPair)
				s.app.Erc20Keeper.SetDenomMap(s.ctx, expPair.Denom, id)
				s.app.Erc20Keeper.SetERC20Map(s.ctx, expPair.GetERC20Contract(), id)

				to = sdk.AccAddress(utiltx.GenerateAddress().Bytes())
			},
			func() {},
			true,
			types.ErrERC20TokenPairDisabled.Error(),
		},
		{
			"fail - minter is not the owner",
			func() {
				expPair.ContractOwner = types.OWNER_MODULE
				expPair.SetOwnerAddress(sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String())
				s.app.Erc20Keeper.SetTokenPair(s.ctx, expPair)
				s.app.Erc20Keeper.SetDenomMap(s.ctx, expPair.Denom, id)
				s.app.Erc20Keeper.SetERC20Map(s.ctx, expPair.GetERC20Contract(), id)

			},
			func() {},
			true,
			authz.ErrNoAuthorizationFound.Error(),
		},
		{
			"pass",
			func() {
				expPair.SetOwnerAddress(sender.String())
				s.app.Erc20Keeper.SetTokenPair(s.ctx, expPair)
				s.app.Erc20Keeper.SetDenomMap(s.ctx, expPair.Denom, id)
				s.app.Erc20Keeper.SetERC20Map(s.ctx, expPair.GetERC20Contract(), id)

				to = sdk.AccAddress(utiltx.GenerateAddress().Bytes())
			},
			func() {},
			false,
			"",
		},
	}	

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			tc.malleate()

			err := s.app.Erc20Keeper.MintCoins(s.ctx, sender, to, math.NewIntFromBigInt(amount), expPair.Erc20Address)
			if tc.expErr {
				s.Require().Error(err, "expected transfer transaction to fail")
				s.Require().Contains(err.Error(), tc.errContains, "expected transfer transaction to fail with specific error")
			} else {
				s.Require().NoError(err, "expected transfer transaction succeeded")
				tc.postCheck()
			}
		})
	}
}

func (s *KeeperTestSuite) TestBurnCoins() {
	sender := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	expPair := types.NewTokenPair(utiltx.GenerateAddress(), "coin", types.OWNER_MODULE)
	expPair.SetOwnerAddress(sender.String())
	amount := big.NewInt(1000000)
	id := expPair.GetID()

	params := types.DefaultParams()
	params.EnableErc20 = true
	s.app.Erc20Keeper.SetParams(s.ctx, params) //nolint:errcheck	

	testcases := []struct {
		name        string
		malleate    func() 
		postCheck   func()
		expErr      bool
		errContains string
	}{
		{
			name: "fail - token pair not found",
			malleate: func() {},
			postCheck: func() {},
			expErr: true,
			errContains: "",
		},
		{
			"fail - pair is not native coin",
			func() {
				expPair.ContractOwner = types.OWNER_EXTERNAL
				s.app.Erc20Keeper.SetTokenPair(s.ctx, expPair)
				s.app.Erc20Keeper.SetDenomMap(s.ctx, expPair.Denom, id)
				s.app.Erc20Keeper.SetERC20Map(s.ctx, expPair.GetERC20Contract(), id)
			},
			func() {},
			true,
			types.ErrERC20TokenPairDisabled.Error(),
		},
		{
			"pass",
			func() {
				expPair.ContractOwner = types.OWNER_MODULE
				if err := s.app.BankKeeper.MintCoins(s.ctx, types.ModuleName, sdk.Coins{{Denom: expPair.Denom, Amount: math.NewIntFromBigInt(amount)}}); err != nil {
					s.FailNow(err.Error())
				}
				if err := s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, sender, sdk.Coins{{Denom: expPair.Denom, Amount: math.NewIntFromBigInt(amount)}}); err != nil {
					s.FailNow(err.Error())
				}
				expPair.SetOwnerAddress(sender.String())
				s.app.Erc20Keeper.SetTokenPair(s.ctx, expPair)
				s.app.Erc20Keeper.SetDenomMap(s.ctx, expPair.Denom, id)
				s.app.Erc20Keeper.SetERC20Map(s.ctx, expPair.GetERC20Contract(), id)
			},
			func() {
				balance := s.app.BankKeeper.GetBalance(s.ctx, sender, expPair.Denom)
				s.Require().Equal(balance.Amount.Int64(), math.NewInt(0).Int64())
			},
			false,
			"",
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			tc.malleate()

			err := s.app.Erc20Keeper.BurnCoins(s.ctx, sender, math.NewIntFromBigInt(amount), expPair.Erc20Address)
			if tc.expErr {
				s.Require().Error(err, "expected transfer transaction to fail")
				s.Require().Contains(err.Error(), tc.errContains, "expected transfer transaction to fail with specific error")
			} else {
				s.Require().NoError(err, "expected transfer transaction succeeded")
				tc.postCheck()
			}
		})
	}
}

func (s *KeeperTestSuite) TestTransferOwnership() {
	sender := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	newOwner := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	expPair := types.NewTokenPair(utiltx.GenerateAddress(), "coin", types.OWNER_MODULE)
	expPair.SetOwnerAddress(sender.String())
	id := expPair.GetID()

	params := types.DefaultParams()
	params.EnableErc20 = true
	s.app.Erc20Keeper.SetParams(s.ctx, params) //nolint:errcheck

	testcases := []struct {
		name        string
		malleate    func()
		postCheck   func()
		expErr      bool
		errContains string
	}{
		{
			"fail - token pair not found",
			func() {},
			func() {},
			true,
			"",
		},
		{
			"fail - pair is not native coin",
			func() {
				expPair.ContractOwner = types.OWNER_EXTERNAL
				s.app.Erc20Keeper.SetTokenPair(s.ctx, expPair)
				s.app.Erc20Keeper.SetDenomMap(s.ctx, expPair.Denom, id)
				s.app.Erc20Keeper.SetERC20Map(s.ctx, expPair.GetERC20Contract(), id)
			},
			func() {},
			true,
			types.ErrERC20TokenPairDisabled.Error(),
		},
		{
			"pass",
			func() {
				expPair.ContractOwner = types.OWNER_MODULE
				expPair.SetOwnerAddress(sender.String())
				s.app.Erc20Keeper.SetTokenPair(s.ctx, expPair)
				s.app.Erc20Keeper.SetDenomMap(s.ctx, expPair.Denom, id)
				s.app.Erc20Keeper.SetERC20Map(s.ctx, expPair.GetERC20Contract(), id)
			},
			func() {},
			false,
			"",
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			tc.malleate()

			err := s.app.Erc20Keeper.TransferOwnership(s.ctx, newOwner, expPair.Denom)
			if tc.expErr {
				s.Require().Error(err, "expected transfer transaction to fail")
				s.Require().Contains(err.Error(), tc.errContains, "expected transfer transaction to fail with specific error")
			} else {
				s.Require().NoError(err, "expected transfer transaction succeeded")
				tc.postCheck()
			}
		})
	}
}
