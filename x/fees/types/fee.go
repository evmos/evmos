package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethermint "github.com/tharsis/ethermint/types"
)

// NewFee returns an instance of Fee. If the provided withdraw address is empty,
// it sets the value to the empty string.
func NewFee(contract common.Address, deployer, withdraw sdk.AccAddress) Fee {
	var withdrawAddr string
	if len(withdraw) > 0 {
		withdrawAddr = withdraw.String()
	}

	return Fee{
		ContractAddress: contract.String(),
		DeployerAddress: deployer.String(),
		WithdrawAddress: withdrawAddr,
	}
}

// GetContractAddr returns the contract address
func (f Fee) GetContractAddr() common.Address {
	return common.HexToAddress(f.ContractAddress)
}

// GetDeployerAddr returns the contract deployer address
func (f Fee) GetDeployerAddr() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(f.DeployerAddress)
}

// GetWithdrawAddr returns the account address to where the funds proceeding
// from the fees will be received. If the withdraw address is not defined, it
// defaults to the deployer address.
func (f Fee) GetWithdrawAddr() sdk.AccAddress {
	if f.WithdrawAddress == "" {
		return nil
	}

	return sdk.MustAccAddressFromBech32(f.WithdrawAddress)
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
