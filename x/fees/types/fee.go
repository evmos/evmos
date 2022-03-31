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
	withdraw sdk.AccAddress,
) FeeContract {
	return FeeContract{
		ContractAddress: contract.String(),
		DeployerAddress: owner.String(),
		WithdrawAddress: withdraw.String(),
	}
}

// Validate performs a stateless validation of a FeeContract
func (i FeeContract) Validate() error {
	if err := ethermint.ValidateAddress(i.ContractAddress); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(i.DeployerAddress); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(i.WithdrawAddress); err != nil {
		return err
	}

	return nil
}
