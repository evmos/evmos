package cli

import (
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

// ParseRegisterCoinProposal reads and parses a ParseRegisterCoinProposal from a file.
func ParseRegisterCoinProposal(cdc codec.JSONCodec, proposalFile string) (types.RegisterCoinProposal, error) {
	proposal := types.RegisterCoinProposal{}

	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err = cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}

// ParseRegisterERC20Proposal reads and parses a RegisterERC20Proposal from a file.
func ParseRegisterERC20Proposal(cdc codec.JSONCodec, proposalFile string) (types.RegisterERC20Proposal, error) {
	proposal := types.RegisterERC20Proposal{}

	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err = cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}
