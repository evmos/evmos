package staking_test

import (
	"testing"

	"github.com/Eidon-AI/eidon-chain/v20/precompiles/staking"
	"github.com/Eidon-AI/eidon-chain/v20/testutil/integration/eidon-chain/factory"
	"github.com/Eidon-AI/eidon-chain/v20/testutil/integration/eidon-chain/grpc"
	testkeyring "github.com/Eidon-AI/eidon-chain/v20/testutil/integration/eidon-chain/keyring"
	"github.com/Eidon-AI/eidon-chain/v20/testutil/integration/eidon-chain/network"
	"github.com/stretchr/testify/suite"
)

type PrecompileTestSuite struct {
	suite.Suite

	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	bondDenom  string
	precompile *staking.Precompile
}

func TestPrecompileUnitTestSuite(t *testing.T) {
	suite.Run(t, new(PrecompileTestSuite))
}

func (s *PrecompileTestSuite) SetupTest() {
	keyring := testkeyring.New(2)
	nw := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(nw)
	txFactory := factory.New(nw, grpcHandler)

	ctx := nw.GetContext()
	sk := nw.App.StakingKeeper
	bondDenom, err := sk.BondDenom(ctx)
	if err != nil {
		panic(err)
	}

	s.bondDenom = bondDenom
	s.factory = txFactory
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.network = nw

	if s.precompile, err = staking.NewPrecompile(
		s.network.App.StakingKeeper,
		s.network.App.AuthzKeeper,
	); err != nil {
		panic(err)
	}
}
