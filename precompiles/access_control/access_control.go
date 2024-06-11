// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package accesscontrol

import (
	"embed"
	"fmt"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	erc20 "github.com/evmos/evmos/v18/precompiles/erc20"
	erc20types "github.com/evmos/evmos/v18/x/erc20/types"
	transferkeeper "github.com/evmos/evmos/v18/x/ibc/transfer/keeper"

	ackeeper "github.com/evmos/evmos/v18/x/access_control/keeper"
)

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

const (
	// abiPath defines the path to the ERC-20 precompile ABI JSON file.
	abiPath = "abi.json"
)

var _ vm.PrecompiledContract = &Precompile{}

var abiInstance abi.ABI

func init() {
	var err error
	abiInstance, err = cmn.LoadABI(f, abiPath)
	if err != nil {
		panic(fmt.Errorf("failed to load abi: %w", err))
	}
}

// GetABI returns the ERC-20 precompile ABI.
func GetABI() abi.ABI {
	return abiInstance
}

// Precompile defines the precompiled contract for ERC-20.
type Precompile struct {
	cmn.Precompile
	ERC20Precompile     *erc20.Precompile
	TokenPair           erc20types.TokenPair
	BankKeeper          bankkeeper.Keeper
	AccessControlKeeper ackeeper.Keeper
}

// NewPrecompile creates a new ERC-20 Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(
	tokenPair erc20types.TokenPair,
	bankKeeper bankkeeper.Keeper,
	authzKeeper authzkeeper.Keeper,
	transferKeeper transferkeeper.Keeper,
	acKeeper ackeeper.Keeper,
) (*Precompile, error) {
	newABI, err := cmn.LoadABI(f, abiPath)
	if err != nil {
		return nil, err
	}

	erc20Precompile, err := erc20.NewPrecompile(tokenPair, bankKeeper, authzKeeper, transferKeeper)
	if err != nil {
		return nil, err
	}

	return &Precompile{
		Precompile: cmn.Precompile{
			ABI:                  newABI,
			AuthzKeeper:          authzKeeper,
			ApprovalExpiration:   cmn.DefaultExpirationDuration,
			KvGasConfig:          storetypes.GasConfig{},
			TransientKVGasConfig: storetypes.GasConfig{},
		},
		ERC20Precompile:     erc20Precompile,
		TokenPair:           tokenPair,
		BankKeeper:          bankKeeper,
		AccessControlKeeper: acKeeper,
	}, nil
}

// Address gets the address from the token pair.
func (p Precompile) Address() common.Address {
	return p.TokenPair.GetERC20Contract()
}

// RequiredGas calculates the contract gas used for the
func (p Precompile) RequiredGas(input []byte) uint64 {
	// Validate input length
	if len(input) < 4 {
		return 0
	}

	methodID := input[:4]
	method, err := p.MethodById(methodID)
	if err != nil {
		return 0
	}

	switch method.Name {
	case MethodHasRole, MethodGetRoleAdmin:
		return 2100 // Estimated gas for read operations
	case MethodBurn, MethodMint:
		return 40000 // Estimated gas for mint and burn operations
	case MethodGrantRole, MethodRenounceRole, MethodRevokeRole:
		return 30000 // Estimated gas for role operations
	// Default to ERC-20 precompile
	default:
		return p.ERC20Precompile.RequiredGas(input)
	}
}

// Run executes the precompiled contract ERC-20 methods defined in the ABI.
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	ctx, stateDB, method, initialGas, args, err := p.RunSetup(evm, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	// This handles any out of gas errors that may occur during the execution of a precompile tx or query.
	// It avoids panics and returns the out of gas error so the EVM can continue gracefully.
	defer cmn.HandleGasError(ctx, contract, initialGas, &err)()

	bz, err = p.HandleMethod(ctx, contract, stateDB, method, args)
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
func (p Precompile) IsTransaction(methodName string) bool {
	switch methodName {
	case MethodBurn, MethodMint, MethodGrantRole, MethodRenounceRole, MethodRevokeRole:
		return true
	default:
		return p.ERC20Precompile.IsTransaction(methodName)
	}
}

// HandleMethod handles the execution of each of the ERC-20 methods.
func (p Precompile) HandleMethod(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) (bz []byte, err error) {
	switch method.Name {
	// Role methods
	case MethodGrantRole:
		return p.GrantRole(ctx, contract, stateDB, args)
	case MethodRevokeRole:
		return p.RevokeRole(ctx, contract, stateDB, args)
	case MethodRenounceRole:
		return p.RenounceRole(ctx, contract, stateDB, args)
	case MethodGetRoleAdmin:
		return p.GetRoleAdmin(ctx, method, args)
	case MethodHasRole:
		return p.HasRole(ctx, method, args)
	// Mint and burn methods
	case MethodMint:
		return p.Mint(ctx, contract, stateDB, args)
	case MethodBurn:
		return p.Burn(ctx, contract, stateDB, args)
	// Default to ERC-20 precompile
	default:
		return p.ERC20Precompile.HandleMethod(ctx, contract, stateDB, method, args)
	}
}
