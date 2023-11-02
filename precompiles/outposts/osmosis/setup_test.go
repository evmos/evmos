// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package osmosis_test

import (
	"testing"

	"github.com/evmos/evmos/v15/precompiles/outposts/osmosis"

	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"

	"github.com/stretchr/testify/suite"
)

const (
	portID    = "transfer"
	channelID = "channel-0"
)

type PrecompileTestSuite struct {
	suite.Suite

	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	precompile *osmosis.Precompile
}

func (s *PrecompileTestSuite) SetupTest() {
	keyring := testkeyring.New(1)
	integrationNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
	txFactory := factory.New(integrationNetwork, grpcHandler)

	precompile, err := osmosis.NewPrecompile(
		portID,
		channelID,
		osmosis.XCSContract,
		integrationNetwork.App.BankKeeper,
		integrationNetwork.App.TransferKeeper,
		integrationNetwork.App.StakingKeeper,
		integrationNetwork.App.Erc20Keeper,
	)
	s.Require().NoError(err)

	s.network = integrationNetwork
	s.factory = txFactory
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.precompile = precompile
}

func TestPrecompileTestSuite(t *testing.T) {
	suite.Run(t, new(PrecompileTestSuite))
}
