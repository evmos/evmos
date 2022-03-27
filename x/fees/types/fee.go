package types

import (
	"github.com/ethereum/go-ethereum/common"
	ethermint "github.com/tharsis/ethermint/types"
)

// NewFee returns an instance of Fee
func NewFee(
	contract common.Address,
) FeeContract {
	return FeeContract{
		Contract: contract.String(),
	}
}

// Validate performs a stateless validation of a FeeContract
func (i FeeContract) Validate() error {
	if err := ethermint.ValidateAddress(i.Contract); err != nil {
		return err
	}

	if err := ethermint.ValidateAddress(i.Owner); err != nil {
		return err
	}

	if err := ethermint.ValidateAddress(i.WithdrawAddress); err != nil {
		return err
	}

	return nil
}
