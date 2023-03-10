package backend

import (
	"fmt"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	goethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"google.golang.org/grpc/metadata"

	"github.com/evmos/evmos/v12/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v12/rpc/backend/mocks"
	utiltx "github.com/evmos/evmos/v12/testutil/tx"
	evmtypes "github.com/evmos/evmos/v12/x/evm/types"
)

func (suite *BackendTestSuite) TestSendTransaction() {
	gasPrice := new(hexutil.Big)
	gas := hexutil.Uint64(1)
	zeroGas := hexutil.Uint64(0)
	toAddr := utiltx.GenerateAddress()
	priv, _ := ethsecp256k1.GenerateKey()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := hexutil.Uint64(1)
	baseFee := sdk.NewInt(1)
	callArgsDefault := evmtypes.TransactionArgs{
		From:     &from,
		To:       &toAddr,
		GasPrice: gasPrice,
		Gas:      &gas,
		Nonce:    &nonce,
	}

	hash := common.Hash{}

	testCases := []struct {
		name         string
		registerMock func()
		args         evmtypes.TransactionArgs
		expHash      common.Hash
		expPass      bool
	}{
		{
			"fail - Can't find account in Keyring",
			func() {},
			evmtypes.TransactionArgs{},
			hash,
			false,
		},
		{
			"fail - Block error can't set Tx defaults",
			func() {
				var header metadata.MD
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				armor := crypto.EncryptArmorPrivKey(priv, "", "eth_secp256k1")
				err := suite.backend.clientCtx.Keyring.ImportPrivKey("test_key", armor, "")
				suite.Require().NoError(err)
				RegisterParams(queryClient, &header, 1)
				RegisterBlockError(client, 1)
			},
			callArgsDefault,
			hash,
			false,
		},
		{
			"fail - Cannot validate transaction gas set to 0",
			func() {
				var header metadata.MD
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				armor := crypto.EncryptArmorPrivKey(priv, "", "eth_secp256k1")
				err := suite.backend.clientCtx.Keyring.ImportPrivKey("test_key", armor, "")
				suite.Require().NoError(err)
				RegisterParams(queryClient, &header, 1)
				_, err = RegisterBlock(client, 1, nil)
				suite.Require().NoError(err)
				_, err = RegisterBlockResults(client, 1)
				suite.Require().NoError(err)
				RegisterBaseFee(queryClient, baseFee)
				RegisterParamsWithoutHeader(queryClient, 1)
			},
			evmtypes.TransactionArgs{
				From:     &from,
				To:       &toAddr,
				GasPrice: gasPrice,
				Gas:      &zeroGas,
				Nonce:    &nonce,
			},
			hash,
			false,
		},
		{
			"fail - Cannot broadcast transaction",
			func() {
				client, txBytes := broadcastTx(suite, priv, baseFee, callArgsDefault)
				RegisterBroadcastTxError(client, txBytes)
			},
			callArgsDefault,
			common.Hash{},
			false,
		},
		{
			"pass - Return the transaction hash",
			func() {
				client, txBytes := broadcastTx(suite, priv, baseFee, callArgsDefault)
				RegisterBroadcastTx(client, txBytes)
			},
			callArgsDefault,
			hash,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock()

			if tc.expPass {
				// Sign the transaction and get the hash
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterParamsWithoutHeader(queryClient, 1)
				ethSigner := ethtypes.LatestSigner(suite.backend.ChainConfig())
				msg := callArgsDefault.ToTransaction()
				err := msg.Sign(ethSigner, suite.backend.clientCtx.Keyring)
				suite.Require().NoError(err)
				tc.expHash = msg.AsTransaction().Hash()
			}
			responseHash, err := suite.backend.SendTransaction(tc.args)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expHash, responseHash)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestSign() {
	from, priv := utiltx.NewAddrKey()
	testCases := []struct {
		name         string
		registerMock func()
		fromAddr     common.Address
		inputBz      hexutil.Bytes
		expPass      bool
	}{
		{
			"fail - can't find key in Keyring",
			func() {},
			from,
			nil,
			false,
		},
		{
			"pass - sign nil data",
			func() {
				armor := crypto.EncryptArmorPrivKey(priv, "", "eth_secp256k1")
				err := suite.backend.clientCtx.Keyring.ImportPrivKey("test_key", armor, "")
				suite.Require().NoError(err)
			},
			from,
			nil,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock()

			responseBz, err := suite.backend.Sign(tc.fromAddr, tc.inputBz)
			if tc.expPass {
				signature, _, err := suite.backend.clientCtx.Keyring.SignByAddress((sdk.AccAddress)(from.Bytes()), tc.inputBz)
				signature[goethcrypto.RecoveryIDOffset] += 27
				suite.Require().NoError(err)
				suite.Require().Equal((hexutil.Bytes)(signature), responseBz)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestSignTypedData() {
	from, priv := utiltx.NewAddrKey()
	testCases := []struct {
		name           string
		registerMock   func()
		fromAddr       common.Address
		inputTypedData apitypes.TypedData
		expPass        bool
	}{
		{
			"fail - can't find key in Keyring",
			func() {},
			from,
			apitypes.TypedData{},
			false,
		},
		{
			"fail - empty TypeData",
			func() {
				armor := crypto.EncryptArmorPrivKey(priv, "", "eth_secp256k1")
				err := suite.backend.clientCtx.Keyring.ImportPrivKey("test_key", armor, "")
				suite.Require().NoError(err)
			},
			from,
			apitypes.TypedData{},
			false,
		},
		// TODO: Generate a TypedData msg
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock()

			responseBz, err := suite.backend.SignTypedData(tc.fromAddr, tc.inputTypedData)

			if tc.expPass {
				sigHash, _, _ := apitypes.TypedDataAndHash(tc.inputTypedData)
				signature, _, err := suite.backend.clientCtx.Keyring.SignByAddress((sdk.AccAddress)(from.Bytes()), sigHash)
				signature[goethcrypto.RecoveryIDOffset] += 27
				suite.Require().NoError(err)
				suite.Require().Equal((hexutil.Bytes)(signature), responseBz)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func broadcastTx(suite *BackendTestSuite, priv *ethsecp256k1.PrivKey, baseFee math.Int, callArgsDefault evmtypes.TransactionArgs) (client *mocks.Client, txBytes []byte) {
	var header metadata.MD
	queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
	client = suite.backend.clientCtx.Client.(*mocks.Client)
	armor := crypto.EncryptArmorPrivKey(priv, "", "eth_secp256k1")
	_ = suite.backend.clientCtx.Keyring.ImportPrivKey("test_key", armor, "")
	RegisterParams(queryClient, &header, 1)
	_, err := RegisterBlock(client, 1, nil)
	suite.Require().NoError(err)
	_, err = RegisterBlockResults(client, 1)
	suite.Require().NoError(err)
	RegisterBaseFee(queryClient, baseFee)
	RegisterParamsWithoutHeader(queryClient, 1)
	ethSigner := ethtypes.LatestSigner(suite.backend.ChainConfig())
	msg := callArgsDefault.ToTransaction()
	err = msg.Sign(ethSigner, suite.backend.clientCtx.Keyring)
	suite.Require().NoError(err)
	tx, _ := msg.BuildTx(suite.backend.clientCtx.TxConfig.NewTxBuilder(), evmtypes.DefaultEVMDenom)
	txEncoder := suite.backend.clientCtx.TxConfig.TxEncoder()
	txBytes, _ = txEncoder(tx)
	return client, txBytes
}
