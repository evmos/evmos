package eip712_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/evmos/v20/ethereum/eip712"
	"github.com/stretchr/testify/require"
)

var (
	aminoCodec        *codec.LegacyAmino
	interfaceRegistry types.InterfaceRegistry
)

func initAddressConfig() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("cosmos", "cosmospub")
	config.SetBech32PrefixForValidator("cosmosvaloper", "cosmosvaloperpub")
	config.SetBech32PrefixForConsensusNode("cosmosvalcons", "cosmosvalconspub")
	config.Seal()
}

func initCodecs() {
	initAddressConfig()

	interfaceRegistry = types.NewInterfaceRegistry()
	aminoCodec = codec.NewLegacyAmino()

	sdk.RegisterLegacyAminoCodec(aminoCodec)
	legacytx.RegisterLegacyAminoCodec(aminoCodec)
	banktypes.RegisterLegacyAminoCodec(aminoCodec)
	banktypes.RegisterInterfaces(interfaceRegistry)

	eip712.SetEncodingConfig(aminoCodec, interfaceRegistry)
}

func TestGetEIP712TypedDataForMsg_AminoSuccess(t *testing.T) {
	initCodecs()

	msg := banktypes.MsgSend{
		FromAddress: "cosmos1qperwt9wrnkg6kzfj7s9wf69w5gk3ya6r7l273",
		ToAddress:   "cosmos1q5x4yng8x6f5v45m59wvndkgj7c32c7dpw2crz",
		Amount:      sdk.Coins{sdk.NewInt64Coin("stake", 1000)},
	}

	msgBytes, err := aminoCodec.MarshalJSON(msg)
	require.NoError(t, err, "failed to marshal MsgSend")

	signDoc := legacytx.StdSignDoc{
		AccountNumber: 1,
		ChainID:       "evmos_9001-1",
		Sequence:      1,
		Msgs:          []json.RawMessage{msgBytes},
		Fee:           json.RawMessage(`{"amount":[{"denom":"stake","amount":"10"}],"gas":"200000"}`),
		Memo:          "test memo",
	}

	signDocBytes, err := aminoCodec.MarshalJSON(signDoc)
	require.NoError(t, err, "failed to marshal Amino sign doc")

	fmt.Printf("SignDoc JSON: %s\n", string(signDocBytes))
}
