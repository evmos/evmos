// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testutil

import (
	"fmt"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	evmtypes "github.com/evmos/evmos/v14/x/evm/types"
)

// CheckVMError is a helper function used to check if the transaction is reverted with the expected error message
// in the VmError field of the MsgEthereumResponse struct.
func CheckVMError(res abci.ResponseDeliverTx, expErrMsg string, args ...interface{}) error {
	if !res.IsOK() {
		return fmt.Errorf("code 0 was expected on response but got code %d", res.Code)
	}
	ethRes, err := evmtypes.DecodeTxResponse(res.Data)
	if err != nil {
		return fmt.Errorf("error occurred while decoding the TxResponse. %s", err)
	}
	expMsg := fmt.Sprintf(expErrMsg, args...)
	if !strings.Contains(ethRes.VmError, expMsg) {
		return fmt.Errorf("unexpected VmError on response. expected error to contain: %s, received: %s", expMsg, ethRes.VmError)
	}
	return nil
}

// CheckEthereumTxFailed checks if there is a VM error in the transaction response and returns the reason.
func CheckEthereumTxFailed(ethRes *evmtypes.MsgEthereumTxResponse) (string, bool) {
	reason := ethRes.VmError
	return reason, reason != ""
}
