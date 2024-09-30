package keeper_test

import (
	"math/big"
	"testing"

	"github.com/evmos/evmos/v20/x/evm/config"
	"github.com/evmos/evmos/v20/x/evm/keeper/testdata"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/ethereum/go-ethereum/common"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	utiltx "github.com/evmos/evmos/v20/testutil/tx"
	evmostypes "github.com/evmos/evmos/v20/types"
	"github.com/evmos/evmos/v20/x/evm/types"
)

func SetupContract(b *testing.B) (*KeeperTestSuite, common.Address) {
	suite := KeeperTestSuite{}
	suite.SetupTest()

	amt := sdk.Coins{evmostypes.NewBaseCoinInt64(1000000000000000000)}
	err := suite.network.App.BankKeeper.MintCoins(suite.network.GetContext(), types.ModuleName, amt)
	require.NoError(b, err)
	err = suite.network.App.BankKeeper.SendCoinsFromModuleToAccount(suite.network.GetContext(), types.ModuleName, suite.keyring.GetAddr(0).Bytes(), amt)
	require.NoError(b, err)

	contractAddr := suite.DeployTestContract(b, suite.network.GetContext(), suite.keyring.GetAddr(0), sdkmath.NewIntWithDecimal(1000, 18).BigInt())
	err = suite.network.NextBlock()
	require.NoError(b, err)

	return &suite, contractAddr
}

func SetupTestMessageCall(b *testing.B) (*KeeperTestSuite, common.Address) {
	suite := KeeperTestSuite{}
	suite.SetupTest()

	amt := sdk.Coins{evmostypes.NewBaseCoinInt64(1000000000000000000)}
	err := suite.network.App.BankKeeper.MintCoins(suite.network.GetContext(), types.ModuleName, amt)
	require.NoError(b, err)
	err = suite.network.App.BankKeeper.SendCoinsFromModuleToAccount(suite.network.GetContext(), types.ModuleName, suite.keyring.GetAddr(0).Bytes(), amt)
	require.NoError(b, err)

	contractAddr := suite.DeployTestMessageCall(b)
	err = suite.network.NextBlock()
	require.NoError(b, err)

	return &suite, contractAddr
}

type TxBuilder func(suite *KeeperTestSuite, contract common.Address) *types.MsgEthereumTx

func DoBenchmark(b *testing.B, txBuilder TxBuilder) {
	suite, contractAddr := SetupContract(b)

	krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
	msg := txBuilder(suite, contractAddr)
	msg.From = suite.keyring.GetAddr(0).Hex()
	err := msg.Sign(ethtypes.LatestSignerForChainID(config.GetChainConfig().ChainID), krSigner)
	require.NoError(b, err)

	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ctx, _ := suite.network.GetContext().CacheContext()

		// deduct fee first
		txData, err := types.UnpackTxData(msg.Data)
		require.NoError(b, err)

		fees := sdk.Coins{sdk.NewCoin(suite.EvmDenom(), sdkmath.NewIntFromBigInt(txData.Fee()))}
		err = authante.DeductFees(suite.network.App.BankKeeper, suite.network.GetContext(), suite.network.App.AccountKeeper.GetAccount(ctx, msg.GetFrom()), fees)
		require.NoError(b, err)

		rsp, err := suite.network.App.EvmKeeper.EthereumTx(ctx, msg)
		require.NoError(b, err)
		require.False(b, rsp.Failed())
	}
}

func BenchmarkTokenTransfer(b *testing.B) {
	erc20Contract, err := testdata.LoadERC20Contract()
	require.NoError(b, err, "failed to load erc20 contract")

	DoBenchmark(b, func(suite *KeeperTestSuite, contract common.Address) *types.MsgEthereumTx {
		input, err := erc20Contract.ABI.Pack("transfer", common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), big.NewInt(1000))
		require.NoError(b, err)
		nonce := suite.network.App.EvmKeeper.GetNonce(suite.network.GetContext(), suite.keyring.GetAddr(0))
		ethTxParams := &types.EvmTxArgs{
			ChainID:  config.GetChainConfig().ChainID,
			Nonce:    nonce,
			To:       &contract,
			Amount:   big.NewInt(0),
			GasLimit: 410000,
			GasPrice: big.NewInt(1),
			Input:    input,
		}
		return types.NewTx(ethTxParams)
	})
}

