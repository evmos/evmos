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
	params := k.evmKeeper.GetParams(ctx)
	evmDenom := params.EvmDenom

	k.IterateTokenPairs(ctx, func(tokenPair types.TokenPair) bool {
		// skip registration if token is native or if it has already been registered
		if tokenPair.ContractOwner != types.OWNER_MODULE ||
			params.IsPrecompileRegistered(tokenPair.Erc20Address) {
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

		address := tokenPair.GetERC20Contract()

		// selfdestruct ERC20 contract
		if err := k.evmKeeper.DeleteAccount(ctx, address); err != nil {
			panic(err)
		}

		precompiles = append(precompiles, precompile)
		return false
	})

	// add the ERC20s to the EVM active and available precompiles
	return k.evmKeeper.AddEVMExtensions(ctx, precompiles...)
}
