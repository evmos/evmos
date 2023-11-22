package erc20

import (
	"errors"

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
