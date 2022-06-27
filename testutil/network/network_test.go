//go:build norace
// +build norace

package network_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/evmos/ethermint/server/config"
	"github.com/evmos/ethermint/testutil/network"

	evmosnetwork "github.com/evmos/evmos/v6/testutil/network"
)

type IntegrationTestSuite struct {
	suite.Suite

	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	var err error
	cfg := evmosnetwork.DefaultConfig()
	cfg.JSONRPCAddress = config.DefaultJSONRPCAddress
	cfg.NumValidators = 1

	s.network, err = network.New(s.T(), s.T().TempDir(), cfg)
	s.Require().NoError(err)
	s.Require().NotNil(s.network)

	_, err = s.network.WaitForHeight(2)
	s.Require().NoError(err)

	if s.network.Validators[0].JSONRPCClient == nil {
		address := fmt.Sprintf("http://%s", s.network.Validators[0].AppConfig.JSONRPC.Address)
		s.network.Validators[0].JSONRPCClient, err = ethclient.Dial(address)
		s.Require().NoError(err)
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestNetwork_Liveness() {
	h, err := s.network.WaitForHeightWithTimeout(10, time.Minute)
	s.Require().NoError(err, "expected to reach 10 blocks; got %d", h)

	latestHeight, err := s.network.LatestHeight()
	s.Require().NoError(err, "latest height failed")
	s.Require().GreaterOrEqual(latestHeight, h)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
