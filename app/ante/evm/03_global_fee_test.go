// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm_test

import (
	sdkmath "cosmossdk.io/math"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/evmos/evmos/v17/app/ante/evm"
)

func (suite *EvmAnteTestSuite) TestGlobalFee() {
	testCases := []struct {
		name              string
		expectedError     error
		txFee             sdkmath.LegacyDec
		globalMinGasPrice sdkmath.LegacyDec
		gasLimit          sdkmath.LegacyDec
	}{
		{
			name:          "success: if globalMinGasPrice is 0, skip check",
			expectedError: nil,
			// values are not used because isLondon is true
			txFee:             sdkmath.LegacyOneDec(),
			globalMinGasPrice: sdkmath.LegacyZeroDec(),
			gasLimit:          sdkmath.LegacyOneDec(),
		},
		{
			name:              "success: fee is greater than global gas price * gas limit",
			expectedError:     nil,
			txFee:             sdkmath.LegacyNewDec(100),
			globalMinGasPrice: sdkmath.LegacyOneDec(),
			gasLimit:          sdkmath.LegacyOneDec(),
		},
		{
			name:              "fail: fee is less than global gas price * gas limit",
			expectedError:     errortypes.ErrInsufficientFee,
			txFee:             sdkmath.LegacyOneDec(),
			globalMinGasPrice: sdkmath.LegacyNewDec(100),
			gasLimit:          sdkmath.LegacyOneDec(),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Function under test
			err := evm.CheckGlobalFee(
				tc.txFee,
				tc.globalMinGasPrice,
				tc.gasLimit,
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
