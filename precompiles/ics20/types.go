// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package ics20

import (
	"fmt"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v14/precompiles/authorization"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
)

// EventIBCTransfer is the event type emitted when a transfer is executed.
type EventIBCTransfer struct {
	Sender        common.Address
	Receiver      common.Hash
	SourcePort    string
	SourceChannel string
	Denom         string
	Amount        *big.Int
	Memo          string
}

// EventTransferAuthorization is the event type emitted when a transfer authorization is created.
type EventTransferAuthorization struct {
	Grantee       common.Address
	Granter       common.Address
	SourcePort    string
	SourceChannel string
	SpendLimit    []cmn.Coin
}

// EventRevokeAuthorization is the event type emitted when a transfer authorization is revoked.
type EventRevokeAuthorization struct {
	Owner   common.Address
	Spender common.Address
}

// DenomTraceResponse defines the data for the denom trace response.
type DenomTraceResponse struct {
	DenomTrace transfertypes.DenomTrace
}

// PageRequest defines the data for the page request.
type PageRequest struct {
	PageRequest query.PageRequest
}

// DenomTracesResponse defines the data for the denom traces response.
type DenomTracesResponse struct {
	DenomTraces  []transfertypes.DenomTrace
	PageResponse query.PageResponse
}

// Allocation defines the spend limit for a particular port and channel
// we need this to be able to unpack to big.Int instead of sdkmath.Int
type Allocation struct {
	SourcePort    string
	SourceChannel string
	SpendLimit    []cmn.Coin
	AllowList     []string
}

// height is a struct used to parse the TimeoutHeight parameter
// used as input in the transfer method
type height struct {
	TimeoutHeight clienttypes.Height
}

// allocs is a struct used to parse the Allocations parameter
// used as input in the transfer authorization method
type allocs struct {
	Allocations []Allocation
}

// NewTransferAuthorization returns a new transfer authorization authz type from the given arguments.
func NewTransferAuthorization(method *abi.Method, args []interface{}) (common.Address, *transfertypes.TransferAuthorization, error) {
	spender, allocations, err := checkTransferAuthzArgs(method, args)
	if err != nil {
		return common.Address{}, nil, err
	}

	transferAuthz := &transfertypes.TransferAuthorization{Allocations: allocations}
	if err = transferAuthz.ValidateBasic(); err != nil {
		return common.Address{}, nil, err
	}

	return spender, transferAuthz, nil
}

// NewMsgTransfer returns a new transfer message from the given arguments.
func NewMsgTransfer(method *abi.Method, args []interface{}) (*transfertypes.MsgTransfer, common.Address, error) {
	if len(args) != 9 {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 9, len(args))
	}

	sourcePort, ok := args[0].(string)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(ErrInvalidSourcePort)
	}

	sourceChannel, ok := args[1].(string)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(ErrInvalidSourceChannel)
	}

	denom, ok := args[2].(string)
	if !ok {
		return nil, common.Address{}, errorsmod.Wrapf(transfertypes.ErrInvalidDenomForTransfer, cmn.ErrInvalidDenom, args[2])
	}

	amount, ok := args[3].(*big.Int)
	if !ok || amount == nil {
		return nil, common.Address{}, errorsmod.Wrapf(transfertypes.ErrInvalidAmount, cmn.ErrInvalidAmount, args[3])
	}

	sender, ok := args[4].(common.Address)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(ErrInvalidSender, args[4])
	}

	receiver, ok := args[5].(string)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(ErrInvalidReceiver, args[5])
	}

	var input height
	heightArg := abi.Arguments{method.Inputs[6]}
	if err := heightArg.Copy(&input, []interface{}{args[6]}); err != nil {
		return nil, common.Address{}, fmt.Errorf("error while unpacking args to TransferInput struct: %s", err)
	}

	timeoutTimestamp, ok := args[7].(uint64)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(ErrInvalidTimeoutTimestamp, args[7])
	}

	memo, ok := args[8].(string)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(ErrInvalidMemo, args[8])
	}

	// Use instance to prevent errors on denom or amount
	token := sdk.Coin{
		Denom:  denom,
		Amount: sdk.NewIntFromBigInt(amount),
	}

	msg := &transfertypes.MsgTransfer{
		SourcePort:       sourcePort,
		SourceChannel:    sourceChannel,
		Token:            token,
		Sender:           sdk.AccAddress(sender.Bytes()).String(), // convert to bech32 format
		Receiver:         receiver,
		TimeoutHeight:    input.TimeoutHeight,
		TimeoutTimestamp: timeoutTimestamp,
		Memo:             memo,
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, common.Address{}, err
	}

	return msg, sender, nil
}

// NewDenomTraceRequest returns a new denom trace request from the given arguments.
func NewDenomTraceRequest(args []interface{}) (*transfertypes.QueryDenomTraceRequest, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid input arguments. Expected 1, got %d", len(args))
	}

	hash, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf(ErrInvalidHash, args[0])
	}

	req := &transfertypes.QueryDenomTraceRequest{
		Hash: hash,
	}

	return req, nil
}

