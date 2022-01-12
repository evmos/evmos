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
				"allocation for denom '%s' (%s) cannot be above allocation limmit '%s' - ", al.Denom, al.Amount, params.AllocationLimit,
			)
		}
	}

	// check if the sum of all allocations for each denom (current + proposed) is
	// < 100%
	incentives := k.GetAllIncentives(ctx)
	if len(incentives) != 0 {

		for _, al := range allocations {
			allocationMeter, ok := k.GetAllocationMeter(ctx, al.Denom)
			// skip if no current allocation exists
			if !ok {
				continue
			}

			allocationSum := allocationMeter.Add(al.Amount)
			if allocationSum.GT(sdk.OneDec()) {
				return nil, sdkerrors.Wrapf(
					types.ErrInternalIncentive,
					"Allocation for denom %s is lager than 100 percent: %v",
					al.Denom, allocationSum,
				)
			}
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
			"unmatching contract '%s' ", contract,
		)
	}

	k.DeleteIncentive(ctx, incentive)
	return nil
}
