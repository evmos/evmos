// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v19

import (
	"slices"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/utils"
	erc20keeper "github.com/evmos/evmos/v19/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v19/x/erc20/types"
	evmkeeper "github.com/evmos/evmos/v19/x/evm/keeper"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
	stakingkeeper "github.com/evmos/evmos/v19/x/staking/keeper"
)

var newExtraEIPs = []string{"evmos_0"}

// CreateUpgradeHandler creates an SDK upgrade handler for v19
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ak authkeeper.AccountKeeper,
	bk bankkeeper.Keeper,
	sk stakingkeeper.Keeper,
	erc20k erc20keeper.Keeper,
	ek *evmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)
		// revenue module is deprecated
		logger.Debug("deleting revenue module from version map...")
		delete(vm, "revenue")

		// run module migrations first.
		// so we wont override erc20 params when running strv2 migration,
		migrationRes, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return migrationRes, err
		}

		ctxCache, writeFn := ctx.CacheContext()
		if err := ReactivateStaticPrecompiles(ctxCache, ek); err == nil {
			writeFn()
		} else {
			logger.Error("error removing outposts", "error", err)
		}

		bondDenom := sk.BondDenom(ctx)

		var wevmosContract common.Address
		switch {
		case utils.IsMainnet(ctx.ChainID()):
			wevmosContract = common.HexToAddress(erc20types.WEVMOSContractMainnet)
		case utils.IsTestnet(ctx.ChainID()):
			wevmosContract = common.HexToAddress(erc20types.WEVMOSContractTestnet)
		default:
			panic("unknown chain id")
		}

		ctxCache, writeFn = ctx.CacheContext()
		if err = RunSTRv2Migration(ctxCache, logger, ak, bk, erc20k, ek, wevmosContract, bondDenom); err == nil {
			writeFn()
		} else {
			logger.Error("error running strv2 migration", "error", err)
		}

		ctxCache, writeFn = ctx.CacheContext()
		if err := EnableCustomEIPs(ctxCache, logger, ek); err == nil {
			writeFn()
		} else {
			logger.Error("error setting new extra EIPs", "error", err)
		}

		if utils.IsTestnet(ctx.ChainID()) {
			ctxCache, writeFn = ctx.CacheContext()
			if err := ReplaceTestnetWevmosContract(ctxCache, erc20k); err == nil {
				writeFn()
			} else {
				logger.Error("error setting wevmos testnet contract", "error", err)
			}
		}

		return migrationRes, err
	}
}

// ReactivateStaticPrecompiles sets ActiveStaticPrecompiles param on the evm
func ReactivateStaticPrecompiles(ctx sdk.Context,
	evmKeeper *evmkeeper.Keeper,
) error {
	params := evmKeeper.GetParams(ctx)
	params.ActiveStaticPrecompiles = evmtypes.DefaultStaticPrecompiles
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
	pair := erc20types.NewTokenPair(wrappedContractAddr, nativeDenom, erc20types.OWNER_MODULE)
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
		*evmKeeper,
		wrappedContractAddr,
		nativeTokenPairs,
	); err != nil {
		return errorsmod.Wrap(err, "failed to convert native coins")
	}

	if err := registerERC20Extensions(ctx, wrappedContractAddr, erc20Keeper, evmKeeper); err != nil {
		return errorsmod.Wrap(err, "failed to register ERC-20 extensions")
	}

	return nil
}

// registerERC20Extensions registers the ERC20 precompiles with the EVM.
func registerERC20Extensions(ctx sdk.Context,
	wrappedContractAddr common.Address,
	erc20Keeper erc20keeper.Keeper,
	evmKeeper *evmkeeper.Keeper,
) error {
	params := erc20Keeper.GetParams(ctx)

	var err error
	erc20Keeper.IterateTokenPairs(ctx, func(tokenPair erc20types.TokenPair) bool {
		// skip registration if token is native or if it has already been registered
		// NOTE: this should handle failure during the selfdestruct
		if tokenPair.ContractOwner != erc20types.OWNER_MODULE ||
			erc20Keeper.IsAvailableERC20Precompile(&params, tokenPair.GetERC20Contract()) {
			return false
		}

		address := tokenPair.GetERC20Contract()
		if !slices.Equal(address.Bytes(), wrappedContractAddr.Bytes()) {
			// Add to existing EVM extensions - except wrappedEvmos which is on NativePrecompiles
			err = erc20Keeper.EnableDynamicPrecompiles(ctx, address)
		}

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

func EnableCustomEIPs(ctx sdk.Context, logger log.Logger, ek *evmkeeper.Keeper) error {
	params := ek.GetParams(ctx)
	extraEIPs := params.ExtraEIPs

	for _, eip := range newExtraEIPs {
		if slices.Contains(extraEIPs, eip) {
			logger.Debug("skipping duplicate EIP", "eip", eip)
		} else {
			extraEIPs = append(extraEIPs, eip)
		}
	}

	params.ExtraEIPs = extraEIPs
	return ek.SetParams(ctx, params)
}

func ReplaceTestnetWevmosContract(ctx sdk.Context, erc erc20keeper.Keeper) error {
	erc20params := erc.GetParams(ctx)
	erc20params.NativePrecompiles = []string{erc20types.WEVMOSContractTestnet}
	return erc.SetParams(ctx, erc20params)
}
