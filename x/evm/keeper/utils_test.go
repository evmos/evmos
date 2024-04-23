package keeper_test

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v18/server/config"
	"github.com/evmos/evmos/v18/testutil"
	"github.com/evmos/evmos/v18/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
	"github.com/stretchr/testify/require"
)

func (suite *KeeperTestSuite) EvmDenom() string {
	ctx := sdk.WrapSDKContext(suite.ctx)
	rsp, _ := suite.queryClient.Params(ctx, &evmtypes.QueryParamsRequest{})
	return rsp.Params.EvmDenom
}

// Commit and begin new block
func (suite *KeeperTestSuite) Commit() {
	var err error
	suite.ctx, err = testutil.CommitAndCreateNewCtx(suite.ctx, suite.app, 0*time.Second, nil)
	suite.Require().NoError(err)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evmtypes.RegisterQueryServer(queryHelper, suite.app.EvmKeeper)
	suite.queryClient = evmtypes.NewQueryClient(queryHelper)
}

func (suite *KeeperTestSuite) StateDB() *statedb.StateDB {
	return statedb.New(suite.ctx, suite.app.EvmKeeper, statedb.NewEmptyTxConfig(common.BytesToHash(suite.ctx.HeaderHash().Bytes())))
}

// DeployTestContract deploy a test erc20 contract and returns the contract address
func (suite *KeeperTestSuite) DeployTestContract(t require.TestingT, owner common.Address, supply *big.Int) common.Address {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	ctorArgs, err := evmtypes.ERC20Contract.ABI.Pack("", owner, supply)
	require.NoError(t, err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	data := evmtypes.ERC20Contract.Bin
	data = append(data, ctorArgs...)
	args, err := json.Marshal(&evmtypes.TransactionArgs{
		From: &suite.address,
		Data: (*hexutil.Bytes)(&data),
	})
	require.NoError(t, err)
	res, err := suite.queryClient.EstimateGas(ctx, &evmtypes.EthCallRequest{
		Args:            args,
		GasCap:          config.DefaultGasCap,
		ProposerAddress: suite.ctx.BlockHeader().ProposerAddress,
	})
	require.NoError(t, err)

	var erc20DeployTx *evmtypes.MsgEthereumTx
	if suite.enableFeemarket {
		ethTxParams := &evmtypes.EvmTxArgs{
			ChainID:   chainID,
			Nonce:     nonce,
			GasLimit:  res.Gas,
			GasFeeCap: suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
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

	erc20DeployTx.From = suite.address.Hex()
	err = erc20DeployTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	require.NoError(t, err)
	rsp, err := suite.app.EvmKeeper.EthereumTx(ctx, erc20DeployTx)
	require.NoError(t, err)
	require.Empty(t, rsp.VmError)
	return crypto.CreateAddress(suite.address, nonce)
}

func (suite *KeeperTestSuite) TransferERC20Token(t require.TestingT, contractAddr, from, to common.Address, amount *big.Int) *evmtypes.MsgEthereumTx {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	transferData, err := evmtypes.ERC20Contract.ABI.Pack("transfer", to, amount)
	require.NoError(t, err)
	args, err := json.Marshal(&evmtypes.TransactionArgs{To: &contractAddr, From: &from, Data: (*hexutil.Bytes)(&transferData)})
	require.NoError(t, err)
	res, err := suite.queryClient.EstimateGas(ctx, &evmtypes.EthCallRequest{
		Args:            args,
		GasCap:          25_000_000,
		ProposerAddress: suite.ctx.BlockHeader().ProposerAddress,
	})
	require.NoError(t, err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	var ercTransferTx *evmtypes.MsgEthereumTx
	if suite.enableFeemarket {
		ethTxParams := &evmtypes.EvmTxArgs{
			ChainID:   chainID,
			Nonce:     nonce,
			To:        &contractAddr,
			GasLimit:  res.Gas,
			GasFeeCap: suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
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

	ercTransferTx.From = suite.address.Hex()
	err = ercTransferTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	require.NoError(t, err)
	rsp, err := suite.app.EvmKeeper.EthereumTx(ctx, ercTransferTx)
	require.NoError(t, err)
	require.Empty(t, rsp.VmError)
	return ercTransferTx
}

// DeployTestMessageCall deploy a test erc20 contract and returns the contract address
func (suite *KeeperTestSuite) DeployTestMessageCall(t require.TestingT) common.Address {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	data := evmtypes.TestMessageCall.Bin
	args, err := json.Marshal(&evmtypes.TransactionArgs{
		From: &suite.address,
		Data: (*hexutil.Bytes)(&data),
	})
	require.NoError(t, err)

	res, err := suite.queryClient.EstimateGas(ctx, &evmtypes.EthCallRequest{
		Args:            args,
		GasCap:          config.DefaultGasCap,
		ProposerAddress: suite.ctx.BlockHeader().ProposerAddress,
	})
	require.NoError(t, err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	var erc20DeployTx *evmtypes.MsgEthereumTx
	if suite.enableFeemarket {
		ethTxParams := &evmtypes.EvmTxArgs{
			ChainID:   chainID,
			Nonce:     nonce,
			GasLimit:  res.Gas,
			Input:     data,
			GasFeeCap: suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
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

	erc20DeployTx.From = suite.address.Hex()
	err = erc20DeployTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	require.NoError(t, err)
	rsp, err := suite.app.EvmKeeper.EthereumTx(ctx, erc20DeployTx)
	require.NoError(t, err)
	require.Empty(t, rsp.VmError)
	return crypto.CreateAddress(suite.address, nonce)
}
