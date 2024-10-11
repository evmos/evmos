package keeper_test

import (
	"encoding/json"
	"math/big"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	servercfg "github.com/evmos/evmos/v20/server/config"
	utiltx "github.com/evmos/evmos/v20/testutil/tx"
	"github.com/evmos/evmos/v20/x/evm/keeper/testdata"
	"github.com/evmos/evmos/v20/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

func (suite *KeeperTestSuite) EvmDenom() string {
	return evmtypes.GetEVMCoinDenom()
}

func (suite *KeeperTestSuite) StateDB() *statedb.StateDB {
	return statedb.New(suite.network.GetContext(), suite.network.App.EvmKeeper, statedb.NewEmptyTxConfig(common.BytesToHash(suite.network.GetContext().HeaderHash())))
}

// DeployTestContract deploy a test erc20 contract and returns the contract address
func (suite *KeeperTestSuite) DeployTestContract(t require.TestingT, ctx sdk.Context, owner common.Address, supply *big.Int) common.Address {
	chainID := evmtypes.GetEthChainConfig().ChainID

	erc20Contract, err := testdata.LoadERC20Contract()
	require.NoError(t, err, "failed to load contract")

	ctorArgs, err := erc20Contract.ABI.Pack("", owner, supply)
	require.NoError(t, err)

	addr := suite.keyring.GetAddr(0)
	nonce := suite.network.App.EvmKeeper.GetNonce(suite.network.GetContext(), addr)

	data := erc20Contract.Bin
	data = append(data, ctorArgs...)
	args, err := json.Marshal(&evmtypes.TransactionArgs{
		From: &addr,
		Data: (*hexutil.Bytes)(&data),
	})
	require.NoError(t, err)
	res, err := suite.network.GetEvmClient().EstimateGas(ctx, &evmtypes.EthCallRequest{
		Args:            args,
		GasCap:          servercfg.DefaultGasCap,
		ProposerAddress: suite.network.GetContext().BlockHeader().ProposerAddress,
	})
	require.NoError(t, err)

	baseFeeRes, err := suite.network.GetEvmClient().BaseFee(ctx, &evmtypes.QueryBaseFeeRequest{})
	require.NoError(t, err)

	var erc20DeployTx *evmtypes.MsgEthereumTx
	if suite.enableFeemarket {
		ethTxParams := &evmtypes.EvmTxArgs{
			ChainID:   chainID,
			Nonce:     nonce,
			GasLimit:  res.Gas,
			GasFeeCap: baseFeeRes.BaseFee.BigInt(),
			GasTipCap: big.NewInt(1),
			Input:     data,
			Accesses:  &ethtypes.AccessList{},
		}
		erc20DeployTx = evmtypes.NewTx(ethTxParams)
	} else {
		ethTxParams := &evmtypes.EvmTxArgs{
			ChainID:  chainID,
			Nonce:    nonce,
			GasLimit: res.Gas,
			Input:    data,
		}
		erc20DeployTx = evmtypes.NewTx(ethTxParams)
	}

	krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
	erc20DeployTx.From = addr.Hex()
	err = erc20DeployTx.Sign(ethtypes.LatestSignerForChainID(chainID), krSigner)
	require.NoError(t, err)
	rsp, err := suite.network.App.EvmKeeper.EthereumTx(ctx, erc20DeployTx)
	require.NoError(t, err)
	require.Empty(t, rsp.VmError)
	return crypto.CreateAddress(addr, nonce)
}

func (suite *KeeperTestSuite) TransferERC20Token(t require.TestingT, contractAddr, from, to common.Address, amount *big.Int) *evmtypes.MsgEthereumTx {
	ctx := suite.network.GetContext()
	chainID := evmtypes.GetEthChainConfig().ChainID

	erc20Contract, err := testdata.LoadERC20Contract()
	require.NoError(t, err, "failed to load contract")

	transferData, err := erc20Contract.ABI.Pack("transfer", to, amount)
	require.NoError(t, err)
	args, err := json.Marshal(&evmtypes.TransactionArgs{To: &contractAddr, From: &from, Data: (*hexutil.Bytes)(&transferData)})
	require.NoError(t, err)
	res, err := suite.network.GetEvmClient().EstimateGas(ctx, &evmtypes.EthCallRequest{
		Args:            args,
		GasCap:          25_000_000,
		ProposerAddress: suite.network.GetContext().BlockHeader().ProposerAddress,
	})
	require.NoError(t, err)

	nonce := suite.network.App.EvmKeeper.GetNonce(suite.network.GetContext(), suite.keyring.GetAddr(0))
	baseFeeRes, err := suite.network.GetEvmClient().BaseFee(ctx, &evmtypes.QueryBaseFeeRequest{})
	require.NoError(t, err)

	var ercTransferTx *evmtypes.MsgEthereumTx
	if suite.enableFeemarket {
		ethTxParams := &evmtypes.EvmTxArgs{
			ChainID:   chainID,
			Nonce:     nonce,
			To:        &contractAddr,
			GasLimit:  res.Gas,
			GasFeeCap: baseFeeRes.BaseFee.BigInt(),
			GasTipCap: big.NewInt(1),
			Input:     transferData,
			Accesses:  &ethtypes.AccessList{},
		}
		ercTransferTx = evmtypes.NewTx(ethTxParams)
	} else {
		ethTxParams := &evmtypes.EvmTxArgs{
			ChainID:  chainID,
			Nonce:    nonce,
			To:       &contractAddr,
			GasLimit: res.Gas,
			Input:    transferData,
		}
		ercTransferTx = evmtypes.NewTx(ethTxParams)
	}

	addr := suite.keyring.GetAddr(0)
	krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
	ercTransferTx.From = addr.Hex()
	err = ercTransferTx.Sign(ethtypes.LatestSignerForChainID(chainID), krSigner)
	require.NoError(t, err)
	rsp, err := suite.network.App.EvmKeeper.EthereumTx(ctx, ercTransferTx)
	require.NoError(t, err)
	require.Empty(t, rsp.VmError)
	return ercTransferTx
}

// DeployTestMessageCall deploy a test erc20 contract and returns the contract address
func (suite *KeeperTestSuite) DeployTestMessageCall(t require.TestingT) common.Address {
	ctx := suite.network.GetContext()
	chainID := evmtypes.GetEthChainConfig().ChainID

	testMsgCall, err := testdata.LoadMessageCallContract()
	require.NoError(t, err)

	data := testMsgCall.Bin
	addr := suite.keyring.GetAddr(0)
	args, err := json.Marshal(&evmtypes.TransactionArgs{
		From: &addr,
		Data: (*hexutil.Bytes)(&data),
	})
	require.NoError(t, err)

	res, err := suite.network.GetEvmClient().EstimateGas(ctx, &evmtypes.EthCallRequest{
		Args:            args,
		GasCap:          servercfg.DefaultGasCap,
		ProposerAddress: suite.network.GetContext().BlockHeader().ProposerAddress,
	})
	require.NoError(t, err)

	nonce := suite.network.App.EvmKeeper.GetNonce(suite.network.GetContext(), addr)
	baseFeeRes, err := suite.network.GetEvmClient().BaseFee(ctx, &evmtypes.QueryBaseFeeRequest{})
	require.NoError(t, err)

	var erc20DeployTx *evmtypes.MsgEthereumTx
	if suite.enableFeemarket {
		ethTxParams := &evmtypes.EvmTxArgs{
			ChainID:   chainID,
			Nonce:     nonce,
			GasLimit:  res.Gas,
			Input:     data,
			GasFeeCap: baseFeeRes.BaseFee.BigInt(),
			Accesses:  &ethtypes.AccessList{},
			GasTipCap: big.NewInt(1),
		}
		erc20DeployTx = evmtypes.NewTx(ethTxParams)
	} else {
		ethTxParams := &evmtypes.EvmTxArgs{
			ChainID:  chainID,
			Nonce:    nonce,
			GasLimit: res.Gas,
			Input:    data,
		}
		erc20DeployTx = evmtypes.NewTx(ethTxParams)
	}

	krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
	erc20DeployTx.From = addr.Hex()
	err = erc20DeployTx.Sign(ethtypes.LatestSignerForChainID(chainID), krSigner)
	require.NoError(t, err)
	rsp, err := suite.network.App.EvmKeeper.EthereumTx(ctx, erc20DeployTx)
	require.NoError(t, err)
	require.Empty(t, rsp.VmError)
	return crypto.CreateAddress(addr, nonce)
}
