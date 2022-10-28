package e2e

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ory/dockertest/v3/docker"

	"github.com/evmos/evmos/v9/tests/e2e/chain"
)

func (s *IntegrationTestSuite) connectIBCChains() {
	s.T().Logf("connecting %s and %s chains via IBC", s.chains[0].ChainMeta.ID, s.chains[1].ChainMeta.ID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.hermesResource.Container.ID,
		User:         "root",
		Cmd: []string{
			"hermes",
			"create",
			"channel",
			s.chains[0].ChainMeta.ID,
			s.chains[1].ChainMeta.ID,
			"--port-a=transfer",
			"--port-b=transfer",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoErrorf(
		err,
		"failed connect chains; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.Require().Containsf(
		errBuf.String(),
		"successfully opened init channel",
		"failed to connect chains via IBC: %s", errBuf.String(),
	)

	s.T().Logf("connected %s and %s chains via IBC", s.chains[0].ChainMeta.ID, s.chains[1].ChainMeta.ID)
}

func (s *IntegrationTestSuite) sendIBC(srcChainID, dstChainID, recipient string, token sdk.Coin) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("sending %s from %s to %s (%s)", token, srcChainID, dstChainID, recipient)

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.hermesResource.Container.ID,
		User:         "root",
		Cmd: []string{
			"hermes",
			"tx",
			"raw",
			"ft-transfer",
			dstChainID,
			srcChainID,
			"transfer",  // source chain port ID
			"channel-0", // since only one connection/channel exists, assume 0
			token.Amount.String(),
			fmt.Sprintf("--denom=%s", token.Denom),
			fmt.Sprintf("--receiver=%s", recipient),
			"--timeout-height-offset=1000",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoErrorf(
		err,
		"failed to send IBC tokens; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.T().Log("successfully sent IBC tokens")
}

func (s *IntegrationTestSuite) submitProposal(c *chain.Chain, upgradeVersion string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("submitting upgrade proposal for chain-id: %s", c.ChainMeta.ID)
	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.valResources[c.ChainMeta.ID][0].Container.ID,
		User:         "root",
		Cmd: []string{
			"/usr/bin/evmosd",
			"--home",
			"/evmos/.evmosd",
			"tx", "gov", "submit-proposal",
			"software-upgrade", upgradeVersion,
			"--title=\"TEST\"",
			"--description=\"test upgrade proposal\"",
			"--upgrade-height=50",
			"--upgrade-info=\"\"",
			fmt.Sprintf("--chain-id=%s", c.ChainMeta.ID),
			"--from=val", "-b=block",
			"--yes", "--keyring-backend=test",
			"--log_format=json",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})

	s.Require().NoErrorf(
		err,
		"failed to submit proposal; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.Require().Truef(
		strings.Contains(outBuf.String(), "code: 0"),
		"tx returned non code 0",
	)

	s.T().Log("successfully submitted proposal")
}

func (s *IntegrationTestSuite) depositProposal(c *chain.Chain) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("depositing to upgrade proposal for chain-id: %s", c.ChainMeta.ID)
	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.valResources[c.ChainMeta.ID][0].Container.ID,
		User:         "root",
		Cmd: []string{
			"/usr/bin/evmosd",
			"--home",
			"/evmos/.evmosd",
			"tx",
			"gov",
			"deposit",
			"1",
			"10000000aevmos",
			"--from=val",
			fmt.Sprintf("--chain-id=%s", c.ChainMeta.ID),
			"-b=block",
			"--yes",
			"--keyring-backend=test",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})

	s.Require().NoErrorf(
		err,
		"failed to deposit to upgrade proposal; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.Require().Truef(
		strings.Contains(outBuf.String(), "code: 0"),
		"tx returned non code 0",
	)

	s.T().Log("successfully deposited to proposal")

}

func (s *IntegrationTestSuite) voteProposal(c *chain.Chain) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("voting for upgrade proposal for chain-id: %s", c.ChainMeta.ID)
	for i := range c.Validators {
		exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
			Context:      ctx,
			AttachStdout: true,
			AttachStderr: true,
			Container:    s.valResources[c.ChainMeta.ID][i].Container.ID,
			User:         "root",
			Cmd: []string{
				"/usr/bin/evmosd",
				"--home",
				"/evmos/.evmosd",
				"tx",
				"gov",
				"vote",
				"1",
				"yes",
				"--from=val",
				fmt.Sprintf("--chain-id=%s", c.ChainMeta.ID),
				"-b=block",
				"--yes",
				"--keyring-backend=test",
			},
		})
		s.Require().NoError(err)

		var (
			outBuf bytes.Buffer
			errBuf bytes.Buffer
		)

		err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
			Context:      ctx,
			Detach:       false,
			OutputStream: &outBuf,
			ErrorStream:  &errBuf,
		})

		s.Require().NoErrorf(
			err,
			"failed to vote for proposal; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
		)

		s.Require().Truef(
			strings.Contains(outBuf.String(), "code: 0"),
			"tx returned non code 0",
		)

		s.T().Logf("successfully voted for proposal on container: %s", s.valResources[c.ChainMeta.ID][i].Container.ID)
	}
}

func (s *IntegrationTestSuite) fundCommunityPool(c *chain.Chain) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("funding community pool for chain-id: %s", c.ChainMeta.ID)
	for i := range c.Validators {
		exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
			Context:      ctx,
			AttachStdout: true,
			AttachStderr: true,
			Container:    s.valResources[c.ChainMeta.ID][i].Container.ID,
			User:         "root",
			Cmd: []string{
				"/usr/bin/evmosd",
				"--home",
				"/evmos/.evmosd",
				"tx",
				"distribution",
				"fund-community-pool",
				"93590289356801768542679aevmos",
				"--from=val",
				fmt.Sprintf("--chain-id=%s", c.ChainMeta.ID),
				"-b=block",
				"--yes",
				"--keyring-backend=test",
			},
		})
		s.Require().NoError(err)

		var (
			outBuf bytes.Buffer
			errBuf bytes.Buffer
		)

		err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
			Context:      ctx,
			Detach:       false,
			OutputStream: &outBuf,
			ErrorStream:  &errBuf,
		})

		s.Require().NoErrorf(
			err,
			"failed to fund community pool; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
		)

		s.Require().Truef(
			strings.Contains(outBuf.String(), "code: 0"),
			"tx returned non code 0"+outBuf.String(),
		)

		s.T().Logf("successfully funded community pool on container: %s", s.valResources[c.ChainMeta.ID][i].Container.ID)
	}
}

func (s *IntegrationTestSuite) chainStatus(containerId string) (int, []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    containerId,
		User:         "root",
		Cmd: []string{
			"/usr/bin/evmosd",
			"--home",
			"/evmos/.evmosd",
			"q",
			"block",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})

	s.Require().NoErrorf(
		err,
		"failed to query height; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	index := strings.Index(outBuf.String(), "\"height\":")
	qq := outBuf.String()[index+10 : index+12]
	h, _ := strconv.Atoi(qq)

	errBufByte := errBuf.Bytes()
	return h, errBufByte

}

func (s *IntegrationTestSuite) migrateGenesis(containerId string) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    containerId,
		User:         "root",
		Cmd: []string{
			"/usr/bin/evmosd",
			"--home",
			"/evmos/.evmosd",
			"migrate",
			s.upgradeParams.PostUpgradeVersion,
			"/evmos/.evmosd/config/genesis.json",
			"--chain-id=evmos_9001-1",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})

	s.Require().NoErrorf(
		err,
		"failed to query height; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	return outBuf.Bytes()

}
