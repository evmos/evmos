package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethermint "github.com/tharsis/ethermint/types"
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
func (i Fee) Validate() error {
	if err := ethermint.ValidateNonZeroAddress(i.ContractAddress); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(i.DeployerAddress); err != nil {
		return err
	}

	if i.WithdrawAddress != "" {
		if _, err := sdk.AccAddressFromBech32(i.WithdrawAddress); err != nil {
			return err
		}
	}

	return nil
}
