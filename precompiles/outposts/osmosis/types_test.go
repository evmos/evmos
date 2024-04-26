package osmosis_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	cmn "github.com/evmos/evmos/v18/precompiles/common"
	osmosisoutpost "github.com/evmos/evmos/v18/precompiles/outposts/osmosis"
	"github.com/evmos/evmos/v18/utils"
)

func TestCreatePacketWithMemo(t *testing.T) {
	t.Parallel()

	contract := "evmos1vl0x3xr0zwgrllhdzxxlkal7txnnk56q3552x7"
	receiver := "evmos1vl0x3xr0zwgrllhdzxxlkal7txnnk56q3552x7"
	doNothing := "do_nothing"

	testCases := []struct {
		name               string
		outputDenom        string
		receiver           string
		contract           string
		slippagePercentage uint8
		windowSeconds      uint64
		onFailedDelivery   interface{}
		nextMemo           string
		expMemo            string
	}{
		{
			name:               "pass - correct string without memo",
			outputDenom:        utils.BaseDenom,
			receiver:           receiver,
			contract:           contract,
			slippagePercentage: 10,
			windowSeconds:      30,
			onFailedDelivery:   doNothing,
			nextMemo:           "",
			expMemo:            "{\"wasm\":{\"contract\":\"evmos1vl0x3xr0zwgrllhdzxxlkal7txnnk56q3552x7\",\"msg\":{\"osmosis_swap\":{\"output_denom\":\"aevmos\",\"slippage\":{\"twap\":{\"slippage_percentage\":\"10\",\"window_seconds\":30}},\"receiver\":\"evmos1vl0x3xr0zwgrllhdzxxlkal7txnnk56q3552x7\",\"on_failed_delivery\":\"do_nothing\"}}}}",
		},
		{
			name:               "pass - correct string with memo",
			outputDenom:        utils.BaseDenom,
			receiver:           receiver,
			contract:           contract,
			slippagePercentage: 10,
			windowSeconds:      30,
			onFailedDelivery:   doNothing,
			nextMemo:           "a next memo",
			expMemo:            "{\"wasm\":{\"contract\":\"evmos1vl0x3xr0zwgrllhdzxxlkal7txnnk56q3552x7\",\"msg\":{\"osmosis_swap\":{\"output_denom\":\"aevmos\",\"slippage\":{\"twap\":{\"slippage_percentage\":\"10\",\"window_seconds\":30}},\"receiver\":\"evmos1vl0x3xr0zwgrllhdzxxlkal7txnnk56q3552x7\",\"on_failed_delivery\":\"do_nothing\",\"next_memo\":\"a next memo\"}}}}",
		},
		{
			name:               "pass - correct string with memo and recovery address",
			outputDenom:        utils.BaseDenom,
			receiver:           receiver,
			contract:           contract,
			slippagePercentage: 10,
			windowSeconds:      30,
			onFailedDelivery:   osmosisoutpost.RecoveryAddress{"osmo1g8j7tgfam7kmj86zks5rcfxruf9lzp87u8mwdf"},
			nextMemo:           "a next memo",
			expMemo:            "{\"wasm\":{\"contract\":\"evmos1vl0x3xr0zwgrllhdzxxlkal7txnnk56q3552x7\",\"msg\":{\"osmosis_swap\":{\"output_denom\":\"aevmos\",\"slippage\":{\"twap\":{\"slippage_percentage\":\"10\",\"window_seconds\":30}},\"receiver\":\"evmos1vl0x3xr0zwgrllhdzxxlkal7txnnk56q3552x7\",\"on_failed_delivery\":{\"local_recovery_addr\":\"osmo1g8j7tgfam7kmj86zks5rcfxruf9lzp87u8mwdf\"},\"next_memo\":\"a next memo\"}}}}",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			packet := osmosisoutpost.CreatePacketWithMemo(
				tc.outputDenom, tc.receiver, tc.contract, tc.slippagePercentage, tc.windowSeconds, tc.onFailedDelivery, tc.nextMemo,
			)
			memo := packet.String()
			require.Equal(t, tc.expMemo, memo)
			err := ValidateAndParseWasmRoutedMemo(memo, tc.receiver)
			require.NoError(t, err, "memo is not a valid wasm routed JSON formatted string")
		})

	}
}

