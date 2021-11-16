package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tharsis/evmos/x/intrarelayer/types"
	"github.com/tharsis/evmos/x/intrarelayer/types/contracts"
)

// RegisterCoin deploys an erc20 contract and creates the token pair for the cosmos coin
func (k Keeper) RegisterCoin(ctx sdk.Context, coinMetadata banktypes.Metadata) (*types.TokenPair, error) {
	params := k.GetParams(ctx)
	if !params.EnableIntrarelayer {
		return nil, sdkerrors.Wrap(types.ErrInternalTokenPair, "intrarelaying is currently disabled by governance")
	}
	if k.IsDenomRegistered(ctx, coinMetadata.Name) {
		return nil, sdkerrors.Wrapf(types.ErrInternalTokenPair, "coin denomination already registered: %s", coinMetadata.Name)
	}

	addr, err := k.DeployERC20Contract(ctx, coinMetadata)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to create wrapped coin denom metadata for ERC20")
	}

	pair := types.NewTokenPair(addr, coinMetadata.Base, true, types.OWNER_MODULE)
	k.SetTokenPair(ctx, pair)
	k.SetDenomMap(ctx, pair.Denom, pair.GetID())
	k.SetERC20Map(ctx, common.HexToAddress(pair.Erc20Address), pair.GetID())

	return &pair, nil
}

// DeployERC20Contract creates and deploys an ERC20 contract on the EVM with the intrarelayer module account as owner
func (k Keeper) DeployERC20Contract(ctx sdk.Context, coinMetadata banktypes.Metadata) (common.Address, error) {
	// meta, found := k.bankKeeper.GetDenomMetaData(ctx, pair.Denom)
	// if !found {
	// 	// metadata already exists; exit
	// 	// TODO: validate that the fields from the ERC20 match the denom metadata's
	// 	return common.Address{}, sdkerrors.Wrapf(types.ErrInternalTokenPair, "coin denomination is not registered")
	// }
	// k.evmKeeper.SetNonce(types.ModuleAddress, 1)

	ctorArgs, err := contracts.ERC20BurnableAndMintableContract.ABI.Pack("", coinMetadata.Name, coinMetadata.Symbol)
	if err != nil {
		return common.Address{}, sdkerrors.Wrapf(err, "coin metadata is invalid  %s", coinMetadata.Name)
	}

	data := make([]byte, len(contracts.ERC20BurnableAndMintableContract.Bin)+len(ctorArgs))
	copy(data[:len(contracts.ERC20BurnableAndMintableContract.Bin)], contracts.ERC20BurnableAndMintableContract.Bin)
	copy(data[len(contracts.ERC20BurnableAndMintableContract.Bin):], ctorArgs)

	nonce, err := k.accountKeeper.GetSequence(ctx, types.ModuleAddress.Bytes())
	if err != nil {
		return common.Address{}, err
	}

	contractAddr := crypto.CreateAddress(types.ModuleAddress, nonce)
	_, err = k.CallEVMWithPayload(ctx, types.ModuleAddress, nil, data)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to deploy contract for %s", coinMetadata.Name)
	}

	return contractAddr, nil
}

// RegisterERC20 creates a cosmos coin and registers the token pair between the coin and the ERC20
func (k Keeper) RegisterERC20(ctx sdk.Context, contract common.Address) (*types.TokenPair, error) {
	params := k.GetParams(ctx)
	if !params.EnableIntrarelayer {
		return nil, sdkerrors.Wrap(types.ErrInternalTokenPair, "intrarelaying is currently disabled by governance")
	}

	if k.IsERC20Registered(ctx, contract) {
		return nil, sdkerrors.Wrapf(types.ErrInternalTokenPair, "token ERC20 contract already registered: %s", contract.String())
	}

	metadata, err := k.CreateCoinMetadata(ctx, contract)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to create wrapped coin denom metadata for ERC20")
	}

	pair := types.NewTokenPair(contract, metadata.Name, true, types.OWNER_EXTERNAL)
	k.SetTokenPair(ctx, pair)
	k.SetDenomMap(ctx, pair.Denom, pair.GetID())
	k.SetERC20Map(ctx, common.HexToAddress(pair.Erc20Address), pair.GetID())
	return nil, nil
}

// CreateCoinMetadata generates the metadata to represent the ERC20 token on evmos.
func (k Keeper) CreateCoinMetadata(ctx sdk.Context, contract common.Address) (*banktypes.Metadata, error) {
	strContract := contract.String()

	erc20Data, err := k.QueryERC20(ctx, contract)
	if err != nil {
		return nil, err
	}

	_, found := k.bankKeeper.GetDenomMetaData(ctx, types.CreateDenom(strContract))
	if found {
		// metadata already exists; exit
		// TODO: validate that the fields from the ERC20 match the denom metadata's
		return nil, sdkerrors.Wrapf(types.ErrInternalTokenPair, "coin denomination already registered")
	}

	if k.IsDenomRegistered(ctx, types.CreateDenom(strContract)) {
		return nil, sdkerrors.Wrapf(types.ErrInternalTokenPair, "coin denomination already registered: %s", erc20Data.Name)
	}

	// create a bank denom metadata based on the ERC20 token ABI details
	metadata := banktypes.Metadata{
		Description: types.CreateDenomDescription(strContract),
		Base:        types.CreateDenom(strContract),
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    types.CreateDenom(strContract),
				Exponent: 0,
			},
			{
				Denom:    erc20Data.Name,
				Exponent: uint32(erc20Data.Decimals),
			},
		},
		Name:    types.CreateDenom(strContract),
		Symbol:  erc20Data.Symbol,
		Display: erc20Data.Name,
	}

	if err := metadata.Validate(); err != nil {
		return nil, sdkerrors.Wrapf(err, "ERC20 token data is invalid for contract %s", strContract)
	}

	k.bankKeeper.SetDenomMetaData(ctx, metadata)

	return &metadata, nil
}

// ToggleRelay toggles relaying for a given token pair
func (k Keeper) ToggleRelay(ctx sdk.Context, token string) (types.TokenPair, error) {
	id := k.GetTokenPairID(ctx, token)

	if len(id) == 0 {
		return types.TokenPair{}, sdkerrors.Wrapf(types.ErrInternalTokenPair, "token %s not registered", token)
	}

	pair, found := k.GetTokenPair(ctx, id)
	if !found {
		return types.TokenPair{}, sdkerrors.Wrapf(types.ErrInternalTokenPair, "not registered")
	}

	pair.Enabled = !pair.Enabled

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
		metadata.Description != types.CreateDenomDescription(erc20Addr.String()) {
		return types.TokenPair{}, sdkerrors.Wrapf(types.ErrInternalTokenPair, "invalid metadata for %s ", pair.Erc20Address)
	}

	// Update the metadata description with the new address
	metadata.Description = types.CreateDenomDescription(newERC20Addr.String())
	k.bankKeeper.SetDenomMetaData(ctx, metadata)
	// Delete old token pair (id is changed because the address was modifed)
	k.DeleteTokenPair(ctx, pair)
	// Update the address
	pair.Erc20Address = newERC20Addr.Hex()
	// Set the new pair
	k.SetTokenPair(ctx, pair)
	// Overwrite the value because id was changed
	k.SetDenomMap(ctx, pair.Denom, pair.GetID())
	// Remove old address
	k.DeleteERC20Map(ctx, erc20Addr)
	// Add the new address
	k.SetERC20Map(ctx, common.HexToAddress(pair.Erc20Address), pair.GetID())
	return pair, nil
}
