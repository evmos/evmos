// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v19

import (
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
		}

		MigrateEthAccountsToBaseAccounts(ctx, ak, ek)

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
