// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gov_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

// callType constants to differentiate between
// the different types of call to the precompile.
type callType int

const (
	directCall callType = iota
	contractCall
)

// CallsData is a helper struct to hold the addresses and ABIs for the
// different contract instances used in the integration tests.
type CallsData struct {
	precompileAddr common.Address
	precompileABI  abi.ABI

	precompileCallerAddr common.Address
	precompileCallerABI  abi.ABI
}

// getTxCallArgs is a helper function to return the correct call arguments and
// transaction data for a given call type.
func (cd CallsData) getTxAndCallArgs(
	callArgs factory.CallArgs,
	txArgs evmtypes.EvmTxArgs,
	callType callType,
	args ...interface{},
) (evmtypes.EvmTxArgs, factory.CallArgs) {
	switch callType {
	case directCall:
		txArgs.To = &cd.precompileAddr
		callArgs.ContractABI = cd.precompileABI
	case contractCall:
		txArgs.To = &cd.precompileCallerAddr
		callArgs.ContractABI = cd.precompileCallerABI
	}

	callArgs.Args = args

	// Setting gas tip cap to zero to have zero gas price and simplify the tests.
	txArgs.GasTipCap = new(big.Int).SetInt64(0)

	return txArgs, callArgs
}
