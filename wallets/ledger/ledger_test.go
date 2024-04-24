// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package ledger_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	gethaccounts "github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v18/app"
	"github.com/evmos/evmos/v18/encoding"
	"github.com/evmos/evmos/v18/ethereum/eip712"
	"github.com/evmos/evmos/v18/wallets/accounts"
	"github.com/evmos/evmos/v18/wallets/ledger"
)

// Test Mnemonic:
// glow spread dentist swamp people siren hint muscle first sausage castle metal cycle abandon accident logic again around mix dial knee organ episode usual

// Load encoding config for sign doc encoding/decoding
func init() {
	config := encoding.MakeConfig(app.ModuleBasics)
	eip712.SetEncodingConfig(config)
	sdk.GetConfig().SetBech32PrefixForAccount("cosmos", "")
}

func (suite *LedgerTestSuite) TestEvmosLedgerDerivation() {
	testCases := []struct {
		name     string
		mockFunc func()
		expPass  bool
	}{
		{
			"fail - no hardware wallets detected",
			func() {},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			derivationFunc := ledger.EvmosLedgerDerivation()
			_, err := derivationFunc()
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *LedgerTestSuite) TestClose() {
	testCases := []struct {
		name     string
		mockFunc func()
		expPass  bool
	}{
		{
			"fail - can't find Ledger device",
			func() {
				suite.ledger.PrimaryWallet = nil
			},
			false,
		},
		{
			"pass - wallet closed successfully",
			func() {
				RegisterClose(suite.mockWallet)
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.mockFunc()
			err := suite.ledger.Close()
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *LedgerTestSuite) TestSignatures() {
	privKey, err := crypto.GenerateKey()
	suite.Require().NoError(err)
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	account := accounts.Account{
		Address:   addr,
		PublicKey: &privKey.PublicKey,
	}

	testCases := []struct {
		name     string
		tx       []byte
		mockFunc func()
		expPass  bool
	}{
		{
			"fail - can't find Ledger device",
			suite.txAmino,
			func() {
				suite.ledger.PrimaryWallet = nil
			},
			false,
		},
		{
			"fail - unable to derive Ledger address",
			suite.txAmino,
			func() {
				RegisterOpen(suite.mockWallet)
				RegisterDeriveError(suite.mockWallet)
			},
			false,
		},
		{
			"fail - error generating signature",
			suite.txAmino,
			func() {
				RegisterOpen(suite.mockWallet)
				RegisterDerive(suite.mockWallet, addr, &privKey.PublicKey)
				RegisterSignTypedDataError(suite.mockWallet, account, suite.txAmino)
			},
			false,
		},
		{
			"pass - test ledger amino signature",
			suite.txAmino,
			func() {
				RegisterOpen(suite.mockWallet)
				RegisterDerive(suite.mockWallet, addr, &privKey.PublicKey)
				RegisterSignTypedData(suite.mockWallet, account, suite.txAmino)
			},
			true,
		},
		{
			"pass - test ledger protobuf signature",
			suite.txProtobuf,
			func() {
				RegisterOpen(suite.mockWallet)
				RegisterDerive(suite.mockWallet, addr, &privKey.PublicKey)
				RegisterSignTypedData(suite.mockWallet, account, suite.txProtobuf)
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.mockFunc()
			_, err := suite.ledger.SignSECP256K1(gethaccounts.DefaultBaseDerivationPath, tc.tx)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *LedgerTestSuite) TestSignatureEquivalence() {
	privKey, err := crypto.GenerateKey()
	suite.Require().NoError(err)
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	account := accounts.Account{
		Address:   addr,
		PublicKey: &privKey.PublicKey,
	}

	testCases := []struct {
		name       string
		txProtobuf []byte
		txAmino    []byte
		mockFunc   func()
		expPass    bool
	}{
		{
			"pass - signatures are equivalent",
			suite.txProtobuf,
			suite.txAmino,
			func() {
				RegisterOpen(suite.mockWallet)
				RegisterDerive(suite.mockWallet, addr, &privKey.PublicKey)
				RegisterSignTypedData(suite.mockWallet, account, suite.txProtobuf)
				RegisterSignTypedData(suite.mockWallet, account, suite.txAmino)
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.mockFunc()
			protoSignature, err := suite.ledger.SignSECP256K1(gethaccounts.DefaultBaseDerivationPath, tc.txProtobuf)
			suite.Require().NoError(err)
			aminoSignature, err := suite.ledger.SignSECP256K1(gethaccounts.DefaultBaseDerivationPath, tc.txAmino)
			suite.Require().NoError(err)
			if tc.expPass {
				suite.Require().Equal(protoSignature, aminoSignature)
			} else {
				suite.Require().NotEqual(protoSignature, aminoSignature)
			}
		})
	}
}

func (suite *LedgerTestSuite) TestGetAddressPubKeySECP256K1() {
	privKey, err := crypto.GenerateKey()
	suite.Require().NoError(err)

	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	expAddr, err := sdk.Bech32ifyAddressBytes("evmos", common.HexToAddress(addr.String()).Bytes())
	suite.Require().NoError(err)

	testCases := []struct {
		name     string
		expPass  bool
		mockFunc func()
	}{
		{
			"fail - can't find Ledger device",
			false,
			func() {
				suite.ledger.PrimaryWallet = nil
			},
		},
		{
			"fail - unable to derive Ledger address",
			false,
			func() {
				RegisterOpen(suite.mockWallet)
				RegisterDeriveError(suite.mockWallet)
			},
		},
		{
			"fail - bech32 prefix empty",
			false,
			func() {
				suite.hrp = ""
				RegisterOpen(suite.mockWallet)
				RegisterDerive(suite.mockWallet, addr, &privKey.PublicKey)
			},
		},
		{
			"pass - get ledger address",
			true,
			func() {
				RegisterOpen(suite.mockWallet)
				RegisterDerive(suite.mockWallet, addr, &privKey.PublicKey)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.mockFunc()
			_, addr, err := suite.ledger.GetAddressPubKeySECP256K1(gethaccounts.DefaultBaseDerivationPath, suite.hrp)
			if tc.expPass {
				suite.Require().NoError(err, "Could not get wallet address")
				suite.Require().Equal(expAddr, addr)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *LedgerTestSuite) TestGetPublicKeySECP256K1() {
	privKey, err := crypto.GenerateKey()
	suite.Require().NoError(err)
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	expPubkeyBz := crypto.FromECDSAPub(&privKey.PublicKey)
	testCases := []struct {
		name     string
		expPass  bool
		mockFunc func()
	}{
		{
			"fail - can't find Ledger device",
			false,
			func() {
				suite.ledger.PrimaryWallet = nil
			},
		},
		{
			"fail - unable to derive Ledger address",
			false,
			func() {
				RegisterOpen(suite.mockWallet)
				RegisterDeriveError(suite.mockWallet)
			},
		},
		{
			"pass - get ledger public key",
			true,
			func() {
				RegisterOpen(suite.mockWallet)
				RegisterDerive(suite.mockWallet, addr, &privKey.PublicKey)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.mockFunc()
			pubKeyBz, err := suite.ledger.GetPublicKeySECP256K1(gethaccounts.DefaultBaseDerivationPath)
			if tc.expPass {
				suite.Require().NoError(err, "Could not get wallet address")
				suite.Require().Equal(expPubkeyBz, pubKeyBz)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
