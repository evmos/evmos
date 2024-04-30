package keeper_test

import (
	"errors"
	"testing"

	errorsmod "cosmossdk.io/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/evmos/evmos/v18/utils"

	"github.com/ethereum/go-ethereum/common"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v7/testing"
	ibcmock "github.com/cosmos/ibc-go/v7/testing/mock"

	testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	utiltx "github.com/evmos/evmos/v18/testutil/tx"

	"github.com/evmos/evmos/v18/x/erc20/types"
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
	networkDenom := utils.BaseDenom

	// Common vars for every test
	sourceChannel := "channel-292"
	destinationChannel := "channel-0"
	amountValue := "100"
	emptyAddress := common.Address{}

	// For Single Hop Coin Test. Generate address for the precompile that will be deployed.
	fakeOsmoDenomTrace := transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "uosmo",
	}
	contractAddr, _ := utils.GetIBCDenomAddress(fakeOsmoDenomTrace.IBCDenom())

	// Set a dummy erc20 pair to test dummy conversion
	erc20Address := utiltx.GenerateAddress()
	erc20TestPair := types.NewTokenPair(erc20Address, types.CreateDenom(erc20Address.String()), types.OWNER_EXTERNAL)
	prefixedErc20Denom := transfertypes.GetDenomPrefix(transfertypes.PortID, sourceChannel) + erc20TestPair.GetDenom()

	testCases := []struct {
		name           string
		transferBytes  []byte
		expectedError  bool
		malleate       func(unitNetwork *network.UnitTestNetwork) error
		precompileAddr common.Address
	}{
		// Test Bad transfer package
		{
			name:          "error - non ics-20 packet",
			transferBytes: ibcgotesting.MockPacketData,
			expectedError: true,
		},
		// Package has an invalid sender
		{
			name: "error - invalid sender",
			transferBytes: newTransferBytes(
				transfertypes.NewFungibleTokenPacketData(
					networkDenom,
					amountValue,
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
					networkDenom,
					amountValue,
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
					networkDenom,
					amountValue,
					keyring.GetAccAddr(0).String(),
					keyring.GetAccAddr(0).String(),
					"",
				),
			),
			expectedError: true,
		},
		// Dont allow conversions from module accounts.
		{
			name: "no-op - receiver is module account",
			transferBytes: newTransferBytes(
				transfertypes.NewFungibleTokenPacketData(
					networkDenom,
					amountValue,
					keyring.GetAccAddr(0).String(),
					authtypes.NewModuleAddress(types.ModuleName).String(),
					"",
				),
			),
			expectedError: false,
		},
		// If transfer coin is evm denom there is no need for any precompile deploy or conversion
		// Since it arrived via IBC aevmos on the other chain will look like transfer/channel-293/aevmos
		// we will decode that and since their destination is == to our source
		// is a native coin
		{
			name: "no-op - evm Denom",
			transferBytes: newTransferBytes(
				transfertypes.NewFungibleTokenPacketData(
					transfertypes.GetDenomPrefix(transfertypes.PortID, sourceChannel)+"aevmos",
					amountValue,
					keyring.GetAccAddr(0).String(),
					keyring.GetAccAddr(1).String(),
					"",
				),
			),
			expectedError: false,
		},
		// If erc20 module is disabled it's not possible to deploy any precompile
		// or convert any coin
		{
			name: "no-op - erc20 module param disabled",
			transferBytes: newTransferBytes(
				transfertypes.NewFungibleTokenPacketData(
					networkDenom,
					amountValue,
					keyring.GetAccAddr(0).String(),
					keyring.GetAccAddr(1).String(),
					"",
				),
			),
			expectedError: false,
			malleate: func(unitNetwork *network.UnitTestNetwork) error {
				params := unitNetwork.App.Erc20Keeper.GetParams(unitNetwork.GetContext())
				params.EnableErc20 = false
				return unitNetwork.App.Erc20Keeper.SetParams(unitNetwork.GetContext(), params)
			},
		},
		// Is single hop and not registered EVM Extension.
		// Should register a new EVM extension
		{
			name: "success - is single hop. Should register erc20 extension",
			transferBytes: newTransferBytes(
				transfertypes.NewFungibleTokenPacketData(
					fakeOsmoDenomTrace.BaseDenom, // "uosmo"
					amountValue,
					keyring.GetAccAddr(0).String(),
					keyring.GetAccAddr(1).String(),
					"",
				),
			),
			expectedError:  false,
			precompileAddr: contractAddr,
		},
		// Is double hop coin - It should not register an Evm extension
		// The base denomination on the source chain, already has a hop
		// Since we are not the origin if this coin, its a double hop
		{
			name: "no-op - is double hop coin",
			transferBytes: newTransferBytes(
				transfertypes.NewFungibleTokenPacketData(
					"transfer/channel-0/ahopcoin",
					amountValue,
					keyring.GetAccAddr(0).String(),
					keyring.GetAccAddr(1).String(),
					"",
				),
			),
			expectedError: false,
		},
		// Registered dummy token pair owned by External.
		// Since balance of the coin is zero, the actual conversion is a no-op
		{
			name: "success (no-op) - native erc20",
			transferBytes: newTransferBytes(
				transfertypes.NewFungibleTokenPacketData(
					prefixedErc20Denom,
					amountValue,
					keyring.GetAccAddr(0).String(),
					keyring.GetAccAddr(1).String(),
					"",
				),
			),
			expectedError: false,
			// Check after execution
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			unitNetwork := network.NewUnitTestNetwork(
				network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
			)

			// Register the token pair on the test network
			HelperRegisterPair(unitNetwork, erc20TestPair)

			if tc.malleate != nil {
				err := tc.malleate(unitNetwork)
				suite.Require().NoError(err, "expected no error setting up test case")
			}

			activeDynamicPrecompilesPre := unitNetwork.App.EvmKeeper.GetParams(unitNetwork.GetContext()).ActiveDynamicPrecompiles

			packet := channeltypes.NewPacket(
				tc.transferBytes,
				1,
				transfertypes.PortID,
				sourceChannel,
				transfertypes.PortID,
				destinationChannel,
				timeoutHeight,
				uint64(0),
			)

			ack := unitNetwork.App.Erc20Keeper.OnRecvPacket(
				unitNetwork.GetContext(),
				packet,
				ibcmock.MockAcknowledgement,
			)

			if tc.expectedError {
				suite.Require().False(ack.Success(), string(ack.Acknowledgement()))
				// Unpacking the error will only return 'Internal error' on evmos related logic
				// We cannot test for each particular error.
				// The specific error cannot be returned on the IBC operation.
				// For more context:
				// https://github.com/cosmos/ibc-go/blob/main/modules/core/04-channel/types/acknowledgement.go#L27-L41
			} else {
				suite.Require().True(ack.Success(), string(ack.Acknowledgement()))
			}

			if tc.precompileAddr != emptyAddress {
				activeDynamicPrecompiles := unitNetwork.App.EvmKeeper.GetParams(unitNetwork.GetContext()).ActiveDynamicPrecompiles
				suite.Require().Contains(activeDynamicPrecompiles, tc.precompileAddr.String())
				if tc.precompileAddr == contractAddr {
					em := unitNetwork.GetContext().EventManager().Events()
					suite.Require().Equal(em[0].Type, types.EventTypeRegisterERC20Extension, "expected emitted event to show registered ERC-20 extension")
					suite.Require().Equal(em[0].Attributes[0].Value, sourceChannel)
					suite.Require().Equal(em[0].Attributes[1].Value, tc.precompileAddr.String())
					suite.Require().Equal(em[0].Attributes[2].Value, fakeOsmoDenomTrace.IBCDenom())
				}

			} else {
				activeDynamicPrecompiles := unitNetwork.App.EvmKeeper.GetParams(unitNetwork.GetContext()).ActiveDynamicPrecompiles
				suite.Require().Equal(
					activeDynamicPrecompilesPre,
					activeDynamicPrecompiles,
					"expected no change in active dynamic precompiles",
				)
			}
		})
	}
}

