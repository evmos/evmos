// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package factory

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/encoding"
	"github.com/evmos/evmos/v16/precompiles/testutil"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	testutiltypes "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

// buildMsgEthereumTx builds an Ethereum transaction from the given arguments and populates the From field.
func buildMsgEthereumTx(txArgs evmtypes.EvmTxArgs, fromAddr common.Address) evmtypes.MsgEthereumTx {
	msgEthereumTx := evmtypes.NewTx(&txArgs)
	msgEthereumTx.From = fromAddr.String()
	return *msgEthereumTx
}

// makeConfig creates an EncodingConfig for testing
func makeConfig(mb module.BasicManager) testutiltypes.TestEncodingConfig {
	return encoding.MakeConfig(mb)
}

// CheckError is a helper function to check if the error is the expected one.
func CheckError(err error, logCheckArgs testutil.LogCheckArgs) error {
	switch {
	case logCheckArgs.ExpPass && err == nil:
		return nil
	case !logCheckArgs.ExpPass && err == nil:
		return errorsmod.Wrap(err, "expected error but got none")
	case logCheckArgs.ExpPass && err != nil:
		return errorsmod.Wrap(err, "expected no error but got one")
	case logCheckArgs.ErrContains == "":
		// NOTE: if err contains is empty, we return the error as it is
		return errorsmod.Wrap(err, "ErrContains needs to be filled")
	case !strings.Contains(err.Error(), logCheckArgs.ErrContains):
		return errorsmod.Wrapf(err, "expected different error; wanted %q", logCheckArgs.ErrContains)
	}

	return nil
}
