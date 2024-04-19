// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package ics20

import (
	"bytes"
	"embed"
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	channelkeeper "github.com/cosmos/ibc-go/v7/modules/core/04-channel/keeper"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v16/precompiles/authorization"
	cmn "github.com/evmos/evmos/v16/precompiles/common"
	transferkeeper "github.com/evmos/evmos/v16/x/ibc/transfer/keeper"
)

var _ vm.PrecompiledContract = &Precompile{}

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

type Precompile struct {
	cmn.Precompile
	transferKeeper transferkeeper.Keeper
	channelKeeper  channelkeeper.Keeper
}

// NewPrecompile creates a new ICS-20 Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(
	transferKeeper transferkeeper.Keeper,
	channelKeeper channelkeeper.Keeper,
	authzKeeper authzkeeper.Keeper,
) (*Precompile, error) {
	abiBz, err := f.ReadFile("abi.json")
	if err != nil {
		return nil, err
	}

	newAbi, err := abi.JSON(bytes.NewReader(abiBz))
	if err != nil {
		return nil, err
	}

	return &Precompile{
		Precompile: cmn.Precompile{
			ABI:                  newAbi,
			AuthzKeeper:          authzKeeper,
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
			ApprovalExpiration:   cmn.DefaultExpirationDuration, // should be configurable in the future.
		},
		transferKeeper: transferKeeper,
		channelKeeper:  channelKeeper,
	}, nil
}

// Address defines the address of the ICS-20 compile contract.
// address: 0x0000000000000000000000000000000000000802
func (Precompile) Address() common.Address {
	return common.HexToAddress("0x0000000000000000000000000000000000000802")
}

// RequiredGas calculates the precompiled contract's base gas rate.
func (p Precompile) RequiredGas(input []byte) uint64 {
	methodID := input[:4]

	method, err := p.MethodById(methodID)
	if err != nil {
		// This should never happen since this method is going to fail during Run
		return 0
	}

	return p.Precompile.RequiredGas(input, p.IsTransaction(method.Name))
}

// Run executes the precompiled contract IBC transfer methods defined in the ABI.
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	ctx, stateDB, method, initialGas, args, err := p.RunSetup(evm, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	// This handles any out of gas errors that may occur during the execution of a precompile tx or query.
	// It avoids panics and returns the out of gas error so the EVM can continue gracefully.
	defer cmn.HandleGasError(ctx, contract, initialGas, &err)()

	if err := stateDB.Commit(); err != nil {
		return nil, err
	}

	switch method.Name {
	// TODO Approval transactions => need cosmos-sdk v0.46 & ibc-go v6.2.0
	// Authorization Methods:
	case authorization.ApproveMethod:
		bz, err = p.Approve(ctx, evm.Origin, stateDB, method, args)
	case authorization.RevokeMethod:
		bz, err = p.Revoke(ctx, evm.Origin, stateDB, method, args)
	case authorization.IncreaseAllowanceMethod:
		bz, err = p.IncreaseAllowance(ctx, evm.Origin, stateDB, method, args)
	case authorization.DecreaseAllowanceMethod:
		bz, err = p.DecreaseAllowance(ctx, evm.Origin, stateDB, method, args)
	// ICS20 transactions
	case TransferMethod:
		bz, err = p.Transfer(ctx, evm.Origin, contract, stateDB, method, args)
	// ICS20 queries
	case DenomTraceMethod:
		bz, err = p.DenomTrace(ctx, contract, method, args)
	case DenomTracesMethod:
		bz, err = p.DenomTraces(ctx, contract, method, args)
	case DenomHashMethod:
		bz, err = p.DenomHash(ctx, contract, method, args)
	case authorization.AllowanceMethod:
		bz, err = p.Allowance(ctx, method, args)
	default:
		return nil, fmt.Errorf(cmn.ErrUnknownMethod, method.Name)
	}

	if err != nil {
		return nil, err
	}

	cost := ctx.GasMeter().GasConsumed() - initialGas

	if !contract.UseGas(cost) {
		return nil, vm.ErrOutOfGas
	}

	return bz, nil
}

// IsTransaction checks if the given method name corresponds to a transaction or query.
//
// Available ics20 transactions are:
//   - Transfer
//
// Available authorization transactions are:
//   - Approve
//   - Revoke
//   - IncreaseAllowance
//   - DecreaseAllowance
func (Precompile) IsTransaction(method string) bool {
	switch method {
	case TransferMethod,
		authorization.ApproveMethod,
		authorization.RevokeMethod,
		authorization.IncreaseAllowanceMethod,
		authorization.DecreaseAllowanceMethod:
		return true
	default:
		return false
	}
}
