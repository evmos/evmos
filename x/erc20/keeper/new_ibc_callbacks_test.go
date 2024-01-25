package keeper_test

import (
	"testing"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v7/testing"
	ibcmock "github.com/cosmos/ibc-go/v7/testing/mock"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	"github.com/stretchr/testify/suite"
)

type Erc20KeeperTestSuite struct {
	suite.Suite
}

func TestErc20TestSuite(t *testing.T) {
	suite.Run(t, &Erc20KeeperTestSuite{})
}

func newTransferBytes(t transfertypes.FungibleTokenPacketData) []byte {
	return transfertypes.ModuleCdc.MustMarshalJSON(&t)
}

func (suite *Erc20KeeperTestSuite) TestOnRecvPacket() {
	keyring := testkeyring.New(2)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

	sourceChannel := "channel-292"
	destinationChannel := "channel-0"
	prefixedDenom := transfertypes.GetDenomPrefix(transfertypes.PortID, sourceChannel) + "aevmos"
	hopCoin := transfertypes.GetDenomPrefix(transfertypes.PortID, destinationChannel) + "ahopcoin"

	// Cant set precompiles
	// contractAddr, _ := utils.GetIBCDenomAddress("coin")
	// unitNetwork.App.PrecompileKeeper.RegisterERC20Extension(unitNetwork.GetContext(), "coin", contractAddr)

	testCases := []struct {
		name                     string
		transferBytes            []byte
		expectedError            bool
		shouldRegisterPrecompile bool
	}{
		// Test Bad transfer package
		{
			name:          "error - non ics-20 packet",
			transferBytes: ibcgotesting.MockPacketData,
			expectedError: false,
		},
		// Package has an invalid sender
		{
			name: "error - invalid sender",
			transferBytes: newTransferBytes(
				transfertypes.NewFungibleTokenPacketData(
					unitNetwork.GetDenom(),
					"100",
					"",
					keyring.GetAccAddr(1).String(),
					"",
				),
			),
			expectedError: true,
		},
		// Package has an invalid receiver
		{
			name: "error - invalid receiver",
			transferBytes: newTransferBytes(
				transfertypes.NewFungibleTokenPacketData(
					unitNetwork.GetDenom(),
					"100",
					keyring.GetAccAddr(0).String(),
					"",
					"",
				),
			),
			expectedError: true,
		},
		// If we received an IBC from non EVM channel the account should be different
		// If its the same, users can have their funds stuck since they dont have access
		// to the same priv key
		{
			name: "error - sender == receiver from non EVM channel",
			transferBytes: newTransferBytes(
				transfertypes.NewFungibleTokenPacketData(
					unitNetwork.GetDenom(),
					"100",
					keyring.GetAccAddr(0).String(),
					keyring.GetAccAddr(0).String(),
					"",
				),
			),
			expectedError: false,
		},
		// TODO: Check why we do this
		{
			name: "no-op - sender is module account",
			transferBytes: newTransferBytes(
				transfertypes.NewFungibleTokenPacketData(
					unitNetwork.GetDenom(),
					"100",
					unitNetwork.App.AccountKeeper.GetModuleAddress("erc20").String(),
					keyring.GetAccAddr(0).String(),
					"",
				),
			),
			expectedError: false,
		},
		// If transfer coin is evm denom there is no need for any precompile deploy or conversion
		{
			name: "no-op - evm Denom",
			transferBytes: newTransferBytes(
				transfertypes.NewFungibleTokenPacketData(
					prefixedDenom,
					"100",
					keyring.GetAccAddr(0).String(),
					keyring.GetAccAddr(1).String(),
					"",
				),
			),
			expectedError: false,
		},
		// TODO: Check if we should check erc20 is enabled
		// If erc20 module is disabled its not possible to deploy or convert
		{
			name: "no-op - erc20 module param disabled",
			transferBytes: newTransferBytes(
				transfertypes.NewFungibleTokenPacketData(
					unitNetwork.GetDenom(),
					"100",
					keyring.GetAccAddr(0).String(),
					keyring.GetAccAddr(1).String(),
					"",
				),
			),
			expectedError: true,
		},
		// is single hop not registered
		{
			name: "success - is single hop. Should register erc20 extension",
			transferBytes: newTransferBytes(
				transfertypes.NewFungibleTokenPacketData(
					"coin",
					"100",
					keyring.GetAccAddr(0).String(),
					keyring.GetAccAddr(1).String(),
					"",
				),
			),
			expectedError:            false,
			shouldRegisterPrecompile: true,
		},
		// TODO: is single hop but erc20 was already registered
		{
			name: "no-op - is single hop -> erc20 extension was already registered",
			transferBytes: newTransferBytes(
				transfertypes.NewFungibleTokenPacketData(
					"uatom",
					"100",
					keyring.GetAccAddr(0).String(),
					keyring.GetAccAddr(1).String(),
					"",
				),
			),
			expectedError: true,
			//
		},
		// TODO: is double hop
		{
			name: "no-op - is double hop coin",
			transferBytes: newTransferBytes(
				transfertypes.NewFungibleTokenPacketData(
					hopCoin,
					"100",
					keyring.GetAccAddr(0).String(),
					keyring.GetAccAddr(1).String(),
					"",
				),
			),
			expectedError: false,
			//
		},
		// TODO: native erc20
		{
			name: "success - native erc20",
			transferBytes: newTransferBytes(
				transfertypes.NewFungibleTokenPacketData(
					unitNetwork.GetDenom(),
					"100",
					keyring.GetAccAddr(0).String(),
					keyring.GetAccAddr(1).String(),
					"",
				),
			),
			expectedError: true,
			// Check after execution
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			packet := channeltypes.NewPacket(tc.transferBytes, 1, transfertypes.PortID, sourceChannel, transfertypes.PortID, destinationChannel, timeoutHeight, uint64(0))
			ack := unitNetwork.App.Erc20Keeper.OnRecvPacket(unitNetwork.GetContext(),
				packet,
				ibcmock.MockAcknowledgement,
			)

			if tc.expectedError {
				suite.Require().False(ack.Success(), string(ack.Acknowledgement()))
				// Unpacking the error will only return 'Internal error' on evmos related logic
				// We cannot test for each particular error
				// https://github.com/cosmos/ibc-go/blob/main/modules/core/04-channel/types/acknowledgement.go#L27-L41
			} else {
				suite.Require().True(ack.Success(), string(ack.Acknowledgement()))
			}
		})
	}
}
