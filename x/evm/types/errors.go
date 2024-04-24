// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package types

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	codeErrInvalidState = uint32(iota) + 2 // NOTE: code 1 is reserved for internal errors
	codeErrInvalidChainConfig
	codeErrZeroAddress
	codeErrCreateDisabled
	codeErrCallDisabled
	codeErrInvalidAmount
	codeErrInvalidGasPrice
	codeErrInvalidGasFee
	codeErrVMExecution
	codeErrInvalidRefund
	codeErrInvalidGasCap
	codeErrInvalidBaseFee
	codeErrGasOverflow
	codeErrInvalidAccount
	codeErrInvalidGasLimit
	codeErrInactivePrecompile
)

var ErrPostTxProcessing = errors.New("failed to execute post processing")

var (
	// ErrInvalidState returns an error resulting from an invalid Storage State.
	ErrInvalidState = errorsmod.Register(ModuleName, codeErrInvalidState, "invalid storage state")

	// ErrInvalidChainConfig returns an error resulting from an invalid ChainConfig.
	ErrInvalidChainConfig = errorsmod.Register(ModuleName, codeErrInvalidChainConfig, "invalid chain configuration")

	// ErrZeroAddress returns an error resulting from an zero (empty) ethereum Address.
	ErrZeroAddress = errorsmod.Register(ModuleName, codeErrZeroAddress, "invalid zero address")

	// ErrCreateDisabled returns an error if the EnableCreate parameter is false.
	ErrCreateDisabled = errorsmod.Register(ModuleName, codeErrCreateDisabled, "EVM Create operation is disabled")

	// ErrCallDisabled returns an error if the EnableCall parameter is false.
	ErrCallDisabled = errorsmod.Register(ModuleName, codeErrCallDisabled, "EVM Call operation is disabled")

	// ErrInvalidAmount returns an error if a tx contains an invalid amount.
	ErrInvalidAmount = errorsmod.Register(ModuleName, codeErrInvalidAmount, "invalid transaction amount")

	// ErrInvalidGasPrice returns an error if an invalid gas price is provided to the tx.
	ErrInvalidGasPrice = errorsmod.Register(ModuleName, codeErrInvalidGasPrice, "invalid gas price")

	// ErrInvalidGasFee returns an error if the tx gas fee is out of bound.
	ErrInvalidGasFee = errorsmod.Register(ModuleName, codeErrInvalidGasFee, "invalid gas fee")

	// ErrVMExecution returns an error resulting from an error in EVM execution.
	ErrVMExecution = errorsmod.Register(ModuleName, codeErrVMExecution, "evm transaction execution failed")

	// ErrInvalidRefund returns an error if a the gas refund value is invalid.
	ErrInvalidRefund = errorsmod.Register(ModuleName, codeErrInvalidRefund, "invalid gas refund amount")

	// ErrInvalidGasCap returns an error if a the gas cap value is negative or invalid
	ErrInvalidGasCap = errorsmod.Register(ModuleName, codeErrInvalidGasCap, "invalid gas cap")

	// ErrInvalidBaseFee returns an error if a the base fee cap value is invalid
	ErrInvalidBaseFee = errorsmod.Register(ModuleName, codeErrInvalidBaseFee, "invalid base fee")

	// ErrGasOverflow returns an error if gas computation overlow/underflow
	ErrGasOverflow = errorsmod.Register(ModuleName, codeErrGasOverflow, "gas computation overflow/underflow")

	// ErrInvalidAccount returns an error if the account is not an EVM compatible account
	ErrInvalidAccount = errorsmod.Register(ModuleName, codeErrInvalidAccount, "account type is not a valid ethereum account")

	// ErrInvalidGasLimit returns an error if gas limit value is invalid
	ErrInvalidGasLimit = errorsmod.Register(ModuleName, codeErrInvalidGasLimit, "invalid gas limit")

	// ErrInactivePrecompile returns an error if a call is made to an inactive precompile
	ErrInactivePrecompile = errorsmod.Register(ModuleName, codeErrInactivePrecompile, "precompile not enabled")
)

// NewExecErrorWithReason unpacks the revert return bytes and returns a wrapped error
// with the return reason.
func NewExecErrorWithReason(revertReason []byte) *RevertError {
	result := common.CopyBytes(revertReason)
	reason, errUnpack := abi.UnpackRevert(result)
	err := errors.New("execution reverted")
	if errUnpack == nil {
		err = fmt.Errorf("execution reverted: %v", reason)
	}
	return &RevertError{
		error:  err,
		reason: hexutil.Encode(result),
	}
}

// RevertError is an API error that encompass an EVM revert with JSON error
// code and a binary data blob.
type RevertError struct {
	error
	reason string // revert reason hex encoded
}

// ErrorCode returns the JSON error code for a revert.
// See: https://github.com/ethereum/wiki/wiki/JSON-RPC-Error-Codes-Improvement-Proposal
func (e *RevertError) ErrorCode() int {
	return 3
}

// ErrorData returns the hex encoded revert reason.
func (e *RevertError) ErrorData() interface{} {
	return e.reason
}
