package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethermint "github.com/tharsis/ethermint/types"
)

// NewFee returns an instance of Fee
func NewFee(
	contract common.Address,
	owner sdk.AccAddress,
	withdrawAddress sdk.AccAddress,
) FeeContract {
	return FeeContract{
		Contract:        contract.String(),
		Owner:           owner.String(),
		WithdrawAddress: withdrawAddress.String(),
	}
}

// Validate performs a stateless validation of a FeeContract
func (i FeeContract) Validate() error {
	if err := ethermint.ValidateAddress(i.Contract); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(i.Owner); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(i.WithdrawAddress); err != nil {
		return err
	}

	return nil
}
