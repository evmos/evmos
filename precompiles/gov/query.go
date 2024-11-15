// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

const (
	// GetVotesMethod defines the method name for the votes precompile request.
	GetVotesMethod = "getVotes"
	// GetVoteMethod defines the method name for the vote precompile request.
	GetVoteMethod = "getVote"
	// GetDepositMethod defines the method name for the deposit precompile request.
	GetDepositMethod = "getDeposit"
	// GetDepositsMethod defines the method name for the deposits precompile request.
	GetDepositsMethod = "getDeposits"
	// GetTallyResultMethod defines the method name for the tally result precompile request.
	GetTallyResultMethod = "getTallyResult"
	// GetProposalMethod defines the method name for the proposal precompile request.
	GetProposalMethod = "getProposal"
	// GetProposalsMethod defines the method name for the proposals precompile request.
	GetProposalsMethod = "getProposals"
)

// GetVotes implements the query logic for getting votes for a proposal.
func (p *Precompile) GetVotes(
	ctx sdk.Context,
	method *abi.Method,
	_ *vm.Contract,
	args []interface{},
) ([]byte, error) {
	queryVotesReq, err := ParseVotesArgs(method, args)
	if err != nil {
		return nil, err
	}

	queryServer := govkeeper.NewQueryServer(&p.govKeeper)
	res, err := queryServer.Votes(ctx, queryVotesReq)
	if err != nil {
		return nil, err
	}

	output := new(VotesOutput).FromResponse(res)
	return method.Outputs.Pack(output.Votes, output.PageResponse)
}

// GetVote implements the query logic for getting votes for a proposal.
func (p *Precompile) GetVote(
	ctx sdk.Context,
	method *abi.Method,
	_ *vm.Contract,
	args []interface{},
) ([]byte, error) {
	queryVotesReq, err := ParseVoteArgs(args)
	if err != nil {
		return nil, err
	}

	queryServer := govkeeper.NewQueryServer(&p.govKeeper)
	res, err := queryServer.Vote(ctx, queryVotesReq)
	if err != nil {
		return nil, err
	}

	output := new(VoteOutput).FromResponse(res)

	return method.Outputs.Pack(output.Vote)
}

// GetDeposit implements the query logic for getting a deposit for a proposal.
func (p *Precompile) GetDeposit(
	ctx sdk.Context,
	method *abi.Method,
	_ *vm.Contract,
	args []interface{},
) ([]byte, error) {
	queryDepositReq, err := ParseDepositArgs(args)
	if err != nil {
		return nil, err
	}

	queryServer := govkeeper.NewQueryServer(&p.govKeeper)
	res, err := queryServer.Deposit(ctx, queryDepositReq)
	if err != nil {
		return nil, err
	}

	output := new(DepositOutput).FromResponse(res)
	return method.Outputs.Pack(output.Deposit)
}

// GetDeposits implements the query logic for getting all deposits for a proposal.
func (p *Precompile) GetDeposits(
	ctx sdk.Context,
	method *abi.Method,
	_ *vm.Contract,
	args []interface{},
) ([]byte, error) {
	queryDepositsReq, err := ParseDepositsArgs(method, args)
	if err != nil {
		return nil, err
	}

	queryServer := govkeeper.NewQueryServer(&p.govKeeper)
	res, err := queryServer.Deposits(ctx, queryDepositsReq)
	if err != nil {
		return nil, err
	}

	output := new(DepositsOutput).FromResponse(res)
	return method.Outputs.Pack(output.Deposits, output.PageResponse)
}

// GetTallyResult implements the query logic for getting the tally result of a proposal.
func (p *Precompile) GetTallyResult(
	ctx sdk.Context,
	method *abi.Method,
	_ *vm.Contract,
	args []interface{},
) ([]byte, error) {
	queryTallyResultReq, err := ParseTallyResultArgs(args)
	if err != nil {
		return nil, err
	}

	queryServer := govkeeper.NewQueryServer(&p.govKeeper)
	res, err := queryServer.TallyResult(ctx, queryTallyResultReq)
	if err != nil {
		return nil, err
	}

	output := new(TallyResultOutput).FromResponse(res)
	return method.Outputs.Pack(output.TallyResult)
}

// GetProposal implements the query logic for getting a proposal
func (p *Precompile) GetProposal(
	ctx sdk.Context,
	method *abi.Method,
	_ *vm.Contract,
	args []interface{},
) ([]byte, error) {
	queryProposalReq, err := ParseProposalArgs(args)
	if err != nil {
		return nil, err
	}

	queryServer := govkeeper.NewQueryServer(&p.govKeeper)
	res, err := queryServer.Proposal(ctx, queryProposalReq)
	if err != nil {
		return nil, err
	}

	output := new(ProposalOutput).FromResponse(res)
	return method.Outputs.Pack(output.Proposal)
}

// GetProposals implements the query logic for getting proposals
func (p *Precompile) GetProposals(
	ctx sdk.Context,
	method *abi.Method,
	_ *vm.Contract,
	args []interface{},
) ([]byte, error) {
	queryProposalsReq, err := ParseProposalsArgs(method, args)
	if err != nil {
		return nil, err
	}

	queryServer := govkeeper.NewQueryServer(&p.govKeeper)
	res, err := queryServer.Proposals(ctx, queryProposalsReq)
	if err != nil {
		return nil, err
	}

	output := new(ProposalsOutput).FromResponse(res)

	return method.Outputs.Pack(output.Proposals, output.PageResponse)
}
