package osmosis_test

import (
	"testing"
	"fmt"

	osmosisoutpost "github.com/evmos/evmos/v15/precompiles/outposts/osmosis"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/stretchr/testify/require"
)

/*
func TestCreateMemo(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name        string
		outputDenom string
		receiver    string
		contract	string
		slippagePercentage uint8
		windowSeconds uint64
		onFailedDelivery string
		expPass     bool
		errContains string
	}{
		{
			name:     "success - create memo",
			outputDenom: "uosmo",
			receiver: "receiveraddress",
			contract: "xcscontract",
			slippagePercentage: 5,
			windowSeconds: 10,
			onFailedDelivery: "",
			expPass:  true,
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			memo, err := osmosisoutpost.CreateMemo(tc.outputDenom, tc.receiver, tc.contract, tc.slippagePercentage, tc.windowSeconds, tc.onFailedDelivery)
			if tc.expPass {
				require.NoError(t, err, "expected no error while creating memo")
				require.NotEmpty(t, memo, "expected memo not to be empty")
			} else {
				require.Error(t, err, "expected error while creating memo")
				require.Contains(t, err.Error(), tc.errContains, "expected different error")
			}
		})
	}
}
*/

func TestValidateSwap(t *testing.T) {
	t.Parallel()

	memo := osmosisoutpost.CreateMemo("atom", "receiveraddress", "xcscontract", 5, 10, "do_nothing")
	portID := "transfer"
	channelID := osmosisoutpost.OsmosisChannelIDMainnet
	osmoVoucher := transfertypes.DenomTrace{
		Path:      fmt.Sprintf("%s/%s", portID, channelID),
		BaseDenom: "osmo",
	}.IBCDenom()
	stakingDenom := "aevmos"

	testcases := []struct {
		name        string
		modifier    func(osmosisoutpost.Memo) osmosisoutpost.Memo
		input string
		expPass     bool
		errContains string
	}{
		{
			name:     "success - create memo with ibc osmo",
			modifier: func(memo osmosisoutpost.Memo) osmosisoutpost.Memo { 
				return memo
			},
			input: transfertypes.DenomTrace{
				Path:      fmt.Sprintf("%s/%s", portID, channelID),
				BaseDenom: "osmo",
			}.IBCDenom(),
			expPass:  true,
		},
		{
			name:     "success - create memo with aevmos",
			modifier: func(memo osmosisoutpost.Memo) osmosisoutpost.Memo { 
				return memo
			},
			input: "aevmos",
			expPass:  true,
		},
		{
			name:     "fail - input and output equal",
			modifier: func(memo osmosisoutpost.Memo) osmosisoutpost.Memo {
				memo.Msg.OsmosisSwap.OutputDenom = osmoVoucher
				return memo
			},
			input: transfertypes.DenomTrace{
				Path:      fmt.Sprintf("%s/%s", portID, channelID),
				BaseDenom: "osmo",
			}.IBCDenom(),
			expPass:  false,
			errContains: osmosisoutpost.ErrInputEqualOutput,
		},
		{
			name:     "fail - input and output equal",
			modifier: func(memo osmosisoutpost.Memo) osmosisoutpost.Memo {
				return memo
			},
			input: "eth",
			expPass:  false,
			errContains: fmt.Sprintf(
				osmosisoutpost.ErrInputTokenNotSupported,
				[]string{stakingDenom, osmoVoucher},
			),
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newMemo := tc.modifier(*memo)

			err := newMemo.ValidateSwap(portID, channelID, stakingDenom, tc.input)

			if tc.expPass {
				require.NoError(t, err, "expected no error while creating memo")
			} else {
				require.Error(t, err, "expected error while validating the memo")
				require.Contains(t, err.Error(), tc.errContains, "expected different error")
			}
		})
	}
}

/*
func TestValidateSwap(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name string
		portID string
		channelID string
		input string
		output string
		stakingDenom string
		slippagePercentage uint64
		windowSeconds uint64
		expPass     bool
		errContains string
	}{
		{
			name:     "fail - input and output cannot be the same",
			portID: "transfer",
			channelID: "channel-0",
			input: "aevmosAddress",
			output: "aevmosAddress",
			stakingDenom: "aevmos",
			slippagePercentage: osmosisoutpost.DefaultSlippagePercentage,
			windowSeconds: osmosisoutpost.DefaultWindowSeconds,
			expPass:  false,
			errContains: fmt.Sprintf(osmosisoutpost.ErrInputEqualOutput),
		},
		{
			name:     "fail - not allowed input",
			portID: "transfer",
			channelID: "channel-0",
			input: "ethAddress",
			output: "uosmoAddress",
			stakingDenom: "aevmos",
			slippagePercentage: osmosisoutpost.DefaultSlippagePercentage,
			windowSeconds: osmosisoutpost.DefaultWindowSeconds,
			expPass:  false,
			errContains: fmt.Sprintf(osmosisoutpost.ErrInputTokenNotSupported, ""),
		},
		{
			name:     "fail - over max slippage percentage",
			portID: "transfer",
			channelID: "channel-0",
			input: "aevmosAddress",
			output: "uosmoAddress",
			stakingDenom: "aevmos",
			slippagePercentage: osmosisoutpost.MaxSlippagePercentage + 1,
			windowSeconds: osmosisoutpost.DefaultWindowSeconds,
			expPass:  false,
			errContains: fmt.Sprintf(osmosisoutpost.ErrInvalidSlippagePercentage, osmosisoutpost.MaxSlippagePercentage),
		},
		{
			name:     "fail - over max window seconds",
			portID: "transfer",
			channelID: "channel-0",
			input: "aevmosAddress",
			output: "uosmoAddress",
			stakingDenom: "aevmos",
			slippagePercentage: osmosisoutpost.DefaultSlippagePercentage,
			windowSeconds: osmosisoutpost.MaxWindowSeconds + 1,
			expPass:  false,
			errContains: fmt.Sprintf(osmosisoutpost.ErrInvalidWindowSeconds, osmosisoutpost.MaxWindowSeconds),
		},
		{
			name:     "pass - correct inputs",
			portID: "transfer",
			channelID: "channel-0",
			input: "aevmosAddress",
			output: "uosmoAddress",
			stakingDenom: "aevmos",
			slippagePercentage: osmosisoutpost.DefaultSlippagePercentage,
			windowSeconds: osmosisoutpost.DefaultWindowSeconds,
			expPass:  true,
			errContains: "",
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := osmosisoutpost.ValidateSwap(tc.portID, tc.channelID, tc.input, tc.output, tc.stakingDenom, tc.slippagePercentage, tc.windowSeconds)
			if tc.expPass {
				require.NoError(t, err, "expected no error while creating memo")
				require.NotEmpty(t, "expected memo not to be empty")
			} else {
				require.Error(t, err, "expected error while creating memo")
				require.Contains(t, err.Error(), tc.errContains, "expected different error")
			}
		})
	}	
}
*/