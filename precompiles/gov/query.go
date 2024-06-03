// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"
)

const (
	// ProposalMethod defines the ABI method name for the Proposal query.
	ProposalMethod = "proposal"
)

// Proposal returns the proposal info.
func (p Precompile) Proposal(
	_ sdk.Context,
	_ *vm.Contract,
	_ *abi.Method,
	_ []interface{},
) ([]byte, error) {
	return nil, nil
}
