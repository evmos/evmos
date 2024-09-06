// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package network

import (
	"time"

	storetypes "cosmossdk.io/store/types"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
)

// NextBlock is a private helper function that runs the EndBlocker logic, commits the changes,
// updates the header and runs the BeginBlocker
func (n *IntegrationNetwork) NextBlock() error {
	return n.NextBlockAfter(time.Second)
}

// NextBlockAfter is a private helper function that runs the FinalizeBlock logic, updates the context and
// commits the changes to have a block time after the given duration.
func (n *IntegrationNetwork) NextBlockAfter(duration time.Duration) error {
	_, err := n.finalizeBlockAndCommit(duration)
	return err
}

// NextBlockWithTxs is a helper function that runs the FinalizeBlock logic
// with the provided tx bytes, updates the context and
// commits the changes to have a block time after the given duration.
func (n *IntegrationNetwork) NextBlockWithTxs(txBytes ...[]byte) (*abcitypes.ResponseFinalizeBlock, error) {
	return n.finalizeBlockAndCommit(time.Second, txBytes...)
}

// finalizeBlockAndCommit is a private helper function that runs the FinalizeBlock logic
// with the provided txBytes, updates the context and
// commits the changes to have a block time after the given duration.
func (n *IntegrationNetwork) finalizeBlockAndCommit(duration time.Duration, txBytes ...[]byte) (*abcitypes.ResponseFinalizeBlock, error) {
	header := n.ctx.BlockHeader()
	// Update block header and BeginBlock
	header.Height++
	header.AppHash = n.app.LastCommitID().Hash
	// Calculate new block time after duration
	newBlockTime := header.Time.Add(duration)
	header.Time = newBlockTime

	// FinalizeBlock to run endBlock, deliverTx & beginBlock logic
	req := buildFinalizeBlockReq(header, n.valSet.Validators, txBytes...)

	res, err := n.app.FinalizeBlock(req)
	if err != nil {
		return nil, err
	}

	newCtx := n.app.BaseApp.NewContextLegacy(false, header)

	// Update context header
	newCtx = newCtx.WithMinGasPrices(n.ctx.MinGasPrices())
	newCtx = newCtx.WithKVGasConfig(n.ctx.KVGasConfig())
	newCtx = newCtx.WithTransientKVGasConfig(n.ctx.TransientKVGasConfig())
	newCtx = newCtx.WithConsensusParams(n.ctx.ConsensusParams())
	// This might have to be changed with time if we want to test gas limits
	newCtx = newCtx.WithBlockGasMeter(storetypes.NewInfiniteGasMeter())
	newCtx = newCtx.WithVoteInfos(req.DecidedLastCommit.GetVotes())
	n.ctx = newCtx

	// commit changes
	_, err = n.app.Commit()

	return res, err
}

// buildFinalizeBlockReq is a helper function to build
// properly the FinalizeBlock request
func buildFinalizeBlockReq(header cmtproto.Header, validators []*cmttypes.Validator, txs ...[]byte) *abcitypes.RequestFinalizeBlock {
	// add validator's commit info to allocate corresponding tokens to validators
	ci := getCommitInfo(validators)
	return &abcitypes.RequestFinalizeBlock{
		Height:             header.Height,
		DecidedLastCommit:  ci,
		Hash:               header.AppHash,
		NextValidatorsHash: header.ValidatorsHash,
		ProposerAddress:    header.ProposerAddress,
		Time:               header.Time,
		Txs:                txs,
	}
}

func getCommitInfo(validators []*cmttypes.Validator) abcitypes.CommitInfo {
	voteInfos := make([]abcitypes.VoteInfo, len(validators))
	for i, val := range validators {
		voteInfos[i] = abcitypes.VoteInfo{
			Validator: abcitypes.Validator{
				Address: val.Address,
				Power:   val.VotingPower,
			},
			BlockIdFlag: cmtproto.BlockIDFlagCommit,
		}
	}
	return abcitypes.CommitInfo{Votes: voteInfos}
}
