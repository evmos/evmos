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

// EventVoteWeighted defines the event data for the VoteWeighted transaction.
type EventVoteWeighted struct {
	Voter      common.Address
	ProposalId uint64 //nolint:revive,stylecheck
	Options    WeightedVoteOptions
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

// WeightedVoteOptions defines a slice of WeightedVoteOption.
type WeightedVoteOptions []WeightedVoteOption

// DepositInput defines the input for the Deposit query.
type DepositInput struct {
	ProposalId uint64 //nolint:revive,stylecheck
	Depositor  common.Address
}

// DepositOutput defines the output for the Deposit query.
type DepositOutput struct {
	Deposit DepositData
}

// DepositsInput defines the input for the Deposits query.
type DepositsInput struct {
	ProposalId uint64 //nolint:revive,stylecheck
	Pagination query.PageRequest
}

// DepositsOutput defines the output for the Deposits query.
type DepositsOutput struct {
	Deposits     []DepositData      `abi:"deposits"`
	PageResponse query.PageResponse `abi:"pageResponse"`
}

// TallyResultOutput defines the output for the TallyResult query.
type TallyResultOutput struct {
	TallyResult TallyResultData
}

// DepositData represents information about a deposit on a proposal
type DepositData struct {
	ProposalId uint64         `abi:"proposalId"` //nolint:revive,stylecheck
	Depositor  common.Address `abi:"depositor"`
	Amount     []cmn.Coin     `abi:"amount"`
}

// TallyResultData represents the tally result of a proposal
type TallyResultData struct {
	Yes        string
	Abstain    string
	No         string
	NoWithVeto string
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

// NewMsgVoteWeighted creates a new MsgVoteWeighted instance.
func NewMsgVoteWeighted(method *abi.Method, args []interface{}) (*govv1.MsgVoteWeighted, common.Address, WeightedVoteOptions, error) {
	if len(args) != 4 {
		return nil, common.Address{}, WeightedVoteOptions{}, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 4, len(args))
	}

	voterAddress, ok := args[0].(common.Address)
	if !ok || voterAddress == (common.Address{}) {
		return nil, common.Address{}, WeightedVoteOptions{}, fmt.Errorf(ErrInvalidVoter, args[0])
	}

	proposalID, ok := args[1].(uint64)
	if !ok {
		return nil, common.Address{}, WeightedVoteOptions{}, fmt.Errorf(ErrInvalidProposalID, args[1])
	}

	// Unpack the input struct
	var options WeightedVoteOptions
	arguments := abi.Arguments{method.Inputs[2]}
	if err := arguments.Copy(&options, []interface{}{args[2]}); err != nil {
		return nil, common.Address{}, WeightedVoteOptions{}, fmt.Errorf("error while unpacking args to Options struct: %s", err)
	}

	weightedOptions := make([]*govv1.WeightedVoteOption, len(options))
	for i, option := range options {
		weightedOptions[i] = &govv1.WeightedVoteOption{
			Option: govv1.VoteOption(option.Option),
			Weight: option.Weight,
		}
	}

	metadata, ok := args[3].(string)
	if !ok {
		return nil, common.Address{}, WeightedVoteOptions{}, fmt.Errorf(ErrInvalidMetadata, args[3])
	}

	msg := &govv1.MsgVoteWeighted{
		ProposalId: proposalID,
		Voter:      sdk.AccAddress(voterAddress.Bytes()).String(),
		Options:    weightedOptions,
		Metadata:   metadata,
	}

	return msg, voterAddress, options, nil
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

// ParseDepositArgs parses the arguments for the Deposit query.
func ParseDepositArgs(args []interface{}) (*govv1.QueryDepositRequest, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	proposalID, ok := args[0].(uint64)
	if !ok {
		return nil, fmt.Errorf(ErrInvalidProposalID, args[0])
	}

	depositor, ok := args[1].(common.Address)
	if !ok {
		return nil, fmt.Errorf(ErrInvalidDepositor, args[1])
	}

	depositorAccAddr := sdk.AccAddress(depositor.Bytes())
	return &govv1.QueryDepositRequest{
		ProposalId: proposalID,
		Depositor:  depositorAccAddr.String(),
	}, nil
}

// ParseDepositsArgs parses the arguments for the Deposits query.
func ParseDepositsArgs(method *abi.Method, args []interface{}) (*govv1.QueryDepositsRequest, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	var input DepositsInput
	if err := method.Inputs.Copy(&input, args); err != nil {
		return nil, fmt.Errorf("error while unpacking args to DepositsInput: %s", err)
	}

	return &govv1.QueryDepositsRequest{
		ProposalId: input.ProposalId,
		Pagination: &input.Pagination,
	}, nil
}

// ParseTallyResultArgs parses the arguments for the TallyResult query.
func ParseTallyResultArgs(args []interface{}) (*govv1.QueryTallyResultRequest, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	proposalID, ok := args[0].(uint64)
	if !ok {
		return nil, fmt.Errorf(ErrInvalidProposalID, args[0])
	}

	return &govv1.QueryTallyResultRequest{
		ProposalId: proposalID,
	}, nil
}

func (do *DepositOutput) FromResponse(res *govv1.QueryDepositResponse) *DepositOutput {
	hexDepositor, err := utils.Bech32ToHexAddr(res.Deposit.Depositor)
	if err != nil {
		return nil
	}
	coins := make([]cmn.Coin, len(res.Deposit.Amount))
	for i, c := range res.Deposit.Amount {
		coins[i] = cmn.Coin{
			Denom:  c.Denom,
			Amount: c.Amount.BigInt(),
		}
	}
	do.Deposit = DepositData{
		ProposalId: res.Deposit.ProposalId,
		Depositor:  hexDepositor,
		Amount:     coins,
	}
	return do
}

func (do *DepositsOutput) FromResponse(res *govv1.QueryDepositsResponse) *DepositsOutput {
	do.Deposits = make([]DepositData, len(res.Deposits))
	for i, d := range res.Deposits {
		hexDepositor, err := utils.Bech32ToHexAddr(d.Depositor)
		if err != nil {
			return nil
		}
		coins := make([]cmn.Coin, len(d.Amount))
		for j, c := range d.Amount {
			coins[j] = cmn.Coin{
				Denom:  c.Denom,
				Amount: c.Amount.BigInt(),
			}
		}
		do.Deposits[i] = DepositData{
			ProposalId: d.ProposalId,
			Depositor:  hexDepositor,
			Amount:     coins,
		}
	}
	if res.Pagination != nil {
		do.PageResponse = query.PageResponse{
			NextKey: res.Pagination.NextKey,
			Total:   res.Pagination.Total,
		}
	}
	return do
}

func (tro *TallyResultOutput) FromResponse(res *govv1.QueryTallyResultResponse) *TallyResultOutput {
	tro.TallyResult = TallyResultData{
		Yes:        res.Tally.YesCount,
		Abstain:    res.Tally.AbstainCount,
		No:         res.Tally.NoCount,
		NoWithVeto: res.Tally.NoWithVetoCount,
	}
	return tro
}

// ProposalOutput defines the output for the Proposal query
type ProposalOutput struct {
	Proposal ProposalData
}

// ProposalsInput defines the input for the Proposals query
type ProposalsInput struct {
	ProposalStatus uint32
	Voter          common.Address
	Depositor      common.Address
	Pagination     query.PageRequest
}

// ProposalsOutput defines the output for the Proposals query
type ProposalsOutput struct {
	Proposals    []ProposalData
	PageResponse query.PageResponse
}

// ProposalData represents a governance proposal
type ProposalData struct {
	Id               uint64          `abi:"id"` //nolint
	Messages         []string        `abi:"messages"`
	Status           uint32          `abi:"status"`
	FinalTallyResult TallyResultData `abi:"finalTallyResult"`
	SubmitTime       uint64          `abi:"submitTime"`
	DepositEndTime   uint64          `abi:"depositEndTime"`
	TotalDeposit     []cmn.Coin      `abi:"totalDeposit"`
	VotingStartTime  uint64          `abi:"votingStartTime"`
	VotingEndTime    uint64          `abi:"votingEndTime"`
	Metadata         string          `abi:"metadata"`
	Title            string          `abi:"title"`
	Summary          string          `abi:"summary"`
	Proposer         common.Address  `abi:"proposer"`
}

// ParseProposalArgs parses the arguments for the Proposal query
func ParseProposalArgs(args []interface{}) (*govv1.QueryProposalRequest, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	proposalID, ok := args[0].(uint64)
	if !ok {
		return nil, fmt.Errorf(ErrInvalidProposalID, args[0])
	}

	return &govv1.QueryProposalRequest{
		ProposalId: proposalID,
	}, nil
}

// ParseProposalsArgs parses the arguments for the Proposals query
func ParseProposalsArgs(method *abi.Method, args []interface{}) (*govv1.QueryProposalsRequest, error) {
	if len(args) != 4 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 4, len(args))
	}

	var input ProposalsInput
	if err := method.Inputs.Copy(&input, args); err != nil {
		return nil, fmt.Errorf("error while unpacking args to ProposalsInput: %s", err)
	}

	voter := ""
	if input.Voter != (common.Address{}) {
		voter = sdk.AccAddress(input.Voter.Bytes()).String()
	}

	depositor := ""
	if input.Depositor != (common.Address{}) {
		depositor = sdk.AccAddress(input.Depositor.Bytes()).String()
	}

	return &govv1.QueryProposalsRequest{
		ProposalStatus: govv1.ProposalStatus(input.ProposalStatus), //nolint:gosec // G115
		Voter:          voter,
		Depositor:      depositor,
		Pagination:     &input.Pagination,
	}, nil
}

