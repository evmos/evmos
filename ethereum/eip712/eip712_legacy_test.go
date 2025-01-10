package eip712_test

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/evmos/v20/ethereum/eip712"
	"github.com/stretchr/testify/require"
)

func TestLegacyWrapTxToTypedData(t *testing.T) {
	// Setup
	cdc := codec.NewProtoCodec(types.NewInterfaceRegistry())

	chainID := uint64(1)
	msg := banktypes.NewMsgSend(
		sdk.AccAddress([]byte("from_address")),
		sdk.AccAddress([]byte("to_address")),
		sdk.Coins{
			sdk.NewInt64Coin("stake", 1000),
		},
	)

	txData := map[string]interface{}{
		"account_number": "1",
		"chain_id":       "cosmoshub-4",
		"fee": map[string]interface{}{
			"amount": []map[string]string{
				{"denom": "stake", "amount": "10"},
			},
			"gas": "200000",
		},
		"memo": "Test Memo",
		"msgs": []interface{}{
			map[string]interface{}{
				"type": "cosmos-sdk/MsgSend",
				"value": map[string]string{
					"from_address": "from_address",
					"to_address":   "to_address",
					"amount":       "1000stake",
				},
			},
		},
		"sequence": "1",
	}

	data, err := json.Marshal(txData)
	require.NoError(t, err)

	feeDelegation := &eip712.FeeDelegationOptions{
		FeePayer: sdk.AccAddress([]byte("fee_payer")),
	}

	// Execute
	typedData, err := eip712.LegacyWrapTxToTypedData(cdc, chainID, msg, data, feeDelegation)

	// Validate
	require.NoError(t, err)
	require.NotNil(t, typedData)

	// Check Domain
	require.Equal(t, "Cosmos Web3", typedData.Domain.Name)
	require.Equal(t, "1.0.0", typedData.Domain.Version)
	require.Equal(t, "cosmos", typedData.Domain.VerifyingContract)

	// Check Message
	message := typedData.Message
	require.Equal(t, "1", message["account_number"])
	require.Equal(t, "cosmoshub-4", message["chain_id"])

	fee := message["fee"].(map[string]interface{})
	require.Equal(t, sdk.AccAddress([]byte("fee_payer")).String(), fee["feePayer"])

	require.Equal(t, "200000", fee["gas"])

	msgs := message["msgs"].([]interface{})
	require.Len(t, msgs, 1)

	msgContent := msgs[0].(map[string]interface{})
	require.Equal(t, "cosmos-sdk/MsgSend", msgContent["type"])
	value := msgContent["value"].(map[string]interface{})
	require.Equal(t, "from_address", value["from_address"])
	require.Equal(t, "to_address", value["to_address"])
	require.Equal(t, "1000stake", value["amount"])

	// Validate types structure
	require.Contains(t, typedData.Types, "EIP712Domain")
	require.Contains(t, typedData.Types, "Tx")
}
