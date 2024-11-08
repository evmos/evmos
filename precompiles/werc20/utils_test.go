// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package werc20_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
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
// different contract instances that are subject to testing here.
type CallsData struct {
	// This field is used to perform transactions that are not relevant for
	// testing purposes like query to the contract.
	sender keyring.Key

	// precompileReverter is used to call into the werc20 interface and
	precompileReverterAddr common.Address
	precompileReverterABI  abi.ABI

	precompileAddr common.Address
	precompileABI  abi.ABI
}

// getTxCallArgs is a helper function to return the correct call arguments and
// transaction data for a given call type.
func (cd CallsData) getTxAndCallArgs(
	callType callType,
	methodName string,
	args ...interface{},
) (evmtypes.EvmTxArgs, factory.CallArgs) {
	txArgs := evmtypes.EvmTxArgs{}
	callArgs := factory.CallArgs{}

	switch callType {
	case directCall:
		txArgs.To = &cd.precompileAddr
		callArgs.ContractABI = cd.precompileABI
	case contractCall:
		txArgs.To = &cd.precompileReverterAddr
		callArgs.ContractABI = cd.precompileReverterABI
	}

	callArgs.MethodName = methodName
	callArgs.Args = args

	// Setting gas tip cap to zero to have zero gas price.
	txArgs.GasTipCap = new(big.Int).SetInt64(0)
	// Gas limit is added only to skip the estimate gas call
	// that makes debugging more complex.
	txArgs.GasLimit = 1_000_000_000_000

	return txArgs, callArgs
}
