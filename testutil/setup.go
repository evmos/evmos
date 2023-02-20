package testutil

import (
	"time"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHeader creates a new Tendermint header for testing purposes.
func NewHeader(
	height int64,
	blockTime time.Time,
	chainID string,
	proposer sdk.ConsAddress,
	appHash,
	validatorHash []byte,
) tmproto.Header {
	return tmproto.Header{
		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		ChainID: chainID,
		Height:  height,
		Time:    blockTime,
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		LastCommitHash:     tmhash.Sum([]byte("commit")),
		DataHash:           tmhash.Sum([]byte("data")),
		ValidatorsHash:     validatorHash,
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		AppHash:            appHash,
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ProposerAddress:    proposer.Bytes(),
	}
}
