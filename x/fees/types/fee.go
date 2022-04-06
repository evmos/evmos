package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethermint "github.com/tharsis/ethermint/types"
)

// NewFee returns an instance of DevFeeInfo
func NewDevFeeInfo(
	contract common.Address,
	owner,
	withdraw sdk.AccAddress,
) DevFeeInfo {
	return DevFeeInfo{
		ContractAddress: contract.String(),
		DeployerAddress: owner.String(),
		WithdrawAddress: withdraw.String(),
	}
}

// Validate performs a stateless validation of a DevFeeInfo
func (i DevFeeInfo) Validate() error {
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