// TestParseSwapPacketData is mainly to test that the returned error of the
// parser is clear and contains the correct data type. For this reason the
// expected error has been hardcoded as a string litera.
func (s *PrecompileTestSuite) TestParseSwapPacketData() {
	sender := common.HexToAddress("sender")
	input := common.HexToAddress("input")
	output := common.HexToAddress("output")
	amount := big.NewInt(3)
	slippagePercentage := uint8(10)
	windowSeconds := uint64(20)
	receiver := "evmos1vl0x3xr0zwgrllhdzxxlkal7txnnk56q3552x7"

	testCases := []struct {
		name        string
		args        []interface{}
		expPass     bool
		errContains string
	}{
		{
			name: "pass - valid payload of type SwapPayload",
			args: []interface{}{
				osmosisoutpost.SwapPacketData{
					ChannelID:          ChannelID,
					XcsContract:        XCSContract,
					Sender:             sender,
					Input:              input,
					Output:             output,
					Amount:             amount,
					SlippagePercentage: slippagePercentage,
					WindowSeconds:      windowSeconds,
					SwapReceiver:       receiver,
				},
			},
			expPass: true,
		},
		{
			name:        "fail - invalid number of args",
			args:        []interface{}{},
			expPass:     false,
			errContains: fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 1, 0),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			method := s.precompile.Methods[osmosisoutpost.SwapMethod]

			swapPacketData, err := osmosisoutpost.ParseSwapPacketData(&method, tc.args)

			if tc.expPass {
				s.NoError(err, "expected no error while creating memo")
				s.Equal(
					osmosisoutpost.SwapPacketData{
						ChannelID:          ChannelID,
						XcsContract:        XCSContract,
						Sender:             sender,
						Input:              input,
						Output:             output,
						Amount:             amount,
						SlippagePercentage: slippagePercentage,
						WindowSeconds:      windowSeconds,
						SwapReceiver:       receiver,
					},
					swapPacketData,
				)
			} else {
				s.Error(err, "expected error while validating the memo")
				s.Contains(err.Error(), tc.errContains, "expected different error")
			}
		})
	}
}

