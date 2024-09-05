// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v192

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/cometbft/cometbft/libs/log"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	erc20keeper "github.com/evmos/evmos/v19/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v19/x/erc20/types"
	evmkeeper "github.com/evmos/evmos/v19/x/evm/keeper"
	"github.com/evmos/evmos/v19/x/evm/statedb"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v19.2
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	erc20k erc20keeper.Keeper,
	ek *evmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		if err := AddCodeToERC20Extensions(ctx, logger, erc20k, ek); err == nil {
			return nil, err
		}

		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// AddCodeToERC20Extensions adds code and code hash to the ERC20 precompiles with the EVM.
func AddCodeToERC20Extensions(
	ctx sdk.Context,
	logger log.Logger,
	erc20Keeper erc20keeper.Keeper,
	evmKeeper *evmkeeper.Keeper,
) (err error) {
	logger.Info("Adding code to erc20 extensions...")
	var (
		// bytecode and codeHash is the same for all IBC coins
		// cause they're all using the same contract
		bytecode = common.FromHex(erc20Bytecode)
		codeHash = crypto.Keccak256(bytecode)
	)

	erc20Keeper.IterateTokenPairs(ctx, func(tokenPair erc20types.TokenPair) bool {
		// Only need to add code to the IBC coins.
		if tokenPair.ContractOwner != erc20types.OWNER_MODULE {
			return false
		}

		contractAddr := common.HexToAddress(tokenPair.Erc20Address)
		// check if code was already stored
		code := evmKeeper.GetCode(ctx, common.Hash(codeHash))
		if len(code) == 0 {
			evmKeeper.SetCode(ctx, codeHash, bytecode)
		}

		var (
			nonce   uint64
			balance = common.Big0
		)
		// keep balance and nonce if account exists
		if acc := evmKeeper.GetAccount(ctx, contractAddr); acc != nil {
			nonce = acc.Nonce
			balance = acc.Balance
		}

		err = evmKeeper.SetAccount(ctx, contractAddr, statedb.Account{
			CodeHash: codeHash,
			Nonce:    nonce,
			Balance:  balance,
		})

		return err != nil
	})

	logger.Info("Done with erc20 extensions")
	return err
}
