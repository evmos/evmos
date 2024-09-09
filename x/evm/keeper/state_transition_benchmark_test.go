package keeper_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	utiltx "github.com/evmos/evmos/v20/testutil/tx"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	"github.com/stretchr/testify/require"
)

var templateAccessListTx = &ethtypes.AccessListTx{
	GasPrice: big.NewInt(1),
	Gas:      21000,
	To:       &common.Address{},
	Value:    big.NewInt(0),
	Data:     []byte{},
}

var templateLegacyTx = &ethtypes.LegacyTx{
	GasPrice: big.NewInt(1),
	Gas:      21000,
	To:       &common.Address{},
	Value:    big.NewInt(0),
	Data:     []byte{},
}

var templateDynamicFeeTx = &ethtypes.DynamicFeeTx{
	GasFeeCap: big.NewInt(10),
	GasTipCap: big.NewInt(2),
	Gas:       21000,
	To:        &common.Address{},
	Value:     big.NewInt(0),
	Data:      []byte{},
}

func newSignedEthTx(
	txData ethtypes.TxData,
	nonce uint64,
	addr sdk.Address,
	krSigner keyring.Signer,
	ethSigner ethtypes.Signer,
) (*ethtypes.Transaction, error) {
	var ethTx *ethtypes.Transaction
	switch txData := txData.(type) {
	case *ethtypes.AccessListTx:
		txData.Nonce = nonce
		ethTx = ethtypes.NewTx(txData)
	case *ethtypes.LegacyTx:
		txData.Nonce = nonce
		ethTx = ethtypes.NewTx(txData)
	case *ethtypes.DynamicFeeTx:
		txData.Nonce = nonce
		ethTx = ethtypes.NewTx(txData)
	default:
		return nil, errors.New("unknown transaction type")
	}

	sig, _, err := krSigner.SignByAddress(addr, ethTx.Hash().Bytes(), signingtypes.SignMode_SIGN_MODE_TEXTUAL)
	if err != nil {
		return nil, err
	}

	ethTx, err = ethTx.WithSignature(ethSigner, sig)
	if err != nil {
		return nil, err
	}

	return ethTx, nil
}

func newEthMsgTx(
	nonce uint64,
	address common.Address,
	krSigner keyring.Signer,
	ethSigner ethtypes.Signer,
	txType byte,
	data []byte,
	accessList ethtypes.AccessList,
) (*evmtypes.MsgEthereumTx, *big.Int, error) {
	var (
		ethTx   *ethtypes.Transaction
		baseFee *big.Int
	)
	switch txType {
	case ethtypes.LegacyTxType:
		templateLegacyTx.Nonce = nonce
		if data != nil {
			templateLegacyTx.Data = data
		}
		ethTx = ethtypes.NewTx(templateLegacyTx)
	case ethtypes.AccessListTxType:
		templateAccessListTx.Nonce = nonce
		if data != nil {
			templateAccessListTx.Data = data
		} else {
			templateAccessListTx.Data = []byte{}
		}

		templateAccessListTx.AccessList = accessList
		ethTx = ethtypes.NewTx(templateAccessListTx)
	case ethtypes.DynamicFeeTxType:
		templateDynamicFeeTx.Nonce = nonce

		if data != nil {
			templateAccessListTx.Data = data
		} else {
			templateAccessListTx.Data = []byte{}
		}
		templateAccessListTx.AccessList = accessList
		ethTx = ethtypes.NewTx(templateDynamicFeeTx)
		baseFee = big.NewInt(3)
	default:
		return nil, baseFee, errors.New("unsupported tx type")
	}

	msg := &evmtypes.MsgEthereumTx{}
	err := msg.FromEthereumTx(ethTx)
	if err != nil {
		return nil, nil, err
	}

	msg.From = address.Hex()

	return msg, baseFee, msg.Sign(ethSigner, krSigner)
}

