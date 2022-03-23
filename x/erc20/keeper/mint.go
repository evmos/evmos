package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/tharsis/evmos/v3/x/erc20/types"
)

// MintingEnabled checks that:
//  - the global parameter for intrarelaying is enabled
//  - minting is enabled for the given (erc20,coin) token pair
//  - recipient address is not on the blocked list
//  - bank module transfers are enabled for the Cosmos coin
func (k Keeper) MintingEnabled(ctx sdk.Context, sender, receiver sdk.AccAddress, token string) (types.TokenPair, error) {
	params := k.GetParams(ctx)
	if !params.EnableErc20 {
		return types.TokenPair{}, sdkerrors.Wrap(types.ErrERC20Disabled, "module is currently disabled by governance")
	}

	id := k.GetTokenPairID(ctx, token)
	if len(id) == 0 {
		return types.TokenPair{}, sdkerrors.Wrapf(types.ErrTokenPairNotFound, "token '%s' not registered by id", token)
	}

	pair, found := k.GetTokenPair(ctx, id)
	if !found {
		return types.TokenPair{}, sdkerrors.Wrapf(types.ErrTokenPairNotFound, "token '%s' not registered", token)
	}

	if !pair.Enabled {
		return types.TokenPair{}, sdkerrors.Wrapf(types.ErrERC20Disabled, "minting token '%s' is not enabled by governance", token)
	}

	if k.bankKeeper.BlockedAddr(receiver.Bytes()) {
		return types.TokenPair{}, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive transactions", receiver)
	}

	// NOTE: ignore amount as only denom is checked on IsSendEnabledCoin
	coin := sdk.Coin{Denom: pair.Denom}

	// check if minting to a recipient address other than the sender is enabled for for the given coin denom
	if !sender.Equals(receiver) && !k.bankKeeper.IsSendEnabledCoin(ctx, coin) {
		return types.TokenPair{}, sdkerrors.Wrapf(banktypes.ErrSendDisabled, "minting '%s' coins to an external address is currently disabled", token)
	}

	return pair, nil
}
