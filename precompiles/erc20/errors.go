// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package erc20

import (
	"errors"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v20/ibc"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

// Errors that have formatted information are defined here as a string.
const (
	ErrIntegerOverflow           = "amount %s causes integer overflow"
	ErrInvalidOwner              = "invalid from address: %s"
	ErrInvalidReceiver           = "invalid to address: %s"
	ErrNoAllowanceForToken       = "allowance for token %s does not exist"
	ErrSubtractMoreThanAllowance = "subtracted value cannot be greater than existing allowance for denom %s: %s > %s"
	ErrCannotReceiveFunds        = "cannot receive funds, received: %s"
)

var (
	// errorSignature are the prefix bytes for the hex-encoded reason string. See UnpackRevert in ABI implementation in Geth.
	errorSignature = crypto.Keccak256([]byte("Error(string)"))

	// Precompile errors
	ErrDecreaseNonPositiveValue = errors.New("cannot decrease allowance with non-positive values")
	ErrIncreaseNonPositiveValue = errors.New("cannot increase allowance with non-positive values")
	ErrNegativeAmount           = errors.New("cannot approve negative values")
	ErrSpenderIsOwner           = errors.New("spender cannot be the owner")

	// ERC20 errors
	ErrDecreasedAllowanceBelowZero  = errors.New("ERC20: decreased allowance below zero")
	ErrInsufficientAllowance        = errors.New("ERC20: insufficient allowance")
	ErrTransferAmountExceedsBalance = errors.New("ERC20: transfer amount exceeds balance")
	ErrOwnableInvalidOwner          = errors.New("ERC20: invalid new owner")
	ErrOwnableUnauthorizedAccount   = errors.New("ERC20: unauthorized account")
	ErrMinterIsNotOwner             = errors.New("ERC20: minter is not the owner")
	ErrSenderIsNotOwner             = errors.New("ERC20: sender is not the owner")
	ErrContractOwnerNotFound        = errors.New("contract owner not found")
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

// ConvertErrToERC20Error is a helper function which maps errors raised by the Cosmos SDK stack
// to the corresponding errors which are raised by an ERC20 contract.
//
// TODO: Create the full RevertError types instead of just the standard error type.
//
// TODO: Return ERC-6093 compliant errors.
func ConvertErrToERC20Error(err error) error {
	switch {
	case strings.Contains(err.Error(), "spendable balance"):
		return ErrTransferAmountExceedsBalance
	case strings.Contains(err.Error(), "minter is not the owner"):
		return ErrMinterIsNotOwner
	case strings.Contains(err.Error(), "sender is not the owner"):
		return ErrSenderIsNotOwner
	case strings.Contains(err.Error(), "contract owner not found"):
		return ErrContractOwnerNotFound
	case strings.Contains(err.Error(), "requested amount is more than spend limit"):
		return ErrInsufficientAllowance
	case strings.Contains(err.Error(), authz.ErrNoAuthorizationFound.Error()):
		return ErrInsufficientAllowance
	case strings.Contains(err.Error(), "subtracted value cannot be greater than existing allowance"):
		return ErrDecreasedAllowanceBelowZero
	case strings.Contains(err.Error(), cmn.ErrIntegerOverflow):
		return vm.ErrExecutionReverted
	case strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "authorization not found"):
		return ErrOwnableUnauthorizedAccount
	case strings.Contains(err.Error(), "invalid owner"):
		return ErrOwnableInvalidOwner
	case errors.Is(err, ibc.ErrNoIBCVoucherDenom) ||
		errors.Is(err, ibc.ErrDenomTraceNotFound) ||
		strings.Contains(err.Error(), "invalid base denomination") ||
		strings.Contains(err.Error(), "display denomination not found") ||
		strings.Contains(err.Error(), "invalid decimals"):
		// NOTE: These are the cases when trying to query metadata of a contract, which has no metadata available.
		// The ERC20 contract raises an "execution reverted" error, without any further information here, which we
		// reproduce (even though it's less verbose than the actual error).
		return vm.ErrExecutionReverted
	default:
		return err
	}
}
