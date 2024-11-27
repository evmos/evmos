package erc20_test

import (
	"math/big"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/precompiles/erc20"
	"github.com/evmos/evmos/v20/precompiles/testutil"
	utiltx "github.com/evmos/evmos/v20/testutil/tx"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
	"github.com/evmos/evmos/v20/x/evm/statedb"
)

var (
	tokenDenom = "xmpl"
	// XMPLCoin is a dummy coin used for testing purposes.
	XMPLCoin = sdk.NewCoins(sdk.NewInt64Coin(tokenDenom, 1e18))
	// toAddr is a dummy address used for testing purposes.
	toAddr = utiltx.GenerateAddress()
)

func (s *PrecompileTestSuite) TestTransfer() {
	method := s.precompile.Methods[erc20.TransferMethod]
	// fromAddr is the address of the keyring account used for testing.
	fromAddr := s.keyring.GetKey(0).Addr
	testcases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func()
		expErr      bool
		errContains string
	}{
		{
			"fail - negative amount",
			func() []interface{} {
				return []interface{}{toAddr, big.NewInt(-1)}
			},
			func() {},
			true,
			"coin -1xmpl amount is not positive",
		},
		{
			"fail - invalid to address",
			func() []interface{} {
				return []interface{}{"", big.NewInt(100)}
			},
			func() {},
			true,
			"invalid to address",
		},
		{
			"fail - invalid amount",
			func() []interface{} {
				return []interface{}{toAddr, ""}
			},
			func() {},
			true,
			"invalid amount",
		},
		{
			"fail - not enough balance",
			func() []interface{} {
				return []interface{}{toAddr, big.NewInt(2e18)}
			},
			func() {},
			true,
			erc20.ErrTransferAmountExceedsBalance.Error(),
		},
		{
			"pass",
			func() []interface{} {
				return []interface{}{toAddr, big.NewInt(100)}
			},
			func() {
				toAddrBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), toAddr.Bytes(), tokenDenom)
				s.Require().Equal(big.NewInt(100), toAddrBalance.Amount.BigInt(), "expected toAddr to have 100 XMPL")
			},
			false,
			"",
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()
			stateDB := s.network.GetStateDB()

			var contract *vm.Contract
			contract, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), fromAddr, s.precompile, 0)

			// Mint some coins to the module account and then send to the from address
			err := s.network.App.BankKeeper.MintCoins(s.network.GetContext(), erc20types.ModuleName, XMPLCoin)
			s.Require().NoError(err, "failed to mint coins")
			err = s.network.App.BankKeeper.SendCoinsFromModuleToAccount(s.network.GetContext(), erc20types.ModuleName, fromAddr.Bytes(), XMPLCoin)
			s.Require().NoError(err, "failed to send coins from module to account")

			_, err = s.precompile.Transfer(ctx, contract, stateDB, &method, tc.malleate())
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

