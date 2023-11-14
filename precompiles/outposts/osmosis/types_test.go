package osmosis_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/cosmos/btcutil/bech32"
	"github.com/ethereum/go-ethereum/common"

	osmosisoutpost "github.com/evmos/evmos/v15/precompiles/outposts/osmosis"
	"github.com/evmos/evmos/v15/utils"
	"github.com/stretchr/testify/require"
)

func TestCreatePacketWithMemo(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		outputDenom        string
		receiver           string
		contract           string
		slippagePercentage uint8
		windowSeconds      uint64
		onFailedDelivery   string
		nextMemo           string
		expNextMemo        bool
	}{
		{
			name:               "pass - correct string without memo",
			outputDenom:        "aevmos",
			receiver:           "receiver",
			contract:           "contract",
			slippagePercentage: 10,
			windowSeconds:      30,
			onFailedDelivery:   "do_nothing",
			nextMemo:           "",
			expNextMemo:        false,
		},
		{
			name:               "pass - correct string with memo",
			outputDenom:        "aevmos",
			receiver:           "receiver",
			contract:           "contract",
			slippagePercentage: 10,
			windowSeconds:      30,
			onFailedDelivery:   "do_nothing",
			nextMemo:           "a next memo",
			expNextMemo:        true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			packet := osmosisoutpost.CreatePacketWithMemo(
				tc.outputDenom, tc.receiver, tc.contract, tc.slippagePercentage, tc.windowSeconds, tc.onFailedDelivery, tc.nextMemo,
			)
			packetString := packet.String()
			err := ValidateAndParseWasmRoutedMemo(packetString, tc.receiver)
			require.NoError(t, err, "memo is not a valid wasm routed JSON formatted string")

			if tc.expNextMemo {
				require.Contains(t, packetString, fmt.Sprintf("\"next_memo\": \"%s\"", tc.nextMemo))
			} else {
				require.NotContains(t, packetString, fmt.Sprintf("next_memo: %s", tc.nextMemo))
			}
		})

	}
}

// TestParseSwapPacketData is mainly to test that the returned error of the
// parser is clear and contains the correct data type. For this reason the
// expected error has been hardcoded as a string litera.
func TestParseSwapPacketData(t *testing.T) {
	t.Parallel()

	testSender := common.HexToAddress("sender")
	testInput := common.HexToAddress("input")
	testOutput := common.HexToAddress("output")
	testAmount := big.NewInt(3)
	testSlippagePercentage := uint8(10)
	testWindowSeconds := uint64(20)
	testReceiver := "cosmos1c2m73hdt6f37w9jqpqps5t3ha3st99dcsp7lf5"

	testCases := []struct {
		name        string
		args        []interface{}
		expPass     bool
		errContains string
	}{
		{
			name: "pass - valid payload",
			args: []interface{}{
				testSender,
				testInput,
				testOutput,
				testAmount,
				testSlippagePercentage,
				testWindowSeconds,
				testReceiver,
			},
			expPass: true,
		}, {
			name: "fail - wrong sender type",
			args: []interface{}{
				"sender",
				testInput,
				testOutput,
				testAmount,
				testSlippagePercentage,
				testWindowSeconds,
				testReceiver,
			},
			expPass:     false,
			errContains: "invalid type for sender: expected common.Address, received string",
		}, {
			name: "fail - wrong input type",
			args: []interface{}{
				testSender,
				"input",
				testOutput,
				testAmount,
				testSlippagePercentage,
				testWindowSeconds,
				testReceiver,
			},
			expPass:     false,
			errContains: "invalid type for input: expected common.Address, received string",
		}, {
			name: "fail - wrong output type",
			args: []interface{}{
				testSender,
				testInput,
				"output",
				testAmount,
				testSlippagePercentage,
				testWindowSeconds,
				testReceiver,
			},
			expPass:     false,
			errContains: "invalid type for output: expected common.Address, received string",
		}, {
			name: "fail - wrong amount type",
			args: []interface{}{
				testSender,
				testInput,
				testOutput,
				3,
				testSlippagePercentage,
				testWindowSeconds,
				testReceiver,
			},
			expPass:     false,
			errContains: "invalid type for amount: expected big.Int, received int",
		}, {
			name: "fail - wrong slippage percentage type",
			args: []interface{}{
				testSender,
				testInput,
				testOutput,
				testAmount,
				10,
				testWindowSeconds,
				testReceiver,
			},
			expPass:     false,
			errContains: "invalid type for slippagePercentage: expected uint8, received int",
		}, {
			name: "fail - wrong window seconds type",
			args: []interface{}{
				testSender,
				testInput,
				testOutput,
				testAmount,
				testSlippagePercentage,
				uint16(20),
				testReceiver,
			},
			expPass:     false,
			errContains: "invalid type for windowSeconds: expected uint64, received uint16",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			swapPacketData, err := osmosisoutpost.ParseSwapPacketData(tc.args)

			if tc.expPass {
				require.NoError(t, err, "expected no error while creating memo")
				require.Equal(
					t,
					osmosisoutpost.SwapPacketData{
						Sender:             testSender,
						Input:              testInput,
						Output:             testOutput,
						Amount:             testAmount,
						SlippagePercentage: testSlippagePercentage,
						WindowSeconds:      testWindowSeconds,
						SwapReceiver:       testReceiver,
					},
					swapPacketData,
				)
			} else {
				require.Error(t, err, "expected error while validating the memo")
				require.Contains(t, err.Error(), tc.errContains, "expected different error")
			}
		})
	}
}

