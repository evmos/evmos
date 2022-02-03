package integration_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"

	"github.com/stretchr/testify/suite"

	"github.com/tharsis/ethermint/server/config"
	"github.com/tharsis/ethermint/testutil/network"

	evmosnetwork "github.com/tharsis/evmos/testutil/network"
	"github.com/tharsis/evmos/x/incentives/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx             context.Context
	cfg             network.Config
	network         *network.Network
	grpcQueryClient types.QueryClient
	// grpcTxClient    types.MsgClient
}

var s *IntegrationTestSuite

func TestIntegration(t *testing.T) {
	s = new(IntegrationTestSuite)
	suite.Run(t, s)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = BeforeSuite(func() {
	s.SetupTest()
})

var _ = AfterSuite(func() {
	s.TearDownSuite()
})

func (s *IntegrationTestSuite) SetupTest() {
	s.T().Log("setting up integration test suite")

	var err error
	cfg := evmosnetwork.DefaultConfig()
	cfg.JSONRPCAddress = config.DefaultJSONRPCAddress
	cfg.NumValidators = 1

	s.ctx = context.Background()
	s.cfg = cfg
	s.network, err = network.New(s.T(), s.T().TempDir(), cfg)
	s.Require().NoError(err)
	s.Require().NotNil(s.network)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	if s.network.Validators[0].JSONRPCClient == nil {
		address := fmt.Sprintf("http://%s", s.network.Validators[0].AppConfig.JSONRPC.Address)
		s.network.Validators[0].JSONRPCClient, err = ethclient.Dial(address)
		s.Require().NoError(err)
	}

	// Create a connection to the gRPC server.
	grpcConn, err := grpc.Dial(
		s.network.Validators[0].AppConfig.GRPC.Address, // gRPC server address.
		grpc.WithInsecure(),                            // nosemgrep
	)
	s.Require().NoError(err)

	s.grpcQueryClient = types.NewQueryClient(grpcConn)

	// FIXME: "unknown service evmos.erc20.v1.Msg"
	// s.grpcTxClient = types.NewMsgClient(grpcConn)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}
