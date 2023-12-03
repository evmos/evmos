package evm_test

import (
	"testing"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"
)

type EvmAnteTestSuite struct {
	suite.Suite

	// To make sure that every tests is run with all the tx types
	ethTxType uint8
}

func TestEvmAnteTestSuite(t *testing.T) {
	suite.Run(t, &EvmAnteTestSuite{
		ethTxType: gethtypes.DynamicFeeTxType,
	})
	suite.Run(t, &EvmAnteTestSuite{
		ethTxType: gethtypes.LegacyTxType,
	})
	suite.Run(t, &EvmAnteTestSuite{
		ethTxType: gethtypes.AccessListTxType,
	})
}
