package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"
	ethermint "github.com/tharsis/ethermint/types"
)

// constants
const (
	ProposalTypeRegisterTokenPair    string = "RegisterTokenPair"
	ProposalTypeEnableTokenRelay     string = "EnableTokenRelay"
	ProposalTypeUpdateTokenPairERC20 string = "UpdateTokenPairERC20"
)

// Implements Proposal Interface
var (
	_ govtypes.Content = &RegisterTokenPairProposal{}
	_ govtypes.Content = &EnableTokenRelayProposal{}
	_ govtypes.Content = &UpdateTokenPairERC20Proposal{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeRegisterTokenPair)
	govtypes.RegisterProposalType(ProposalTypeEnableTokenRelay)
	govtypes.RegisterProposalType(ProposalTypeUpdateTokenPairERC20)
	govtypes.RegisterProposalTypeCodec(&RegisterTokenPairProposal{}, "intrarelayer/RegisterTokenPairProposal")
	govtypes.RegisterProposalTypeCodec(&EnableTokenRelayProposal{}, "intrarelayer/EnableTokenRelayProposal")
	govtypes.RegisterProposalTypeCodec(&UpdateTokenPairERC20Proposal{}, "intrarelayer/UpdateTokenPairERC20Proposal")
}

// NewRegisterTokenPairProposal returns new instance of TokenPairProposal
func NewRegisterTokenPairProposal(title, description string, pair TokenPair) govtypes.Content {
	return &RegisterTokenPairProposal{
		Title:       title,
		Description: description,
		TokenPair:   pair,
	}
}

// ProposalRoute returns router key for this proposal
func (*RegisterTokenPairProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for this proposal
func (*RegisterTokenPairProposal) ProposalType() string {
	return ProposalTypeRegisterTokenPair
}

// ValidateBasic performs a stateless check of the proposal fields
func (rtbp *RegisterTokenPairProposal) ValidateBasic() error {
	if err := rtbp.TokenPair.Validate(); err != nil {
		return err
	}

	return govtypes.ValidateAbstract(rtbp)
}

// NewEnableTokenRelayProposal returns new instance of TokenPairProposal
func NewEnableTokenRelayProposal(title, description string, token string) govtypes.Content {
	return &EnableTokenRelayProposal{
		Title:       title,
		Description: description,
		Token:       token,
	}
}

// ProposalRoute returns router key for this proposal
func (*EnableTokenRelayProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for this proposal
func (*EnableTokenRelayProposal) ProposalType() string {
	return ProposalTypeEnableTokenRelay
}

// ValidateBasic performs a stateless check of the proposal fields
func (etrp *EnableTokenRelayProposal) ValidateBasic() error {
	// check if the token is a hex address, if not, check if it is a valid SDK
	// denom
	if err := ethermint.ValidateAddress(etrp.Token); err != nil {
		return sdk.ValidateDenom(etrp.Token)
	}

	return govtypes.ValidateAbstract(etrp)
}

// NewUpdateTokenPairERC20Proposal returns new instance of TokenPairProposal
func NewUpdateTokenPairERC20Proposal(title, description string, erc20Addr, newERC20Addr common.Address) govtypes.Content {
	return &UpdateTokenPairERC20Proposal{
		Title:           title,
		Description:     description,
		Erc20Address:    erc20Addr.String(),
		NewErc20Address: newERC20Addr.String(),
	}
}

// ProposalRoute returns router key for this proposal
func (*UpdateTokenPairERC20Proposal) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for this proposal
func (*UpdateTokenPairERC20Proposal) ProposalType() string {
	return ProposalTypeUpdateTokenPairERC20
}

// ValidateBasic performs a stateless check of the proposal fields
func (p *UpdateTokenPairERC20Proposal) ValidateBasic() error {
	if err := ethermint.ValidateAddress(p.Erc20Address); err != nil {
		return sdkerrors.Wrap(err, "ERC20 address")
	}

	if err := ethermint.ValidateAddress(p.NewErc20Address); err != nil {
		return sdkerrors.Wrap(err, "new ERC20 address")
	}

	return govtypes.ValidateAbstract(p)
}

// GetERC20Address returns the common.Address representation of the ERC20 hex address
func (p UpdateTokenPairERC20Proposal) GetERC20Address() common.Address {
	return common.HexToAddress(p.Erc20Address)
}

// GetNewERC20Address returns the common.Address representation of the new ERC20 hex address
func (p UpdateTokenPairERC20Proposal) GetNewERC20Address() common.Address {
	return common.HexToAddress(p.NewErc20Address)
}