func BenchmarkEmitLogs(b *testing.B) {
	erc20Contract, err := testdata.LoadERC20Contract()
	require.NoError(b, err, "failed to load erc20 contract")

	DoBenchmark(b, func(suite *KeeperTestSuite, contract common.Address) *types.MsgEthereumTx {
		input, err := erc20Contract.ABI.Pack("benchmarkLogs", big.NewInt(1000))
		require.NoError(b, err)
		nonce := suite.network.App.EvmKeeper.GetNonce(suite.network.GetContext(), suite.keyring.GetAddr(0))
		ethTxParams := &types.EvmTxArgs{
			ChainID:  config.GetChainConfig().ChainID,
			Nonce:    nonce,
			To:       &contract,
			Amount:   big.NewInt(0),
			GasLimit: 4100000,
			GasPrice: big.NewInt(1),
			Input:    input,
		}
		return types.NewTx(ethTxParams)
	})
}

func BenchmarkTokenTransferFrom(b *testing.B) {
	erc20Contract, err := testdata.LoadERC20Contract()
	require.NoError(b, err)

	DoBenchmark(b, func(suite *KeeperTestSuite, contract common.Address) *types.MsgEthereumTx {
		input, err := erc20Contract.ABI.Pack("transferFrom", suite.keyring.GetAddr(0), common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), big.NewInt(0))
		require.NoError(b, err)
		nonce := suite.network.App.EvmKeeper.GetNonce(suite.network.GetContext(), suite.keyring.GetAddr(0))
		ethTxParams := &types.EvmTxArgs{
			ChainID:  config.GetChainConfig().ChainID,
			Nonce:    nonce,
			To:       &contract,
			Amount:   big.NewInt(0),
			GasLimit: 410000,
			GasPrice: big.NewInt(1),
			Input:    input,
		}
		return types.NewTx(ethTxParams)
	})
}

func BenchmarkTokenMint(b *testing.B) {
	erc20Contract, err := testdata.LoadERC20Contract()
	require.NoError(b, err, "failed to load erc20 contract")

	DoBenchmark(b, func(suite *KeeperTestSuite, contract common.Address) *types.MsgEthereumTx {
		input, err := erc20Contract.ABI.Pack("mint", common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), big.NewInt(1000))
		require.NoError(b, err)
		nonce := suite.network.App.EvmKeeper.GetNonce(suite.network.GetContext(), suite.keyring.GetAddr(0))
		ethTxParams := &types.EvmTxArgs{
			ChainID:  config.GetChainConfig().ChainID,
			Nonce:    nonce,
			To:       &contract,
			Amount:   big.NewInt(0),
			GasLimit: 410000,
			GasPrice: big.NewInt(1),
			Input:    input,
		}
		return types.NewTx(ethTxParams)
	})
}

func BenchmarkMessageCall(b *testing.B) {
	suite, contract := SetupTestMessageCall(b)

	messageCallContract, err := testdata.LoadMessageCallContract()
	require.NoError(b, err, "failed to load message call contract")

	input, err := messageCallContract.ABI.Pack("benchmarkMessageCall", big.NewInt(10000))
	require.NoError(b, err)
	nonce := suite.network.App.EvmKeeper.GetNonce(suite.network.GetContext(), suite.keyring.GetAddr(0))
	ethTxParams := &types.EvmTxArgs{
		ChainID:  config.GetChainConfig().ChainID,
		Nonce:    nonce,
		To:       &contract,
		Amount:   big.NewInt(0),
		GasLimit: 25000000,
		GasPrice: big.NewInt(1),
		Input:    input,
	}
	msg := types.NewTx(ethTxParams)

	msg.From = suite.keyring.GetAddr(0).Hex()
	krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
	err = msg.Sign(ethtypes.LatestSignerForChainID(config.GetChainConfig().ChainID), krSigner)
	require.NoError(b, err)

	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ctx, _ := suite.network.GetContext().CacheContext()

		// deduct fee first
		txData, err := types.UnpackTxData(msg.Data)
		require.NoError(b, err)

		fees := sdk.Coins{sdk.NewCoin(suite.EvmDenom(), sdkmath.NewIntFromBigInt(txData.Fee()))}
		err = authante.DeductFees(suite.network.App.BankKeeper, suite.network.GetContext(), suite.network.App.AccountKeeper.GetAccount(ctx, msg.GetFrom()), fees)
		require.NoError(b, err)

		rsp, err := suite.network.App.EvmKeeper.EthereumTx(ctx, msg)
		require.NoError(b, err)
		require.False(b, rsp.Failed())
	}
}
