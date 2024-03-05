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
//
//	commits the changes to have a block time after the given duration.
func (n *IntegrationNetwork) NextBlockAfter(duration time.Duration) error {
	header := n.ctx.BlockHeader()
	// Update block header and BeginBlock
	header.Height++
	header.AppHash = n.app.LastCommitID().Hash
	// Calculate new block time after duration
	newBlockTime := header.Time.Add(duration)
	header.Time = newBlockTime

	// add validator's commit info to allocate corresponding tokens to validators
	ci := getCommitInfo(n.valSet.Validators)

	// FinalizeBlock to run endBlock, deliverTx & beginBlock logic
	req := &abcitypes.RequestFinalizeBlock{
		Height:             n.app.LastBlockHeight() + 1,
		DecidedLastCommit:  ci,
		Hash:               header.AppHash,
		NextValidatorsHash: n.valSet.Hash(),
		ProposerAddress:    n.valSet.Proposer.Address,
		Time:               newBlockTime,
	}

	if _, err := n.app.FinalizeBlock(req); err != nil {
		return err
	}

	newCtx := n.app.BaseApp.NewContextLegacy(false, header)

	// Update context header
	newCtx = newCtx.WithMinGasPrices(n.ctx.MinGasPrices())
	newCtx = newCtx.WithKVGasConfig(n.ctx.KVGasConfig())
	newCtx = newCtx.WithTransientKVGasConfig(n.ctx.TransientKVGasConfig())
	newCtx = newCtx.WithConsensusParams(n.ctx.ConsensusParams())
	// This might have to be changed with time if we want to test gas limits
	newCtx = newCtx.WithBlockGasMeter(storetypes.NewInfiniteGasMeter())
	newCtx = newCtx.WithVoteInfos(ci.GetVotes())
	n.ctx = newCtx

	// commit changes
	_, err := n.app.Commit()

	return err
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
