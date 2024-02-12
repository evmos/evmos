// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package encoding

import (
	"cosmossdk.io/x/tx/signing"
	amino "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	sdktestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/gogoproto/proto"

	enccodec "github.com/evmos/evmos/v16/encoding/codec"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

// MakeConfig creates an EncodingConfig for testing
// and registers the interfaces
func MakeConfig(mb module.BasicManager) sdktestutil.TestEncodingConfig {
	ec := encodingConfig()
	enccodec.RegisterLegacyAminoCodec(ec.Amino)
	mb.RegisterLegacyAminoCodec(ec.Amino)
	enccodec.RegisterInterfaces(ec.InterfaceRegistry)
	mb.RegisterInterfaces(ec.InterfaceRegistry)
	// This is needed for the EIP712 txs because currently is using
	// the deprecated method legacytx.StdSignBytes
	legacytx.RegressionTestingAminoCodec = ec.Amino
	return ec
}

// encodingConfig creates a new EncodingConfig and returns it
func encodingConfig() sdktestutil.TestEncodingConfig {
	cdc := amino.NewLegacyAmino()
	signingOptions := signing.Options{
		AddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
		},
		ValidatorAddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
		},
	}
	signingOptions.DefineCustomGetSigners(
		evmtypes.MsgEthereumTxCustomGetSigner.MsgType,
		evmtypes.MsgEthereumTxCustomGetSigner.Fn,
	)

	interfaceRegistry, _ := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles:     proto.HybridResolver,
		SigningOptions: signingOptions,
	})
	codec := amino.NewProtoCodec(interfaceRegistry)

	return sdktestutil.TestEncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             codec,
		TxConfig:          tx.NewTxConfig(codec, tx.DefaultSignModes),
		Amino:             cdc,
	}
}