func TestValidateMemo(t *testing.T) {
	t.Parallel()

	receiver := "evmos1vl0x3xr0zwgrllhdzxxlkal7txnnk56q3552x7"
	contract := "osmo1a34wxsxjwvtz3ua4hnkh4lv3d4qrgry0fhkasppplphwu5k538tqcyms9x"
	onFailedDelivery := "do_nothing"
	slippagePercentage := uint8(10)
	windowSeconds := uint64(30)
	// Variable used for the memo that are not parameters for the tests.
	output := "output"
	nextMemo := ""

	testCases := []struct {
		name               string
		receiver           string
		contractAddress    string
		onFailedDelivery   string
		slippagePercentage uint8
		windowSeconds      uint64
		expPass            bool
		errContains        string
	}{
		{
			name:               "success - valid packet",
			receiver:           receiver,
			contractAddress:    contract,
			onFailedDelivery:   onFailedDelivery,
			slippagePercentage: slippagePercentage,
			windowSeconds:      windowSeconds,
			expPass:            true,
		}, {
			name:               "fail - not evmos bech32",
			receiver:           "cosmos1c2m73hdt6f37w9jqpqps5t3ha3st99dcsp7lf5",
			contractAddress:    contract,
			onFailedDelivery:   onFailedDelivery,
			slippagePercentage: slippagePercentage,
			windowSeconds:      windowSeconds,
			expPass:            false,
			errContains:        fmt.Sprintf(osmosisoutpost.ErrReceiverAddress, "not a valid evmos address"),
		}, {
			name:               "fail - not bech32",
			receiver:           "cosmos",
			contractAddress:    contract,
			onFailedDelivery:   onFailedDelivery,
			slippagePercentage: slippagePercentage,
			windowSeconds:      windowSeconds,
			expPass:            false,
			errContains:        fmt.Sprintf(osmosisoutpost.ErrReceiverAddress, "not a valid evmos address"),
		}, {
			name:               "fail - empty receiver",
			receiver:           "",
			contractAddress:    contract,
			onFailedDelivery:   onFailedDelivery,
			slippagePercentage: slippagePercentage,
			windowSeconds:      windowSeconds,
			expPass:            false,
			errContains:        fmt.Sprintf(osmosisoutpost.ErrReceiverAddress, "not a valid evmos address"),
		}, {
			name:               "fail - on failed delivery empty",
			receiver:           receiver,
			contractAddress:    contract,
			onFailedDelivery:   "",
			slippagePercentage: slippagePercentage,
			windowSeconds:      windowSeconds,
			expPass:            false,
			errContains:        fmt.Sprintf(osmosisoutpost.ErrEmptyOnFailedDelivery),
		}, {
			name:               "fail - over max slippage percentage",
			receiver:           receiver,
			contractAddress:    contract,
			onFailedDelivery:   onFailedDelivery,
			slippagePercentage: osmosisoutpost.MaxSlippagePercentage + 1,
			windowSeconds:      windowSeconds,
			expPass:            false,
			errContains:        fmt.Sprintf(osmosisoutpost.ErrSlippagePercentage),
		}, {
			name:               "fail - zero slippage percentage",
			receiver:           receiver,
			contractAddress:    contract,
			onFailedDelivery:   onFailedDelivery,
			slippagePercentage: 0,
			windowSeconds:      windowSeconds,
			expPass:            false,
			errContains:        fmt.Sprintf(osmosisoutpost.ErrSlippagePercentage),
		}, {
			name:               "fail - over max window seconds",
			receiver:           receiver,
			contractAddress:    contract,
			onFailedDelivery:   onFailedDelivery,
			slippagePercentage: slippagePercentage,
			windowSeconds:      osmosisoutpost.MaxWindowSeconds + 1,
			expPass:            false,
			errContains:        fmt.Sprintf(osmosisoutpost.ErrWindowSeconds),
		}, {
			name:               "fail - zero window seconds",
			receiver:           receiver,
			contractAddress:    contract,
			onFailedDelivery:   onFailedDelivery,
			slippagePercentage: slippagePercentage,
			windowSeconds:      0,
			expPass:            false,
			errContains:        fmt.Sprintf(osmosisoutpost.ErrWindowSeconds),
		}, {
			name:               "fail - empty contract address",
			receiver:           receiver,
			contractAddress:    "",
			onFailedDelivery:   onFailedDelivery,
			slippagePercentage: slippagePercentage,
			windowSeconds:      windowSeconds,
			expPass:            false,
			errContains:        fmt.Sprintf(osmosisoutpost.ErrEmptyContractAddress),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			packet := osmosisoutpost.CreatePacketWithMemo(
				output, tc.receiver, tc.contractAddress, tc.slippagePercentage, tc.windowSeconds, tc.onFailedDelivery, nextMemo,
			)

			err := packet.Validate()

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

	portID := "transfer"
	channelID := "channel-0"
	uosmosDenom := utils.ComputeIBCDenom(portID, channelID, osmosisoutpost.OsmosisDenom)
	validInputs := []string{utils.BaseDenom, uosmosDenom}

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
			inputDenom:   utils.BaseDenom,
			outputDenom:  uosmosDenom,
			stakingDenom: utils.BaseDenom,
			portID:       portID,
			channelID:    channelID,
			expPass:      true,
		},
		{
			name:         "fail - input equal to output aevmos",
			inputDenom:   utils.BaseDenom,
			outputDenom:  utils.BaseDenom,
			stakingDenom: utils.BaseDenom,
			portID:       portID,
			channelID:    channelID,
			expPass:      false,
			errContains:  fmt.Sprintf(osmosisoutpost.ErrInputEqualOutput, utils.BaseDenom),
		},
		{
			name:         "fail - input equal to output ibc osmo",
			inputDenom:   uosmosDenom,
			outputDenom:  uosmosDenom,
			stakingDenom: utils.BaseDenom,
			portID:       portID,
			channelID:    channelID,
			expPass:      false,
			errContains:  fmt.Sprintf(osmosisoutpost.ErrInputEqualOutput, uosmosDenom),
		},
		{
			name:         "fail - invalid input",
			inputDenom:   "token",
			outputDenom:  uosmosDenom,
			stakingDenom: utils.BaseDenom,
			portID:       portID,
			channelID:    channelID,
			expPass:      false,
			errContains:  fmt.Sprintf(osmosisoutpost.ErrDenomNotSupported, validInputs),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			evmosChannel := osmosisoutpost.NewIBCChannel(tc.portID, tc.channelID)

			err := osmosisoutpost.ValidateInputOutput(tc.inputDenom, tc.outputDenom, tc.stakingDenom, evmosChannel)
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

	address := "osmo1c2m73hdt6f37w9jqpqps5t3ha3st99dcc6d0lx"
	testCases := []struct {
		name    string
		address string
		expRes  interface{}
	}{
		{
			name:    "receiver osmo bech32",
			address: address,
			expRes:  osmosisoutpost.RecoveryAddress{address},
		},
		{
			name:    "use default do_nothing",
			address: "not_bech_32",
			expRes:  osmosisoutpost.DefaultOnFailedDelivery,
		},
		{
			name:    "convert receiver to osmo bech32",
			address: "cosmos1c2m73hdt6f37w9jqpqps5t3ha3st99dcsp7lf5",
			expRes:  osmosisoutpost.RecoveryAddress{address},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			onFailedDelivery := osmosisoutpost.CreateOnFailedDeliveryField(tc.address)
			require.Equal(t, onFailedDelivery, tc.expRes)
		})
	}
}

func TestConvertToOsmosisRepresentation(t *testing.T) {
	t.Parallel()

	portID := "transfer"
	channelID := "channel-0"
	osmoIBCDenom := utils.ComputeIBCDenom(portID, channelID, osmosisoutpost.OsmosisDenom)
	evmosChannel := osmosisoutpost.NewIBCChannel(portID, channelID)
	osmosisChannel := osmosisoutpost.NewIBCChannel(portID, channelID)

	testCases := []struct {
		name        string
		denom       string
		expPass     bool
		expDenom    string
		errContains string
	}{
		{
			name:     "pass - correct conversion of aevmos",
			denom:    utils.BaseDenom,
			expPass:  true,
			expDenom: "ibc/8EAC8061F4499F03D2D1419A3E73D346289AE9DB89CAB1486B72539572B1915E",
		}, {
			name:     "pass - correct conversion of ibc uosmo",
			denom:    osmoIBCDenom,
			expPass:  true,
			expDenom: "uosmo",
		}, {
			name:        "fail - not allowed token",
			denom:       "token",
			expPass:     false,
			errContains: fmt.Sprintf(osmosisoutpost.ErrDenomNotSupported, []string{utils.BaseDenom, osmoIBCDenom}),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			denom, err := osmosisoutpost.ConvertToOsmosisRepresentation(tc.denom, utils.BaseDenom, evmosChannel, osmosisChannel)
			if tc.expPass {
				require.NoError(t, err, "expected no error while creating memo")
				require.Equal(t, denom, tc.expDenom)
			} else {
				require.Error(t, err, "expected error while validating the memo")
				require.Contains(t, err.Error(), tc.errContains, "expected different error")
			}
		})
	}
}

