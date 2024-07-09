// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v18/precompiles/erc20"
	"github.com/evmos/evmos/v18/precompiles/werc20"
	"github.com/evmos/evmos/v18/x/erc20/types"
	"github.com/evmos/evmos/v18/x/evm/core/vm"
)

// RegisterERC20Extensions registers the ERC20 precompiles with the EVM.
func (k Keeper) RegisterERC20Extensions(ctx sdk.Context) error {
	precompiles := make([]vm.PrecompiledContract, 0)
	params := k.evmKeeper.GetParams(ctx)
	evmDenom := params.EvmDenom

	var err error
	k.IterateTokenPairs(ctx, func(tokenPair types.TokenPair) bool {
		// skip registration if token is native or if it has already been registered
		// NOTE: this should handle failure during the selfdestruct
		if tokenPair.ContractOwner != types.OWNER_MODULE ||
			k.evmKeeper.IsAvailablePrecompile(tokenPair.GetERC20Contract()) {
			return false
		}

		var precompile vm.PrecompiledContract

		if tokenPair.Denom == evmDenom {
			precompile, err = werc20.NewPrecompile(tokenPair, k.bankKeeper, k.authzKeeper, *k.transferKeeper)
		} else {
			precompile, err = erc20.NewPrecompile(tokenPair, k.bankKeeper, k.authzKeeper, *k.transferKeeper)
		}

		if err != nil {
			err = errorsmod.Wrapf(err, "failed to instantiate ERC-20 precompile for denom %s", tokenPair.Denom)
			return true
		}

		address := tokenPair.GetERC20Contract()

		// try selfdestruct ERC20 contract

		// NOTE(@fedekunze): From now on, the contract address will map to a precompile instead
		// of the ERC20MinterBurner contract. We try to force a selfdestruct to remove the unnecessary
		// code and storage from the state machine. In any case, the precompiles are handled in the EVM
		// before the regular contracts so not removing them doesn't create any issues in the implementation.
		err = k.evmKeeper.DeleteAccount(ctx, address)
		if err != nil {
			err = errorsmod.Wrapf(err, "failed to selfdestruct account %s", address)
			return true
		}

		precompiles = append(precompiles, precompile)
		return false
	})

	if err != nil {
		return err
	}

	// add the ERC20s to the EVM active and available precompiles
	return k.evmKeeper.AddEVMExtensions(ctx, precompiles...)
}
