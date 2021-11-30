package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/common"
	ethermint "github.com/tharsis/ethermint/types"
)

// constants
const (
	ProposalTypeRegisterCoin         string = "RegisterCoin"
	ProposalTypeRegisterERC20        string = "RegisterERC20"
	ProposalTypeToggleTokenRelay     string = "ToggleTokenRelay" // #nosec
	ProposalTypeUpdateTokenPairERC20 string = "UpdateTokenPairERC20"
)

// Implements Proposal Interface
var (
	_ govtypes.Content = &RegisterCoinProposal{}
	_ govtypes.Content = &RegisterERC20Proposal{}
	_ govtypes.Content = &ToggleTokenRelayProposal{}
	_ govtypes.Content = &UpdateTokenPairERC20Proposal{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeRegisterCoin)
	govtypes.RegisterProposalType(ProposalTypeRegisterERC20)
	govtypes.RegisterProposalType(ProposalTypeToggleTokenRelay)
	govtypes.RegisterProposalType(ProposalTypeUpdateTokenPairERC20)
	govtypes.RegisterProposalTypeCodec(&RegisterCoinProposal{}, "intrarelayer/RegisterCoinProposal")
	govtypes.RegisterProposalTypeCodec(&RegisterERC20Proposal{}, "intrarelayer/RegisterERC20Proposal")
	govtypes.RegisterProposalTypeCodec(&ToggleTokenRelayProposal{}, "intrarelayer/ToggleTokenRelayProposal")
	govtypes.RegisterProposalTypeCodec(&UpdateTokenPairERC20Proposal{}, "intrarelayer/UpdateTokenPairERC20Proposal")
}

func CreateDenomDescription(address string) string {
	return fmt.Sprintf("Cosmos coin token representation of %s", address)
}

func CreateDenom(address string) string {
	return fmt.Sprintf("%s/%s", ModuleName, address)
}

// NewRegisterCoinProposal returns new instance of RegisterCoinProposal
func NewRegisterCoinProposal(title, description string, coinMetadata banktypes.Metadata) govtypes.Content {
	return &RegisterCoinProposal{
		Title:       title,
		Description: description,
		Metadata:    coinMetadata,
	}
}

// ProposalRoute returns router key for this proposal
func (*RegisterCoinProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for this proposal
func (*RegisterCoinProposal) ProposalType() string {
	return ProposalTypeRegisterCoin
}

// ValidateBasic performs a stateless check of the proposal fields
func (rtbp *RegisterCoinProposal) ValidateBasic() error {
	if err := rtbp.Metadata.Validate(); err != nil {
		return err
	}

	if err := ibctransfertypes.ValidateIBCDenom(rtbp.Metadata.Base); err != nil {
		return err
	}

	if err := validateIBC(rtbp.Metadata); err != nil {
		return err
	}

	return govtypes.ValidateAbstract(rtbp)
}

func validateIBC(metadata banktypes.Metadata) error {
	// Check ibc/ denom
	denomSplit := strings.SplitN(metadata.Base, "/", 2)

	if denomSplit[0] == metadata.Base && strings.TrimSpace(metadata.Base) != "" {
		// Not IBC
		return nil
	}

	if len(denomSplit) != 2 || denomSplit[0] != ibctransfertypes.DenomPrefix {
		// NOTE: should be unaccessible (covered on ValidateIBCDenom)
		return fmt.Errorf("invalid metadata. %s denomination should be prefixed with the format 'ibc/", metadata.Base)
	}

	if !strings.Contains(metadata.Name, "channel-") {
		return fmt.Errorf("invalid metadata (Name) for ibc. %s should include channel", metadata.Name)
	}

	if !strings.HasPrefix(metadata.Symbol, "ibc") {
		return fmt.Errorf("invalid metadata (Symbol) for ibc. %s should include \"ibc\" prefix", metadata.Symbol)
	}

	return nil
}

// ValidateIntrarelayerDenom checks if a denom is a valid intrarelayer/
// denomination
func ValidateIntrarelayerDenom(denom string) error {
	denomSplit := strings.SplitN(denom, "/", 2)

	if len(denomSplit) != 2 || denomSplit[0] != ModuleName {
		return fmt.Errorf("invalid denom. %s denomination should be prefixed with the format 'intrarelayer/", denom)
	}

	return ethermint.ValidateAddress(denomSplit[1])
}

// NewRegisterERC20Proposal returns new instance of RegisterERC20Proposal
func NewRegisterERC20Proposal(title, description, erc20Addr string) govtypes.Content {
	return &RegisterERC20Proposal{
		Title:        title,
		Description:  description,
		Erc20Address: erc20Addr,
	}
}

// ProposalRoute returns router key for this proposal
func (*RegisterERC20Proposal) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for this proposal
func (*RegisterERC20Proposal) ProposalType() string {
	return ProposalTypeRegisterERC20
}

// ValidateBasic performs a stateless check of the proposal fields
func (rtbp *RegisterERC20Proposal) ValidateBasic() error {
	if err := ethermint.ValidateAddress(rtbp.Erc20Address); err != nil {
		return sdkerrors.Wrap(err, "ERC20 address")
	}
	return govtypes.ValidateAbstract(rtbp)
}

// NewToggleTokenRelayProposal returns new instance of ToggleTokenRelayProposal
func NewToggleTokenRelayProposal(title, description string, token string) govtypes.Content {
	return &ToggleTokenRelayProposal{
		Title:       title,
		Description: description,
		Token:       token,
	}
}

// ProposalRoute returns router key for this proposal
func (*ToggleTokenRelayProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for this proposal
func (*ToggleTokenRelayProposal) ProposalType() string {
	return ProposalTypeToggleTokenRelay
}

// ValidateBasic performs a stateless check of the proposal fields
func (etrp *ToggleTokenRelayProposal) ValidateBasic() error {
	// check if the token is a hex address, if not, check if it is a valid SDK
	// denom
	if err := ethermint.ValidateAddress(etrp.Token); err != nil {
		if err := sdk.ValidateDenom(etrp.Token); err != nil {
			return err
		}
	}

	return govtypes.ValidateAbstract(etrp)
}

// NewUpdateTokenPairERC20Proposal returns new instance of UpdateTokenPairERC20Proposal
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