func HelperRegisterDummyCoin(unitNetwork *network.UnitTestNetwork, denom string) types.TokenPair {
	// Set a dummy coin pair to test dummy conversion
	dummyAddress := utiltx.GenerateAddress()
	testPair := types.NewTokenPair(dummyAddress, denom, types.OWNER_MODULE)
	HelperRegisterPair(unitNetwork, testPair)
	return testPair
}

func HelperRegisterPair(unitNetwork *network.UnitTestNetwork, erc20TestPair types.TokenPair) {
	unitNetwork.App.Erc20Keeper.SetTokenPair(unitNetwork.GetContext(), erc20TestPair)
	unitNetwork.App.Erc20Keeper.SetDenomMap(unitNetwork.GetContext(), erc20TestPair.Denom, erc20TestPair.GetID())
	unitNetwork.App.Erc20Keeper.SetERC20Map(unitNetwork.GetContext(), common.HexToAddress(erc20TestPair.Erc20Address), erc20TestPair.GetID())
}

func (suite *Erc20KeeperTestSuite) TestConvertCoinToERC20FromPacket() {
	keyring := testkeyring.New(2)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	sender := keyring.GetAccAddr(0).String()
	receiver := keyring.GetAccAddr(1).String()
	amountValue := "100"

	// Set a dummy erc20 pair to test dummy conversion
	erc20Address := utiltx.GenerateAddress()
	erc20TestPair := types.NewTokenPair(erc20Address, types.CreateDenom(erc20Address.String()), types.OWNER_EXTERNAL)
	HelperRegisterPair(unitNetwork, erc20TestPair)
	// Set a dummy aevmos pair and native coin to test dummy conversion
	dummyEvmosAddress := HelperRegisterDummyCoin(unitNetwork, "aevmos")
	dummyCoinAddress := HelperRegisterDummyCoin(unitNetwork, "acoin")

	testCases := []struct {
		name          string
		transfer      transfertypes.FungibleTokenPacketData
		expectedError error
	}{
		// If sender is invalid, nothing can be done. It should error
		{
			name:          "error - invalid sender",
			transfer:      transfertypes.FungibleTokenPacketData{Sender: ""},
			expectedError: errors.New("empty address string is not allowed"),
		},
		// If coin is not registered, no conversion is attempted
		// No error is returned
		{
			name: "noError - pair not registered",
			transfer: transfertypes.NewFungibleTokenPacketData(
				"nocoin",
				amountValue,
				sender,
				receiver,
				"",
			),
			expectedError: nil,
		},
		// If coin is aevmos, it should have been handled by the precompile
		// No conversion necessary at this point
		{
			name: "noError - is bondDenom",
			transfer: transfertypes.NewFungibleTokenPacketData(
				dummyEvmosAddress.Denom,
				amountValue,
				sender,
				receiver,
				"",
			),
			expectedError: nil,
		},
		// If coin is native, it should have been handled by the precompile
		// No conversion necessary at this point
		{
			name: "noError - is native Coin",
			transfer: transfertypes.NewFungibleTokenPacketData(
				dummyCoinAddress.Denom,
				amountValue,
				sender,
				receiver,
				"",
			),
			expectedError: nil,
		},

		{
			name: "no-op - is native Erc20 - sender is module address",
			transfer: transfertypes.NewFungibleTokenPacketData(
				erc20TestPair.Denom,
				amountValue,
				unitNetwork.App.AccountKeeper.GetModuleAddress("erc20").String(),
				receiver,
				"",
			),
			expectedError: nil,
		},
		// Erc20 is registered but there is no balance on the account
		// It should error when attempting to convert coin -> erc20
		{
			name: "error - is native Erc20 - no coin balance for conversion",
			transfer: transfertypes.NewFungibleTokenPacketData(
				erc20TestPair.Denom,
				amountValue,
				sender,
				receiver,
				"",
			),
			expectedError: errorsmod.Wrap(types.ErrEVMCall, "failed to retrieve balance"),
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := unitNetwork.App.Erc20Keeper.ConvertCoinToERC20FromPacket(
				unitNetwork.GetContext(),
				tc.transfer,
			)

			if tc.expectedError != nil {
				suite.Require().Equal(tc.expectedError.Error(), err.Error())
			} else {
				suite.Require().Equal(tc.expectedError, err)
			}
		})
	}
}

