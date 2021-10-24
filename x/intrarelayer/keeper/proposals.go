package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

// RegisterTokenPair registers token pair by coin denom and ERC20 contract
// address. This function fails if the mapping ERC20 <--> cosmos coin already exists.
func (k Keeper) RegisterTokenPair(ctx sdk.Context, pair types.TokenPair) error {
	params := k.GetParams(ctx)
	if !params.EnableIntrarelayer {
		return sdkerrors.Wrap(types.ErrInternalTokenPair, "intrarelaying is currently disabled by governance")
	}

	erc20 := pair.GetERC20Contract()
	if k.IsERC20Registered(ctx, erc20) {
		return sdkerrors.Wrapf(types.ErrInternalTokenPair, "token ERC20 contract already registered: %s", pair.Erc20Address)
	}

	if k.IsDenomRegistered(ctx, pair.Denom) {
		return sdkerrors.Wrapf(types.ErrInternalTokenPair, "coin denomination already registered: %s", pair.Denom)
	}

	// create metadata if not already stored
	if err := k.CreateMetadata(ctx, pair); err != nil {
		return sdkerrors.Wrap(err, "failed to create wrapped coin denom metadata for ERC20")
	}

	k.SetTokenPair(ctx, pair)
	return nil
}

func (k Keeper) CreateMetadata(ctx sdk.Context, pair types.TokenPair) error {
	// TODO: replace for HasDenomMetaData once available
	_, found := k.bankKeeper.GetDenomMetaData(ctx, pair.Denom)
	if found {
		// metadata already exists; exit
		// TODO: validate that the fields from the ERC20 match the denom metadata's
		return sdkerrors.Wrapf(types.ErrInternalTokenPair, "coin denomination already registered")
	}

	// if cosmos denom doesn't exist
	// TODO: query the contract and supply

	// TODO: retrieve the vars below from the ERC20 contract ABI
	symbol := "ERC"
	decimals := uint32(18)
	token := "erc20"

	// create a bank denom metadata based on the ERC20 token ABI details
	metadata := banktypes.Metadata{
		Description: fmt.Sprintf("Cosmos coin token representation of %s", pair.Erc20Address),
		Base:        pair.Denom,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    pair.Denom,
				Exponent: 0,
			},
			{
				Denom:    token,
				Exponent: decimals,
			},
		},
		Name:    pair.Erc20Address,
		Symbol:  symbol,
		Display: token,
	}

	if err := metadata.Validate(); err != nil {
		return sdkerrors.Wrapf(err, "ERC20 token data is invalid for contract %s", pair.Erc20Address)
	}

	k.bankKeeper.SetDenomMetaData(ctx, metadata)
	return nil
}

// EnableRelay enables relaying for a given token pair
func (k Keeper) EnableRelay(ctx sdk.Context, token string) (types.TokenPair, error) {
	id := k.GetTokenPairID(ctx, token)

	if len(id) == 0 {
		return types.TokenPair{}, sdkerrors.Wrapf(types.ErrInternalTokenPair, "token %s not registered", token)
	}

	pair, found := k.GetTokenPair(ctx, id)
	if !found {
		return types.TokenPair{}, sdkerrors.Wrapf(types.ErrInternalTokenPair, "not registered")
	}

	pair.Enabled = true

	k.SetTokenPair(ctx, pair)
	return pair, nil
}

// UpdateTokenPairERC20 updates the ERC20 token address for the registered token pair
func (k Keeper) UpdateTokenPairERC20(ctx sdk.Context, erc20Addr, newERC20Addr common.Address) (types.TokenPair, error) {
	id := k.GetERC20Map(ctx, erc20Addr)
	if len(id) == 0 {
		return types.TokenPair{}, sdkerrors.Wrapf(types.ErrInternalTokenPair, "token %s not registered", erc20Addr)
	}

	pair, found := k.GetTokenPair(ctx, id)
	if !found {
		return types.TokenPair{}, sdkerrors.Wrapf(types.ErrInternalTokenPair, "not registered")
	}

	pair.Erc20Address = newERC20Addr.Hex()
	k.SetTokenPair(ctx, pair)
	return pair, nil
}
