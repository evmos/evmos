package cli

import (
	"io/ioutil"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/codec"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// PARSING METADATA ACCORDING TO PROPOSAL STRUCT IN GOVTYPES TYPE IN UNIGOV

// ParseRegisterCoinProposal reads and parses a ParseRegisterCoinProposal from a file.
func ParseMetadata(cdc codec.JSONCodec, metadataFile string) (govtypes.Proposal, error) {
	propMetaData := govtypes.Proposal{}

	contents, err := ioutil.ReadFile(filepath.Clean(metadataFile))
	if err != nil {
		return propMetaData, err
	}

	if err = cdc.UnmarshalJSON(contents, &propMetaData); err != nil {
		return propMetaData, err
	}

	return propMetaData, nil
}