// NewDenomTracesRequest returns a new denom traces request from the given arguments.
func NewDenomTracesRequest(method *abi.Method, args []interface{}) (*transfertypes.QueryDenomTracesRequest, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	var pageRequest PageRequest
	if err := method.Inputs.Copy(&pageRequest, args); err != nil {
		return nil, fmt.Errorf("error while unpacking args to PageRequest: %w", err)
	}

	req := &transfertypes.QueryDenomTracesRequest{
		Pagination: &pageRequest.PageRequest,
	}

	return req, nil
}

// NewDenomHashRequest returns a new denom hash request from the given arguments.
func NewDenomHashRequest(args []interface{}) (*transfertypes.QueryDenomHashRequest, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid input arguments. Expected 1, got %d", len(args))
	}

	trace, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid denom trace")
	}

	req := &transfertypes.QueryDenomHashRequest{
		Trace: trace,
	}

	return req, nil
}

// checkRevokeArgs checks if the given arguments are valid for the Revoke tx.
func checkRevokeArgs(args []interface{}) (common.Address, error) {
	if len(args) != 1 {
		return common.Address{}, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	spender, ok := args[0].(common.Address)
	if !ok || spender == (common.Address{}) {
		return common.Address{}, fmt.Errorf(authorization.ErrInvalidGranter, args[0])
	}

	return spender, nil
}

// checkAllowanceArgs checks if the given arguments are valid for the DecreaseAllowance and IncreaseAllowance txs.
func checkAllowanceArgs(args []interface{}) (common.Address, string, string, string, *big.Int, error) {
	if len(args) != 5 {
		return common.Address{}, "", "", "", nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 5, len(args))
	}

	spender, ok := args[0].(common.Address)
	if !ok || spender == (common.Address{}) {
		return common.Address{}, "", "", "", nil, fmt.Errorf(authorization.ErrInvalidGranter, args[0])
	}

	sourcePort, ok := args[1].(string)
	if !ok {
		return common.Address{}, "", "", "", nil, fmt.Errorf(ErrInvalidSourcePort)
	}

	sourceChannel, ok := args[2].(string)
	if !ok {
		return common.Address{}, "", "", "", nil, fmt.Errorf(ErrInvalidSourceChannel)
	}

	denom, ok := args[3].(string)
	if !ok {
		return common.Address{}, "", "", "", nil, errorsmod.Wrapf(transfertypes.ErrInvalidDenomForTransfer, cmn.ErrInvalidDenom, args[2])
	}

	amount, ok := args[4].(*big.Int)
	if !ok || amount == nil {
		return common.Address{}, "", "", "", nil, errorsmod.Wrapf(transfertypes.ErrInvalidAmount, cmn.ErrInvalidAmount, args[3])
	}

	return spender, sourcePort, sourceChannel, denom, amount, nil
}

// checkTransferArgs checks if the given arguments are valid for the Transfer Approve tx.
func checkTransferAuthzArgs(method *abi.Method, args []interface{}) (common.Address, []transfertypes.Allocation, error) {
	if len(args) != 2 {
		return common.Address{}, nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	spender, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, nil, fmt.Errorf(authorization.ErrInvalidGranter, args[0])
	}

	var input allocs
	allocArg := abi.Arguments{method.Inputs[1]}
	if err := allocArg.Copy(&input, []interface{}{args[1]}); err != nil {
		return common.Address{}, nil, fmt.Errorf("error while unpacking args to AuthInput struct: %s", err)
	}

	allocations := make([]transfertypes.Allocation, len(input.Allocations))
	for i, a := range input.Allocations {
		spendLimit := make(sdk.Coins, len(a.SpendLimit))
		for is, sl := range a.SpendLimit {
			spendLimit[is] = sdk.Coin{
				Amount: math.NewIntFromBigInt(sl.Amount),
				Denom:  sl.Denom,
			}
		}

		allocations[i] = transfertypes.Allocation{
			SourcePort:    a.SourcePort,
			SourceChannel: a.SourceChannel,
			SpendLimit:    spendLimit,
		}
	}

	return spender, allocations, nil
}

// checkAllocationExists checks if the given authorization allocation matches the given arguments.
func checkAllocationExists(allocations []transfertypes.Allocation, sourcePort, sourceChannel, denom string) (spendLimit sdk.Coin, allocationIdx int, err error) {
	var found bool
	spendLimit = sdk.Coin{Denom: denom, Amount: sdk.ZeroInt()}

	for i, allocation := range allocations {
		if allocation.SourcePort != sourcePort || allocation.SourceChannel != sourceChannel {
			continue
		}

		found, spendLimit = allocation.SpendLimit.Find(denom)
		if !found {
			return spendLimit, 0, fmt.Errorf(ErrNoMatchingAllocation, sourcePort, sourceChannel, denom)
		}

		return spendLimit, i, nil
	}

	return spendLimit, 0, fmt.Errorf(ErrNoMatchingAllocation, sourcePort, sourceChannel, denom)
}