func (s *PrecompileTestSuite) TestTransferFrom() {
	var (
		ctx  sdk.Context
		stDB *statedb.StateDB
	)
	method := s.precompile.Methods[erc20.TransferFromMethod]
	// owner of the tokens
	owner := s.keyring.GetKey(0)
	// spender of the tokens
	spender := s.keyring.GetKey(1)

	testcases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func()
		expErr      bool
		errContains string
	}{
		{
			"fail - negative amount",
			func() []interface{} {
				return []interface{}{owner.Addr, toAddr, big.NewInt(-1)}
			},
			func() {},
			true,
			"coin -1xmpl amount is not positive",
		},
		{
			"fail - invalid from address",
			func() []interface{} {
				return []interface{}{"", toAddr, big.NewInt(100)}
			},
			func() {},
			true,
			"invalid from address",
		},
		{
			"fail - invalid to address",
			func() []interface{} {
				return []interface{}{owner.Addr, "", big.NewInt(100)}
			},
			func() {},
			true,
			"invalid to address",
		},
		{
			"fail - invalid amount",
			func() []interface{} {
				return []interface{}{owner.Addr, toAddr, ""}
			},
			func() {},
			true,
			"invalid amount",
		},
		{
			"fail - not enough allowance",
			func() []interface{} {
				return []interface{}{owner.Addr, toAddr, big.NewInt(100)}
			},
			func() {},
			true,
			erc20.ErrInsufficientAllowance.Error(),
		},
		{
			"fail - not enough balance",
			func() []interface{} {
				expiration := time.Now().Add(time.Hour)
				err := s.network.App.AuthzKeeper.SaveGrant(
					ctx,
					spender.AccAddr,
					owner.AccAddr,
					&banktypes.SendAuthorization{SpendLimit: sdk.Coins{sdk.Coin{Denom: s.tokenDenom, Amount: math.NewInt(5e18)}}},
					&expiration,
				)
				s.Require().NoError(err, "failed to save grant")

				return []interface{}{owner.Addr, toAddr, big.NewInt(2e18)}
			},
			func() {},
			true,
			erc20.ErrTransferAmountExceedsBalance.Error(),
		},
		{
			"pass - spend on behalf of other account",
			func() []interface{} {
				expiration := time.Now().Add(time.Hour)
				err := s.network.App.AuthzKeeper.SaveGrant(
					ctx,
					spender.AccAddr,
					owner.AccAddr,
					&banktypes.SendAuthorization{SpendLimit: sdk.Coins{sdk.Coin{Denom: tokenDenom, Amount: math.NewInt(300)}}},
					&expiration,
				)
				s.Require().NoError(err, "failed to save grant")

				return []interface{}{owner.Addr, toAddr, big.NewInt(100)}
			},
			func() {
				toAddrBalance := s.network.App.BankKeeper.GetBalance(ctx, toAddr.Bytes(), tokenDenom)
				s.Require().Equal(big.NewInt(100), toAddrBalance.Amount.BigInt(), "expected toAddr to have 100 XMPL")
			},
			false,
			"",
		},
		{
			"pass - spend on behalf of own account",
			func() []interface{} {
				// Mint some coins to the module account and then send to the spender address
				err := s.network.App.BankKeeper.MintCoins(ctx, erc20types.ModuleName, XMPLCoin)
				s.Require().NoError(err, "failed to mint coins")
				err = s.network.App.BankKeeper.SendCoinsFromModuleToAccount(ctx, erc20types.ModuleName, spender.AccAddr, XMPLCoin)
				s.Require().NoError(err, "failed to send coins from module to account")

				// NOTE: no authorization is necessary to spend on behalf of the same account
				return []interface{}{spender.Addr, toAddr, big.NewInt(100)}
			},
			func() {
				toAddrBalance := s.network.App.BankKeeper.GetBalance(ctx, toAddr.Bytes(), tokenDenom)
				s.Require().Equal(big.NewInt(100), toAddrBalance.Amount.BigInt(), "expected toAddr to have 100 XMPL")
			},
			false,
			"",
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx = s.network.GetContext()
			stDB = s.network.GetStateDB()

			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(s.T(), ctx, spender.Addr, s.precompile, 0)

			// Mint some coins to the module account and then send to the from address
			err := s.network.App.BankKeeper.MintCoins(ctx, erc20types.ModuleName, XMPLCoin)
			s.Require().NoError(err, "failed to mint coins")
			err = s.network.App.BankKeeper.SendCoinsFromModuleToAccount(ctx, erc20types.ModuleName, owner.AccAddr, XMPLCoin)
			s.Require().NoError(err, "failed to send coins from module to account")

			_, err = s.precompile.TransferFrom(ctx, contract, stDB, &method, tc.malleate())
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

func (s *PrecompileTestSuite) TestMint() {
	method := s.precompile.Methods[erc20.MintMethod]
	sender := s.keyring.GetKey(0)
	spender := s.keyring.GetKey(1)

	testcases := []struct {
		name        string
		malleate    func() ([]interface{}, erc20types.TokenPair)
		postCheck   func()
		expErr      bool
		errContains string
	}{
		{
			"fail - negative amount",
			func() ([]interface{}, erc20types.TokenPair) {
				tokenPair := erc20types.NewTokenPair(utiltx.GenerateAddress(), s.tokenDenom, erc20types.OWNER_MODULE)
				tokenPair.SetOwnerAddress(sender.AccAddr.String())
				s.network.App.Erc20Keeper.SetTokenPair(s.network.GetContext(), tokenPair)
				s.network.App.Erc20Keeper.SetDenomMap(s.network.GetContext(), tokenPair.Denom, tokenPair.GetID())
				s.network.App.Erc20Keeper.SetERC20Map(s.network.GetContext(), tokenPair.GetERC20Contract(), tokenPair.GetID())
				return []interface{}{toAddr, big.NewInt(-1)}, tokenPair
			},
			func() {},
			true,
			"-1xmpl: invalid coins",
		},
		{
			"fail - invalid to address",
			func() ([]interface{}, erc20types.TokenPair) {
				tokenPair := erc20types.NewTokenPair(utiltx.GenerateAddress(), s.tokenDenom, erc20types.OWNER_MODULE)
				tokenPair.SetOwnerAddress(sender.AccAddr.String())
				s.network.App.Erc20Keeper.SetTokenPair(s.network.GetContext(), tokenPair)
				s.network.App.Erc20Keeper.SetDenomMap(s.network.GetContext(), tokenPair.Denom, tokenPair.GetID())
				s.network.App.Erc20Keeper.SetERC20Map(s.network.GetContext(), tokenPair.GetERC20Contract(), tokenPair.GetID())
				return []interface{}{"", big.NewInt(100)}, tokenPair
			},
			func() {},
			true,
			"invalid to address",
		},
		{
			"fail - invalid amount",
			func() ([]interface{}, erc20types.TokenPair) {
				tokenPair := erc20types.NewTokenPair(utiltx.GenerateAddress(), s.tokenDenom, erc20types.OWNER_MODULE)
				tokenPair.SetOwnerAddress(sender.AccAddr.String())
				s.network.App.Erc20Keeper.SetTokenPair(s.network.GetContext(), tokenPair)
				s.network.App.Erc20Keeper.SetDenomMap(s.network.GetContext(), tokenPair.Denom, tokenPair.GetID())
				s.network.App.Erc20Keeper.SetERC20Map(s.network.GetContext(), tokenPair.GetERC20Contract(), tokenPair.GetID())
				return []interface{}{toAddr, ""}, tokenPair
			},
			func() {},
			true,
			"invalid amount",
		},
		{
			"fail - minter is not the owner",
			func() ([]interface{}, erc20types.TokenPair) {
				tokenPair := erc20types.NewTokenPair(utiltx.GenerateAddress(), s.tokenDenom, erc20types.OWNER_MODULE)
				tokenPair.SetOwnerAddress(sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String())
				s.network.App.Erc20Keeper.SetTokenPair(s.network.GetContext(), tokenPair)
				s.network.App.Erc20Keeper.SetDenomMap(s.network.GetContext(), tokenPair.Denom, tokenPair.GetID())
				s.network.App.Erc20Keeper.SetERC20Map(s.network.GetContext(), tokenPair.GetERC20Contract(), tokenPair.GetID())
				return []interface{}{spender.Addr, big.NewInt(100)}, tokenPair
			},
			func() {},
			true,
			erc20types.ErrMinterIsNotOwner.Error(),
		},
		{
			"pass",
			func() ([]interface{}, erc20types.TokenPair) {
				tokenPair := erc20types.NewTokenPair(utiltx.GenerateAddress(), s.tokenDenom, erc20types.OWNER_MODULE)
				tokenPair.SetOwnerAddress(sender.AccAddr.String())
				s.network.App.Erc20Keeper.SetTokenPair(s.network.GetContext(), tokenPair)
				s.network.App.Erc20Keeper.SetDenomMap(s.network.GetContext(), tokenPair.Denom, tokenPair.GetID())
				s.network.App.Erc20Keeper.SetERC20Map(s.network.GetContext(), tokenPair.GetERC20Contract(), tokenPair.GetID())

				coins := sdk.Coins{{Denom: tokenDenom, Amount: math.NewInt(100)}}
				err := s.network.App.BankKeeper.MintCoins(s.network.GetContext(), erc20types.ModuleName, coins)
				s.Require().NoError(err, "failed to mint coins")
				err = s.network.App.BankKeeper.SendCoinsFromModuleToAccount(s.network.GetContext(), erc20types.ModuleName, sdk.AccAddress(toAddr.Bytes()), coins)
				s.Require().NoError(err, "failed to send coins from module to account")
				return []interface{}{spender.Addr, big.NewInt(100)}, tokenPair
			},
			func() {
				toAddrBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), toAddr.Bytes(), tokenDenom)
				s.Require().Equal(big.NewInt(100), toAddrBalance.Amount.BigInt(), "expected toAddr to have 100 XMPL")
			},
			false,
			"",
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()
			stateDB := s.network.GetStateDB()

			args, tokenPair := tc.malleate()

			precompile, err := setupERC20PrecompileForTokenPair(*s.network, tokenPair)
			s.Require().NoError(err, "failed to set up %q erc20 precompile", tokenPair.Denom)

			var contract *vm.Contract
			contract, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), sender.Addr, precompile, 0)

			// Mint some coins to the module account and then send to the from address
			err = s.network.App.BankKeeper.MintCoins(s.network.GetContext(), erc20types.ModuleName, XMPLCoin)
			s.Require().NoError(err, "failed to mint coins")
			err = s.network.App.BankKeeper.SendCoinsFromModuleToAccount(s.network.GetContext(), erc20types.ModuleName, sender.AccAddr, XMPLCoin)
			s.Require().NoError(err, "failed to send coins from module to account")

			_, err = precompile.Mint(ctx, contract, stateDB, &method, args)
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

func (s *PrecompileTestSuite) TestBurn() {
	method := s.precompile.Methods[erc20.BurnMethod]
	from := s.keyring.GetKey(0)
	testcases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func()
		expErr      bool
		errContains string
	}{
		{
			"pass",
			func() []interface{} {
				return []interface{}{big.NewInt(100)}
			},
			func() {
				fromBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), from.AccAddr, tokenDenom)
				s.Require().Equal(big.NewInt(0), fromBalance.Amount.BigInt(), "expected fromAddr to have 0 XMPL")
			},
			false,
			"",
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()
			stateDB := s.network.GetStateDB()
			coins := sdk.Coins{{Denom: tokenDenom, Amount: math.NewInt(100)}}

			tokenPair := erc20types.NewTokenPair(utiltx.GenerateAddress(), s.tokenDenom, erc20types.OWNER_MODULE)
			tokenPair.SetOwnerAddress(from.AccAddr.String())
			s.network.App.Erc20Keeper.SetTokenPair(s.network.GetContext(), tokenPair)
			s.network.App.Erc20Keeper.SetDenomMap(s.network.GetContext(), tokenPair.Denom, tokenPair.GetID())
			s.network.App.Erc20Keeper.SetERC20Map(s.network.GetContext(), tokenPair.GetERC20Contract(), tokenPair.GetID())

			precompile, err := setupERC20PrecompileForTokenPair(*s.network, tokenPair)
			s.Require().NoError(err, "failed to set up %q erc20 precompile", tokenPair.Denom)

			var contract *vm.Contract
			contract, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), from.Addr, precompile, 0)

			// Mint some coins to the module account and then send to the from address
			err = s.network.App.BankKeeper.MintCoins(s.network.GetContext(), erc20types.ModuleName, coins)
			s.Require().NoError(err, "failed to mint coins")
			err = s.network.App.BankKeeper.SendCoinsFromModuleToAccount(s.network.GetContext(), erc20types.ModuleName, from.AccAddr, coins)
			s.Require().NoError(err, "failed to send coins from module to account")

			_, err = precompile.Burn(ctx, contract, stateDB, &method, tc.malleate())
			if tc.expErr {
				s.Require().Error(err, "expected burn transaction to fail")
				s.Require().Contains(err.Error(), tc.errContains, "expected burn transaction to fail with specific error")
			} else {
				s.Require().NoError(err, "expected transfer transaction succeeded")
				tc.postCheck()
			}
		})
	}
}

