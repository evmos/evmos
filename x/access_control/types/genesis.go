// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package types

import (
	"fmt"

	evmostypes "github.com/evmos/evmos/v18/types"
)

type ContractAccounts struct {
	Contract string
	Accounts []string
}

// DefaultGenesisState sets default genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		ContractOwner:   []ContractAccount{},
		PausedContracts: []string{},
	}
}

// TODO: define alias to remove dup code
func (gs GenesisState) Validate() error {
	seenContractsOwner := make(map[string]bool)
	seenPausedContracts := make(map[string]bool)

	for _, co := range gs.ContractOwner {
		if seenContractsOwner[co.Contract] {
			return fmt.Errorf("duplicate contract owner %s", co.Contract)
		}

		if err := co.Validate(); err != nil {
			return err
		}

		seenContractsOwner[co.Contract] = true
	}

	for _, pc := range gs.PausedContracts {
		if seenPausedContracts[pc] {
			return fmt.Errorf("duplicate paused contract %s", pc)
		}

		if err := evmostypes.ValidateNonZeroAddress(pc); err != nil {
			return err
		}

		seenPausedContracts[pc] = true
	}

	return nil
}

func (cacc ContractAccount) Validate() error {
	if err := evmostypes.ValidateNonZeroAddress(cacc.Contract); err != nil {
		return err
	}

	return evmostypes.ValidateNonZeroAddress(cacc.Account)
}

func (cacc ContractAccounts) Validate() error {
	if err := evmostypes.ValidateNonZeroAddress(cacc.Contract); err != nil {
		return err
	}

	seenAccounts := make(map[string]bool)
	for _, account := range cacc.Accounts {
		if seenAccounts[account] {
			return fmt.Errorf("duplicate account %s", account)
		}

		if err := evmostypes.ValidateNonZeroAddress(account); err != nil {
			return err
		}

		seenAccounts[account] = true
	}

	return nil
}
