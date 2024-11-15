// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm_test

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v20/app/ante/evm"
	testkeyring "github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

type validateMsgParams struct {
	evmParams evmtypes.Params
	from      sdktypes.AccAddress
	txData    evmtypes.TxData
}

func (suite *EvmAnteTestSuite) TestValidateMsg() {
	keyring := testkeyring.New(2)

	testCases := []struct {
		name              string
		expectedError     error
		getFunctionParams func() validateMsgParams
	}{
		{
			name:          "fail: invalid from address, should be nil",
			expectedError: errortypes.ErrInvalidRequest,
			getFunctionParams: func() validateMsgParams {
				return validateMsgParams{
					evmParams: evmtypes.DefaultParams(),
					txData:    nil,
					from:      keyring.GetAccAddr(0),
				}
			},
		},
		{
			name:          "success: transfer with default params",
			expectedError: nil,
			getFunctionParams: func() validateMsgParams {
				txArgs := getTxByType("transfer", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)
				return validateMsgParams{
					evmParams: evmtypes.DefaultParams(),
					txData:    txData,
					from:      nil,
				}
			},
		},
		{
			name:          "success: transfer with disable call and create",
			expectedError: evmtypes.ErrCallDisabled,
			getFunctionParams: func() validateMsgParams {
				txArgs := getTxByType("transfer", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)

				params := evmtypes.DefaultParams()
				params.AccessControl.Call.AccessType = evmtypes.AccessTypeRestricted
				params.AccessControl.Create.AccessType = evmtypes.AccessTypeRestricted

				return validateMsgParams{
					evmParams: params,
					txData:    txData,
					from:      nil,
				}
			},
		},
		{
			name:          "success: call with default params",
			expectedError: nil,
			getFunctionParams: func() validateMsgParams {
				txArgs := getTxByType("call", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)
				return validateMsgParams{
					evmParams: evmtypes.DefaultParams(),
					txData:    txData,
					from:      nil,
				}
			},
		},
		{
			name:          "success: call tx with disabled create",
			expectedError: nil,
			getFunctionParams: func() validateMsgParams {
				txArgs := getTxByType("call", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)

				params := evmtypes.DefaultParams()
				params.AccessControl.Create.AccessType = evmtypes.AccessTypeRestricted

				return validateMsgParams{
					evmParams: params,
					txData:    txData,
					from:      nil,
				}
			},
		},
		{
			name:          "fail: call tx with disabled call",
			expectedError: evmtypes.ErrCallDisabled,
			getFunctionParams: func() validateMsgParams {
				txArgs := getTxByType("call", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)

				params := evmtypes.DefaultParams()
				params.AccessControl.Call.AccessType = evmtypes.AccessTypeRestricted

				return validateMsgParams{
					evmParams: params,
					txData:    txData,
					from:      nil,
				}
			},
		},
		{
			name:          "success: create with default params",
			expectedError: nil,
			getFunctionParams: func() validateMsgParams {
				txArgs := getTxByType("create", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)
				return validateMsgParams{
					evmParams: evmtypes.DefaultParams(),
					txData:    txData,
					from:      nil,
				}
			},
		},
		{
			name:          "success: create with disable call",
			expectedError: nil,
			getFunctionParams: func() validateMsgParams {
				txArgs := getTxByType("create", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)

				params := evmtypes.DefaultParams()
				params.AccessControl.Call.AccessType = evmtypes.AccessTypeRestricted

				return validateMsgParams{
					evmParams: params,
					txData:    txData,
					from:      nil,
				}
			},
		},
		{
			name:          "fail: create with disable create",
			expectedError: evmtypes.ErrCreateDisabled,
			getFunctionParams: func() validateMsgParams {
				txArgs := getTxByType("create", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)

				params := evmtypes.DefaultParams()
				params.AccessControl.Create.AccessType = evmtypes.AccessTypeRestricted

				return validateMsgParams{
					evmParams: params,
					txData:    txData,
					from:      nil,
				}
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			params := tc.getFunctionParams()

			// Function under test
			err := evm.ValidateMsg(
				params.evmParams,
				params.txData,
				params.from,
			)

			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.expectedError.Error())
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func getTxByType(typeTx string, recipient common.Address) evmtypes.EvmTxArgs {
	switch typeTx {
	case "call":
		return evmtypes.EvmTxArgs{
			To:    &recipient,
			Input: []byte("call bytes"),
		}
	case "create":
		return evmtypes.EvmTxArgs{
			Input: []byte("create bytes"),
		}
	case "transfer":
		return evmtypes.EvmTxArgs{
			To:     &recipient,
			Amount: big.NewInt(100),
		}
	default:
		panic("invalid type")
	}
}

func (suite *EvmAnteTestSuite) TestCheckTxFee() {
	// amount represents 1 token in the 18 decimals representation.
	amount := math.NewInt(1e18)
	gasLimit := uint64(1e6)

	testCases := []struct {
		name       string
		txFee      *big.Int
		txGasLimit uint64
		expError   error
	}{
		{
			name:       "pass",
			txFee:      big.NewInt(amount.Int64()),
			txGasLimit: gasLimit,
			expError:   nil,
		},
		{
			name:       "fail: not enough tx fees",
			txFee:      big.NewInt(amount.Int64() - 1),
			txGasLimit: gasLimit,
			expError:   errortypes.ErrInvalidRequest,
		},
	}

	for _, decimals := range []evmtypes.Decimals{
		evmtypes.SixDecimals,
		evmtypes.EighteenDecimals,
	} {
		for _, tc := range testCases {
			suite.Run(fmt.Sprintf("%d decimals, %s", decimals, tc.name), func() {
				// Call the configurator to set the EVM coin required for the
				// function to be tested.
				configurator := evmtypes.NewEVMConfigurator()
				configurator.ResetTestConfig()
				suite.Require().NoError(configurator.WithEVMCoinInfo(types.BaseDenom, uint8(decimals)).Configure())

				// If decimals is not 18 decimals, we have to convert txFeeInfo to original
				// decimals representation.
				originalAmount := amount
				evmCoinDecimal := evmtypes.GetEVMCoinDecimals()
				originalAmount = originalAmount.Quo(evmCoinDecimal.ConversionFactor())

				coins := sdktypes.Coins{sdktypes.Coin{Denom: "aevmos", Amount: originalAmount}}

				// This struct should hold values in the original representation
				txFeeInfo := &tx.Fee{
					Amount:   coins,
					GasLimit: gasLimit,
				}

				// Function under test
				err := evm.CheckTxFee(txFeeInfo, tc.txFee, tc.txGasLimit)

				if tc.expError != nil {
					suite.Require().Error(err)
					suite.Contains(err.Error(), tc.expError.Error())
				} else {
					suite.Require().NoError(err)
				}
			})
		}
	}
}
