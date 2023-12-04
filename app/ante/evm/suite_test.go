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

func (suite *EvmAnteTestSuite) getTxTypeTestName() string {
	switch suite.ethTxType {
	case gethtypes.DynamicFeeTxType:
		return "DynamicFeeTxType"
	case gethtypes.LegacyTxType:
		return "LegacyTxType"
	case gethtypes.AccessListTxType:
		return "AccessListTxType"
	default:
		panic("unknown tx type")
	}
}
