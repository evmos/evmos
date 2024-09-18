// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gov

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/utils"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	cmn "github.com/evmos/evmos/v20/precompiles/common"
)

// EventVote defines the event data for the Vote transaction.
type EventVote struct {
	Voter      common.Address
	ProposalId uint64 //nolint:revive,stylecheck
	Option     uint8
}

// VotesInput defines the input for the Votes query.
type VotesInput struct {
	ProposalId uint64 //nolint:revive,stylecheck
	Pagination query.PageRequest
}

// VotesOutput defines the output for the Votes query.
type VotesOutput struct {
	Votes        []WeightedVote
	PageResponse query.PageResponse
}

// VoteOutput is the output response returned by the vote query method.
type VoteOutput struct {
	Vote WeightedVote
}

// WeightedVote defines a struct of vote for vote split.
type WeightedVote struct {
	ProposalId uint64 //nolint:revive,stylecheck
	Voter      common.Address
	Options    []WeightedVoteOption
	Metadata   string
}

// WeightedVoteOption defines a unit of vote for vote split.
type WeightedVoteOption struct {
	Option uint8
	Weight string
}

// NewMsgVote creates a new MsgVote instance.
func NewMsgVote(args []interface{}) (*govv1.MsgVote, common.Address, error) {
	if len(args) != 4 {
		return nil, common.Address{}, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 4, len(args))
	}

	voterAddress, ok := args[0].(common.Address)
	if !ok || voterAddress == (common.Address{}) {
		return nil, common.Address{}, fmt.Errorf(ErrInvalidVoter, args[0])
	}

	proposalID, ok := args[1].(uint64)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(ErrInvalidProposalID, args[1])
	}

	option, ok := args[2].(uint8)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(ErrInvalidOption, args[2])
	}

	metadata, ok := args[3].(string)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(ErrInvalidMetadata, args[3])
	}

	msg := &govv1.MsgVote{
		ProposalId: proposalID,
		Voter:      sdk.AccAddress(voterAddress.Bytes()).String(),
		Option:     govv1.VoteOption(option),
		Metadata:   metadata,
	}

	return msg, voterAddress, nil
}

// ParseVotesArgs parses the arguments for the Votes query.
func ParseVotesArgs(method *abi.Method, args []interface{}) (*govv1.QueryVotesRequest, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	var input VotesInput
	if err := method.Inputs.Copy(&input, args); err != nil {
		return nil, fmt.Errorf("error while unpacking args to VotesInput: %s", err)
	}

	return &govv1.QueryVotesRequest{
		ProposalId: input.ProposalId,
		Pagination: &input.Pagination,
	}, nil
}

func (vo *VotesOutput) FromResponse(res *govv1.QueryVotesResponse) *VotesOutput {
	vo.Votes = make([]WeightedVote, len(res.Votes))
	for i, v := range res.Votes {
		hexAddr, err := utils.Bech32ToHexAddr(v.Voter)
		if err != nil {
			return nil
		}
		options := make([]WeightedVoteOption, len(v.Options))
		for j, opt := range v.Options {
			options[j] = WeightedVoteOption{
				Option: uint8(opt.Option), //nolint:gosec // G115
				Weight: opt.Weight,
			}
		}
		vo.Votes[i] = WeightedVote{
			ProposalId: v.ProposalId,
			Voter:      hexAddr,
			Options:    options,
			Metadata:   v.Metadata,
		}
	}
	if res.Pagination != nil {
		vo.PageResponse = query.PageResponse{
			NextKey: res.Pagination.NextKey,
			Total:   res.Pagination.Total,
		}
	}
	return vo
}

// ParseVoteArgs parses the arguments for the Votes query.
func ParseVoteArgs(args []interface{}) (*govv1.QueryVoteRequest, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	proposalID, ok := args[0].(uint64)
	if !ok {
		return nil, fmt.Errorf(ErrInvalidProposalID, args[0])
	}

	voter, ok := args[1].(common.Address)
	if !ok {
		return nil, fmt.Errorf(ErrInvalidVoter, args[1])
	}

	voterAccAddr := sdk.AccAddress(voter.Bytes())
	return &govv1.QueryVoteRequest{
		ProposalId: proposalID,
		Voter:      voterAccAddr.String(),
	}, nil
}

func (vo *VoteOutput) FromResponse(res *govv1.QueryVoteResponse) *VoteOutput {
	hexVoter, err := utils.Bech32ToHexAddr(res.Vote.Voter)
	if err != nil {
		return nil
	}
	vo.Vote.Voter = hexVoter
	vo.Vote.Metadata = res.Vote.Metadata
	vo.Vote.ProposalId = res.Vote.ProposalId

	options := make([]WeightedVoteOption, len(res.Vote.Options))
	for j, opt := range res.Vote.Options {
		options[j] = WeightedVoteOption{
			Option: uint8(opt.Option), //nolint:gosec // G115
			Weight: opt.Weight,
		}
	}
	vo.Vote.Options = options
	return vo
}
