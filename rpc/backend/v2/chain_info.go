// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package v2

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	rpctypes "github.com/evmos/evmos/v20/rpc/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v20/x/feemarket/types"
)

// ChainID is the EIP-155 replay-protection chain id for the current ethereum chain config.
func (b *Backend) ChainID() *big.Int {
	return b.chainID
}

// ChainConfig returns a default ethereum chain configuration
func (b *Backend) ChainConfig() *params.ChainConfig {
	return evmtypes.DefaultChainConfig().EthereumConfig(b.chainID)
}

// GlobalMinGasPrice returns MinGasPrice param from FeeMarket
func (b *Backend) GlobalMinGasPrice() (math.LegacyDec, error) {
	res, err := b.queryClient.FeeMarket.Params(b.ctx, &feemarkettypes.QueryParamsRequest{})
	if err != nil {
		return math.LegacyZeroDec(), err
	}
	return res.Params.MinGasPrice, nil
}

// BaseFee returns the base fee incorporated the latest block
// vote extension
func (b *Backend) ProcessEVMParams(injectedTx sdk.Tx) (*big.Int, error) {
	baseFee := common.Big0
	return baseFee, nil
}

// CurrentHeader returns the latest block header
// This will return error as per node configuration
// if the ABCI responses are discarded ('discard_abci_responses' config param)
func (b *Backend) CurrentHeader() (*ethtypes.Header, error) {
	return b.HeaderByNumber(rpctypes.EthLatestBlockNumber)
}

// PendingTransactions returns the transactions that are in the transaction pool
// and have a from address that is one of the accounts this node manages.
func (b *Backend) PendingTransactions() ([]*ethtypes.Transaction, error) {
	res, err := b.rpcClient.UnconfirmedTxs(b.ctx, nil)
	if err != nil {
		return nil, err
	}

	pending := make([]*ethtypes.Transaction, 0, len(res.Txs))
	for _, txBz := range res.Txs {
		var tx ethtypes.Transaction
		err := tx.UnmarshalBinary(txBz)
		if err != nil {
			return nil, err
		}

		pending = append(pending, &tx)
	}

	return pending, nil
}

// GetCoinbase is the address that staking rewards will be send to (alias for Etherbase).
func (b *Backend) GetCoinbase() (common.Address, error) {
	status, err := b.rpcClient.Status(b.ctx)
	if err != nil {
		return common.Address{}, err
	}

	req := &evmtypes.QueryValidatorAccountRequest{
		ConsAddress: sdk.ConsAddress(status.ValidatorInfo.Address).String(),
	}

	res, err := b.queryClient.ValidatorAccount(b.ctx, req)
	if err != nil {
		return common.Address{}, err
	}

	address, err := sdk.AccAddressFromBech32(res.AccountAddress)
	if err != nil {
		return common.Address{}, err
	}

	return common.BytesToAddress(address), nil
}

