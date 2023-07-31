package e2e

import (
	"context"
	"strings"
)

// TestUpgrade tests if an Evmos node can be upgraded from one version to another.
// It iterates through the list of scheduled upgrades, that are defined using the input
// arguments to the make command. The function then submits a proposal to upgrade the chain,
// and finally upgrades the chain.
// If the chain can be restarted after the upgrade(s), the test passes.
func (s *IntegrationTestSuite) TestUpgrade() {
	versionMap := map[string]string{
		"v12.1.0": "v12.1.5",
	}
	for idx, version := range s.upgradeParams.Versions {
		if overwriteTag, ok := versionMap[version.UpgradeName]; ok {
			version.ImageTag = overwriteTag
			version.UpgradeName = overwriteTag
		}
		if idx == 0 {
			// start initial node
			s.runInitialNode(version)
			continue
		}
		s.T().Logf("(upgrade %d): UPGRADING TO %s WITH PROPOSAL NAME %s", idx, version.ImageTag, version.UpgradeName)
		s.proposeUpgrade(version.UpgradeName, version.ImageTag)
		s.voteForProposal(idx)
		s.upgrade(version.ImageName, version.ImageTag)
	}
	s.T().Logf("SUCCESS")
}

// TestCLITxs executes different types of transactions against an Evmos node
// using the CLI client. The node used for the test has the latest changes introduced.
func (s *IntegrationTestSuite) TestCLITxs() {
	// start a node
	s.runNodeWithCurrentChanges()

	testCases := []struct {
		name      string
		cmd       func() (string, error)
		expPass   bool
		expErrMsg string
	}{
		{
			name: "fail - submit upgrade proposal, invalid flags combination (gas-prices & fees)",
			cmd: func() (string, error) {
				return s.upgradeManager.CreateSubmitProposalExec(
					"v11.0.0",
					s.upgradeParams.ChainID,
					5000,
					true,
					"--fees=5000000000aevmos",
					"--gas-prices=50000aevmos",
				)
			},
			expPass:   false,
			expErrMsg: "cannot provide both fees and gas prices",
		},
		{
			name: "fail - submit upgrade proposal, no fees & insufficient gas",
			cmd: func() (string, error) {
				return s.upgradeManager.CreateSubmitProposalExec(
					"v11.0.0",
					s.upgradeParams.ChainID,
					5000,
					true,
					"--gas=50000",
				)
			},
			expPass:   false,
			expErrMsg: "gas prices too low",
		},
		{
			name: "fail - submit upgrade proposal, insufficient fees",
			cmd: func() (string, error) {
				return s.upgradeManager.CreateSubmitProposalExec(
					"v11.0.0",
					s.upgradeParams.ChainID,
					5000,
					true,
					"--fees=10aevmos",
					"--gas=50000",
				)
			},
			expPass:   false,
			expErrMsg: "insufficient fee",
		},
		{
			name: "fail - submit upgrade proposal, insufficient gas",
			cmd: func() (string, error) {
				return s.upgradeManager.CreateSubmitProposalExec(
					"v11.0.0",
					s.upgradeParams.ChainID,
					5000,
					true,
					"--fees=500000000000aevmos",
					"--gas=1000",
				)
			},
			expPass:   false,
			expErrMsg: "out of gas",
		},
		{
			name: "success - submit upgrade proposal, defined fees & gas",
			cmd: func() (string, error) {
				return s.upgradeManager.CreateSubmitProposalExec(
					"v11.0.0",
					s.upgradeParams.ChainID,
					5000,
					true,
					"--fees=10000000000000000aevmos",
					"--gas=1500000",
				)
			},
			expPass: true,
		},
		{
			name: "success - submit upgrade proposal, using gas & gas-prices",
			cmd: func() (string, error) {
				return s.upgradeManager.CreateSubmitProposalExec(
					"v11.0.0",
					s.upgradeParams.ChainID,
					5000,
					true,
					"--gas-prices=1000000000aevmos",
					"--gas=1500000",
				)
			},
			expPass: true,
		},
		{
			name: "fail - vote upgrade proposal, insufficient fees",
			cmd: func() (string, error) {
				return s.upgradeManager.CreateVoteProposalExec(
					s.upgradeParams.ChainID,
					1,
					"--fees=10aevmos",
					"--gas=500000",
				)
			},
			expPass:   false,
			expErrMsg: "insufficient fee",
		},
		{
			name: "success - vote upgrade proposal (using gas 'auto' and specific fees)",
			cmd: func() (string, error) {
				return s.upgradeManager.CreateVoteProposalExec(
					s.upgradeParams.ChainID,
					1,
					"--gas=auto",
					"--gas-adjustment=1.5",
					"--fees=10000000000000000aevmos",
				)
			},
			expPass: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			exec, err := tc.cmd()
			s.Require().NoError(err)

			outBuf, errBuf, err := s.upgradeManager.RunExec(ctx, exec)
			s.Require().NoError(err)

			if tc.expPass {
				s.Require().Truef(
					strings.Contains(outBuf.String(), "code: 0"),
					"tx returned non code 0:\nstdout: %s\nstderr: %s", outBuf.String(), errBuf.String(),
				)
			} else {
				s.Require().Truef(
					strings.Contains(outBuf.String(), tc.expErrMsg) || strings.Contains(errBuf.String(), tc.expErrMsg),
					"tx returned non code 0 but with unexpected error:\nstdout: %s\nstderr: %s", outBuf.String(), errBuf.String(),
				)
			}
		})
	}
}
