package proposals

import (
	"cosmossdk.io/log"
	cometabci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ProposalHandler is responsible primarily for:
//  1. Filling a proposal with transactions.
//  2. Injecting vote extensions into the proposal (if vote extensions are enabled).
//  3. Verifying that the vote extensions injected are valid.
//
// To verify the validity of the vote extensions, the proposal handler will
// call the validateVoteExtensionsFn. This function is responsible for verifying
// that the vote extensions included in the proposal are valid and compose a
// supermajority of signatures and vote extensions for the current block.
// The given VoteExtensionCodec must be the same used by the VoteExtensionHandler,
// the extended commit is decoded in accordance with the given ExtendedCommitCodec.
type ProposalHandler struct {
	logger log.Logger

	// prepareProposalHandler fills a proposal with transactions.
	prepareProposalHandler sdk.PrepareProposalHandler

	// processProposalHandler processes transactions in a proposal.
	processProposalHandler sdk.ProcessProposalHandler

	// validateVoteExtensionsFn validates the vote extensions included in a proposal.
	validateVoteExtensionsFn ve.ValidateVoteExtensionsFn

	// voteExtensionCodec is used to decode vote extensions.
	voteExtensionCodec codec.VoteExtensionCodec

	// extendedCommitCodec is used to decode extended commit info.
	extendedCommitCodec codec.ExtendedCommitCodec

	// retainOracleDataInWrappedHandler is a flag that determines whether the
	// proposal handler should pass the injected extended commit info to the
	// wrapped proposal handler.
	retainOracleDataInWrappedHandler bool
}

// PrepareProposalHandler returns a PrepareProposalHandler that will be called
// by base app when a new block proposal is requested. The PrepareProposalHandler
// will first fill the proposal with transactions. Then, if vote extensions are
// enabled, the handler will inject the extended commit info into the proposal.
func (h *ProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *cometabci.RequestPrepareProposal) (resp *cometabci.ResponsePrepareProposal, err error) {
		return &cometabci.ResponsePrepareProposal{Txs: make([][]byte, 0)}, nil
	}
}