// FeeHistory returns data relevant for fee estimation based on the specified range of blocks.
func (b *Backend) FeeHistory(
	userBlockCount rpc.DecimalOrHex, // number blocks to fetch, maximum is 100
	lastBlock rpc.BlockNumber, // the block to start search , to oldest
	rewardPercentiles []float64, // percentiles to fetch reward
) (*rpctypes.FeeHistoryResult, error) {
	blockEnd := int64(lastBlock) //#nosec G115 G701 -- checked for int overflow already

	if blockEnd < 0 {
		blockNumber, err := b.BlockNumber()
		if err != nil {
			return nil, err
		}
		blockEnd = int64(blockNumber) //#nosec G115 G701 -- checked for int overflow already
	}

	blocks := int64(userBlockCount)                     // #nosec G115 G701 -- checked for int overflow already
	maxBlockCount := int64(b.cfg.JSONRPC.FeeHistoryCap) // #nosec G115 G701 -- checked for int overflow already
	if blocks > maxBlockCount {
		return nil, fmt.Errorf("FeeHistory user block count %d higher than %d", blocks, maxBlockCount)
	}

	if blockEnd+1 < blocks {
		blocks = blockEnd + 1
	}
	// Ensure not trying to retrieve before genesis.
	blockStart := blockEnd + 1 - blocks
	oldestBlock := (*hexutil.Big)(big.NewInt(blockStart))

	// prepare space
	reward := make([][]*hexutil.Big, blocks)
	rewardCount := len(rewardPercentiles)
	for i := 0; i < int(blocks); i++ {
		reward[i] = make([]*hexutil.Big, rewardCount)
	}

	thisBaseFee := make([]*hexutil.Big, blocks+1)
	thisGasUsedRatio := make([]float64, blocks)

	// rewards should only be calculated if reward percentiles were included
	calculateRewards := rewardCount != 0

	// fetch block
	for blockID := blockStart; blockID <= blockEnd; blockID++ {
		index := int32(blockID - blockStart) // #nosec G701 G115
		// tendermint block
		tendermintblock, err := b.TendermintBlockByNumber(rpctypes.BlockNumber(blockID))
		if tendermintblock == nil {
			return nil, err
		}

		// eth block
		ethBlock, err := b.GetBlockByNumber(rpctypes.BlockNumber(blockID), true)
		if ethBlock == nil {
			return nil, err
		}

		// tendermint block result
		tendermintBlockResult, err := b.TendermintBlockResultByNumber(&tendermintblock.Block.Height)
		if tendermintBlockResult == nil {
			b.logger.Debug("block result not found", "height", tendermintblock.Block.Height, "error", err.Error())
			return nil, err
		}

		oneFeeHistory := rpctypes.OneFeeHistory{}
		err = b.processBlock(tendermintblock, &ethBlock, rewardPercentiles, tendermintBlockResult, &oneFeeHistory)
		if err != nil {
			return nil, err
		}

		// copy
		thisBaseFee[index] = (*hexutil.Big)(oneFeeHistory.BaseFee)
		thisBaseFee[index+1] = (*hexutil.Big)(oneFeeHistory.NextBaseFee)
		thisGasUsedRatio[index] = oneFeeHistory.GasUsedRatio
		if calculateRewards {
			for j := 0; j < rewardCount; j++ {
				reward[index][j] = (*hexutil.Big)(oneFeeHistory.Reward[j])
				if reward[index][j] == nil {
					reward[index][j] = (*hexutil.Big)(big.NewInt(0))
				}
			}
		}
	}

	feeHistory := rpctypes.FeeHistoryResult{
		OldestBlock:  oldestBlock,
		BaseFee:      thisBaseFee,
		GasUsedRatio: thisGasUsedRatio,
	}

	if calculateRewards {
		feeHistory.Reward = reward
	}

	return &feeHistory, nil
}

// SuggestGasTipCap returns the suggested tip cap
// Although we don't support tx prioritization yet, but we return a positive value to help client to
// mitigate the base fee changes.
func (b *Backend) SuggestGasTipCap(baseFee *big.Int) (*big.Int, error) {
	if baseFee == nil {
		// london hardfork not enabled or feemarket not enabled
		return big.NewInt(0), nil
	}

	params, err := b.queryClient.FeeMarket.Params(b.ctx, &feemarkettypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}
	// calculate the maximum base fee delta in current block, assuming all block gas limit is consumed
	// ```
	// GasTarget = GasLimit / ElasticityMultiplier
	// Delta = BaseFee * (GasUsed - GasTarget) / GasTarget / Denominator
	// ```
	// The delta is at maximum when `GasUsed` is equal to `GasLimit`, which is:
	// ```
	// MaxDelta = BaseFee * (GasLimit - GasLimit / ElasticityMultiplier) / (GasLimit / ElasticityMultiplier) / Denominator
	//          = BaseFee * (ElasticityMultiplier - 1) / Denominator
	// ```t
	maxDelta := baseFee.Int64() * (int64(params.Params.ElasticityMultiplier) - 1) / int64(params.Params.BaseFeeChangeDenominator) // #nosec G701
	if maxDelta < 0 {
		// impossible if the parameter validation passed.
		maxDelta = 0
	}
	return big.NewInt(maxDelta), nil
}
