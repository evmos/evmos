package osmosis_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	osmosisoutpost "github.com/evmos/evmos/v15/precompiles/outposts/osmosis"
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
		expMemo            bool
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
			expMemo:            false,
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
			expMemo:            true,
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

			if tc.expMemo {
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
		}, {
			name: "fail - receiver not bech32",
			args: []interface{}{
				testSender,
				testInput,
				testOutput,
				testAmount,
				testSlippagePercentage,
				testWindowSeconds,
				"address",
			},
			expPass: false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _, _, _, _, _, _, err := osmosisoutpost.ParseSwapPacketData(tc.args)

			if tc.expPass {
				require.NoError(t, err, "expected no error while creating memo")
			} else {
				require.Error(t, err, "expected error while validating the memo")
				require.Contains(t, err.Error(), tc.errContains, "expected different error")
			}
		})
	}
}

func TestValidatePacket(t *testing.T) {
	t.Parallel()

	portID := "transfer"
	channelID := "channel-0"
	osmoVoucher := transfertypes.DenomTrace{
		Path:      fmt.Sprintf("%s/%s", portID, channelID),
		BaseDenom: osmosisoutpost.OsmosisDenom,
	}.IBCDenom()
	stakingDenom := "aevmos"
	slippagePercentage := uint8(10)
	windowSeconds := uint64(30)

	testCases := []struct {
		name               string
		input              string
		outputDenom        string
		slippagePercentage uint8
		windowSeconds      uint64
		expPass            bool
		errContains        string
	}{
		{
			name:               "success - valid packet",
			input:              osmoVoucher,
			outputDenom:        stakingDenom,
			slippagePercentage: slippagePercentage,
			windowSeconds:      windowSeconds,
			expPass:            true,
		}, {
			name:               "success - valid tokens aevmos for uosmo",
			input:              stakingDenom,
			outputDenom:        osmoVoucher,
			slippagePercentage: slippagePercentage,
			windowSeconds:      windowSeconds,
			expPass:            true,
		}, {
			name:               "fail - input equal to output aevmos",
			input:              stakingDenom,
			outputDenom:        stakingDenom,
			slippagePercentage: slippagePercentage,
			windowSeconds:      windowSeconds,
			expPass:            false,
			errContains:        osmosisoutpost.ErrInputEqualOutput,
		}, {
			name:               "fail - input equal to output uosmos",
			input:              osmoVoucher,
			outputDenom:        osmoVoucher,
			slippagePercentage: slippagePercentage,
			windowSeconds:      windowSeconds,
			expPass:            false,
			errContains:        osmosisoutpost.ErrInputEqualOutput,
		}, {
			name:               "fail - input not supported",
			input:              "btc",
			outputDenom:        stakingDenom,
			slippagePercentage: slippagePercentage,
			windowSeconds:      windowSeconds,
			expPass:            false,
			errContains: fmt.Sprintf(
				osmosisoutpost.ErrInputTokenNotSupported,
				[]string{stakingDenom, osmoVoucher},
			),
		}, {
			name:               "fail - over max slippage",
			input:              osmoVoucher,
			outputDenom:        stakingDenom,
			slippagePercentage: osmosisoutpost.MaxSlippagePercentage + 1,
			windowSeconds:      windowSeconds,
			expPass:            false,
			errContains:        fmt.Sprintf(osmosisoutpost.ErrMaxSlippagePercentage),
		}, {
			name:               "fail - over max window seconds",
			input:              osmoVoucher,
			outputDenom:        stakingDenom,
			slippagePercentage: slippagePercentage,
			windowSeconds:      osmosisoutpost.MaxWindowSeconds + 1,
			expPass:            false,
			errContains:        fmt.Sprintf(osmosisoutpost.ErrMaxWindowSeconds),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			receiver := "receiver"
			nextMemo := ""
			onFailedDelivery := "do_nothing"
			contract := "xcsContract"

			packet := osmosisoutpost.CreatePacketWithMemo(
				tc.outputDenom, receiver, contract, tc.slippagePercentage, tc.windowSeconds, onFailedDelivery, nextMemo,
			)

			err := packet.Validate(tc.input, stakingDenom, portID, channelID)

			if tc.expPass {
				require.NoError(t, err, "expected no error while creating memo")
			} else {
				require.Error(t, err, "expected error while validating the memo")
				require.Contains(t, err.Error(), tc.errContains, "expected different error")
			}
		})
	}
}
