// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gov

import (
	"errors"
	"fmt"

	cmn "github.com/evmos/evmos/v20/precompiles/common"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"

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
	// DepositMethod defines the ABI method name for the gov Deposit transaction.
	DepositMethod = "deposit"
	// CancelProposalMethod defines the ABI method name for the gov CancelProposal transaction.
	CancelProposalMethod = "cancelProposal"
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

// Deposit defines a method to add a deposit to a proposal
func (p Precompile) Deposit(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, depositorAddr, err := NewMsgDeposit(method, args)
	if err != nil {
		return nil, err
	}

	// Deposit must have length 1
	if len(msg.Amount) != 1 {
		return nil, errors.New(ErrInvalidDeposit)
	}

	// If the contract is the depositor, we don't need an origin check
	// Otherwise check if the origin matches the depositor address
	isOriginCaller := contract.CallerAddress == origin
	isContractDepositor := contract.CallerAddress == depositorAddr && !isOriginCaller
	if !isContractDepositor && origin != depositorAddr {
		return nil, fmt.Errorf(ErrDifferentOriginDepositor, origin.String(), depositorAddr.String())
	}

	msgSrv := govkeeper.NewMsgServerImpl(&p.govKeeper)
	if _, err = msgSrv.Deposit(ctx, msg); err != nil {
		return nil, err
	}

	if err = p.EmitDepositEvent(ctx, stateDB, msg.ProposalId, depositorAddr, msg.Amount); err != nil {
		return nil, err
	}

	if !isOriginCaller {
		// NOTE: This ensures that the changes in the bank keeper are correctly mirrored to the EVM stateDB
		// when calling the precompile from a smart contract
		// This prevents the stateDB from overwriting the changed balance in the bank keeper when committing the EVM state.
		// Need to scale the amount to 18 decimals for the EVM balance change entry

		scaledAmt := evmtypes.ConvertAmountTo18DecimalsBigInt(msg.Amount[0].Amount.BigInt())
		p.SetBalanceChangeEntries(cmn.NewBalanceChangeEntry(depositorAddr, scaledAmt, cmn.Sub))

	}
	return method.Outputs.Pack(true)
}

// CancelProposal defines a method to cancel a proposal
func (p Precompile) CancelProposal(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, proposerAddr, err := NewMsgCancelProposal(args)
	if err != nil {
		return nil, err
	}

	// If the contract is the proposer, we don't need an origin check
	// Otherwise check if the origin matches the proposer address
	isContractProposer := contract.CallerAddress == proposerAddr && contract.CallerAddress != origin
	if !isContractProposer && origin != proposerAddr {
		return nil, fmt.Errorf(ErrDifferentOriginProposer, origin.String(), proposerAddr.String())
	}

	msgSrv := govkeeper.NewMsgServerImpl(&p.govKeeper)
	resp, err := msgSrv.CancelProposal(ctx, msg)
	if err != nil {
		return nil, err
	}

	if err = p.EmitCancelProposalEvent(ctx, stateDB, msg.ProposalId, proposerAddr); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(
		true,
		uint64(resp.CanceledTime.Unix()), //nolint:gosec // G115
		resp.CanceledHeight,
	)
}
