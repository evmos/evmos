// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package keeper

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v15/x/evm/statedb"
	"github.com/evmos/evmos/v15/x/evm/types"
)

// EVMConfig creates the EVMConfig based on current state
func (k *Keeper) EVMConfig(ctx sdk.Context, proposerAddress sdk.ConsAddress, chainID *big.Int) (*statedb.EVMConfig, error) {
	params := k.GetParams(ctx)
	ethCfg := params.ChainConfig.EthereumConfig(chainID)

	// get the coinbase address from the block proposer
	coinbase, err := k.GetCoinbaseAddress(ctx, proposerAddress)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to obtain coinbase address")
	}

	baseFee := k.GetBaseFee(ctx, ethCfg)
	return &statedb.EVMConfig{
		Params:      params,
		ChainConfig: ethCfg,
		CoinBase:    coinbase,
		BaseFee:     baseFee,
	}, nil
}

// TxConfig loads `TxConfig` from current transient storage
func (k *Keeper) TxConfig(ctx sdk.Context, txHash common.Hash) statedb.TxConfig {
	return statedb.NewTxConfig(
		common.BytesToHash(ctx.HeaderHash()), // BlockHash
		txHash,                               // TxHash
		uint(k.GetTxIndexTransient(ctx)),     // TxIndex
		uint(k.GetLogSizeTransient(ctx)),     // LogIndex
	)
}

// VMConfig creates an EVM configuration from the debug setting and the extra EIPs enabled on the
// module parameters. The config generated uses the default JumpTable from the EVM.
func (k Keeper) VMConfig(ctx sdk.Context, _ core.Message, cfg *statedb.EVMConfig, tracer vm.EVMLogger) vm.Config {
	noBaseFee := true
	if types.IsLondon(cfg.ChainConfig, ctx.BlockHeight()) {
		noBaseFee = k.feeMarketKeeper.GetParams(ctx).NoBaseFee
	}

	var debug bool
	if _, ok := tracer.(types.NoOpTracer); !ok {
		debug = true
	}

	return vm.Config{
		Debug:     debug,
		Tracer:    tracer,
		NoBaseFee: noBaseFee,
		ExtraEips: cfg.Params.EIPs(),
	}
}
