package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (k Keeper) AutoRegisterCoin(ctx sdk.Context, traceDenom, baseDenom string) error {
	metadata := banktypes.Metadata{
		Description: "auto registered ERC20 for IBC token " + traceDenom,
		DenomUnits:  []*banktypes.DenomUnit{},
		Base:        traceDenom,
		//FIXME: need a define a better way to get the display name and the decimals
		Display: baseDenom,
		Name:    "ERC20 of " + baseDenom,
		Symbol:  baseDenom,
	}

	pair, err := k.RegisterCoin(ctx, metadata)
	ctx.Logger().Info("registerd coin", "metadata", metadata, "erc20addr", pair.Erc20Address, "denom", pair.Denom)
	return err
}
