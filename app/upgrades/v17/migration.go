// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v17

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	erc20keeper "github.com/evmos/evmos/v16/x/erc20/keeper"
)

// ConvertToNativeCoinExtensions converts all the registered ERC20 tokens of Cosmos native token pairs
// back to the native representation and registers the (W)ERC20 precompiles for each token pair.
func ConvertToNativeCoinExtensions(
	ctx sdk.Context,
	logger log.Logger,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
	wrappedContractAddr common.Address,
) error {
	// NOTE (@fedekunze): first we must convert the all the registered tokens.
	// If we do it the other way around, the conversion will fail since there won't
	// be any contract code due to the selfdestruct.
	if err := ConvertERC20Coins(ctx, logger, accountKeeper, bankKeeper, erc20Keeper, wrappedContractAddr); err != nil {
		return errorsmod.Wrap(err, "failed to convert native coins")
	}

	// Instantiate the (W)ERC20 Precompile for each registered IBC Coin

	// IMPORTANT (@fedekunze): This logic needs to be included on EVERY UPGRADE
	// from now on because the AvailablePrecompiles function does not have access
	// to the state (in this case, the registered token pairs).
	if err := erc20Keeper.RegisterERC20Extensions(ctx); err != nil {
		return errorsmod.Wrap(err, "failed to register ERC-20 extensions")
	}

	return nil
}
