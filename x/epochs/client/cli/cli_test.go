package cli_test

import (
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"

	"github.com/tharsis/ethermint/testutil/network"

	evmosnetwork "github.com/tharsis/evmos/testutil/network"
	"github.com/tharsis/evmos/x/epochs/client/cli"
	"github.com/tharsis/evmos/x/epochs/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewIntegrationTestSuite(cfg network.Config) *IntegrationTestSuite {
	return &IntegrationTestSuite{cfg: cfg}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	var err error
	s.cfg = evmosnetwork.DefaultConfig()
	s.cfg.NumValidators = 1

	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)
	s.Require().NotNil(s.network)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestGetCmdCurrentEpoch() {
	val := s.network.Validators[0]

	testCases := []struct {
		name       string
		identifier string
		expectErr  bool
		respType   proto.Message
	}{
		{
			"query weekly epoch number",
			"week",
			false,
			&types.QueryCurrentEpochResponse{},
		},
		{
			"query unavailable epoch number",
			"unavailable",
			true,
			&types.QueryCurrentEpochResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdCurrentEpoch()
			clientCtx := val.ClientCtx

			args := []string{
				tc.identifier,
			}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdEpochsInfos() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		expectErr bool
		respType  proto.Message
	}{
		{
			"query epoch infos",
			false, &types.QueryEpochsInfoResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdEpochsInfos()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{})
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
			}
		})
	}
}
