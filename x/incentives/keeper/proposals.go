package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tharsis/evmos/x/incentives/types"
)

// RegisterIncentive creates an incentive for a contract
func (k Keeper) RegisterIncentive(
	ctx sdk.Context,
	allocations sdk.DecCoins,
	contract common.Address,
	epochs uint32,
) (*types.Incentive, error) {
	// TODO check if sum of all active incentived contracts' allocation is < 100%

	// check if the Incentives are globally enabled
	params := k.GetParams(ctx)
	if !params.EnableIncentives {
		return nil, sdkerrors.Wrap(
			types.ErrInternalIncentive,
			"incentives are currently disabled by governance",
		)
	}

	// check if the incentive is already registered
	if k.IsIncentiveRegistered(ctx, contract) {
		return nil, sdkerrors.Wrapf(
			types.ErrInternalIncentive,
			"incentive already registered: %s", contract,
		)
	}

	// check if the balance is > 0 for coins other than the mint denomination
	for _, al := range allocations {
		// TODO: Skip if al.Denom == the mint denomination
		if !k.bankKeeper.HasSupply(ctx, al.Denom) {
			return nil, sdkerrors.Wrapf(
				sdkerrors.ErrInvalidCoins,
				"base denomination '%s' cannot have a supply of 0", al.Denom,
			)
		}
	}

	// create incentive and set to store
	incentive := types.NewIncentive(contract, allocations, epochs)
	k.SetIncentive(ctx, incentive)

	return &incentive, nil
}

// RegisterIncentive deletes the incentive for a contract
func (k Keeper) CancelIncentive(
	ctx sdk.Context,
	contract common.Address,
) error {
	incentive, found := k.GetIncentive(ctx, contract)
	if !found {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"unmatching contract '%s' cannot have a supply of 0", contract,
		)
	}

	k.DeleteIncentive(ctx, incentive)
	return nil
}