func TestValidateOsmosisContractAddress(t *testing.T) {
	testCases := []struct {
		name            string
		contractAddress string
		expPass         bool
		errContains     string
	}{
		{
			name:            "fail - empty contract address",
			contractAddress: "",
			expPass:         false,
			errContains:     fmt.Sprintf(osmosisoutpost.ErrInvalidContractAddress),
		},
		{
			name:            "pass - not contract address",
			contractAddress: "osmo1qql8ag4cluz6r4dz28p3w00dnc9w8ueuhnecd2",
			expPass:         false,
			errContains:     fmt.Sprintf(osmosisoutpost.ErrInvalidContractAddress),
		},
		{
			name:            "fail - not osmosis smart contract",
			contractAddress: "evmos18rj46qcpr57m3qncrj9cuzm0gn3km08w5jxxlnw002c9y7xex5xsu74ytz",
			expPass:         false,
			errContains:     fmt.Sprintf(osmosisoutpost.ErrInvalidContractAddress),
		},
		{
			name:            "pass - valid contract address",
			contractAddress: "osmo1a34wxsxjwvtz3ua4hnkh4lv3d4qrgry0fhkasppplphwu5k538tqcyms9x",
			expPass:         true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			err := osmosisoutpost.ValidateOsmosisContractAddress(tc.contractAddress)
			if tc.expPass {
				require.NoError(t, err, "expected no error while validating the contract address")
			} else {
				require.Error(t, err, "expected error while validating the contract address")
				require.Contains(t, err.Error(), tc.errContains)
			}
		})

	}
}
