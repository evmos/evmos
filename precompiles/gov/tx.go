// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gov

import (
	"fmt"

	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"
)

const (
	// VoteMethod defines the ABI method name for the gov Vote transaction.
	VoteMethod = "vote"
)

// Vote claims the rewards accumulated by a delegator from multiple or all validators.
func (p Precompile) Vote(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, voterHexAddr, err := NewMsgVote(args)
	if err != nil {
		return nil, err
	}

	// If the contract is the voter, we don't need an origin check
	// Otherwise check if the origin matches the delegator address
	isContractVoter := contract.CallerAddress == voterHexAddr
	if !isContractVoter && origin != voterHexAddr {
		return nil, fmt.Errorf(ErrDifferentOrigin, origin.String(), voterHexAddr.String())
	}

	msgSrv := govkeeper.NewMsgServerImpl(&p.govKeeper)
	if _, err = msgSrv.Vote(sdk.WrapSDKContext(ctx), msg); err != nil {
		return nil, err
	}

	if err = p.EmitVoteEvent(ctx, stateDB, voterHexAddr, msg.ProposalId, uint8(msg.Option)); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}
