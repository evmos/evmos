package e2e

import (
	"context"
	"strings"

	"github.com/evmos/evmos/v11/tests/e2e/upgrade"
)

// TestUpgrade tests if an Evmos node can be upgraded from one version to another.
// It iterates through the list of scheduled upgrades, that are defined using the input
// arguments to the make command. The function then submits a proposal to upgrade the chain,
// and finally upgrades the chain.
// If the chain can be restarted after the upgrade(s), the test passes.
func (s *IntegrationTestSuite) TestUpgrade() {
	for idx, version := range s.upgradeParams.Versions {
		if idx == 0 {
			// start initial node
			s.runInitialNode(version, registryDockerFile)
			continue
		}
		s.T().Logf("(upgrade %d): UPGRADING TO %s WITH PROPOSAL NAME %s", idx, version.ImageTag, version.UpgradeName)
		s.proposeUpgrade(version.UpgradeName, version.ImageTag)
		s.voteForProposal(idx)
		s.upgrade(version.ImageName, version.ImageTag)
	}
	s.T().Logf("SUCCESS")
}

func (s *IntegrationTestSuite) TestCLITxs() {
	mainBranch := upgrade.VersionConfig{
		ImageTag:  "main",
		ImageName: "evmos",
	}

	s.runInitialNode(mainBranch, repoDockerFile)

	testCases := []struct {
		name      string
		cmd       func() (string, error)
		expPass   bool
		expErrMsg string
	}{
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
			expErrMsg: "insufficient fee", // TODO when the PR https://github.com/evmos/evmos/pull/1386 is merged, this may fail with out of gas error
			// when the PR https://github.com/evmos/cosmos-sdk/pull/8 on cosmos-sdk and included on this repo, will get an error that cannot define gas flag when using fees=auto (which is the default)
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
		// TODO uncomment this test when the PR https://github.com/evmos/evmos/pull/1386 is merged - will fail with error saying cannot use --gas flag with fees=auto
		// {
		// 	name: "fail - submit upgrade proposal, no fees (defaults to 'auto') & sufficient gas",
		// 	cmd: func() (string, error) {
		// 		return s.upgradeManager.CreateSubmitProposalExec(
		// 			"v11.0.0",
		// 			s.upgradeParams.ChainID,
		// 			5000,
		// 			true,
		// 			"--gas=1500000",
		// 		)
		// 	},
		// 	expPass: false,
		// },
		// TODO uncomment these tests when the PR https://github.com/evmos/cosmos-sdk/pull/8 on cosmos-sdk is merged and that version is used on Evmos
		// {
		// 	name: "success - submit upgrade proposal, no fees (defaults to 'auto')",
		// 	cmd: func() (string, error) {
		// 		return s.upgradeManager.CreateSubmitProposalExec(
		// 			"v11.0.0",
		// 			s.upgradeParams.ChainID,
		// 			5000,
		// 			true,
		// 		)
		// 	},
		// 	expPass: true,
		// },
		// {
		// 	name: "success - submit upgrade proposal, gas 'auto'",
		// 	cmd: func() (string, error) {
		// 		return s.upgradeManager.CreateSubmitProposalExec(
		// 			"v11.0.0",
		// 			s.upgradeParams.ChainID,
		// 			5000,
		// 			true,
		// 			"--gas=auto",
		// 		)
		// 	},
		// 	expPass: true,
		// },
		// {
		// 	name: "success - submit upgrade proposal, fees 'auto'",
		// 	cmd: func() (string, error) {
		// 		return s.upgradeManager.CreateSubmitProposalExec(
		// 			"v11.0.0",
		// 			s.upgradeParams.ChainID,
		// 			5000,
		// 			true,
		// 			"--fees=auto",
		// 		)
		// 	},
		// 	expPass: true,
		// },
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
			name: "fail - vote upgrade proposal, insufficient gas",
			cmd: func() (string, error) {
				return s.upgradeManager.CreateVoteProposalExec(
					s.upgradeParams.ChainID,
					1,
					"--fees=10000000000000000aevmos",
					"--gas=100",
				)
			},
			expPass:   false,
			expErrMsg: "out of gas",
		},
		{
			name: "success - vote upgrade proposal, defined gas and fees",
			cmd: func() (string, error) {
				return s.upgradeManager.CreateVoteProposalExec(
					s.upgradeParams.ChainID,
					1,
					"--fees=10000000000000000aevmos",
					"--gas=500000",
				)
			},
			expPass: true,
		},
		// TODO uncomment these tests when the PR https://github.com/evmos/cosmos-sdk/pull/8 on cosmos-sdk is merged and that version is used on Evmos
		// {
		// 	name: "success - vote upgrade proposal, gas 'auto'",
		// 	cmd: func() (string, error) {
		// 		return s.upgradeManager.CreateVoteProposalExec(
		// 		s.upgradeParams.ChainID,
		// 		1,
		// 		"--gas=auto",
		// 	  )
		// 	},
		// 	expPass:   true,
		// },
		// {
		// 	name: "success - vote upgrade proposal, fees 'auto'",
		// 	cmd: func() (string, error) {
		// 		return s.upgradeManager.CreateVoteProposalExec(
		// 		s.upgradeParams.ChainID,
		// 		1,
		// 		"--fees=auto",
		// 	)
		// 	},
		// 	expPass:   true,
		// },
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
					strings.Contains(outBuf.String(), tc.expErrMsg),
					"tx returned code 0 but with unexpected error:\nstdout: %s\nstderr: %s", outBuf.String(), errBuf.String(),
				)
			}
		})
	}
}
