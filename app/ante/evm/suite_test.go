package evm_test

import (
	"testing"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"
)

// EvmAnteTestSuite aims to test all EVM ante handler unit functions.
// NOTE: the suite only holds properties related to global execution parameters
// (what type of tx to run the tests with) not independent tests values.
type EvmAnteTestSuite struct {
	suite.Suite

	// To make sure that every tests is run with all the tx types
	ethTxType int
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
