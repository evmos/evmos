// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/evmos/evmos/v19/x/erc20/types"
)

// MintingEnabled checks that:
//   - the global parameter for erc20 conversion is enabled
//   - minting is enabled for the given (erc20,coin) token pair
//   - recipient address is not on the blocked list
//   - bank module transfers are enabled for the Cosmos coin
func (k Keeper) MintingEnabled(
	ctx sdk.Context,
	sender, receiver sdk.AccAddress,
	token string,
) (types.TokenPair, error) {
	if !k.IsERC20Enabled(ctx) {
		return types.TokenPair{}, errorsmod.Wrap(
			types.ErrERC20Disabled, "module is currently disabled by governance",
		)
	}

	id := k.GetTokenPairID(ctx, token)
	if len(id) == 0 {
		return types.TokenPair{}, errorsmod.Wrapf(
			types.ErrTokenPairNotFound, "token '%s' not registered by id", token,
		)
	}

	pair, found := k.GetTokenPair(ctx, id)
	if !found {
		return types.TokenPair{}, errorsmod.Wrapf(
			types.ErrTokenPairNotFound, "token '%s' not registered", token,
		)
	}

	if !pair.Enabled {
		return types.TokenPair{}, errorsmod.Wrapf(
			types.ErrERC20TokenPairDisabled, "minting token '%s' is not enabled by governance", token,
		)
	}

	if k.bankKeeper.BlockedAddr(receiver.Bytes()) {
		return types.TokenPair{}, errorsmod.Wrapf(
			errortypes.ErrUnauthorized, "%s is not allowed to receive transactions", receiver,
		)
	}

	// NOTE: ignore amount as only denom is checked on IsSendEnabledCoin
	coin := sdk.Coin{Denom: pair.Denom}

	// check if minting to a recipient address other than the sender is enabled
	// for for the given coin denom
	if !sender.Equals(receiver) && !k.bankKeeper.IsSendEnabledCoin(ctx, coin) {
		return types.TokenPair{}, errorsmod.Wrapf(
			banktypes.ErrSendDisabled, "minting '%s' coins to an external address is currently disabled", token,
		)
	}

	return pair, nil
}

// MintCoins mints the provided amount of coins to the given address.
func (k Keeper) MintCoins(ctx sdk.Context, sender, to sdk.AccAddress, amount math.Int, token string) error {
	ctx.Logger().Error("MintCoins", "token", token, "sender", sender.String(), "to", to.String(), "amount", amount.String())
	pair, err := k.MintingEnabled(ctx, sender, to, token)
	if err != nil {
		ctx.Logger().Error("MintingEnabled", "error", err)
		return err
	}

	if !pair.IsNativeCoin() {
		return errorsmod.Wrapf(types.ErrERC20TokenPairDisabled, "minting token '%s' is not enabled by governance", token)
	}

	contractOwnerAddr, err := sdk.AccAddressFromBech32(pair.OwnerAddress)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid owner address")
	}
	if !sender.Equals(contractOwnerAddr) {
		return errorsmod.Wrapf(authz.ErrNoAuthorizationFound, "minter is not the owner")
	}

	coins := sdk.Coins{{Denom: pair.Denom, Amount: amount}}
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, coins)
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, to, coins)
	if err != nil {
		return err
	}

	// if p.tokenPair.Denom == utils.BaseDenom {
	// 	p.SetBalanceChangeEntries(
	// 		cmn.NewBalanceChangeEntry(to, coins.AmountOf(utils.BaseDenom).BigInt(), cmn.Add))
	// }

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyAction, types.TypeMsgMint),
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amount.String()),
		),
	)

	return nil
}

// BurnCoins burns the provided amount of coins from the given address.
func (k Keeper) BurnCoins(ctx sdk.Context, sender sdk.AccAddress, amount math.Int, token string) error {
	pair, found := k.GetTokenPair(ctx, k.GetTokenPairID(ctx, token))
	if !found {
		return errorsmod.Wrapf(types.ErrTokenPairNotFound, "token '%s' not registered", token)
	}

	if !pair.IsNativeCoin() {
		return errorsmod.Wrapf(types.ErrERC20TokenPairDisabled, "burning token '%s' is not enabled by governance", token)
	}

	coins := sdk.Coins{{Denom: pair.Denom, Amount: amount}}

	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, coins)
	if err != nil {
		return err
	}

	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins)
	if err != nil {
		return err
	}

	// if p.tokenPair.Denom == utils.BaseDenom {
	// 	p.SetBalanceChangeEntries(
	// 		cmn.NewBalanceChangeEntry(burnerAddr, coins.AmountOf(utils.BaseDenom).BigInt(), cmn.Sub))
	// }

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyAction, types.TypeMsgBurn),
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amount.String()),
		),
	)
	return nil
}

// TransferOwnership transfers ownership of the token to the new owner
func (k Keeper) TransferOwnership(ctx sdk.Context, newOwner sdk.AccAddress, token string) error {
	pair, found := k.GetTokenPair(ctx, k.GetTokenPairID(ctx, token))
	if !found {
		return errorsmod.Wrapf(types.ErrTokenPairNotFound, "token '%s' not registered", token)
	}

	if !pair.IsNativeCoin() {
		return errorsmod.Wrapf(types.ErrERC20TokenPairDisabled, "transferring ownership of token '%s' is not enabled by governance", token)
	}

	k.SetTokenPairOwnerAddress(ctx, pair, newOwner.String())

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyAction, types.TypeMsgTransferOwnership),
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyNewOwner, newOwner.String()),
		),
	)

	return nil
}

func (k Keeper) GetOwnerAddress(ctx sdk.Context, contractAddress string) string {
	pair, found := k.GetTokenPair(ctx, k.GetTokenPairID(ctx, contractAddress))
	if !found {
		return ""
	}

	return pair.OwnerAddress
}
