// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package evidence

import (
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
	_ *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, err := NewMsgSubmitEvidence(origin, args)
	if err != nil {
		return nil, err
	}

	msgServer := evidencekeeper.NewMsgServerImpl(p.evidenceKeeper)
	res, err := msgServer.SubmitEvidence(ctx, msg)
	if err != nil {
		return nil, err
	}

	if err = p.EmitSubmitEvidenceEvent(ctx, stateDB, origin, res.Hash); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}
