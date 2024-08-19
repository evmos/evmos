// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package auctions

import (
	"embed"
	"fmt"

	erc20keeper "github.com/evmos/evmos/v19/x/erc20/keeper"

	auctionskeeper "github.com/evmos/evmos/v19/x/auctions/keeper"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/ethereum/go-ethereum/common"
	contractutils "github.com/evmos/evmos/v19/contracts/utils"
	cmn "github.com/evmos/evmos/v19/precompiles/common"
	"github.com/evmos/evmos/v19/x/evm/core/vm"
)

var _ vm.PrecompiledContract = &Precompile{}

const PrecompileAddress string = "0x0000000000000000000000000000000000000900"

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

// Precompile defines the precompiled contract for auctions.
type Precompile struct {
	cmn.Precompile
	auctionsKeeper auctionskeeper.Keeper
	erc20Keeper    erc20keeper.Keeper
}

// NewPrecompile creates a new auctions Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(
	auctionsKeeper auctionskeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
	authzKeeper authzkeeper.Keeper,
) (*Precompile, error) {
	newAbi, err := contractutils.LoadABI(f, "abi.json")
	if err != nil {
		return nil, fmt.Errorf("error loading the auction ABI %s", err)
	}

	p := &Precompile{
		Precompile: cmn.Precompile{
			ABI:                  newAbi,
			AuthzKeeper:          authzKeeper,
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
			ApprovalExpiration:   cmn.DefaultExpirationDuration, // should be configurable in the future.
		},
		erc20Keeper:    erc20Keeper,
		auctionsKeeper: auctionsKeeper,
	}

	// SetAddress defines the address of the auction compile contract.
	p.SetAddress(common.HexToAddress(PrecompileAddress))

	return p, nil
}

// RequiredGas calculates the precompiled contract's base gas rate.
func (p Precompile) RequiredGas(input []byte) uint64 {
	// NOTE: This check avoid panicking when trying to decode the method ID
	if len(input) < 4 {
		return 0
	}

	methodID := input[:4]

	method, err := p.MethodById(methodID)
	if err != nil {
		// This should never happen since this method is going to fail during Run
		return 0
	}

	return p.Precompile.RequiredGas(input, p.IsTransaction(method.Name))
}

// Run executes the precompiled contract auction methods defined in the ABI.
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	ctx, stateDB, snapshot, method, initialGas, args, err := p.RunSetup(evm, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	// This handles any out of gas errors that may occur during the execution of a precompile tx or query.
	// It avoids panics and returns the out of gas error so the EVM can continue gracefully.
	defer cmn.HandleGasError(ctx, contract, initialGas, &err)()

	fmt.Println("msg", method.Name)
	switch method.Name {
	// Auction transactions
	case DepositCoinMethod:
		bz, err = p.DepositCoin(ctx, evm.Origin, contract, stateDB, method, args)
	case BidMethod:
		bz, err = p.Bid(ctx, evm.Origin, contract, stateDB, method, args)
	case AuctionInfoMethod:
		bz, err = p.AuctionInfo(ctx, contract, method, args)
	}

	if err != nil {
		return nil, err
	}

	cost := ctx.GasMeter().GasConsumed() - initialGas

	if !contract.UseGas(cost) {
		return nil, vm.ErrOutOfGas
	}

	if err := p.AddJournalEntries(stateDB, snapshot); err != nil {
		return nil, err
	}

	return bz, nil
}

// IsTransaction checks if the given method name corresponds to a transaction.
//
// Available auction transactions are:
//   - Bid
//   - DepositCoin
func (Precompile) IsTransaction(methodName string) bool {
	switch methodName {
	case BidMethod, DepositCoinMethod:
		return true
	default:
		return false
	}
}