func (po *ProposalOutput) FromResponse(res *govv1.QueryProposalResponse) *ProposalOutput {
	msgs := make([]string, len(res.Proposal.Messages))
	for i, msg := range res.Proposal.Messages {
		msgs[i] = msg.TypeUrl
	}

	coins := make([]cmn.Coin, len(res.Proposal.TotalDeposit))
	for i, c := range res.Proposal.TotalDeposit {
		coins[i] = cmn.Coin{
			Denom:  c.Denom,
			Amount: c.Amount.BigInt(),
		}
	}

	proposer, err := utils.Bech32ToHexAddr(res.Proposal.Proposer)
	if err != nil {
		return nil
	}

	po.Proposal = ProposalData{
		Id:       res.Proposal.Id,
		Messages: msgs,
		Status:   uint32(res.Proposal.Status), //nolint:gosec // G115
		FinalTallyResult: TallyResultData{
			Yes:        res.Proposal.FinalTallyResult.YesCount,
			Abstain:    res.Proposal.FinalTallyResult.AbstainCount,
			No:         res.Proposal.FinalTallyResult.NoCount,
			NoWithVeto: res.Proposal.FinalTallyResult.NoWithVetoCount,
		},
		SubmitTime:      uint64(res.Proposal.SubmitTime.Unix()),     //nolint:gosec // G115
		DepositEndTime:  uint64(res.Proposal.DepositEndTime.Unix()), //nolint:gosec // G115
		TotalDeposit:    coins,
		VotingStartTime: uint64(res.Proposal.VotingStartTime.Unix()), //nolint:gosec // G115
		VotingEndTime:   uint64(res.Proposal.VotingEndTime.Unix()),   //nolint:gosec // G115
		Metadata:        res.Proposal.Metadata,
		Title:           res.Proposal.Title,
		Summary:         res.Proposal.Summary,
		Proposer:        proposer,
	}
	return po
}

