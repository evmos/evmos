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
	erc20keeper "github.com/evmos/evmos/v18/x/erc20/keeper"
	"github.com/evmos/evmos/v18/x/erc20/types"
	evmkeeper "github.com/evmos/evmos/v18/x/evm/keeper"
)

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
		wrappedContractAddr,
		nativeTokenPairs,
	); err != nil {
		return errorsmod.Wrap(err, "failed to convert native coins")
	}

	// Register the ERC-20 extensions for the native token pairs and delete the old contract code.
	return RegisterERC20Extensions(
		ctx, erc20Keeper, evmKeeper,
	)
}
