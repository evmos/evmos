// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v19

import (
	"fmt"
	"slices"

	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/ethereum/go-ethereum/common"
	evmostypes "github.com/evmos/evmos/v18/types"
	evmkeeper "github.com/evmos/evmos/v18/x/evm/keeper"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
)

const (
	StrideOutpostAddress  = "0x0000000000000000000000000000000000000900"
	OsmosisOutpostAddress = "0x0000000000000000000000000000000000000901"
)

var newExtraEIPs = []int64{0o000, 0o001, 0o002}

// CreateUpgradeHandler creates an SDK upgrade handler for v19
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ak authkeeper.AccountKeeper,
	ek *evmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)
		// revenue module is deprecated
		logger.Debug("deleting revenue module from version map...")
		delete(vm, "revenue")

		ctxCache, writeFn := ctx.CacheContext()
		if err := RemoveOutpostsFromEvmParams(ctxCache, ek); err == nil {
			writeFn()
		} else {
			logger.Debug("error removing outposts")
		}

		MigrateEthAccountsToBaseAccounts(ctx, ak, ek)

		ctxCache, writeFn = ctx.CacheContext()
		if err := EnableCustomEIPs(ctxCache, logger, ek); err == nil {
			writeFn()
		} else {
			logger.Debug("error setting new extra EIPs")
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
	for _, precompile := range params.ActivePrecompiles {
		if precompile != OsmosisOutpostAddress &&
			precompile != StrideOutpostAddress {
			newActivePrecompiles = append(newActivePrecompiles, precompile)
		}
	}
	params.ActivePrecompiles = newActivePrecompiles
	return evmKeeper.SetParams(ctx, params)
}

// MigrateEthAccountsToBaseAccounts is used to store the code hash of the associated
// smart contracts in the dedicated store in the EVM module and convert the former
// EthAccounts to standard Cosmos SDK accounts.
func MigrateEthAccountsToBaseAccounts(ctx sdk.Context, ak authkeeper.AccountKeeper, ek *evmkeeper.Keeper) {
	ak.IterateAccounts(ctx, func(account authtypes.AccountI) (stop bool) {
		ethAcc, ok := account.(*evmostypes.EthAccount)
		if !ok {
			return false
		}

		// NOTE: we only need to add store entries for smart contracts
		codeHashBytes := common.HexToHash(ethAcc.CodeHash).Bytes()
		if !evmtypes.IsEmptyCodeHash(codeHashBytes) {
			ek.SetCodeHash(ctx, ethAcc.EthAddress().Bytes(), codeHashBytes)
		}

		// Set the base account in the account keeper instead of the EthAccount
		ak.SetAccount(ctx, ethAcc.BaseAccount)

		return false
	})
}

func EnableCustomEIPs(ctx sdk.Context, logger log.Logger, ek *evmkeeper.Keeper) error {
	params := ek.GetParams(ctx)
	extraEIPs := params.ExtraEIPs

	for _, eip := range newExtraEIPs {
		if slices.Contains(extraEIPs, eip) {
			logger.Debug(fmt.Sprintf("Skipping EIP %d because duplicate", eip))
		} else {
			extraEIPs = append(extraEIPs, eip)
		}
	}

	params.ExtraEIPs = extraEIPs
	return ek.SetParams(ctx, params)
}
