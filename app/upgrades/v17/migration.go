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
	"github.com/evmos/evmos/v16/x/erc20/types"
	evmkeeper "github.com/evmos/evmos/v16/x/evm/keeper"
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
		*evmKeeper,
		wrappedContractAddr,
		nativeTokenPairs,
	); err != nil {
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

// LogTokenPairBalances logs the total balances of each token pair.
func LogTokenPairBalances(
	ctx sdk.Context,
	logger log.Logger,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
	// evmKeeper evmkeeper.Keeper,
) error {
	//chainID, err := strconv.Atoi(ctx.ChainID())
	//if err != nil {
	//	return errorsmod.Wrap(err, "failed to convert chainID to int")
	//}

	tokenPairs := erc20Keeper.GetTokenPairs(ctx)
	for _, tokenPair := range tokenPairs {
		//// Log the total balance of the token pair
		//ret, err := evmKeeper.EthCall(sdk.WrapSDKContext(ctx), &evmtypes.EthCallRequest{
		//	ChainId: int64(chainID),
		//})
		//if err != nil {
		//	logger.Error(
		//		fmt.Sprintf("failed to get total supply for token pair %q", tokenPair.Denom),
		//		"error",
		//		err.Error(),
		//	)
		//}

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
