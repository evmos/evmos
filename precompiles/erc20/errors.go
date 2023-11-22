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

var (
	ErrInsufficientAllowance        = errors.New("ERC20: insufficient allowance")
	ErrTransferAmountExceedsBalance = errors.New("ERC20: transfer amount exceeds balance")
)

// BuildExecRevertedErr returns a mocked error that should align with the
// behavior of the original ERC20 Solidity implementation.
//
// FIXME: This is not yet producing the correct reason bytes.
func BuildExecRevertedErr(reason string) (error, error) {
	// The reason bytes are prefixed with this byte array -> see UnpackRevert in ABI implementation in Geth.
	prefixBytes := crypto.Keccak256([]byte("Error(string)"))

	// This is reverse-engineering the ABI encoding of the revert reason.
	typ, _ := abi.NewType("string", "", nil)
	packedReason, err := (abi.Arguments{{Type: typ}}).Pack(reason)
	if err != nil {
		return nil, errors.New("failed to pack revert reason")
	}

	var reasonBytes []byte
	reasonBytes = append(reasonBytes, prefixBytes...)
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