func (s *PrecompileTestSuite) TestTransferOwnership() {
	method := s.precompile.Methods[erc20.TransferOwnershipMethod]
	from := s.keyring.GetKey(0)
	newOwner := common.Address(utiltx.GenerateAddress().Bytes())

	testcases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func()
		expErr      bool
		errContains string
	}{
		{
			name: "fail - invalid number of arguments",
			malleate: func() []interface{} {
				return []interface{}{}
			},
			expErr:      true,
			errContains: "invalid number of arguments; expected 1; got: 0",
		},
		{
			name: "fail - invalid address",
			malleate: func() []interface{} {
				return []interface{}{"invalid"}
			},
			expErr:      true,
			errContains: "invalid new owner address",
		},
		{
			name: "pass",
			malleate: func() []interface{} {
				return []interface{}{newOwner}
			},
			postCheck: func() {},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()
			stateDB := s.network.GetStateDB()

			tokenPair := erc20types.NewTokenPair(utiltx.GenerateAddress(), s.tokenDenom, erc20types.OWNER_MODULE)
			tokenPair.SetOwnerAddress(from.AccAddr.String())
			s.network.App.Erc20Keeper.SetTokenPair(s.network.GetContext(), tokenPair)
			s.network.App.Erc20Keeper.SetDenomMap(s.network.GetContext(), tokenPair.Denom, tokenPair.GetID())
			s.network.App.Erc20Keeper.SetERC20Map(s.network.GetContext(), tokenPair.GetERC20Contract(), tokenPair.GetID())

			precompile, err := setupERC20PrecompileForTokenPair(*s.network, tokenPair)
			s.Require().NoError(err, "failed to set up %q erc20 precompile", tokenPair.Denom)

			var contract *vm.Contract
			contract, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), from.Addr, s.precompile, 0)

			_, err = precompile.TransferOwnership(ctx, contract, stateDB, &method, tc.malleate())
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}
