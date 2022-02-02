package integration_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/tharsis/ethermint/server/config"
	"github.com/tharsis/ethermint/testutil/network"

	evmosnetwork "github.com/tharsis/evmos/testutil/network"
	"github.com/tharsis/evmos/x/erc20/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx             context.Context
	cfg             network.Config
	network         *network.Network
	grpcQueryClient types.QueryClient
	grpcTxClient    types.MsgClient
}

func TestIntegration(t *testing.T) {
	// TODO fix on ethermint
	// t.Skip()
	// suite.Run(t, new(IntegrationTestSuite))

	s := new(IntegrationTestSuite)
	s.SetupTest() // reset

	// Run Ginkgo Testing Suite
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")

	s.TearDownSuite()
}

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
	s.grpcTxClient = types.NewMsgClient(grpcConn)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")

	// This is important and must be called to ensure other tests can create
	// a network!
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestLiveness() {
	// test the gRPC query client to check if everything's ok
	resParams, err := s.grpcQueryClient.Params(s.ctx, &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().NotNil(resParams)

	// FIXME: enable
	// res, err := s.grpcTxClient.ConvertCoin(s.ctx, &types.MsgConvertCoin{})
	// s.Require().NoError(err)
	// s.Require().NotNil(res)
}
