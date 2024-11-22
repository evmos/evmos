// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

const (
	// GetSigningInfoMethod defines the ABI method name for the slashing SigningInfo query
	GetSigningInfoMethod = "getSigningInfo"
	// GetSigningInfosMethod defines the ABI method name for the slashing SigningInfos query
	GetSigningInfosMethod = "getSigningInfos"
	// GetParamsMethod defines the ABI method name for the slashing Params query
	GetParamsMethod = "getParams"
)

// GetSigningInfo implements the query to get a validator's signing info.
func (p *Precompile) GetSigningInfo(
	ctx sdk.Context,
	method *abi.Method,
	_ *vm.Contract,
	args []interface{},
) ([]byte, error) {
	req, err := ParseSigningInfoArgs(args)
	if err != nil {
		return nil, err
	}

	res, err := p.slashingKeeper.SigningInfo(ctx, req)
	if err != nil {
		return nil, err
	}

	out := new(SigningInfoOutput).FromResponse(res)
	return method.Outputs.Pack(out.SigningInfo)
}

// GetSigningInfos implements the query to get signing info for all validators.
func (p *Precompile) GetSigningInfos(
	ctx sdk.Context,
	method *abi.Method,
	_ *vm.Contract,
	args []interface{},
) ([]byte, error) {
	req, err := ParseSigningInfosArgs(method, args)
	if err != nil {
		return nil, err
	}

	res, err := p.slashingKeeper.SigningInfos(ctx, req)
	if err != nil {
		return nil, err
	}

	out := new(SigningInfosOutput).FromResponse(res)
	return method.Outputs.Pack(out.SigningInfos, out.PageResponse)
}

// GetParams implements the query to get the slashing parameters.
func (p *Precompile) GetParams(
	ctx sdk.Context,
	method *abi.Method,
	_ *vm.Contract,
	_ []interface{},
) ([]byte, error) {
	res, err := p.slashingKeeper.Params(ctx, &types.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}

	out := new(ParamsOutput).FromResponse(res)
	return method.Outputs.Pack(out.Params)
}
