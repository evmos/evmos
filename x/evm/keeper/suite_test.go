package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// EvmKeeperTestSuite aims to test all EVM Keeper unit functions.
type EvmKeeperTestSuite struct {
	suite.Suite
}

func TestEvmKeeperTestSuite(t *testing.T) {
	suite.Run(t, &EvmKeeperTestSuite{})
}
