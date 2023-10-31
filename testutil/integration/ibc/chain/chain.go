// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package chain

import (
	"time"

	tmtypes "github.com/cometbft/cometbft/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/cosmos/ibc-go/v7/testing/simapp"
)

// IBCChain defines the required methods needed for a testing IBC chain that complies
// with the ibctesting chain struct.
type Chain interface {
	// GetContext returns the current context for the application.
	GetContext() sdktypes.Context
	// GetSimApp returns the SimApp to allow usage ofnon-interface fields.
	GetSimApp() *simapp.SimApp
	// QueryProof performs an abci query with the given key and returns the proto encoded merkle proof
	// for the query and the height at which the proof will succeed on a tendermint verifier.
	QueryProof(key []byte) ([]byte, clienttypes.Height)
	// QueryProofAtHeight performs an abci query with the given key and returns the proto encoded merkle proof
	// for the query and the height at which the proof will succeed on a tendermint verifier. Only the IBC
	// store is supported
	QueryProofAtHeight(key []byte, height int64) ([]byte, clienttypes.Height)
	// QueryProofForStore performs an abci query with the given key and returns the proto encoded merkle proof
	// for the query and the height at which the proof will succeed on a tendermint verifier.
	QueryProofForStore(storeKey string, key []byte, height int64) ([]byte, clienttypes.Height)
	// QueryUpgradeProof performs an abci query with the given key and returns the proto encoded merkle proof
	// for the query and the height at which the proof will succeed on a tendermint verifier.
	QueryUpgradeProof(key []byte, height uint64) ([]byte, clienttypes.Height)
	// QueryConsensusStateProof performs an abci query for a consensus state
	// stored on the given clientID. The proof and consensusHeight are returned.
	QueryConsensusStateProof(clientID string) ([]byte, clienttypes.Height)
	// NextBlock sets the last header to the current header and increments the current header to be
	// at the next block height. It does not update the time as that is handled by the Coordinator.
	// It will call Endblock and Commit and apply the validator set changes to the next validators
	// of the next block being created. This follows the Tendermint protocol of applying valset changes
	// returned on block `n` to the validators of block `n+2`.
	// It calls BeginBlock with the new block created before returning.
	NextBlock()
	// GetClientState retrieves the client state for the provided clientID. The client is
	// expected to exist otherwise testing will fail.
	GetClientState(clientID string) exported.ClientState
	// GetConsensusState retrieves the consensus state for the provided clientID and height.
	// It will return a success boolean depending on if consensus state exists or not.
	GetConsensusState(clientID string, height exported.Height) (exported.ConsensusState, bool)
	// GetValsAtHeight will return the trusted validator set of the chain for the given trusted height. It will return
	// a success boolean depending on if the validator set exists or not at that height.
	GetValsAtHeight(trustedHeight int64) (*tmtypes.ValidatorSet, bool)
	// GetAcknowledgement retrieves an acknowledgement for the provided packet. If the
	// acknowledgement does not exist then testing will fail.
	GetAcknowledgement(packet exported.PacketI) []byte
	// GetPrefix returns the prefix for used by a chain in connection creation
	GetPrefix() commitmenttypes.MerklePrefix
	// ConstructUpdateTMClientHeader will construct a valid 07-tendermint Header to update the
	// light client on the source chain.
	ConstructUpdateTMClientHeader(counterparty *ibctesting.TestChain, clientID string) (*ibctm.Header, error)
	// ConstructUpdateTMClientHeader will construct a valid 07-tendermint Header to update the
	ConstructUpdateTMClientHeaderWithTrustedHeight(counterparty *ibctesting.TestChain, clientID string, trustedHeight clienttypes.Height) (*ibctm.Header, error) // light client on the source chain.
	// ExpireClient fast forwards the chain's block time by the provided amount of time which will
	// expire any clients with a trusting period less than or equal to this amount of time.
	ExpireClient(amount time.Duration)
	// CurrentTMClientHeader creates a TM header using the current header parameters
	// on the chain. The trusted fields in the header are set to nil.
	CurrentTMClientHeader() *ibctm.Header
	// GetChannelCapability returns the channel capability for the given portID and channelID.
	// The capability must exist, otherwise testing will fail.
	GetChannelCapability(portID, channelID string) *capabilitytypes.Capability
	// GetTimeoutHeight is a convenience function which returns a IBC packet timeout height
	// to be used for testing. It returns the current IBC height + 100 blocks
	GetTimeoutHeight() clienttypes.Height
}
