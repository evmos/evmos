// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package ics20_test

import (
	"testing"
	"time"

	evmosibc "github.com/evmos/evmos/v18/ibc/testing"
	"github.com/evmos/evmos/v18/precompiles/ics20"
	commonnetwork "github.com/evmos/evmos/v18/testutil/integration/common/network"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v18/testutil/integration/ibc/coordinator"
	"github.com/evmos/evmos/v18/x/evm/statedb"

	"github.com/stretchr/testify/suite"
)

// TODO remove
var s *PrecompileTestSuite

type PrecompileTestSuite struct {
	suite.Suite

	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring
	bondDenom   string

	precompile *ics20.Precompile
	stateDB    *statedb.StateDB

	coordinator  *coordinator.IntegrationCoordinator
	transferPath *evmosibc.Path

	defaultExpirationDuration time.Time

	// TODO remove??
	suiteIBCTesting bool
}

func TestPrecompileTestSuite(t *testing.T) {
	s := new(PrecompileTestSuite)
	suite.Run(t, s)
}

func (s *PrecompileTestSuite) SetupTest() {
	keyring := testkeyring.New(2)
	nw := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	precompile, err := ics20.NewPrecompile(nw.App.TransferKeeper, nw.App.IBCKeeper.ChannelKeeper, nw.App.AuthzKeeper)
	if err != nil {
		panic(err)
	}

	s.precompile = precompile

	grpcHandler := grpc.NewIntegrationHandler(nw)

	s.network = nw
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.precompile = precompile

	IBCCoordinator := coordinator.NewIntegrationCoordinator(
		s.T(),
		[]commonnetwork.Network{s.network},
	)

	acc, err := s.grpcHandler.GetAccount(s.keyring.GetAccAddr(0).String())
	if err != nil {
		panic(err)
	}

	IBCCoordinator.SetDefaultSignerForChain(s.network.GetChainID(), s.keyring.GetPrivKey(0), acc)
	chainA := s.network.GetChainID()
	chainB := IBCCoordinator.GetDummyChainsIDs()[0]
	IBCCoordinator.Setup(chainA, chainB)

	err = IBCCoordinator.CommitAll()
	if err != nil {
		panic(err)
	}

	s.coordinator = IBCCoordinator
	s.transferPath = IBCCoordinator.GetPath(chainA, chainB)
}
