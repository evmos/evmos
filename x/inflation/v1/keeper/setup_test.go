package keeper_test

import (
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
)

var denomMint string

type KeeperTestSuite struct {
	suite.Suite

	network *network.UnitTestNetwork
	handler grpc.Handler
	keyring keyring.Keyring
	factory factory.TxFactory
}

func (suite *KeeperTestSuite) SetupTest() {

	keys := keyring.New(2)
	nw := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keys.GetAllAccAddrs()...),
	)
	gh := grpc.NewIntegrationHandler(nw)
	tf := factory.New(nw, gh)
	var err error
	denomMint, err = sdk.GetBaseDenom()
	suite.Require().NoError(err)
	suite.network = nw
	suite.factory = tf
	suite.handler = gh
	suite.keyring = keys
}
