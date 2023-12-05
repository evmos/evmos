// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v16/precompiles/erc20"
	"github.com/evmos/evmos/v16/precompiles/werc20"
	"github.com/evmos/evmos/v16/utils"
	"github.com/evmos/evmos/v16/x/erc20/types"
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

// RegisterPrecompileForCoin deploys an erc20 precompile contract for an IBC voucher that is
// and creates the token pair for the existing Cosmos coin
func (k Keeper) RegisterPrecompileForCoin(
	ctx sdk.Context,
	coin sdk.Coin,
	pair types.TokenPair,
) error {
	denomAddr, err := utils.GetIBCDenomAddress(coin.Denom)
	if err != nil {
		return err
	}

	// Truncate to 20 bytes (40 hex characters)
	truncatedAddr := denomAddr[:20]
	params := k.evmKeeper.GetParams(ctx)
	found := params.IsPrecompileRegistered(common.BytesToAddress(truncatedAddr).String())
	if !found {
		// Register a new precompile address
		newPrecompile, err := erc20.NewPrecompile(pair, k.bankKeeper, k.authzKeeper, *k.transferKeeper)
		if err != nil {
			return err
		}

		err = k.evmKeeper.AddEVMExtensions(ctx, newPrecompile)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: Are we going to use this to register a token pair automatically for newly seen native vouchers ?
// RegisterTokenPairForNativeCoin creates a token pair for the native coin for a newly seen denom together with a
// precompile instance for the token pair.
func (k Keeper) RegisterTokenPairForNativeCoin(
	ctx sdk.Context,
	coinMetadata banktypes.Metadata,
) (*types.TokenPair, error) {
	// Check if denomination is already registered
	if k.IsDenomRegistered(ctx, coinMetadata.Name) {
		return nil, errorsmod.Wrapf(
			types.ErrTokenPairAlreadyExists, "coin denomination already registered: %s", coinMetadata.Name,
		)
	}

	// Check if the coin exists by ensuring the supply is set
	if !k.bankKeeper.HasSupply(ctx, coinMetadata.Base) {
		return nil, errorsmod.Wrapf(
			errortypes.ErrInvalidCoins, "base denomination '%s' cannot have a supply of 0", coinMetadata.Base,
		)
	}

	if err := k.verifyMetadata(ctx, coinMetadata); err != nil {
		return nil, errorsmod.Wrapf(
			types.ErrInternalTokenPair, "coin metadata is invalid %s", coinMetadata.Name,
		)
	}

	prefix := transfertypes.DenomPrefix + "/"
	if len(coinMetadata.Base) < len(prefix) {
		return nil, errorsmod.Wrapf(transfertypes.ErrInvalidDenomForTransfer, "invalid denom %s", coinMetadata.Base)
	}

	hexBz, err := hex.DecodeString(coinMetadata.Base[len(prefix):])
	if err != nil {
		return nil, errorsmod.Wrapf(transfertypes.ErrInvalidDenomForTransfer, "invalid hex %s", coinMetadata.Base)
	}

	addr := common.BytesToAddress(hexBz)

	pair := types.NewTokenPair(addr, coinMetadata.Base, types.OWNER_MODULE)

	precompile, err := erc20.NewPrecompile(pair, k.bankKeeper, k.authzKeeper, *k.transferKeeper)
	if err != nil {
		return nil, err
	}

	if err := k.evmKeeper.AddEVMExtensions(ctx, precompile); err != nil {
		return nil, err
	}

	k.SetTokenPair(ctx, pair)
	k.SetDenomMap(ctx, pair.Denom, pair.GetID())
	k.SetERC20Map(ctx, common.HexToAddress(pair.Erc20Address), pair.GetID())

	return &pair, nil
}
