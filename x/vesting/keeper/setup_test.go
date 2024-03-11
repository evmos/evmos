package keeper_test

import (

	//nolint:revive // dot imports are fine for Ginkgo
	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
)

var (
	contract  common.Address
	contract2 common.Address
)

var (
	erc20Name     = "Coin Token"
	erc20Symbol   = "CTKN"
	erc20Name2    = "Coin Token 2"
	erc20Symbol2  = "CTKN2"
	erc20Decimals = uint8(18)
)

type KeeperTestSuite struct {
	suite.Suite

	network *network.UnitTestNetwork
	handler grpc.Handler
	keyring keyring.Keyring
	factory factory.TxFactory
}

func (s *KeeperTestSuite) SetupTest() {
	keys := keyring.New(2)
	nw := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keys.GetAllAccAddrs()...),
	)
	gh := grpc.NewIntegrationHandler(nw)
	tf := factory.New(nw, gh)

	s.network = nw
	s.factory = tf
	s.handler = gh
	s.keyring = keys
}
