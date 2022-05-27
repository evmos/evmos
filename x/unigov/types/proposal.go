package types

import (
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeLendingMarket string = "Lending-Market"
	ProposalTypeTreasury string  = "Treasury"
	MaxDescriptionLength      int    = 1000
	MaxTitleLength            int    = 140
)

var (
	_ govtypes.Content = &LendingMarketProposal{}
	_ govtypes.Content = &TreasuryProposal{}
)

//Register Compound Proposal type as a valid proposal type in goveranance module
func init() {
	govtypes.RegisterProposalType(ProposalTypeLendingMarket)
	govtypes.RegisterProposalType(ProposalTypeTreasury) 
	govtypes.RegisterProposalTypeCodec(&LendingMarketProposal{}, "unigov/LendingMarketProposal")
	govtypes.RegisterProposalTypeCodec(&TreasuryProposal{}, "unigov/TreasuryProposal")
}

func NewLendingMarketProposal(title, description string, m *LendingMarketMetadata) govtypes.Content {
	return &LendingMarketProposal{
		Title:       title,
		Description: description,
		Metadata:    m,
	}
}

func NewTreasuryProposal(title, description string, tm *TreasuryProposalMetadata) govtypes.Content {
	return &TreasuryProposal{
		Title:        title,
		Description:  description,
		Metadata:     tm,
	}
}

func (*TreasuryProposal) ProposalRoute() string {return RouterKey}

func (*TreasuryProposal) ProposalType() string {
	return ProposalTypeTreasury
}

func (*LendingMarketProposal) ProposalRoute() string { return RouterKey }

func (*LendingMarketProposal) ProposalType() string {
	return ProposalTypeLendingMarket
}

func (lm *LendingMarketProposal) ValidateBasic() error {
	if err := govtypes.ValidateAbstract(lm); err != nil {
		return err
	}

	m := lm.GetMetadata()
	
	cd, vals, sigs := len(m.GetCalldatas()), len(m.GetValues()), len(m.GetSignatures())

	if cd != vals {
		return sdkerrors.Wrapf(govtypes.ErrInvalidProposalContent, "proposal array arguments must be same length")
	}

	if vals != sigs {
		return sdkerrors.Wrapf(govtypes.ErrInvalidProposalContent, "proposal array arguments must be same length")
	}
	return nil
}


func (tp *TreasuryProposal) ValidateBasic() error {
	if err := govtypes.ValidateAbstract(tp); err != nil {
		return err
	}

	tm := tp.GetMetadata()
	s := strings.ToLower(tm.GetDenom())
	
	if s != "canto" && s != "note" {
		return sdkerrors.Wrapf(govtypes.ErrInvalidProposalContent, "%s is not a valid denom string", tm.GetDenom())
	}
	
	return nil
}

func (tp *TreasuryProposal) FromTreasuryToLendingMarket() *LendingMarketProposal {
	m := tp.GetMetadata()
	
	lm := LendingMarketMetadata{
		Account: []string{m.GetRecipient()},
		PropId: m.GetPropID(), 
		Values: []uint64{m.GetAmount()},
		Calldatas: nil,
		Signatures: []string{m.GetDenom()},
	}
	
	return &LendingMarketProposal{
		Title: tp.GetTitle(),
		Description: tp.GetDescription(),
		Metadata: &lm,
	}
}
