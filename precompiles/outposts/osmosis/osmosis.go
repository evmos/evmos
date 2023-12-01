// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package osmosis

import (
	"embed"
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channelkeeper "github.com/cosmos/ibc-go/v7/modules/core/04-channel/keeper"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v16/precompiles/common"
	"github.com/evmos/evmos/v16/precompiles/ics20"
	erc20keeper "github.com/evmos/evmos/v16/x/erc20/keeper"
	transferkeeper "github.com/evmos/evmos/v16/x/ibc/transfer/keeper"
)

const (
	// OsmosisPrefix represents the human readable part for bech32 addresses on the Osmosis chain.
	OsmosisPrefix = "osmo"

	// OsmosisOutpostAddress is the address of the Osmosis outpost precompile.
	OsmosisOutpostAddress = "0x0000000000000000000000000000000000000901"

	// XCSContract address for Osmosis testnet.
	XCSContractTestnet = "osmo18rj46qcpr57m3qncrj9cuzm0gn3km08w5jxxlnw002c9y7xex5xsu74ytz"
	// XCSContract address for Osmosis mainnet.
	XCSContractMainnet = ""
)

var _ vm.PrecompiledContract = &Precompile{}

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

// Precompile is the structure that defines the Osmosis outpost precompile extending
// the common Precompile type.
type Precompile struct {
	cmn.Precompile
	// IBC
	portID           string
	channelID        string
	timeoutHeight    clienttypes.Height
	timeoutTimestamp uint64

	// Osmosis
	osmosisXCSContract string

	// Keepers
	bankKeeper     bankkeeper.Keeper
	transferKeeper transferkeeper.Keeper
	stakingKeeper  stakingkeeper.Keeper
	erc20Keeper    erc20keeper.Keeper
	channelKeeper  channelkeeper.Keeper
}

// NewPrecompile creates a new Osmosis outpost Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(
	authzKeeper authzkeeper.Keeper,
	portID, channelID string,
	osmosisXCSContract string,
	bankKeeper bankkeeper.Keeper,
	transferKeeper transferkeeper.Keeper,
	stakingKeeper stakingkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
	channelKeeper channelkeeper.Keeper,
) (*Precompile, error) {
	newAbi, err := LoadABI()
	if err != nil {
		return nil, err
	}

	err = ValidateOsmosisContractAddress(osmosisXCSContract)
	if err != nil {
		return nil, err
	}

	return &Precompile{
		Precompile: cmn.Precompile{
			ABI:                  newAbi,
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
			ApprovalExpiration:   cmn.DefaultExpirationDuration,
			AuthzKeeper:          authzKeeper,
		},
		portID:             portID,
		channelID:          channelID,
		timeoutHeight:      clienttypes.NewHeight(ics20.DefaultTimeoutHeight, ics20.DefaultTimeoutHeight),
		timeoutTimestamp:   ics20.DefaultTimeoutTimestamp,
		osmosisXCSContract: osmosisXCSContract,
		bankKeeper:         bankKeeper,
		transferKeeper:     transferKeeper,
		stakingKeeper:      stakingKeeper,
		erc20Keeper:        erc20Keeper,
		channelKeeper:      channelKeeper,
	}, nil
}

// LoadABI loads the Osmosis outpost ABI from the embedded abi.json file
// for the Osmosis outpost precompile.
func LoadABI() (abi.ABI, error) {
	return cmn.LoadABI(f, "abi.json")
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

// Run executes the precompiled contract Swap method.
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	ctx, stateDB, method, initialGas, args, err := p.RunSetup(evm, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	// This handles any out of gas errors that may occur during the execution of a precompile tx or query.
	// It avoids panics and returns the out of gas error so the EVM can continue gracefully.
	defer cmn.HandleGasError(ctx, contract, initialGas, &err)()

	switch method.Name {
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
