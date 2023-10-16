// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package factory

import (
	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"

	errorsmod "cosmossdk.io/errors"
	amino "github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	testutiltypes "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	enccodec "github.com/evmos/evmos/v15/encoding/codec"
	"github.com/evmos/evmos/v15/testutil/tx"
	evmostypes "github.com/evmos/evmos/v15/types"
)

// buildMsgEthereumTx builds an Ethereum transaction from the given arguments and populates the From field.
func buildMsgEthereumTx(txArgs evmtypes.EvmTxArgs, fromAddr common.Address) (evmtypes.MsgEthereumTx, error) {
	msgEthereumTx := evmtypes.NewTx(&txArgs)
	msgEthereumTx.From = fromAddr.String()

	// Validate the transaction to avoid unrealistic behavior
	err := msgEthereumTx.ValidateBasic()
	if err != nil {
		return evmtypes.MsgEthereumTx{}, errorsmod.Wrap(err, "failed to validate transaction")
	}
	return *msgEthereumTx, nil
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
