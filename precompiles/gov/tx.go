// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gov

import (
	"fmt"
	"strings"

	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

const (
	// VoteMethod defines the ABI method name for the gov Vote transaction.
	VoteMethod = "vote"
	// VoteWeightedMethod defines the ABI method name for the gov VoteWeighted transaction.
	VoteWeightedMethod = "voteWeighted"
	// UpdateParamsMethod defines the ABI method name for the gov UpdateParams transaction.
	UpdateParamsMethod = "updateParams"
)

// Vote defines a method to add a vote on a specific proposal.
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
	// Otherwise check if the origin matches the voter address
	isContractVoter := contract.CallerAddress == voterHexAddr && contract.CallerAddress != origin
	if !isContractVoter && origin != voterHexAddr {
		return nil, fmt.Errorf(ErrDifferentOrigin, origin.String(), voterHexAddr.String())
	}

	msgSrv := govkeeper.NewMsgServerImpl(&p.govKeeper)
	if _, err = msgSrv.Vote(ctx, msg); err != nil {
		return nil, err
	}

	if err = p.EmitVoteEvent(ctx, stateDB, voterHexAddr, msg.ProposalId, int32(msg.Option)); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// VoteWeighted defines a method to add a vote on a specific proposal.
func (p Precompile) VoteWeighted(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, voterHexAddr, options, err := NewMsgVoteWeighted(method, args)
	if err != nil {
		return nil, err
	}

	// If the contract is the voter, we don't need an origin check
	// Otherwise check if the origin matches the voter address
	isContractVoter := contract.CallerAddress == voterHexAddr && contract.CallerAddress != origin
	if !isContractVoter && origin != voterHexAddr {
		return nil, fmt.Errorf(ErrDifferentOrigin, origin.String(), voterHexAddr.String())
	}

	msgSrv := govkeeper.NewMsgServerImpl(&p.govKeeper)
	if _, err = msgSrv.VoteWeighted(ctx, msg); err != nil {
		return nil, err
	}

	if err = p.EmitVoteWeightedEvent(ctx, stateDB, voterHexAddr, msg.ProposalId, options); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// UpdateParams updates the governance parameters
func (p Precompile) UpdateParams(
	ctx sdk.Context,
	origin common.Address,
	_ *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	authority := p.govKeeper.GetAuthority()

	bech32Prefix := strings.SplitN(authority, "1", 2)[0]
	if bech32Prefix == authority {
		return nil, fmt.Errorf("invalid bech32 address: %s", authority)
	}

	addressBz, err := sdk.GetFromBech32(authority, bech32Prefix)
	if err != nil {
		return nil, err
	}
	authorityAddress := common.BytesToAddress(addressBz)

	// TODO: Not sure if the caller authority will be origin or msg.sender ?
	// Would this require changes in the SDK so that transactions from the gov handler go through the Precompile ?
	if origin != authorityAddress {
		return nil, fmt.Errorf(ErrInvalidAuthority, origin)
	}

	params, err := NewMsgUpdateParams(authority, args)
	if err != nil {
		return nil, err
	}

	msgSrv := govkeeper.NewMsgServerImpl(&p.govKeeper)
	if _, err = msgSrv.UpdateParams(ctx, params); err != nil {
		return nil, err
	}

	if err = p.EmitUpdateParamsEvent(ctx, stateDB, params.Params); err != nil {
		fmt.Println("error emitting update params event", err)
		return nil, err
	}

	return method.Outputs.Pack(true)
}
