// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package osmosis

import (
	"bytes"
	"embed"
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
	"github.com/evmos/evmos/v15/precompiles/ics20"
	erc20keeper "github.com/evmos/evmos/v15/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v15/x/erc20/types"
	transferkeeper "github.com/evmos/evmos/v15/x/ibc/transfer/keeper"
)

const (
	// OsmosisOutpostAddress is the address of the Osmosis outpost precompile
	OsmosisOutpostAddress = "0x0000000000000000000000000000000000000901"
)

var _ vm.PrecompiledContract = &Precompile{}

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

// Precompile is the structure that define the Osmosis outpost precompile extending
// the common Precompile type.
type Precompile struct {
	cmn.Precompile
	// IBC
	portID        string
	channelID     string
	timeoutHeight clienttypes.Height

	// Osmosis
	osmosisXCSContract string

	// Keepers
	bankKeeper     erc20types.BankKeeper
	transferKeeper transferkeeper.Keeper
	erc20Keeper    erc20keeper.Keeper
}

// NewPrecompile creates a new Osmosis outpost Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(
	portID, channelID string,
	osmosisXCSContract string,
	bankKeeper erc20types.BankKeeper,
	transferKeeper transferkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
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
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
			ApprovalExpiration:   cmn.DefaultExpirationDuration, // should be configurable in the future.
		},
		portID:             portID,
		channelID:          channelID,
		timeoutHeight:      clienttypes.NewHeight(ics20.DefaultTimeoutHeight, ics20.DefaultTimeoutHeight),
		osmosisXCSContract: osmosisXCSContract,
		transferKeeper:     transferKeeper,
		bankKeeper:         bankKeeper,
		erc20Keeper:        erc20Keeper,
	}, nil
}

// Address defines the address of the Osmosis outpost precompile contract.
func (Precompile) Address() common.Address {
	return common.HexToAddress(OsmosisOutpostAddress)
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

// IsTransaction checks if the given method name corresponds to a transaction or query.
func (Precompile) IsTransaction(method string) bool {
	switch method {
	case SwapMethod:
		return true
	default:
		return false
	}
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
	// Osmosis Outpost Methods:
	case SwapMethod:
		bz, err = p.Swap(ctx, evm.Origin, stateDB, contract, method, args)
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
