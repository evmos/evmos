package keeper_test

import (
	"testing"

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

var s *KeeperTestSuite

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)
}
