// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm_test

import (
	sdkmath "cosmossdk.io/math"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/evmos/evmos/v16/app/ante/evm"
)

func (suite *EvmAnteTestSuite) TestMempoolFee() {
	testCases := []struct {
		name          string
		expectedError error
		isLondon      bool
		txFee         sdkmath.LegacyDec
		minGasPrice   sdkmath.LegacyDec
		gasLimit      sdkmath.LegacyDec
	}{
		{
			name:          "success: if London fork is enabled, skip check",
			expectedError: nil,
			isLondon:      true,
			// values are not used because isLondon is true
			txFee:       sdkmath.LegacyOneDec(),
			minGasPrice: sdkmath.LegacyOneDec(),
			gasLimit:    sdkmath.LegacyOneDec(),
		},
		{
			name:          "success: fee is greater than min gas price * gas limit",
			expectedError: nil,
			isLondon:      false,
			txFee:         sdkmath.LegacyNewDec(100),
			minGasPrice:   sdkmath.LegacyOneDec(),
			gasLimit:      sdkmath.LegacyOneDec(),
		},
		{
			name:          "fail: fee is less than min gas price * gas limit",
			expectedError: errortypes.ErrInsufficientFee,
			isLondon:      false,
			txFee:         sdkmath.LegacyOneDec(),
			minGasPrice:   sdkmath.LegacyNewDec(100),
			gasLimit:      sdkmath.LegacyOneDec(),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Function under test
			err := evm.CheckMempoolFee(
				tc.txFee,
				tc.minGasPrice,
				tc.gasLimit,
				tc.isLondon,
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
