// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package ics20

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v14/precompiles/authorization"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
)

const (
	// DenomTraceMethod defines the ABI method name for the ICS20 DenomTrace
	// query.
	DenomTraceMethod = "denomTrace"
	// DenomTracesMethod defines the ABI method name for the ICS20 DenomTraces
	// query.
	DenomTracesMethod = "denomTraces"
	// DenomHashMethod defines the ABI method name for the ICS20 DenomHash
	// query.
	DenomHashMethod = "denomHash"
)

// DenomTrace returns the requested denomination trace information.
func (p Precompile) DenomTrace(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	req, err := NewDenomTraceRequest(args)
	if err != nil {
		return nil, err
	}

	res, err := p.transferKeeper.DenomTrace(sdk.WrapSDKContext(ctx), req)
	if err != nil {
		// if the trace does not exist, return empty array
		if strings.Contains(err.Error(), ErrTraceNotFound) {
			return method.Outputs.Pack(transfertypes.DenomTrace{})
		}
		return nil, err
	}

	return method.Outputs.Pack(*res.DenomTrace)
}

// DenomTraces returns the requested denomination traces information.
func (p Precompile) DenomTraces(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	req, err := NewDenomTracesRequest(method, args)
	if err != nil {
		return nil, err
	}

	res, err := p.transferKeeper.DenomTraces(sdk.WrapSDKContext(ctx), req)
	if err != nil {
		return nil, err
	}

	return method.Outputs.Pack(res.DenomTraces, res.Pagination)
}

// DenomHash returns the denom hash (in hex format) of the denomination trace information.
func (p Precompile) DenomHash(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	req, err := NewDenomHashRequest(args)
	if err != nil {
		return nil, err
	}

	res, err := p.transferKeeper.DenomHash(sdk.WrapSDKContext(ctx), req)
	if err != nil {
		// if the denom hash does not exist, return empty string
		if strings.Contains(err.Error(), ErrTraceNotFound) {
			return method.Outputs.Pack("")
		}
		return nil, err
	}

	return method.Outputs.Pack(res.Hash)
}

// Allowance returns the remaining allowance of a spender.
func (p Precompile) Allowance(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// append here the msg type. Will always be the TransferMsg
	// for this precompile
	args = append(args, TransferMsg)

	owner, spender, msg, err := authorization.CheckAllowanceArgs(args)
	if err != nil {
		return nil, err
	}

	msgAuthz, _ := p.AuthzKeeper.GetAuthorization(ctx, spender.Bytes(), owner.Bytes(), msg)

	if msgAuthz == nil {
		// return empty array
		return method.Outputs.Pack([]Allocation{})
	}

	transferAuthz, ok := msgAuthz.(*transfertypes.TransferAuthorization)
	if !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "transfer authorization", &transfertypes.TransferAuthorization{}, transferAuthz)
	}

	// need to convert to ics20.Allocation (uses big.Int)
	// because ibc Allocation has sdkmath.Int
	allocs := make([]Allocation, len(transferAuthz.Allocations))
	for i, a := range transferAuthz.Allocations {
		spendLimit := make([]cmn.Coin, len(a.SpendLimit))
		for j, c := range a.SpendLimit {
			spendLimit[j] = cmn.Coin{
				Denom:  c.Denom,
				Amount: c.Amount.BigInt(),
			}
		}

		allocs[i] = Allocation{
			SourcePort:    a.SourcePort,
			SourceChannel: a.SourceChannel,
			SpendLimit:    spendLimit,
			AllowList:     a.AllowList,
		}
	}

	return method.Outputs.Pack(allocs)
}
