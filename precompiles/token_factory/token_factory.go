// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package tokenfactory

import (
	"bytes"
	"embed"

	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	erc20Keeper "github.com/evmos/evmos/v18/x/erc20/keeper"
	transferkeeper "github.com/evmos/evmos/v18/x/ibc/transfer/keeper"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	ackeeper "github.com/evmos/evmos/v18/x/access_control/keeper"
)

var _ vm.PrecompiledContract = &Precompile{}

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

type Precompile struct {
	cmn.Precompile
	authzKeeper    authzkeeper.Keeper
	accountKeeper  authkeeper.AccountKeeper
	bankKeeper     bankkeeper.Keeper
	evmKeeper      EVMKeeper
	erc20Keeper    erc20Keeper.Keeper
	transferKeeper transferkeeper.Keeper
	acKeeper       ackeeper.Keeper
}

// NewPrecompile creates a new TokenFactory Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(
	authzKeeper authzkeeper.Keeper,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	evmKeeper EVMKeeper,
	erc20Keeper erc20Keeper.Keeper,
	transferKeeper transferkeeper.Keeper,
	acKeeper ackeeper.Keeper,
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
		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
		evmKeeper:      evmKeeper,
		erc20Keeper:    erc20Keeper,
		transferKeeper: transferKeeper,
		acKeeper:       acKeeper,
	}, nil
}

// Address defines the address of the Token Factory compile contract.
// address: 0x0000000000000000000000000000000000000900
// TODO: Update the address to the one we decide.
func (Precompile) Address() common.Address {
	return common.HexToAddress("0x0000000000000000000000000000000000000900")
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
	// Token Factory transactions
	case MethodCreateERC20:
		bz, err = p.CreateERC20(ctx, contract, stateDB, method, args)
	case MethodCreate2ERC20:
		bz, err = p.Create2ERC20(ctx, contract, stateDB, method, args)
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
// Available token factory transactions are:
//   - CreateERC20
//   - Create2ERC20
func (Precompile) IsTransaction(method string) bool {
	switch method {
	case MethodCreateERC20, MethodCreate2ERC20:
		return true
	default:
		return false
	}
}
