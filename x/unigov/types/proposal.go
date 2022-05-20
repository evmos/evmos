package types

import (
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeLendingMarket string = "Lending-Market"
)

var (
	_ govtypes.Content = &LendingMarketProposal{}
)

//Register Compound Proposal type as a valid proposal type in goveranance module
func init() {
	govtypes.RegisterProposalType(ProposalTypeLendingMarket)
	govtypes.RegisterProposalTypeCodec(&LendingMarketProposal{}, "unigov/LendingMarketProposal")
}

func NewLendingMarketProposal(title, description string, propMetaData govtypes.Proposal) govtypes.Content {
	return &LendingMarketProposal{
		Title:       title,
		Description: description,
		Proposal:    propMetaData,
	}
}

func (*LendingMarketProposal) ProposalRoute() string { return RouterKey }

func (*LendingMarketProposal) ProposalType() string {
	return ProposalTypeLendingMarket
}

func (lm *LendingMarketProposal) ValidateBasic() error {
	return govtypes.ValidateAbstract(lm)
}
