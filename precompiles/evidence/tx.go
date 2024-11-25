// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package evidence

import (
	"fmt"

	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

// SubmitEvidence implements the evidence submission logic for the evidence precompile.
func (p Precompile) SubmitEvidence(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, submitterAddr, err := NewMsgSubmitEvidence(args)
	if err != nil {
		return nil, err
	}

	// If the contract is the submitter, we don't need an origin check
	// Otherwise check if the origin matches the submitter address
	isContractSubmitter := contract.CallerAddress == submitterAddr && contract.CallerAddress != origin
	if !isContractSubmitter && origin != submitterAddr {
		return nil, fmt.Errorf(ErrOriginDifferentFromSubmitter, origin.String(), submitterAddr.String())
	}

	msgServer := evidencekeeper.NewMsgServerImpl(p.evidenceKeeper)
	res, err := msgServer.SubmitEvidence(ctx, msg)
	if err != nil {
		return nil, err
	}

	if err = p.EmitSubmitEvidenceEvent(ctx, stateDB, submitterAddr, res.Hash); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}
