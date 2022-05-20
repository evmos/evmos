package cli

import (
	"io/ioutil"
	"path/filepath"

	"github.com/Canto-Network/canto/v3/x/unigov/types"
	"github.com/cosmos/cosmos-sdk/codec"
)

// PARSING METADATA ACCORDING TO PROPOSAL STRUCT IN GOVTYPES TYPE IN UNIGOV

// ParseRegisterCoinProposal reads and parses a ParseRegisterCoinProposal from a file.
func ParseMetadata(cdc codec.JSONCodec, metadataFile string) (types.LendingMarketProposal, error) {
	propMetaData := types.LendingMarketProposal{}

	contents, err := ioutil.ReadFile(filepath.Clean(metadataFile))
	if err != nil {
		return propMetaData, err
	}

	if err = cdc.UnmarshalJSON(contents, &propMetaData); err != nil {
		return propMetaData, err
	}

	return propMetaData, nil
}
