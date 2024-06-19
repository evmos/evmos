// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v19

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/ethereum/go-ethereum/common"
	erc20keeper "github.com/evmos/evmos/v18/x/erc20/keeper"
	"github.com/evmos/evmos/v18/x/erc20/types"
	evmkeeper "github.com/evmos/evmos/v18/x/evm/keeper"
)

const (
	StrideOutpostAddress  = "0x0000000000000000000000000000000000000900"
	OsmosisOutpostAddress = "0x0000000000000000000000000000000000000901"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v19
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	evmKeeper *evmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)
		// revenue module is deprecated
		logger.Debug("deleting revenue module from version map...")
		delete(vm, "revenue")

		ctxCache, writeFn := ctx.CacheContext()
		if err := RemoveOutpostsFromEvmParams(ctxCache, evmKeeper); err == nil {
			writeFn()
		}

		// Leave modules as-is to avoid running InitGenesis.
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func RemoveOutpostsFromEvmParams(ctx sdk.Context,
	evmKeeper *evmkeeper.Keeper,
) error {
	params := evmKeeper.GetParams(ctx)
	newActivePrecompiles := make([]string, 0)
	for _, precompile := range params.ActiveStaticPrecompiles {
		if precompile != OsmosisOutpostAddress &&
			precompile != StrideOutpostAddress {
			newActivePrecompiles = append(newActivePrecompiles, precompile)
		}
	}
	params.ActiveStaticPrecompiles = newActivePrecompiles
	return evmKeeper.SetParams(ctx, params)
}

// RunSTRv2Migration converts all the registered ERC-20 tokens of Cosmos native token pairs
// back to the native representation and registers the WEVMOS token as an ERC-20 token pair.
func RunSTRv2Migration(
	ctx sdk.Context,
	logger log.Logger,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
	evmKeeper *evmkeeper.Keeper,
	wrappedContractAddr common.Address,
	nativeDenom string,
) error {
	// NOTE: it's necessary to register the WEVMOS token as a native token pair before adding
	// the dynamic EVM extensions (which is relying on the registered token pairs).
	pair := types.NewTokenPair(wrappedContractAddr, nativeDenom, types.OWNER_MODULE)
	erc20Keeper.SetToken(ctx, pair)

	// Filter all token pairs for the ones that are for Cosmos native coins.
	nativeTokenPairs := getNativeTokenPairs(ctx, erc20Keeper)

	// NOTE (@fedekunze): first we must convert the all the registered tokens.
	// If we do it the other way around, the conversion will fail since there won't
	// be any contract code due to the selfdestruct.
	if err := ConvertERC20Coins(
		ctx,
		logger,
		accountKeeper,
		bankKeeper,
		erc20Keeper,
		*evmKeeper,
		wrappedContractAddr,
		nativeTokenPairs,
	); err != nil {
		return errorsmod.Wrap(err, "failed to convert native coins")
	}

	if err := RegisterERC20Extensions(ctx, erc20Keeper, evmKeeper); err != nil {
		return errorsmod.Wrap(err, "failed to register ERC-20 extensions")
	}

	return nil
}

// RegisterERC20Extensions registers the ERC20 precompiles with the EVM.
func RegisterERC20Extensions(ctx sdk.Context,
	erc20Keeper erc20keeper.Keeper,
	evmKeeper *evmkeeper.Keeper,
) error {
	params := erc20Keeper.GetParams(ctx)

	var err error
	erc20Keeper.IterateTokenPairs(ctx, func(tokenPair types.TokenPair) bool {
		// skip registration if token is native or if it has already been registered
		// NOTE: this should handle failure during the selfdestruct
		if tokenPair.ContractOwner != types.OWNER_MODULE ||
			erc20Keeper.IsAvailableERC20Precompile(&params, tokenPair.GetERC20Contract()) {
			return false
		}

		address := tokenPair.GetERC20Contract()
		// Add to existing EVM extensions
		err = erc20Keeper.EnableDynamicPrecompiles(ctx, address)
		if err != nil {
			return true
		}

		// try selfdestruct ERC20 contract

		// NOTE(@fedekunze): From now on, the contract address will map to a precompile instead
		// of the ERC20MinterBurner contract. We try to force a selfdestruct to remove the unnecessary
		// code and storage from the state machine. In any case, the precompiles are handled in the EVM
		// before the regular contracts so not removing them doesn't create any issues in the implementation.
		err = evmKeeper.DeleteAccount(ctx, address)
		if err != nil {
			err = errorsmod.Wrapf(err, "failed to selfdestruct account %s", address)
			return true
		}

		return false
	})

	return err
}

// LogTokenPairBalances logs the total balances of each token pair.
func LogTokenPairBalances(
	ctx sdk.Context,
	logger log.Logger,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
) error {
	tokenPairs := erc20Keeper.GetTokenPairs(ctx)
	for _, tokenPair := range tokenPairs {
		bankSupply := bankKeeper.GetSupply(ctx, tokenPair.Denom)

		logger.Info(
			"token pair balances",
			"token_pair", tokenPair.Denom,
			//// TODO: add ERC-20 supply by calling EthCall
			//"erc20 supply", totalSupply,
			"bank supply", bankSupply.Amount.String(),
		)
	}

	return nil
}
