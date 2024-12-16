// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm_test

import (
	"fmt"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	evmante "github.com/evmos/evmos/v20/app/ante/evm"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

func (suite *EvmAnteTestSuite) TestUpdateCumulativeGasWanted() {
	keyring := testkeyring.New(1)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithChainID(suite.chainID),
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

	testCases := []struct {
		name                string
		msgGasWanted        uint64
		maxTxGasWanted      uint64
		cumulativeGasWanted uint64
		getCtx              func() sdktypes.Context
		expectedResponse    uint64
	}{
		{
			name:                "when is NOT checkTx and cumulativeGasWanted is 0, returns msgGasWanted",
			msgGasWanted:        100,
			maxTxGasWanted:      150,
			cumulativeGasWanted: 0,
			getCtx: func() sdktypes.Context {
				return unitNetwork.GetContext().WithIsCheckTx(false)
			},
			expectedResponse: 100,
		},
		{
			name:                "when is NOT checkTx and cumulativeGasWanted has value, returns cumulativeGasWanted + msgGasWanted",
			msgGasWanted:        100,
			maxTxGasWanted:      150,
			cumulativeGasWanted: 50,
			getCtx: func() sdktypes.Context {
				return unitNetwork.GetContext().WithIsCheckTx(false)
			},
			expectedResponse: 150,
		},
		{
			name:                "when is checkTx, maxTxGasWanted is not 0 and msgGasWanted > maxTxGasWanted, returns cumulativeGasWanted + maxTxGasWanted",
			msgGasWanted:        200,
			maxTxGasWanted:      100,
			cumulativeGasWanted: 50,
			getCtx: func() sdktypes.Context {
				return unitNetwork.GetContext().WithIsCheckTx(true)
			},
			expectedResponse: 150,
		},
		{
			name:                "when is checkTx, maxTxGasWanted is not 0 and msgGasWanted < maxTxGasWanted, returns cumulativeGasWanted + msgGasWanted",
			msgGasWanted:        50,
			maxTxGasWanted:      100,
			cumulativeGasWanted: 50,
			getCtx: func() sdktypes.Context {
				return unitNetwork.GetContext().WithIsCheckTx(true)
			},
			expectedResponse: 100,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Function under test
			gasWanted := evmante.UpdateCumulativeGasWanted(
				tc.getCtx(),
				tc.msgGasWanted,
				tc.maxTxGasWanted,
				tc.cumulativeGasWanted,
			)

			suite.Require().Equal(tc.expectedResponse, gasWanted)
		})
	}
}

// NOTE: claim rewards are not tested since there is an independent suite to test just that
func (suite *EvmAnteTestSuite) TestConsumeGasAndEmitEvent() {
	keyring := testkeyring.New(1)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithChainID(suite.chainID),
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(unitNetwork)

	testCases := []struct {
		name          string
		expectedError string
		feesAmt       math.Int
		getSender     func() sdktypes.AccAddress
	}{
		{
			name:    "success: fees are zero and event emitted",
			feesAmt: math.NewInt(0),
			getSender: func() sdktypes.AccAddress {
				// Return prefunded sender
				return keyring.GetKey(0).AccAddr
			},
		},
		{
			name:    "success: there are non zero fees, user has sufficient bank balances and event emitted",
			feesAmt: math.NewInt(1000),
			getSender: func() sdktypes.AccAddress {
				// Return prefunded sender
				return keyring.GetKey(0).AccAddr
			},
		},
		{
			name:          "fail: insufficient user balance, event is NOT emitted",
			expectedError: "failed to deduct transaction costs from user balance",
			feesAmt:       math.NewInt(1000),
			getSender: func() sdktypes.AccAddress {
				// Return unfunded account
				index := keyring.AddKey()
				return keyring.GetKey(index).AccAddr
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("%v_%v_%v", evmtypes.GetTxTypeName(suite.ethTxType), suite.chainID, tc.name), func() {
			keepers := &evmante.ConsumeGasKeepers{
				Evm: unitNetwork.App.EvmKeeper,
			}
			sender := tc.getSender()

			resp, err := grpcHandler.GetBalanceFromEVM(sender)
			suite.Require().NoError(err)
			prevBalance, ok := math.NewIntFromString(resp.Balance)
			suite.Require().True(ok)

			suite.Require().NoError(err)

			evmDecimals := evmtypes.GetEVMCoinDecimals()
			feesAmt := tc.feesAmt.Mul(evmDecimals.ConversionFactor())
			fees := sdktypes.NewCoins(sdktypes.NewCoin(unitNetwork.GetBaseDenom(), feesAmt))

			// Function under test
			err = evmante.ConsumeFeesAndEmitEvent(
				unitNetwork.GetContext(),
				keepers,
				fees,
				sender,
			)

			if tc.expectedError != "" {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.expectedError)

				// Check events are not present
				events := unitNetwork.GetContext().EventManager().Events()
				suite.Require().Zero(len(events), "required no events to be emitted")
			} else {
				suite.Require().NoError(err)

				// Check fees are deducted
				resp, err := grpcHandler.GetBalanceFromEVM(sender)
				suite.Require().NoError(err)
				afterBalance, ok := math.NewIntFromString(resp.Balance)
				suite.Require().True(ok)

				suite.Require().NoError(err)
				expectedBalance := prevBalance.Sub(feesAmt)
				suite.Require().True(expectedBalance.Equal(afterBalance), "expected different balance after fees deduction")

				// Event to be emitted
				expectedEvent := sdktypes.NewEvent(
					sdktypes.EventTypeTx,
					sdktypes.NewAttribute(sdktypes.AttributeKeyFee, fees.String()),
				)
				// Check events are present
				events := unitNetwork.GetContext().EventManager().Events()
				suite.Require().NotZero(len(events))
				suite.Require().Contains(
					events,
					expectedEvent,
					"expected different events after fees deduction",
				)
			}

			// Reset the context
			err = unitNetwork.NextBlock()
			suite.Require().NoError(err)
		})
	}
}
