// Copyright 2021 Evmos Foundation
// This file is part of Evmos' Evmos packages.
//
// The Evmos packages is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/ethermint/blob/main/LICENSE
package ethante

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tx "github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	"github.com/evmos/evmos/v11/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v11/x/evm/types"
	evm "github.com/evmos/evmos/v11/x/evm/vm"
	feemarkettypes "github.com/evmos/evmos/v11/x/feemarket/types"
)

// DynamicFeeEVMKeeper is a subset of EVMKeeper interface that supports dynamic fee checker
type DynamicFeeEVMKeeper interface {
	ChainID() *big.Int
	GetParams(ctx sdk.Context) evmtypes.Params
	GetBaseFee(ctx sdk.Context, ethCfg *params.ChainConfig) *big.Int
}

// EVMKeeper defines the expected keeper interface used on the Eth AnteHandler
type EVMKeeper interface {
	statedb.Keeper
	DynamicFeeEVMKeeper

	NewEVM(ctx sdk.Context, msg core.Message, cfg *statedb.EVMConfig, tracer vm.EVMLogger, stateDB vm.StateDB) evm.EVM
	DeductTxCostsFromUserBalance(ctx sdk.Context, fees sdk.Coins, from common.Address) error
	GetBalance(ctx sdk.Context, addr common.Address) *big.Int
	ResetTransientGasUsed(ctx sdk.Context)
	GetTxIndexTransient(ctx sdk.Context) uint64
	GetChainConfig(ctx sdk.Context) evmtypes.ChainConfig
	GetAllowUnprotectedTxs(ctx sdk.Context) bool
	GetEVMDenom(ctx sdk.Context) string
	GetEnableCall(ctx sdk.Context) bool
	GetEnableCreate(ctx sdk.Context) bool
}

type protoTxProvider interface {
	GetProtoTx() *tx.Tx
}

// FeeMarketKeeper defines the expected keeper interface used on the AnteHandler
type FeeMarketKeeper interface {
	GetParams(ctx sdk.Context) (params feemarkettypes.Params)
	AddTransientGasWanted(ctx sdk.Context, gasWanted uint64) (uint64, error)
	GetBaseFeeEnabled(ctx sdk.Context) bool
}
