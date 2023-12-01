// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package osmosis_test

import (
	"testing"

	"github.com/evmos/evmos/v16/precompiles/outposts/osmosis"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	"github.com/stretchr/testify/suite"
)

const (
	PortID    = "transfer"
	ChannelID = "channel-0"
)

type PrecompileTestSuite struct {
	suite.Suite

	unitNetwork *network.UnitTestNetwork
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	precompile *osmosis.Precompile
}

func TestPrecompileTestSuite(t *testing.T) {
	suite.Run(t, new(PrecompileTestSuite))
}

func (s *PrecompileTestSuite) SetupTest() {
	keyring := testkeyring.New(2)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

	precompile, err := osmosis.NewPrecompile(
		unitNetwork.App.AuthzKeeper,
		PortID,
		ChannelID,
		osmosis.XCSContractTestnet,
		unitNetwork.App.BankKeeper,
		unitNetwork.App.TransferKeeper,
		unitNetwork.App.StakingKeeper,
		unitNetwork.App.Erc20Keeper,
		unitNetwork.App.IBCKeeper.ChannelKeeper,
	)
	s.Require().NoError(err, "expected no error during precompile creation")

	grpcHandler := grpc.NewIntegrationHandler(unitNetwork)

	s.unitNetwork = unitNetwork
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.precompile = precompile
}
