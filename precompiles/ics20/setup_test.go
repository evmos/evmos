// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package ics20_test

import (
	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	"testing"
	"time"

	"github.com/evmos/evmos/v16/precompiles/ics20"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/suite"
)

var s *PrecompileTestSuite

type PrecompileTestSuite struct {
	suite.Suite

	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	bondDenom  string
	precompile *ics20.Precompile

	defaultExpirationDuration time.Time
	suiteIBCTesting           bool
}

func TestPrecompileTestSuite(t *testing.T) {
	suite.Run(t, new(PrecompileTestSuite))

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "ICS20 Precompile Suite")
}

func (s *PrecompileTestSuite) SetupTest() {
	keyring := testkeyring.New(2)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

	precompile, err := ics20.NewPrecompile(
		unitNetwork.App.TransferKeeper,
		unitNetwork.App.IBCKeeper.ChannelKeeper,
		unitNetwork.App.AuthzKeeper,
	)
	s.Require().NoError(err, "expected no error during precompile creation")

	grpcHandler := grpc.NewIntegrationHandler(unitNetwork)

	ctx := unitNetwork.GetContext()
	bondDenom, err := unitNetwork.App.StakingKeeper.BondDenom(ctx)
	s.Require().NoError(err, "expected no error during bond denom retrieval")

	s.bondDenom = bondDenom
	s.network = unitNetwork
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.precompile = precompile
}
