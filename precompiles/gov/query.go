package gov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

const (
	// VotesMethodRequest defines the method name for the votes precompile request.
	VotesMethodRequest = "votes"
	// VoteMethodRequest defines the method name for the vote precompile request.
	VoteMethodRequest
)

// Votes implements the query logic for getting votes for a proposal.
func (p *Precompile) Votes(
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

// VoteRequest implements the query logic for getting votes for a proposal.
func (p *Precompile) VoteRequest(
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
	fmt.Println(res, err)
	if err != nil {
		return nil, err
	}

	fmt.Println(res, err)
	output := new(SingleVote).FromResponse(res)

	return method.Outputs.Pack(output)
}
