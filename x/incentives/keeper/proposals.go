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
	contract common.Address,
	allocations sdk.DecCoins,
	epochs uint32,
) (*types.Incentive, error) {

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
	mintDenom := k.mintKeeper.GetParams(ctx).MintDenom
	for _, al := range allocations {
		if al.Denom != mintDenom {
			if !k.bankKeeper.HasSupply(ctx, al.Denom) {
				return nil, sdkerrors.Wrapf(
					sdkerrors.ErrInvalidCoins,
					"base denomination '%s' cannot have a supply of 0", al.Denom,
				)
			}
		}

		// check if each allocation is below the allocation limit
		if al.Amount.GT(params.AllocationLimit) {
			return nil, sdkerrors.Wrapf(
				types.ErrInternalIncentive,
				"allocation for denom '%s' (%s) cannot be above allocation limit (%s)", al.Denom, al.Amount, params.AllocationLimit,
			)
		}
	}

	// Iterate over allocations to update allocation meters
	allocationMeters := []sdk.DecCoin{}
	for _, al := range allocations {
		allocationMeter, _ := k.GetAllocationMeter(ctx, al.Denom)
		// check if the sum of all allocations (current + proposed) exceeds 100%
		allocationSum := allocationMeter.Amount.Add(al.Amount)
		if allocationSum.GT(sdk.OneDec()) {
			return nil, sdkerrors.Wrapf(
				types.ErrInternalIncentive,
				"allocation for denom %s is lager than 100 percent: %v",
				al.Denom, allocationSum,
			)
		}

		// build new allocation meter
		newAllocationMeter := sdk.DecCoin{
			Denom:  al.Denom,
			Amount: allocationSum,
		}
		allocationMeters = append(allocationMeters, newAllocationMeter)
	}

	// create incentive and set to store
	incentive := types.NewIncentive(contract, allocations, epochs)
	k.SetIncentive(ctx, incentive)

	// Update allocation meters
	for _, am := range allocationMeters {
		k.SetAllocationMeter(ctx, am)
	}

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
			"unmatching contract '%s' ", contract,
		)
	}

	k.DeleteIncentiveAndUpdateAllocationMeters(ctx, incentive)

	return nil
}
