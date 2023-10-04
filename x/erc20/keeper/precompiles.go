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
	logger := ctx.Logger()

	k.IterateTokenPairs(ctx, func(tokenPair types.TokenPair) bool {
		// skip registration if token is native or if it has already been registered
		// NOTE: this should handle failure during the selfdestruct
		if tokenPair.ContractOwner != types.OWNER_MODULE ||
			params.IsPrecompileRegistered(tokenPair.Erc20Address) {
			return false
		}

		var (
			err        error
			precompile vm.PrecompiledContract
		)

		if tokenPair.Denom == evmDenom {
			precompile, err = werc20.NewPrecompile(tokenPair, k.bankKeeper, k.authzKeeper)
		} else {
			precompile, err = erc20.NewPrecompile(tokenPair, k.bankKeeper, k.authzKeeper)
		}

		if err != nil {
			panic(fmt.Errorf("failed to instantiate ERC-20 precompile for denom %s: %w", tokenPair.Denom, err))
		}

		address := tokenPair.GetERC20Contract()

		// try selfdestruct ERC20 contract

		// NOTE(@fedekunze): From now on, the contract address will map to a precompile instead
		// of the ERC20MinterBurner contract. We try to force a selfdestruct to remove the unnecessary
		// code and storage from the state machine. In any case, the precompiles are handled in the EVM
		// before the regular contracts so not removing them doesn't create any issues in the implementation.
		if err := k.evmKeeper.DeleteAccount(ctx, address); err != nil {
			logger.Debug("failed to selfdestruct account", "error", err)
		}

		precompiles = append(precompiles, precompile)
		return false
	})

	// add the ERC20s to the EVM active and available precompiles
	return k.evmKeeper.AddEVMExtensions(ctx, precompiles...)
}
