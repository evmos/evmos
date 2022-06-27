package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethermint "github.com/evmos/ethermint/types"
)

// NewFee returns an instance of Fee
func NewFee(contract common.Address, deployer, withdraw sdk.AccAddress) Fee {
	return Fee{
		ContractAddress: contract.String(),
		DeployerAddress: deployer.String(),
		WithdrawAddress: withdraw.String(),
	}
}

// Validate performs a stateless validation of a Fee
func (f Fee) Validate() error {
	if err := ethermint.ValidateNonZeroAddress(f.ContractAddress); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(f.DeployerAddress); err != nil {
		return err
	}

	if f.WithdrawAddress != "" {
		if _, err := sdk.AccAddressFromBech32(f.WithdrawAddress); err != nil {
			return err
		}
	}

	return nil
}
