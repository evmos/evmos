package types

import(
	"strings"
	
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	ProposalTypeLendingMarket string = "Lending-Market"
	MaxDescriptionLength int = 1000
	MaxTitleLength int = 140
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
		Title: title,
		Description: description,
		proposal: propMetaData,
	}
}

func (lm *LendingMarketProposal) GetTitle() string {
	return lm.title;
}

func (lm *LendingMarketProposal) GetDescription() string {
	return lm.desc;
}

func (*LendingMarketProposal) ProposalRoute() string {return RouterKey}


func (*LendingMarketProposal) ProposalType() string {
	return ProposalTypeLendingMarket
}

func (lm *LendingMarketProposal) ValidateBasic() error {
	if err := govtypes.ValidateAbstract(lm); err != nil {
		return err
	}

	
}

func (lm *LendingMarketProposal) String() string {
	return lm.GetTitle()
}
