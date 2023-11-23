// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package erc20

import (
	"errors"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

// Errors that have formatted information are defined here as a string.
const (
	ErrIntegerOverflow           = "amount %s causes integer overflow"
	ErrNoAllowanceForToken       = "allowance for token %s does not exist"
	ErrSubtractMoreThanAllowance = "subtracted value cannot be greater than existing allowance for denom %s: %s > %s"
)

var (
	// errorSignature are the prefix bytes for the hex-encoded reason string. See UnpackRevert in ABI implementation in Geth.
	errorSignature = crypto.Keccak256([]byte("Error(string)"))

	// Precompile errors
	ErrDecreaseNonPositiveValue = errors.New("cannot decrease allowance with non-positive values")
	ErrDenomTraceNotFound       = errors.New("denom trace not found")
	ErrIncreaseNonPositiveValue = errors.New("cannot increase allowance with non-positive values")
	ErrNegativeAmount           = errors.New("cannot approve negative values")
	ErrNoIBCVoucherDenom        = errors.New("denom is not an IBC voucher")

	// ErrInsufficientAllowance is returned by ERC20 smart contracts in case
	// no authorization is found or the granted amount is not sufficient.
	ErrInsufficientAllowance = errors.New("ERC20: insufficient allowance")
	// ErrTransferAmountExceedsBalance is returned by ERC20 smart contracts in
	// case a transfer is attempted, that sends more than the sender's balance.
	ErrTransferAmountExceedsBalance = errors.New("ERC20: transfer amount exceeds balance")
)

// BuildExecRevertedErr returns a mocked error that should align with the
// behavior of the original ERC20 Solidity implementation.
//
// FIXME: This is not yet producing the correct reason bytes. Will fix on a follow up PR.
func BuildExecRevertedErr(reason string) (error, error) {
	// This is reverse-engineering the ABI encoding of the revert reason.
	typ, err := abi.NewType("string", "", nil)
	if err != nil {
		return nil, err
	}

	packedReason, err := (abi.Arguments{{Type: typ}}).Pack(reason)
	if err != nil {
		return nil, errors.New("failed to pack revert reason")
	}

	var reasonBytes []byte
	reasonBytes = append(reasonBytes, errorSignature...)
	reasonBytes = append(reasonBytes, packedReason...)

	return evmtypes.NewExecErrorWithReason(reasonBytes), nil
}

// convertErrToERC20Error is a helper function which maps errors raised by the Cosmos SDK stack
// to the corresponding errors which are raised by an ERC20 contract.
//
// TODO: Create the full RevertError types instead of just the standard error type.
//
// TODO: Return ERC-6093 compliant errors.
func convertErrToERC20Error(err error) error {
	switch {
	case strings.Contains(err.Error(), "spendable balance"):
		return ErrTransferAmountExceedsBalance
	case strings.Contains(err.Error(), "requested amount is more than spend limit"):
		return ErrInsufficientAllowance
	case strings.Contains(err.Error(), authz.ErrNoAuthorizationFound.Error()):
		return ErrInsufficientAllowance
	default:
		return err
	}
}
