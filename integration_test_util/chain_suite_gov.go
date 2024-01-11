package integration_test_util

//goland:noinspection SpellCheckingInspection
import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypeslegacy "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	itutiltypes "github.com/evmos/evmos/v16/integration_test_util/types"
	"math"
	"time"
)

// TxFullGov submit gov proposal, full vote Yes and wait gov passed.
func (suite *ChainIntegrationTestSuite) TxFullGov(proposer *itutiltypes.TestAccount, newProposalContent govtypeslegacy.Content) uint64 {
	suite.Require().NotNil(proposer)

	depositAmount := sdk.NewInt(int64(0.1 * math.Pow10(18)))
	msg, err := govtypeslegacy.NewMsgSubmitProposal(newProposalContent, sdk.NewCoins(
		sdk.NewCoin(suite.ChainConstantsConfig.GetMinDenom(), depositAmount),
	), proposer.GetCosmosAddress())
	suite.Require().NoError(err)

	_, _, err = suite.DeliverTx(suite.CurrentContext, proposer, nil, msg)
	suite.Require().NoError(err)

	suite.Commit()

	proposal := suite.QueryLatestGovProposal(proposer)
	suite.Require().NotNilf(proposal, "proposal could not be found")

	suite.Require().Equalf(govv1types.StatusVotingPeriod, proposal.Status, "proposal must be in voting period")

	suite.TxAllVote(proposal.Id, govv1types.OptionYes)
	suite.Commit()
	if suite.HasTendermint() {
		time.Sleep(itutiltypes.TendermintGovVotingPeriod + 200*time.Millisecond)
	}

	var proposalById *govv1types.Proposal

	for proposalById == nil || proposalById.Status == govv1types.StatusDepositPeriod || proposalById.Status == govv1types.StatusVotingPeriod {
		proposalById = suite.QueryGovProposalById(proposal.Id)
		suite.Commit()
	}

	suite.Require().Equalf(govv1types.StatusPassed, proposalById.Status, "proposal must be passed")

	return proposal.Id
}

// TxVote submits a vote on given proposal.
func (suite *ChainIntegrationTestSuite) TxVote(voter *itutiltypes.TestAccount, proposalId uint64, option govv1types.VoteOption) error {
	suite.Require().NotNil(voter)

	_, _, err := suite.DeliverTx(suite.CurrentContext, voter, nil, &govv1types.MsgVote{
		ProposalId: proposalId,
		Voter:      voter.GetCosmosAddress().String(),
		Option:     option,
	})

	return err
}

// TxAllVote using all accounts, each submits a vote on given proposal.
func (suite *ChainIntegrationTestSuite) TxAllVote(proposalId uint64, option govv1types.VoteOption) {
	voted := make(map[string]bool)

	for _, voter := range append(suite.WalletAccounts, suite.ValidatorAccounts...) {
		if voted[voter.GetCosmosAddress().String()] {
			continue
		}
		err := suite.TxVote(voter, proposalId, option)
		suite.Require().NoErrorf(err, "voter %s could not vote", voter.GetCosmosAddress().String())
		voted[voter.GetCosmosAddress().String()] = true
	}

	return
}

// QueryLatestGovProposal returns the latest gov proposal submitted by given proposer.
func (suite *ChainIntegrationTestSuite) QueryLatestGovProposal(proposer *itutiltypes.TestAccount) *govv1types.Proposal {
	suite.Require().NotNil(proposer)

	resProposals, err := suite.QueryClients.GovV1.Proposals(suite.CurrentContext, &govv1types.QueryProposalsRequest{
		Depositor: proposer.GetCosmosAddress().String(),
	})
	suite.Require().NoError(err)
	suite.Require().NotNilf(resProposals, "proposal could not be found")
	if len(resProposals.Proposals) < 1 {
		return nil
	}
	if len(resProposals.Proposals) == 1 {
		return resProposals.Proposals[0]
	}

	var latestProposal *govv1types.Proposal
	for _, proposal := range resProposals.Proposals {
		if latestProposal == nil || latestProposal.Id < proposal.Id {
			latestProposal = proposal
		}
	}
	return latestProposal
}

// QueryGovProposalById returns the gov proposal of the given proposal id.
func (suite *ChainIntegrationTestSuite) QueryGovProposalById(id uint64) *govv1types.Proposal {
	resProposal, err := suite.QueryClients.GovV1.Proposal(suite.CurrentContext, &govv1types.QueryProposalRequest{
		ProposalId: id,
	})
	suite.Require().NoError(err)
	suite.Require().NotNilf(resProposal, "proposal could not be found")
	suite.Require().NotNilf(resProposal.Proposal, "proposal could not be found")
	return resProposal.Proposal
}
