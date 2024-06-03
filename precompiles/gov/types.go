// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
)

// EventSetWithdrawAddress defines the event data for the SetWithdrawAddress transaction.
type EventSetWithdrawAddress struct {
	Caller            common.Address
	WithdrawerAddress string
}

// EventVote defines the event data for the Vote transaction.
type EventVote struct {
	Voter      common.Address
	ProposalId uint64
	Option     uint8
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

	proposalId, ok := args[1].(uint64)
	if !ok {
		return nil, common.Address{}, fmt.Errorf(ErrInvalidProposalId, args[1])
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
		ProposalId: proposalId,
		Voter:      sdk.AccAddress(voterAddress.Bytes()).String(),
		Option:     govv1.VoteOption(option),
		Metadata:   metadata,
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, common.Address{}, err
	}

	return msg, voterAddress, nil
}
