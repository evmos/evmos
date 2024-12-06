package evm_test

import (
	"testing"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v20/utils"
	"github.com/stretchr/testify/suite"
)

// EvmAnteTestSuite aims to test all EVM ante handler unit functions.
// NOTE: the suite only holds properties related to global execution parameters
// (what type of tx to run the tests with) not independent tests values.
type EvmAnteTestSuite struct {
	suite.Suite

	// To make sure that every tests is run with all the tx types
	ethTxType int
	chainID   string
}

func TestEvmAnteTestSuite(t *testing.T) {
	txTypes := []int{gethtypes.DynamicFeeTxType, gethtypes.LegacyTxType, gethtypes.AccessListTxType}
	chainIDs := []string{utils.MainnetChainID + "-1", utils.SixDecChainID + "-1"}
	for _, txType := range txTypes {
		for _, chainID := range chainIDs {
			suite.Run(t, &EvmAnteTestSuite{
				ethTxType: txType,
				chainID:   chainID,
			})
		}
	}
}
