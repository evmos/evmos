package types

import(
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeLendingMarket string = "Lending-Market"
)

var (
	_ govtypes.Content = &CompoundProposal{}
)


//Register Compound Proposal type as a valid proposal type in goveranance module
func init() {
	govtypes.RegisterProposalType(ProposalTypeLendingMarket)
	govtypes.RegisterProposalTypeCodec(&LendingMarketProposal{}, "unigov/LendingMarketProposal")
}

func NewCompoundProposal(title, description string, propMetaData govtypes.Proposal) govtypes.Content {
	return &CompoundProposal{
		Title: title,
		Description: description,
		proposal: propMetaData,
	}
}

func (*CompoundProposal) ProposalRoute() string {return RouterKey}


func (cp *CompoundProposal) ProposalType() string {
	return ProposalTypeLendingMarket
}