func TestValidateMemo(t *testing.T) {
	t.Parallel()

	receiver := "cosmos1c2m73hdt6f37w9jqpqps5t3ha3st99dcsp7lf5"
	onFailedDelivery := "do_nothing"
	slippagePercentage := uint8(10)
	windowSeconds := uint64(30)

	testCases := []struct {
		name               string
		receiver           string
		onFailedDelivery   string
		slippagePercentage uint8
		windowSeconds      uint64
		expPass            bool
		errContains        string
	}{
		{
			name:               "success - valid packet",
			receiver:           receiver,
			onFailedDelivery:   onFailedDelivery,
			slippagePercentage: slippagePercentage,
			windowSeconds:      windowSeconds,
			expPass:            true,
		}, {
			name:               "fail - empty receiver",
			receiver:           "",
			onFailedDelivery:   onFailedDelivery,
			slippagePercentage: slippagePercentage,
			windowSeconds:      windowSeconds,
			expPass:            false,
			errContains:        fmt.Sprint(bech32.ErrInvalidLength(len("")).Error()),
		}, {
			name:               "fail - on failed delivery empty",
			receiver:           receiver,
			onFailedDelivery:   "",
			slippagePercentage: slippagePercentage,
			windowSeconds:      windowSeconds,
			expPass:            false,
			errContains:        fmt.Sprintf(osmosisoutpost.ErrEmptyOnFailedDelivery),
		}, {
			name:               "fail - over max slippage percentage",
			receiver:           receiver,
			onFailedDelivery:   onFailedDelivery,
			slippagePercentage: osmosisoutpost.MaxSlippagePercentage + 1,
			windowSeconds:      windowSeconds,
			expPass:            false,
			errContains:        fmt.Sprintf(osmosisoutpost.ErrSlippagePercentage),
		}, {
			name:               "fail - zero slippage percentage",
			receiver:           receiver,
			onFailedDelivery:   onFailedDelivery,
			slippagePercentage: 0,
			windowSeconds:      windowSeconds,
			expPass:            false,
			errContains:        fmt.Sprintf(osmosisoutpost.ErrSlippagePercentage),
		}, {
			name:               "fail - over max window seconds",
			receiver:           receiver,
			onFailedDelivery:   onFailedDelivery,
			slippagePercentage: slippagePercentage,
			windowSeconds:      osmosisoutpost.MaxWindowSeconds + 1,
			expPass:            false,
			errContains:        fmt.Sprintf(osmosisoutpost.ErrWindowSeconds),
		}, {
			name:               "fail - zero window seconds",
			receiver:           receiver,
			onFailedDelivery:   onFailedDelivery,
			slippagePercentage: slippagePercentage,
			windowSeconds:      0,
			expPass:            false,
			errContains:        fmt.Sprintf(osmosisoutpost.ErrWindowSeconds),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Variable used for the memo that are not parameters for the tests.
			output := "output"
			nextMemo := ""
			contract := "contract"

			packet := osmosisoutpost.CreatePacketWithMemo(
				output, tc.receiver, contract, tc.slippagePercentage, tc.windowSeconds, tc.onFailedDelivery, nextMemo,
			)

			err := packet.Memo.Validate()

			if tc.expPass {
				require.NoError(t, err, "expected no error while creating memo")
			} else {
				require.Error(t, err, "expected error while validating the memo")
				require.Contains(t, err.Error(), tc.errContains, "expected different error")
			}
		})
	}
}

