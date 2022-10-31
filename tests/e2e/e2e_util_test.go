package e2e

import (
	"bytes"
	"context"
	"time"

	"github.com/ory/dockertest/v3/docker"
)

// func (s *IntegrationTestSuite) connectIBCChains() {
// 	s.T().Logf("connecting %s and %s chains via IBC", s.chains[0].ChainMeta.ID, s.chains[1].ChainMeta.ID)

// 	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
// 	defer cancel()

// 	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
// 		Context:      ctx,
// 		AttachStdout: true,
// 		AttachStderr: true,
// 		Container:    s.hermesResource.Container.ID,
// 		User:         "root",
// 		Cmd: []string{
// 			"hermes",
// 			"create",
// 			"channel",
// 			s.chains[0].ChainMeta.ID,
// 			s.chains[1].ChainMeta.ID,
// 			"--port-a=transfer",
// 			"--port-b=transfer",
// 		},
// 	})
// 	s.Require().NoError(err)

// 	var (
// 		outBuf bytes.Buffer
// 		errBuf bytes.Buffer
// 	)

// 	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
// 		Context:      ctx,
// 		Detach:       false,
// 		OutputStream: &outBuf,
// 		ErrorStream:  &errBuf,
// 	})
// 	s.Require().NoErrorf(
// 		err,
// 		"failed connect chains; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
// 	)

// 	s.Require().Containsf(
// 		errBuf.String(),
// 		"successfully opened init channel",
// 		"failed to connect chains via IBC: %s", errBuf.String(),
// 	)

// 	s.T().Logf("connected %s and %s chains via IBC", s.chains[0].ChainMeta.ID, s.chains[1].ChainMeta.ID)
// }

// func (s *IntegrationTestSuite) sendIBC(srcChainID, dstChainID, recipient string, token sdk.Coin) {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
// 	defer cancel()

// 	s.T().Logf("sending %s from %s to %s (%s)", token, srcChainID, dstChainID, recipient)

// 	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
// 		Context:      ctx,
// 		AttachStdout: true,
// 		AttachStderr: true,
// 		Container:    s.hermesResource.Container.ID,
// 		User:         "root",
// 		Cmd: []string{
// 			"hermes",
// 			"tx",
// 			"raw",
// 			"ft-transfer",
// 			dstChainID,
// 			srcChainID,
// 			"transfer",  // source chain port ID
// 			"channel-0", // since only one connection/channel exists, assume 0
// 			token.Amount.String(),
// 			fmt.Sprintf("--denom=%s", token.Denom),
// 			fmt.Sprintf("--receiver=%s", recipient),
// 			"--timeout-height-offset=1000",
// 		},
// 	})
// 	s.Require().NoError(err)

// 	var (
// 		outBuf bytes.Buffer
// 		errBuf bytes.Buffer
// 	)

// 	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
// 		Context:      ctx,
// 		Detach:       false,
// 		OutputStream: &outBuf,
// 		ErrorStream:  &errBuf,
// 	})
// 	s.Require().NoErrorf(
// 		err,
// 		"failed to send IBC tokens; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
// 	)

// 	s.T().Log("successfully sent IBC tokens")
// }

// func (s *IntegrationTestSuite) fundCommunityPool(c *chain.Chain) {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
// 	defer cancel()

// 	s.T().Logf("funding community pool for chain-id: %s", c.ChainMeta.ID)
// 	for i := range c.Validators {
// 		exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
// 			Context:      ctx,
// 			AttachStdout: true,
// 			AttachStderr: true,
// 			Container:    s.valResources[c.ChainMeta.ID][i].Container.ID,
// 			User:         "root",
// 			Cmd: []string{
// 				"/usr/bin/evmosd",
// 				"--home",
// 				"/evmos/.evmosd",
// 				"tx",
// 				"distribution",
// 				"fund-community-pool",
// 				"93590289356801768542679aevmos",
// 				"--from=val",
// 				fmt.Sprintf("--chain-id=%s", c.ChainMeta.ID),
// 				"-b=block",
// 				"--yes",
// 				"--keyring-backend=test",
// 			},
// 		})
// 		s.Require().NoError(err)

// 		var (
// 			outBuf bytes.Buffer
// 			errBuf bytes.Buffer
// 		)

// 		err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
// 			Context:      ctx,
// 			Detach:       false,
// 			OutputStream: &outBuf,
// 			ErrorStream:  &errBuf,
// 		})

// 		s.Require().NoErrorf(
// 			err,
// 			"failed to fund community pool; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
// 		)

// 		s.Require().Truef(
// 			strings.Contains(outBuf.String(), "code: 0"),
// 			"tx returned non code 0"+outBuf.String(),
// 		)

// 		s.T().Logf("successfully funded community pool on container: %s", s.valResources[c.ChainMeta.ID][i].Container.ID)
// 	}
// }

func (s *IntegrationTestSuite) migrateGenesis() []byte {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exec, err := s.upgradeManager.Client().CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.upgradeManager.ContainerID(),
		User:         "root",
		Cmd: []string{
			"evmosd",
			"migrate",
			s.upgradeParams.TargetVersion,
			"/evmos/.evmosd/config/genesis.json",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	outBuf, errBuf, err = s.upgradeManager.RunExec(ctx, exec.ID)

	s.Require().NoErrorf(
		err,
		"failed to query height; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	return outBuf.Bytes()

}
