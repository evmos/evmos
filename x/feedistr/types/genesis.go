package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethermint "github.com/tharsis/ethermint/types"
)

func NewGenesisState(params Params, withdrawAddresses []ContractWithdrawAddress) GenesisState {
	return GenesisState{
		Params:            params,
		WithdrawAddresses: withdrawAddresses,
	}
}

func (gs GenesisState) Validate() error {
	seenContracts := make(map[string]bool)

	for _, wa := range gs.WithdrawAddresses {
		if seenContracts[wa.ContractAddress] {
			return fmt.Errorf("contract address duplicated on genesis '%s'", wa.ContractAddress)
		}

		if err := wa.Validate(); err != nil {
			return err
		}

		seenContracts[wa.ContractAddress] = true
	}

	return gs.Params.Validate()
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs a stateless validation of the fields.
func (cwa ContractWithdrawAddress) Validate() error {
	if err := ethermint.ValidateAddress(cwa.ContractAddress); err != nil {
		return sdkerrors.Wrap(err, "smart contract address")
	}
	if err := ethermint.ValidateAddress(cwa.WithdrawalAddress); err != nil {
		return sdkerrors.Wrap(err, "withdrawal address")
	}
	return nil
}
