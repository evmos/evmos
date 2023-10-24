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

	testCases := []struct {
		name        string
		args        []interface{}
		expPass     bool
		errContains string
	}{
		{
			name: "pass - valid payload",
			args: []interface{}{
				common.HexToAddress("sender"),
				common.HexToAddress("input"),
				common.HexToAddress("output"),
				big.NewInt(3),
				uint8(10),
				uint64(20),
				"cosmos1c2m73hdt6f37w9jqpqps5t3ha3st99dcsp7lf5",
			},
			expPass: true,
		}, {
			name: "fail - wrong sender type",
			args: []interface{}{
				"sender",
				common.HexToAddress("input"),
				common.HexToAddress("output"),
				big.NewInt(3),
				uint8(10),
				uint64(20),
				"cosmos1c2m73hdt6f37w9jqpqps5t3ha3st99dcsp7lf5",
			},
			expPass:     false,
			errContains: "invalid type for sender: expected common.Address, received string",
		}, {
			name: "fail - wrong input type",
			args: []interface{}{
				common.HexToAddress("sender"),
				"input",
				common.HexToAddress("output"),
				big.NewInt(3),
				uint8(10),
				uint64(20),
				"cosmos1c2m73hdt6f37w9jqpqps5t3ha3st99dcsp7lf5",
			},
			expPass:     false,
			errContains: "invalid type for input: expected common.Address, received string",
		}, {
			name: "fail - wrong output type",
			args: []interface{}{
				common.HexToAddress("sender"),
				common.HexToAddress("input"),
				"output",
				big.NewInt(3),
				uint8(10),
				uint64(20),
				"cosmos1c2m73hdt6f37w9jqpqps5t3ha3st99dcsp7lf5",
			},
			expPass:     false,
			errContains: "invalid type for output: expected common.Address, received string",
		}, {
			name: "fail - wrong amount type",
			args: []interface{}{
				common.HexToAddress("sender"),
				common.HexToAddress("input"),
				common.HexToAddress("output"),
				3,
				uint8(10),
				uint64(20),
				"cosmos1c2m73hdt6f37w9jqpqps5t3ha3st99dcsp7lf5",
			},
			expPass:     false,
			errContains: "invalid type for amount: expected big.Int, received int",
		}, {
			name: "fail - wrong slippage percentage type",
			args: []interface{}{
				common.HexToAddress("input"),
				common.HexToAddress("output"),
				common.HexToAddress("output"),
				big.NewInt(3),
				10,
				uint64(20),
				"cosmos1c2m73hdt6f37w9jqpqps5t3ha3st99dcsp7lf5",
			},
			expPass:     false,
			errContains: "invalid type for slippagePercentage: expected uint8, received int",
		}, {
			name: "fail - wrong window seconds type",
			args: []interface{}{
				common.HexToAddress("sender"),
				common.HexToAddress("input"),
				common.HexToAddress("output"),
				big.NewInt(3),
				uint8(10),
				uint16(20),
				"cosmos1c2m73hdt6f37w9jqpqps5t3ha3st99dcsp7lf5",
			},
			expPass:     false,
			errContains: "invalid type for windowSeconds: expected uint64, received uint16",
		}, {
			name: "fail - receiver not bech32",
			args: []interface{}{
				common.HexToAddress("sender"),
				common.HexToAddress("input"),
				common.HexToAddress("output"),
				big.NewInt(3),
				uint8(10),
				uint16(20),
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

func TestValidateSwapTokens(t *testing.T) {
	t.Parallel()

	portID := "transfer"
	channelID := "channel-0"
	osmoVoucher := transfertypes.DenomTrace{
		Path:      fmt.Sprintf("%s/%s", portID, channelID),
		BaseDenom: osmosisoutpost.OsmosisDenom,
	}.IBCDenom()
	stakingDenom := "aevmos"

	testCases := []struct {
		name        string
		input       string
		output      string
		expPass     bool
		errContains string
	}{
		{
			name:    "success - valid tokens uosmo for aevmos",
			input:   osmoVoucher,
			output:  stakingDenom,
			expPass: true,
		}, {
			name:    "success - valid tokens aevmos for uosmo",
			input:   stakingDenom,
			output:  osmoVoucher,
			expPass: true,
		}, {
			name:        "fail - input equal to output aevmos",
			input:       stakingDenom,
			output:      stakingDenom,
			expPass:     false,
			errContains: osmosisoutpost.ErrInputEqualOutput,
		}, {
			name:        "fail - input equal to output uosmos",
			input:       osmoVoucher,
			output:      osmoVoucher,
			expPass:     false,
			errContains: osmosisoutpost.ErrInputEqualOutput,
		}, {
			name:    "fail - input not supported",
			input:   "btc",
			output:  stakingDenom,
			expPass: false,
			errContains: fmt.Sprintf(
				osmosisoutpost.ErrInputTokenNotSupported,
				[]string{stakingDenom, osmoVoucher},
			),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := osmosisoutpost.ValidateSwapTokens(tc.input, tc.output, stakingDenom, portID, channelID)

			if tc.expPass {
				require.NoError(t, err, "expected no error while creating memo")
			} else {
				require.Error(t, err, "expected error while validating the memo")
				require.Contains(t, err.Error(), tc.errContains, "expected different error")
			}
		})
	}
}

func TestValidateSwapParameters(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		slippagePercentage uint8
		windowSeconds      uint64
		expPass            bool
		errContains        string
	}{
		{
			name:               "success - valid parameters",
			slippagePercentage: 10,
			windowSeconds:      30,
			expPass:            true,
		}, {
			name:               "fail - over max slippage",
			slippagePercentage: osmosisoutpost.MaxSlippagePercentage + 1,
			windowSeconds:      30,
			expPass:            false,
			errContains:        fmt.Sprintf(osmosisoutpost.ErrMaxSlippagePercentage),
		}, {
			name:               "fail - over max window seconds",
			slippagePercentage: 10,
			windowSeconds:      osmosisoutpost.MaxWindowSeconds + 1,
			expPass:            false,
			errContains:        fmt.Sprintf(osmosisoutpost.ErrMaxWindowSeconds),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := osmosisoutpost.ValidateSwapParameters(tc.slippagePercentage, tc.windowSeconds)

			if tc.expPass {
				require.NoError(t, err, "expected no error while creating memo")
			} else {
				require.Error(t, err, "expected error while validating the memo")
				require.Contains(t, err.Error(), tc.errContains, "expected different error")
			}
		})
	}
}
