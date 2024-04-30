// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package erc20

import (
	"embed"
	"fmt"

	cmn "github.com/evmos/evmos/v18/precompiles/common"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	auth "github.com/evmos/evmos/v18/precompiles/authorization"
	erc20types "github.com/evmos/evmos/v18/x/erc20/types"
	transferkeeper "github.com/evmos/evmos/v18/x/ibc/transfer/keeper"
)

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

var _ vm.PrecompiledContract = &Precompile{}

var abiInstance abi.ABI

func init() {
	var err error
	abiInstance, err = cmn.LoadABI(f, abiPath)
	if err != nil {
		panic(fmt.Errorf("failed to load abi: %w", err))
	}
}

const (
	// abiPath defines the path to the ERC-20 precompile ABI JSON file.
	abiPath = "abi.json"

	// NOTE: These gas values have been derived from tests that have been concluded on a testing branch, which
	// is not being merged to the main branch. The reason for this was to not clutter the repository with the
	// necessary tests for this use case.
	//
	// The results can be inspected here:
	// https://github.com/evmos/evmos/blob/malte/erc20-gas-tests/precompiles/erc20/plot_gas_values.ipynb

	GasTransfer          = 9_000
	GasTransferFrom      = 30_500
	GasApprove           = 8_100
	GasIncreaseAllowance = 8_580
	GasDecreaseAllowance = 3_620
	GasName              = 3_421
	GasSymbol            = 3_464
	GasDecimals          = 427
	GasTotalSupply       = 2_480
	GasBalanceOf         = 2_870
	GasAllowance         = 3_225
)

// GetABI returns the ERC-20 precompile ABI.
func GetABI() abi.ABI {
	return abiInstance
}

// Precompile defines the precompiled contract for ERC-20.
type Precompile struct {
	cmn.Precompile
	tokenPair      erc20types.TokenPair
	bankKeeper     bankkeeper.Keeper
	transferKeeper transferkeeper.Keeper
}

// NewPrecompile creates a new ERC-20 Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(
	tokenPair erc20types.TokenPair,
	bankKeeper bankkeeper.Keeper,
	authzKeeper authzkeeper.Keeper,
	transferKeeper transferkeeper.Keeper,
) (*Precompile, error) {
	return &Precompile{
		Precompile: cmn.Precompile{
			ABI:                  abiInstance,
			AuthzKeeper:          authzKeeper,
			ApprovalExpiration:   cmn.DefaultExpirationDuration,
			KvGasConfig:          storetypes.GasConfig{},
			TransientKVGasConfig: storetypes.GasConfig{},
		},
		tokenPair:      tokenPair,
		bankKeeper:     bankKeeper,
		transferKeeper: transferKeeper,
	}, nil
}

// Address defines the address of the ERC-20 precompile contract.
func (p Precompile) Address() common.Address {
	return p.tokenPair.GetERC20Contract()
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

	// TODO: these values were obtained from Remix using the ERC20.sol from OpenZeppelin.
	// We should execute the transactions using the ERC20MinterBurnerDecimals.sol from Evmos testnet
	// to ensure parity in the values.
	switch method.Name {
	// ERC-20 transactions
	case TransferMethod:
		return GasTransfer
	case TransferFromMethod:
		return GasTransferFrom
	case auth.ApproveMethod:
		return GasApprove
	case auth.IncreaseAllowanceMethod:
		return GasIncreaseAllowance
	case auth.DecreaseAllowanceMethod:
		return GasDecreaseAllowance
	// ERC-20 queries
	case NameMethod:
		return GasName
	case SymbolMethod:
		return GasSymbol
	case DecimalsMethod:
		return GasDecimals
	case TotalSupplyMethod:
		return GasTotalSupply
	case BalanceOfMethod:
		return GasBalanceOf
	case auth.AllowanceMethod:
		return GasAllowance
	default:
		return 0
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
func (Precompile) IsTransaction(methodName string) bool {
	switch methodName {
	case TransferMethod,
		TransferFromMethod,
		auth.ApproveMethod,
		auth.IncreaseAllowanceMethod,
		auth.DecreaseAllowanceMethod:
		return true
	default:
		return false
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
	// ERC-20 transactions
	case TransferMethod:
		bz, err = p.Transfer(ctx, contract, stateDB, method, args)
	case TransferFromMethod:
		bz, err = p.TransferFrom(ctx, contract, stateDB, method, args)
	case auth.ApproveMethod:
		bz, err = p.Approve(ctx, contract, stateDB, method, args)
	case auth.IncreaseAllowanceMethod:
		bz, err = p.IncreaseAllowance(ctx, contract, stateDB, method, args)
	case auth.DecreaseAllowanceMethod:
		bz, err = p.DecreaseAllowance(ctx, contract, stateDB, method, args)
	// ERC-20 queries
	case NameMethod:
		bz, err = p.Name(ctx, contract, stateDB, method, args)
	case SymbolMethod:
		bz, err = p.Symbol(ctx, contract, stateDB, method, args)
	case DecimalsMethod:
		bz, err = p.Decimals(ctx, contract, stateDB, method, args)
	case TotalSupplyMethod:
		bz, err = p.TotalSupply(ctx, contract, stateDB, method, args)
	case BalanceOfMethod:
		bz, err = p.BalanceOf(ctx, contract, stateDB, method, args)
	case auth.AllowanceMethod:
		bz, err = p.Allowance(ctx, contract, stateDB, method, args)
	default:
		return nil, fmt.Errorf(cmn.ErrUnknownMethod, method.Name)
	}

	return bz, err
}
