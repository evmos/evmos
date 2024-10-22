package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/Eidon-AI/eidon-chain/v20/testutil/integration/eidon-chain/factory"
	"github.com/Eidon-AI/eidon-chain/v20/testutil/integration/eidon-chain/grpc"
	testkeyring "github.com/Eidon-AI/eidon-chain/v20/testutil/integration/eidon-chain/keyring"
	"github.com/Eidon-AI/eidon-chain/v20/testutil/integration/eidon-chain/network"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	suite.Suite

	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	denom string
}

// SetupTest setup test environment
func (suite *KeeperTestSuite) SetupTest() {
	keyring := testkeyring.New(2)
	nw := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		network.WithCustomBaseAppOpts(baseapp.SetMinGasPrices("10aeidon-chain")),
	)
	grpcHandler := grpc.NewIntegrationHandler(nw)
	txFactory := factory.New(nw, grpcHandler)

	ctx := nw.GetContext()
	sk := nw.App.StakingKeeper
	bondDenom, err := sk.BondDenom(ctx)
	if err != nil {
		panic(err)
	}

	suite.denom = bondDenom
	suite.factory = txFactory
	suite.grpcHandler = grpcHandler
	suite.keyring = keyring
	suite.network = nw
}
