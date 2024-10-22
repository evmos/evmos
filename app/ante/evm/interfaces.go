// Copyright Tharsis Labs Ltd.(Eidon-chain)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/Eidon-AI/eidon-chain/blob/main/LICENSE)
package evm

import (
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/Eidon-AI/eidon-chain/v20/x/evm/core/vm"

	"github.com/Eidon-AI/eidon-chain/v20/x/evm/statedb"
	evmtypes "github.com/Eidon-AI/eidon-chain/v20/x/evm/types"
	feemarkettypes "github.com/Eidon-AI/eidon-chain/v20/x/feemarket/types"
)

// EVMKeeper defines the expected keeper interface used on the AnteHandler
type EVMKeeper interface { //nolint: revive
	statedb.Keeper

	NewEVM(ctx sdk.Context, msg core.Message, cfg *statedb.EVMConfig, tracer vm.EVMLogger, stateDB vm.StateDB) *vm.EVM
	DeductTxCostsFromUserBalance(ctx sdk.Context, fees sdk.Coins, from common.Address) error
	GetBalance(ctx sdk.Context, addr common.Address) *big.Int
	ResetTransientGasUsed(ctx sdk.Context)
	GetTxIndexTransient(ctx sdk.Context) uint64
	GetParams(ctx sdk.Context) evmtypes.Params
	// GetBaseFee returns the BaseFee param from the fee market module
	// adapted according to the evm denom decimals
	GetBaseFee(ctx sdk.Context) *big.Int
	// GetMinGasPrice returns the MinGasPrice param from the fee market module
	// adapted according to the evm denom decimals
	GetMinGasPrice(ctx sdk.Context) math.LegacyDec
}

type FeeMarketKeeper interface {
	GetParams(ctx sdk.Context) (params feemarkettypes.Params)
	AddTransientGasWanted(ctx sdk.Context, gasWanted uint64) (uint64, error)
	GetBaseFeeEnabled(ctx sdk.Context) bool
	GetBaseFee(ctx sdk.Context) math.LegacyDec
}

type protoTxProvider interface {
	GetProtoTx() *tx.Tx
}