func (po *ProposalsOutput) FromResponse(res *govv1.QueryProposalsResponse) *ProposalsOutput {
	po.Proposals = make([]ProposalData, len(res.Proposals))
	for i, p := range res.Proposals {
		msgs := make([]string, len(p.Messages))
		for j, msg := range p.Messages {
			msgs[j] = msg.TypeUrl
		}

		coins := make([]cmn.Coin, len(p.TotalDeposit))
		for j, c := range p.TotalDeposit {
			coins[j] = cmn.Coin{
				Denom:  c.Denom,
				Amount: c.Amount.BigInt(),
			}
		}

		proposer, err := utils.Bech32ToHexAddr(p.Proposer)
		if err != nil {
			return nil
		}

		po.Proposals[i] = ProposalData{
			Id:       p.Id,
			Messages: msgs,
			Status:   uint32(p.Status), //nolint:gosec // G115
			FinalTallyResult: TallyResultData{
				Yes:        p.FinalTallyResult.YesCount,
				Abstain:    p.FinalTallyResult.AbstainCount,
				No:         p.FinalTallyResult.NoCount,
				NoWithVeto: p.FinalTallyResult.NoWithVetoCount,
			},
			SubmitTime:      uint64(p.SubmitTime.Unix()),     //nolint:gosec // G115
			DepositEndTime:  uint64(p.DepositEndTime.Unix()), //nolint:gosec // G115
			TotalDeposit:    coins,
			VotingStartTime: uint64(p.VotingStartTime.Unix()), //nolint:gosec // G115
			VotingEndTime:   uint64(p.VotingEndTime.Unix()),   //nolint:gosec // G115
			Metadata:        p.Metadata,
			Title:           p.Title,
			Summary:         p.Summary,
			Proposer:        proposer,
		}
	}

	if res.Pagination != nil {
		po.PageResponse = query.PageResponse{
			NextKey: res.Pagination.NextKey,
			Total:   res.Pagination.Total,
		}
	}
	return po
}
