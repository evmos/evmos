// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package stride

import (
	"embed"
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v16/precompiles/common"
	"github.com/evmos/evmos/v16/precompiles/ics20"
	erc20keeper "github.com/evmos/evmos/v16/x/erc20/keeper"
	transferkeeper "github.com/evmos/evmos/v16/x/ibc/transfer/keeper"
)

var _ vm.PrecompiledContract = &Precompile{}

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

type Precompile struct {
	cmn.Precompile
	portID         string
	channelID      string
	timeoutHeight  clienttypes.Height
	transferKeeper transferkeeper.Keeper
	erc20Keeper    erc20keeper.Keeper
	stakingKeeper  stakingkeeper.Keeper
}

// NewPrecompile creates a new Stride outpost Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(
	portID, channelID string,
	transferKeeper transferkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
	authzKeeper authzkeeper.Keeper,
	stakingKeeper stakingkeeper.Keeper,
) (*Precompile, error) {
	abi, err := LoadABI()
	if err != nil {
		return nil, err
	}

	return &Precompile{
		Precompile: cmn.Precompile{
			ABI:                  abi,
			AuthzKeeper:          authzKeeper,
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
			ApprovalExpiration:   cmn.DefaultExpirationDuration, // should be configurable in the future.
		},
		portID:         portID,
		channelID:      channelID,
		timeoutHeight:  clienttypes.NewHeight(ics20.DefaultTimeoutHeight, ics20.DefaultTimeoutHeight),
		transferKeeper: transferKeeper,
		erc20Keeper:    erc20Keeper,
		stakingKeeper:  stakingKeeper,
	}, nil
}

// LoadABI loads the Stride outpost ABI from the embedded abi.json file
// for the Stride outpost precompile.
func LoadABI() (abi.ABI, error) {
	return cmn.LoadABI(f, "abi.json")
}

// Address defines the address of the Stride Outpost precompile contract.
func (Precompile) Address() common.Address {
	return common.HexToAddress("0x0000000000000000000000000000000000000900")
}

// IsStateful returns true since the precompile contract has access to the
// chain state.
func (Precompile) IsStateful() bool {
	return true
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

	switch method.Name {
	// Stride Outpost Methods:
	case LiquidStakeMethod:
		bz, err = p.LiquidStake(ctx, evm.Origin, stateDB, contract, method, args)
	case RedeemStakeMethod:
		bz, err = p.RedeemStake(ctx, evm.Origin, stateDB, contract, method, args)
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
func (Precompile) IsTransaction(method string) bool {
	switch method {
	case LiquidStakeMethod, RedeemStakeMethod:
		return true
	default:
		return false
	}
}
