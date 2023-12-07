// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package factory

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/precompiles/testutil"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"

	errorsmod "cosmossdk.io/errors"
	amino "github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	testutiltypes "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	enccodec "github.com/evmos/evmos/v16/encoding/codec"
	"github.com/evmos/evmos/v16/testutil/tx"
	evmostypes "github.com/evmos/evmos/v16/types"
)

// buildMsgEthereumTx builds an Ethereum transaction from the given arguments and populates the From field.
func buildMsgEthereumTx(txArgs evmtypes.EvmTxArgs, fromAddr common.Address) evmtypes.MsgEthereumTx {
	msgEthereumTx := evmtypes.NewTx(&txArgs)
	msgEthereumTx.From = fromAddr.String()
	return *msgEthereumTx
}

// signMsgEthereumTx signs a MsgEthereumTx with the provided private key and chainID.
func signMsgEthereumTx(msgEthereumTx evmtypes.MsgEthereumTx, privKey cryptotypes.PrivKey, chainID string) (evmtypes.MsgEthereumTx, error) {
	ethChainID, err := evmostypes.ParseChainID(chainID)
	if err != nil {
		return evmtypes.MsgEthereumTx{}, errorsmod.Wrapf(err, "failed to parse chainID: %v", chainID)
	}

	signer := ethtypes.LatestSignerForChainID(ethChainID)
	err = msgEthereumTx.Sign(signer, tx.NewSigner(privKey))
	if err != nil {
		return evmtypes.MsgEthereumTx{}, errorsmod.Wrap(err, "failed to sign transaction")
	}

	// Validate the transaction to avoid unrealistic behavior
	if err = msgEthereumTx.ValidateBasic(); err != nil {
		return evmtypes.MsgEthereumTx{}, errorsmod.Wrap(err, "failed to validate transaction")
	}
	return msgEthereumTx, nil
}

// makeConfig creates an EncodingConfig for testing
func makeConfig(mb module.BasicManager) testutiltypes.TestEncodingConfig {
	cdc := amino.NewLegacyAmino()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	codec := amino.NewProtoCodec(interfaceRegistry)

	encodingConfig := testutiltypes.TestEncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             codec,
		TxConfig:          authtx.NewTxConfig(codec, authtx.DefaultSignModes),
		Amino:             cdc,
	}

	enccodec.RegisterLegacyAminoCodec(encodingConfig.Amino)
	mb.RegisterLegacyAminoCodec(encodingConfig.Amino)
	enccodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	mb.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
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
