// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm_test

import (
	"fmt"
	"math/big"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/app/ante/evm"
	testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
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

type validatePermissionArgs struct {
	txData    evmtypes.TxData
	evmParams evmtypes.Params
	from      common.Address
}

func (suite *EvmAnteTestSuite) TestValidatePermissions() {
	// Setup
	keyring := testkeyring.New(2)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

	testCases := []struct {
		name              string
		expectedError     error
		getFunctionParams func() validatePermissionArgs
	}{
		{
			name:          "success: transfer with default params",
			expectedError: nil,
			getFunctionParams: func() validatePermissionArgs {
				from := keyring.GetAddr(0)
				txArgs := getTxByType("transfer", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)
				return validatePermissionArgs{
					evmParams: evmtypes.DefaultParams(),
					txData:    txData,
					from:      from,
				}
			},
		},
		{
			name:          "success: transfer with disable call and create",
			expectedError: nil,
			getFunctionParams: func() validatePermissionArgs {
				txArgs := getTxByType("transfer", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)
				from := keyring.GetAddr(0)
				params := evmtypes.DefaultParams()
				params.AccessControl.Call.AccessType = evmtypes.AccessTypeRestricted
				params.AccessControl.Create.AccessType = evmtypes.AccessTypeRestricted

				return validatePermissionArgs{
					evmParams: params,
					txData:    txData,
					from:      from,
				}
			},
		},
		{
			name:          "success: call with default params",
			expectedError: nil,
			getFunctionParams: func() validatePermissionArgs {
				from := keyring.GetAddr(0)
				txArgs := getTxByType("call", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)
				return validatePermissionArgs{
					evmParams: evmtypes.DefaultParams(),
					txData:    txData,
					from:      from,
				}
			},
		},
		{
			name:          "success: call tx with disabled create",
			expectedError: nil,
			getFunctionParams: func() validatePermissionArgs {
				from := keyring.GetAddr(0)
				txArgs := getTxByType("call", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)

				params := evmtypes.DefaultParams()
				params.AccessControl.Create.AccessType = evmtypes.AccessTypeRestricted

				return validatePermissionArgs{
					evmParams: params,
					txData:    txData,
					from:      from,
				}
			},
		},
		{
			name:          "fail: call tx with disabled call",
			expectedError: evmtypes.ErrCallDisabled,
			getFunctionParams: func() validatePermissionArgs {
				from := keyring.GetAddr(0)
				txArgs := getTxByType("call", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)

				params := evmtypes.DefaultParams()
				params.AccessControl.Call.AccessType = evmtypes.AccessTypeRestricted

				return validatePermissionArgs{
					evmParams: params,
					txData:    txData,
					from:      from,
				}
			},
		},
		{
			name:          "success: call tx with whitelisted address create",
			expectedError: nil,
			getFunctionParams: func() validatePermissionArgs {
				from := keyring.GetAddr(0)
				txArgs := getTxByType("call", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)

				params := evmtypes.DefaultParams()
				params.AccessControl.Call.AccessType = evmtypes.AccessTypePermissioned
				params.AccessControl.Call.WhitelistAddresses = []string{from.Hex()}

				return validatePermissionArgs{
					evmParams: params,
					txData:    txData,
					from:      from,
				}
			},
		},
		{
			name:          "fail: call tx without whitelisted address call",
			expectedError: evmtypes.ErrCallDisabled,
			getFunctionParams: func() validatePermissionArgs {
				from := keyring.GetAddr(0)
				txArgs := getTxByType("call", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)

				params := evmtypes.DefaultParams()
				params.AccessControl.Call.AccessType = evmtypes.AccessTypePermissioned

				return validatePermissionArgs{
					evmParams: params,
					txData:    txData,
					from:      from,
				}
			},
		},
		{
			name:          "success: create with default params",
			expectedError: nil,
			getFunctionParams: func() validatePermissionArgs {
				from := keyring.GetAddr(0)
				txArgs := getTxByType("create", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)
				return validatePermissionArgs{
					evmParams: evmtypes.DefaultParams(),
					txData:    txData,
					from:      from,
				}
			},
		},
		{
			name:          "success: create with disable call",
			expectedError: nil,
			getFunctionParams: func() validatePermissionArgs {
				from := keyring.GetAddr(0)
				txArgs := getTxByType("create", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)

				params := evmtypes.DefaultParams()
				params.AccessControl.Call.AccessType = evmtypes.AccessTypeRestricted

				return validatePermissionArgs{
					evmParams: params,
					txData:    txData,
					from:      from,
				}
			},
		},
		{
			name:          "fail: create with disable create",
			expectedError: evmtypes.ErrCreateDisabled,
			getFunctionParams: func() validatePermissionArgs {
				from := keyring.GetAddr(0)
				txArgs := getTxByType("create", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)

				params := evmtypes.DefaultParams()
				params.AccessControl.Create.AccessType = evmtypes.AccessTypeRestricted

				return validatePermissionArgs{
					evmParams: params,
					txData:    txData,
					from:      from,
				}
			},
		},
		{
			name:          "fail: create without whitelisted create address",
			expectedError: evmtypes.ErrCreateDisabled,
			getFunctionParams: func() validatePermissionArgs {
				from := keyring.GetAddr(0)
				txArgs := getTxByType("create", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)

				params := evmtypes.DefaultParams()
				params.AccessControl.Create.AccessType = evmtypes.AccessTypePermissioned

				return validatePermissionArgs{
					evmParams: params,
					txData:    txData,
					from:      from,
				}
			},
		},
		{
			name:          "succeed: create with whitelisted create address",
			expectedError: nil,
			getFunctionParams: func() validatePermissionArgs {
				from := keyring.GetAddr(0)
				txArgs := getTxByType("create", keyring.GetAddr(1))
				txData, err := txArgs.ToTxData()
				suite.Require().NoError(err)

				params := evmtypes.DefaultParams()
				params.AccessControl.Create.AccessType = evmtypes.AccessTypePermissioned
				params.AccessControl.Create.WhitelistAddresses = []string{from.Hex()}

				return validatePermissionArgs{
					evmParams: params,
					txData:    txData,
					from:      from,
				}
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("%v_%v", evmtypes.GetTxTypeName(suite.ethTxType), tc.name), func() {
			// Perform test logic
			args := tc.getFunctionParams()

			//  Function to be tested
			err := evm.ValidatePermission(
				unitNetwork.GetContext(),
				args.txData,
				unitNetwork.App.AccountKeeper,
				args.evmParams,
				args.from,
			)

			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.expectedError.Error())
			} else {
				suite.Require().NoError(err)
			}
			// Clean block for next test
			err = unitNetwork.NextBlock()
			suite.Require().NoError(err)
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