func newNativeMessage(
	nonce uint64,
	blockHeight int64,
	address common.Address,
	cfg *params.ChainConfig,
	krSigner keyring.Signer,
	ethSigner ethtypes.Signer,
	txType byte,
	data []byte,
	accessList ethtypes.AccessList,
) (core.Message, error) {
	msgSigner := ethtypes.MakeSigner(cfg, big.NewInt(blockHeight))

	msg, baseFee, err := newEthMsgTx(nonce, address, krSigner, ethSigner, txType, data, accessList)
	if err != nil {
		return nil, err
	}

	m, err := msg.AsMessage(msgSigner, baseFee)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func BenchmarkApplyTransaction(b *testing.B) { //nolint:dupl
	suite := KeeperTestSuite{enableLondonHF: true}
	suite.SetupTest()

	ethSigner := ethtypes.LatestSignerForChainID(suite.network.App.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		addr := suite.keyring.GetAddr(0)
		krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
		tx, err := newSignedEthTx(templateAccessListTx,
			suite.network.App.EvmKeeper.GetNonce(suite.network.GetContext(), addr),
			sdk.AccAddress(addr.Bytes()),
			krSigner,
			ethSigner,
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, err := suite.network.App.EvmKeeper.ApplyTransaction(suite.network.GetContext(), tx)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

func BenchmarkApplyTransactionWithLegacyTx(b *testing.B) { //nolint:dupl
	suite := KeeperTestSuite{enableLondonHF: true}
	suite.SetupTest()

	ethSigner := ethtypes.LatestSignerForChainID(suite.network.App.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		addr := suite.keyring.GetAddr(0)
		krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
		tx, err := newSignedEthTx(templateLegacyTx,
			suite.network.App.EvmKeeper.GetNonce(suite.network.GetContext(), addr),
			sdk.AccAddress(addr.Bytes()),
			krSigner,
			ethSigner,
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, err := suite.network.App.EvmKeeper.ApplyTransaction(suite.network.GetContext(), tx)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

func BenchmarkApplyTransactionWithDynamicFeeTx(b *testing.B) {
	suite := KeeperTestSuite{enableFeemarket: true, enableLondonHF: true}
	suite.SetupTest()

	ethSigner := ethtypes.LatestSignerForChainID(suite.network.App.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		addr := suite.keyring.GetAddr(0)
		krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
		tx, err := newSignedEthTx(templateDynamicFeeTx,
			suite.network.App.EvmKeeper.GetNonce(suite.network.GetContext(), addr),
			sdk.AccAddress(addr.Bytes()),
			krSigner,
			ethSigner,
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, err := suite.network.App.EvmKeeper.ApplyTransaction(suite.network.GetContext(), tx)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

//nolint:all
func BenchmarkApplyMessage(b *testing.B) {
	suite := KeeperTestSuite{enableLondonHF: true}
	suite.SetupTest()

	params := suite.network.App.EvmKeeper.GetParams(suite.network.GetContext())
	ethCfg := params.ChainConfig.EthereumConfig(suite.network.App.EvmKeeper.ChainID())
	signer := ethtypes.LatestSignerForChainID(suite.network.App.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		addr := suite.keyring.GetAddr(0)
		krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
		m, err := newNativeMessage(
			suite.network.App.EvmKeeper.GetNonce(suite.network.GetContext(), addr),
			suite.network.GetContext().BlockHeight(),
			addr,
			ethCfg,
			krSigner,
			signer,
			ethtypes.AccessListTxType,
			nil,
			nil,
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, err := suite.network.App.EvmKeeper.ApplyMessage(suite.network.GetContext(), m, nil, true)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

//nolint:all
func BenchmarkApplyMessageWithLegacyTx(b *testing.B) {
	suite := KeeperTestSuite{enableLondonHF: true}
	suite.SetupTest()

	params := suite.network.App.EvmKeeper.GetParams(suite.network.GetContext())
	ethCfg := params.ChainConfig.EthereumConfig(suite.network.App.EvmKeeper.ChainID())
	signer := ethtypes.LatestSignerForChainID(suite.network.App.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		addr := suite.keyring.GetAddr(0)
		krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
		m, err := newNativeMessage(
			suite.network.App.EvmKeeper.GetNonce(suite.network.GetContext(), addr),
			suite.network.GetContext().BlockHeight(),
			addr,
			ethCfg,
			krSigner,
			signer,
			ethtypes.LegacyTxType,
			nil,
			nil,
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, err := suite.network.App.EvmKeeper.ApplyMessage(suite.network.GetContext(), m, nil, true)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

func BenchmarkApplyMessageWithDynamicFeeTx(b *testing.B) {
	suite := KeeperTestSuite{enableFeemarket: true, enableLondonHF: true}
	suite.SetupTest()

	params := suite.network.App.EvmKeeper.GetParams(suite.network.GetContext())
	ethCfg := params.ChainConfig.EthereumConfig(suite.network.App.EvmKeeper.ChainID())
	signer := ethtypes.LatestSignerForChainID(suite.network.App.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		addr := suite.keyring.GetAddr(0)
		krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
		m, err := newNativeMessage(
			suite.network.App.EvmKeeper.GetNonce(suite.network.GetContext(), addr),
			suite.network.GetContext().BlockHeight(),
			addr,
			ethCfg,
			krSigner,
			signer,
			ethtypes.DynamicFeeTxType,
			nil,
			nil,
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, err := suite.network.App.EvmKeeper.ApplyMessage(suite.network.GetContext(), m, nil, true)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}