func TestValidateInputOutput(t *testing.T) {
	t.Parallel()

	aevmosDenom := "aevmos"
	stakingDenom := "aevmos"
	portID := "transfer"
	channelID := "channel-0"
	uosmosDenom := utils.ComputeIBCDenom(portID, channelID, osmosisoutpost.OsmosisDenom)
	validInputs := []string{aevmosDenom, uosmosDenom}

	testCases := []struct {
		name         string
		inputDenom   string
		outputDenom  string
		stakingDenom string
		portID       string
		channelID    string
		expPass      bool
		errContains  string
	}{
		{
			name:         "pass - correct input and output",
			inputDenom:   aevmosDenom,
			outputDenom:  uosmosDenom,
			stakingDenom: stakingDenom,
			portID:       portID,
			channelID:    channelID,
			expPass:      true,
		},
		{
			name:         "fail - input equal to output aevmos",
			inputDenom:   aevmosDenom,
			outputDenom:  aevmosDenom,
			stakingDenom: stakingDenom,
			portID:       portID,
			channelID:    channelID,
			expPass:      false,
			errContains:  fmt.Sprintf(osmosisoutpost.ErrInputEqualOutput, aevmosDenom),
		},
		{
			name:         "fail - input equal to output ibc osmo",
			inputDenom:   uosmosDenom,
			outputDenom:  uosmosDenom,
			stakingDenom: stakingDenom,
			portID:       portID,
			channelID:    channelID,
			expPass:      false,
			errContains:  fmt.Sprintf(osmosisoutpost.ErrInputEqualOutput, uosmosDenom),
		},
		{
			name:         "fail - invalid input",
			inputDenom:   "token",
			outputDenom:  uosmosDenom,
			stakingDenom: stakingDenom,
			portID:       portID,
			channelID:    channelID,
			expPass:      false,
			errContains:  fmt.Sprintf(osmosisoutpost.ErrInputTokenNotSupported, validInputs),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := osmosisoutpost.ValidateInputOutput(tc.inputDenom, tc.outputDenom, tc.stakingDenom, tc.portID, tc.channelID)
			if tc.expPass {
				require.NoError(t, err, "expected no error while creating memo")
			} else {
				require.Error(t, err, "expected error while validating the memo")
				require.Contains(t, err.Error(), tc.errContains, "expected different error")
			}
		})
	}
}

func TestCreateOnFailedDeliveryField(t *testing.T) {
	t.Parallel()

	receiver := "osmo1c2m73hdt6f37w9jqpqps5t3ha3st99dcc6d0lx"
	testCases := []struct {
		name     string
		receiver string
		expRes   string
	}{
		{
			name:     "pass - receiver osmo bech32",
			receiver: receiver,
			expRes:   receiver,
		},
		{
			name:     "pass - receiver osmo bech32",
			receiver: "receiver",
			expRes:   osmosisoutpost.DefaultOnFailedDelivery,
		},
		{
			name:     "pass - convert receiver to osmo bech32",
			receiver: "cosmos1c2m73hdt6f37w9jqpqps5t3ha3st99dcsp7lf5",
			expRes:   receiver,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			onFailedDelivery := osmosisoutpost.CreateOnFailedDeliveryField(tc.receiver)

			require.Contains(t, onFailedDelivery, tc.expRes)
		})
	}
}
