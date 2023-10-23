package osmosis_test

import (
	"fmt"
	"testing"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	osmosisoutpost "github.com/evmos/evmos/v15/precompiles/outposts/osmosis"
	"github.com/stretchr/testify/require"
)

func TestCreatePacketWithMemo(t *testing.T) {
	t.Parallel()

	packet := osmosisoutpost.CreatePacketWithMemo("aevmos", "receiver", "contract", 10, 30, "osmoAddress")

	jsonPacket, err := packet.ConvertToJSON()
	require.NoError(t, err, "expected no error while creating memo")
	require.NotEmpty(t, jsonPacket, "expected memo not to be empty")
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
