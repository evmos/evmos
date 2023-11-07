package erc20_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v15/precompiles/erc20"
	"github.com/evmos/evmos/v15/precompiles/testutil"
	utiltx "github.com/evmos/evmos/v15/testutil/tx"
	erc20types "github.com/evmos/evmos/v15/x/erc20/types"
)

var (
	// XMPLCoin is a dummy coin used for testing purposes.
	XMPLCoin = sdk.NewCoins(sdk.NewCoin("xmpl", sdk.NewInt(1e18)))
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
			"-1xmpl: invalid coins",
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
			"spendable balance 1000000000000000000xmpl is smaller than 2000000000000000000xmpl: insufficient funds",
		},
		{
			"pass",
			func() []interface{} {
				return []interface{}{toAddr, big.NewInt(100)}
			},
			func() {
				toAddrBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), toAddr.Bytes(), "xmpl")
				s.Require().Equal(big.NewInt(100), toAddrBalance.Amount.BigInt(), "expected toAddr to have 100 XMPL")
			},
			false,
			"",
		},
	}

	//nolint: dupl
	for _, tc := range testcases {
		tc := tc
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
	method := s.precompile.Methods[erc20.TransferFromMethod]
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
				return []interface{}{fromAddr, toAddr, big.NewInt(-1)}
			},
			func() {},
			true,
			"-1xmpl: invalid coins",
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
				return []interface{}{fromAddr, "", big.NewInt(100)}
			},
			func() {},
			true,
			"invalid to address",
		},
		{
			"fail - invalid amount",
			func() []interface{} {
				return []interface{}{fromAddr, toAddr, ""}
			},
			func() {},
			true,
			"invalid amount",
		},
		{
			"fail - not enough balance",
			func() []interface{} {
				return []interface{}{fromAddr, toAddr, big.NewInt(2e18)}
			},
			func() {},
			true,
			"spendable balance 1000000000000000000xmpl is smaller than 2000000000000000000xmpl: insufficient funds",
		},
		{
			"pass",
			func() []interface{} {
				return []interface{}{fromAddr, toAddr, big.NewInt(100)}
			},
			func() {
				toAddrBalance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), toAddr.Bytes(), "xmpl")
				s.Require().Equal(big.NewInt(100), toAddrBalance.Amount.BigInt(), "expected toAddr to have 100 XMPL")
			},
			false,
			"",
		},
	}

	//nolint: dupl
	for _, tc := range testcases {
		tc := tc
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

			_, err = s.precompile.TransferFrom(ctx, contract, stateDB, &method, tc.malleate())
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