func (suite *Erc20KeeperTestSuite) TestOnTimeoutPacket() {
	keyring := testkeyring.New(2)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	sender := keyring.GetAccAddr(0).String()
	receiver := keyring.GetAccAddr(1).String()

	testCases := []struct {
		name          string
		transfer      transfertypes.FungibleTokenPacketData
		expectedError error
	}{
		{
			name:          "error - invalid sender",
			transfer:      transfertypes.FungibleTokenPacketData{Sender: ""},
			expectedError: errors.New("empty address string is not allowed"),
		},
		{
			name: "noError - pair not registered",
			transfer: transfertypes.NewFungibleTokenPacketData(
				"nocoin",
				"100",
				sender,
				receiver,
				"",
			),
			expectedError: nil,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := unitNetwork.App.Erc20Keeper.OnTimeoutPacket(
				unitNetwork.GetContext(),
				channeltypes.Packet{},
				tc.transfer,
			)

			if tc.expectedError != nil {
				suite.Require().Equal(tc.expectedError.Error(), err.Error())
			} else {
				suite.Require().Equal(tc.expectedError, err)
			}
		})
	}
}

func (suite *Erc20KeeperTestSuite) TestOnAcknowledgementPacket() {
	keyring := testkeyring.New(2)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	sender := keyring.GetAccAddr(0).String()
	receiver := keyring.GetAccAddr(1).String()

	testCases := []struct {
		name          string
		transfer      transfertypes.FungibleTokenPacketData
		expectedError error
		ack           channeltypes.Acknowledgement
	}{
		// If nothing failed on IBC tx, we will receive a success ack
		// No-op is executed
		{
			name:          "no - error  - ack is Success",
			transfer:      transfertypes.FungibleTokenPacketData{Sender: ""},
			expectedError: nil,
			ack:           ibcmock.MockAcknowledgement,
		},
		// In case of any error we should rollback via ConvertCoinToERC20FromPacket
		// Tests are copied from  TestConvertCoinToERC20FromPacket
		{
			name:          "error - invalid sender",
			transfer:      transfertypes.FungibleTokenPacketData{Sender: ""},
			expectedError: errors.New("empty address string is not allowed"),
			ack:           channeltypes.NewErrorAcknowledgement(errors.New("")),
		},
		{
			name: "noError - pair not registered",
			transfer: transfertypes.NewFungibleTokenPacketData(
				"nocoin",
				"100",
				sender,
				receiver,
				"",
			),
			ack:           channeltypes.NewErrorAcknowledgement(errors.New("")),
			expectedError: nil,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := unitNetwork.App.Erc20Keeper.OnAcknowledgementPacket(
				unitNetwork.GetContext(),
				channeltypes.Packet{},
				tc.transfer,
				tc.ack,
			)

			if tc.expectedError != nil {
				suite.Require().Equal(tc.expectedError.Error(), err.Error())
			} else {
				suite.Require().Equal(tc.expectedError, err)
			}
		})
	}
}
