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

func CreateDenomDescription(address string) string {
	return fmt.Sprintf("Cosmos coin token representation of %s", address)
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
	contract := pair.GetERC20Contract()

	erc20Data, err := k.QueryERC20(ctx, contract)

	if err != nil {
		return err
	}

	// create a bank denom metadata based on the ERC20 token ABI details
	metadata := banktypes.Metadata{
		Description: CreateDenomDescription(pair.Erc20Address),
		Base:        pair.Denom,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    pair.Denom,
				Exponent: 0,
			},
			{
				Denom:    erc20Data.Name,
				Exponent: uint32(erc20Data.Decimals),
			},
		},
		Name:    pair.Erc20Address,
		Symbol:  erc20Data.Symbol,
		Display: erc20Data.Name,
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

	// Get current stored metadata
	metadata, found := k.bankKeeper.GetDenomMetaData(ctx, pair.Denom)
	if !found {
		return types.TokenPair{}, sdkerrors.Wrapf(types.ErrInternalTokenPair, "could not get metadata for %s", pair.Denom)

	}
	// Get new erc20 values
	erc20Data, err := k.QueryERC20(ctx, newERC20Addr)
	if err != nil {
		return types.TokenPair{}, sdkerrors.Wrapf(types.ErrInternalTokenPair, "could not get token %s erc20Data", newERC20Addr.String())
	}
	// Compare
	if len(metadata.DenomUnits) != 2 {
		return types.TokenPair{}, sdkerrors.Wrapf(types.ErrInternalTokenPair, "invalid metadata for %s ", pair.Erc20Address)
	}

	if metadata.Display != erc20Data.Name ||
		metadata.Symbol != erc20Data.Symbol ||
		metadata.DenomUnits[1].Denom != erc20Data.Name ||
		metadata.DenomUnits[1].Exponent != uint32(erc20Data.Decimals) ||
		metadata.Description != CreateDenomDescription(erc20Addr.String()) {
		return types.TokenPair{}, sdkerrors.Wrapf(types.ErrInternalTokenPair, "invalid metadata for %s ", pair.Erc20Address)
	}

	// Update the metadata description with the new address
	metadata.Description = CreateDenomDescription(newERC20Addr.String())
	k.bankKeeper.SetDenomMetaData(ctx, metadata)
	pair.Erc20Address = newERC20Addr.Hex()
	k.SetTokenPair(ctx, pair)
	return pair, nil
}
