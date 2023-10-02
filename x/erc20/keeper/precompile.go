package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v14/precompiles/erc20"
	"github.com/evmos/evmos/v14/precompiles/werc20"
	"github.com/evmos/evmos/v14/x/erc20/types"
)

func (k Keeper) RegisterERC20Extensions(ctx sdk.Context) error {
	precompiles := make([]vm.PrecompiledContract, 0)
	evmDenom := k.evmKeeper.GetParams(ctx).EvmDenom

	k.IterateTokenPairs(ctx, func(tokenPair types.TokenPair) bool {
		if tokenPair.ContractOwner != types.OWNER_MODULE {
			return false
		}

		var (
			err        error
			precompile vm.PrecompiledContract
		)

		if tokenPair.Denom == evmDenom {
			precompile, err = werc20.NewPrecompile(tokenPair, k.bankKeeper, k, k.authzKeeper)
		} else {
			precompile, err = erc20.NewPrecompile(tokenPair, k.bankKeeper, k, k.authzKeeper)
		}

		if err != nil {
			panic(fmt.Errorf("failed to load precompile for denom %s: %w", tokenPair.Denom, err))
		}

		precompiles = append(precompiles, precompile)
		return false
	})

	return k.evmKeeper.AddEVMExtensions(ctx, precompiles...)
}
