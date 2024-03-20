// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v17

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	erc20keeper "github.com/evmos/evmos/v16/x/erc20/keeper"
	"github.com/evmos/evmos/v16/x/erc20/types"
	evmkeeper "github.com/evmos/evmos/v16/x/evm/keeper"
)

// RegisterERC20Extensions registers the ERC20 precompiles with the EVM.
func RegisterERC20Extensions(
	ctx sdk.Context,
	erc20Keeper erc20keeper.Keeper,
	evmKeeper *evmkeeper.Keeper,
) error {
	precompiles := make([]string, 0)

	var err error
	erc20Keeper.IterateTokenPairs(ctx, func(tokenPair types.TokenPair) bool {
		// skip registration if token is native or if it has already been registered
		// NOTE: this should handle failure during the selfdestruct
		if !tokenPair.IsNativeCoin() ||
			evmKeeper.IsAvailableDynamicPrecompile(ctx, tokenPair.GetErc20Address()) {
			return false
		}

		address := tokenPair.GetERC20Contract()

		// try to self-destruct the old ERC20 contract
		// NOTE(@fedekunze): From now on, the contract address will map to a precompile instead
		// of the ERC20MinterBurner contract. We try to force a self-destruction to remove the unnecessary
		// code and storage from the state machine.
		// In any case, the precompiles are handled in the EVM
		// before the regular contracts so not removing them doesn't create any issues in the implementation.
		err = evmKeeper.DeleteAccount(ctx, address)
		if err != nil {
			err = errorsmod.Wrapf(err, "failed to selfdestruct account %s", address)
			return true
		}

		precompiles = append(precompiles, address.String())
		return false
	})

	if err != nil {
		return err
	}

	// add the ERC20s to the EVM active and available precompiles
	return evmKeeper.EnableDynamicPrecompiles(ctx, precompiles...)
}
