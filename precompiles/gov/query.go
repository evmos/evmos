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
