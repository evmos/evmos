package tokenfactory

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core/vm"
)

// EVMKeeper defines the expected EVM keeper.
type EVMKeeper interface {
	AddDynamicPrecompiles(ctx sdk.Context, precompiles ...vm.PrecompiledContract) error
}
