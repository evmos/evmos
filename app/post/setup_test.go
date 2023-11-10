// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package post_test

import (
	"testing"

	"github.com/evmos/evmos/v15/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"

	"github.com/stretchr/testify/suite"
)

type PostTestSuite struct {
	suite.Suite

	network     *network.UnitTestNetwork
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring
}

func (s *PostTestSuite) SetupTest() {
	keyring := testkeyring.New(1)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(unitNetwork)

	s.network = unitNetwork
	s.grpcHandler = grpcHandler
	s.keyring = keyring
}

func TestPrecompileTestSuite(t *testing.T) {
	suite.Run(t, new(PostTestSuite))
}
