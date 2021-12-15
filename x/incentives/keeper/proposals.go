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
	contract string,
	epochs uint32,
) (*types.Incentive, error) {
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
	incentive := types.NewIncentive(common.HexToAddress(contract), allocations, epochs)
	k.SetIncentive(ctx, incentive)

	return &incentive, nil
}

// RegisterIncentive deletes the incentive for a contract
func (k Keeper) CancelIncentive(ctx sdk.Context, contract string) error {

	return nil
}
