package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	"github.com/tharsis/evmos/v3/x/fees/types"
)

// RegisterIncentive creates an incentive for a contract
func (k Keeper) RegisterContract(
	ctx sdk.Context,
	contract common.Address,
) (*types.FeeContract, error) {
	// Check if the Incentives are globally enabled
	params := k.GetParams(ctx)
	if !params.EnableFees {
		return nil, sdkerrors.Wrap(
			types.ErrInternalFee,
			"incentives are currently disabled by governance",
		)
	}

	// Check if contract exists
	acc := k.evmKeeper.GetAccountWithoutBalance(ctx, contract)
	if acc == nil || !acc.IsContract() {
		return nil, sdkerrors.Wrapf(
			types.ErrInternalFee,
			"contract doesn't exist: %s", contract,
		)
	}

	// Check if the incentive is already registered
	if k.IsContractRegistered(ctx, contract) {
		return nil, sdkerrors.Wrapf(
			types.ErrInternalFee,
			"incentive already registered: %s", contract,
		)
	}

	// create incentive and set to store
	fee := types.NewFee(contract)
	fee.StartTime = ctx.BlockTime()
	k.SetFeeContract(ctx, fee)

	return &fee, nil
}

// RegisterFeeContract deletes the fee for a contract
func (k Keeper) CancelContract(
	ctx sdk.Context,
	contract common.Address,
) error {
	// Check if the fees are globally enabled
	params := k.GetParams(ctx)
	if !params.EnableFees {
		return sdkerrors.Wrap(
			types.ErrInternalFee,
			"incentives are currently disabled by governance",
		)
	}

	_, found := k.GetFee(ctx, contract)
	if !found {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"unmatching contract '%s' ", contract,
		)
	}

	return nil
}
