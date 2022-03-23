package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/mock"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"github.com/tharsis/evmos/v3/testutil"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v3/testing"
	ibcmock "github.com/cosmos/ibc-go/v3/testing/mock"

	claimstypes "github.com/tharsis/evmos/v3/x/claims/types"
	incentivestypes "github.com/tharsis/evmos/v3/x/incentives/types"
	"github.com/tharsis/evmos/v3/x/recovery/keeper"
	"github.com/tharsis/evmos/v3/x/recovery/types"
	vestingtypes "github.com/tharsis/evmos/v3/x/vesting/types"
)

func (suite *KeeperTestSuite) TestOnRecvPacket() {
	// secp256k1 account
	secpPk := secp256k1.GenPrivKey()
	secpAddr := sdk.AccAddress(secpPk.PubKey().Address())
	secpAddrEvmos := secpAddr.String()
	secpAddrCosmos := sdk.MustBech32ifyAddressBytes(sdk.Bech32MainPrefix, secpAddr)

	// ethsecp256k1 account
	ethPk, err := ethsecp256k1.GenerateKey()
	suite.Require().Nil(err)
	ethsecpAddr := sdk.AccAddress(ethPk.PubKey().Address())
	ethsecpAddrEvmos := sdk.AccAddress(ethPk.PubKey().Address()).String()
	ethsecpAddrCosmos := sdk.MustBech32ifyAddressBytes(sdk.Bech32MainPrefix, ethsecpAddr)

	// Setup Cosmos <=> Evmos IBC relayer
	denom := "uatom"
	sourceChannel := "channel-292"
	evmosChannel := claimstypes.DefaultAuthorizedChannels[1]
	path := fmt.Sprintf("%s/%s", transfertypes.PortID, evmosChannel)

	timeoutHeight := clienttypes.NewHeight(0, 100)
	disabledTimeoutTimestamp := uint64(0)
	mockPacket := channeltypes.NewPacket(ibcgotesting.MockPacketData, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, disabledTimeoutTimestamp)
	packet := mockPacket
	expAck := ibcmock.MockAcknowledgement

	testCases := []struct {
		name        string
		malleate    func()
		ackSuccess  bool
		expRecovery bool
	}{
		{
			"continue - params disabled",
			func() {
				params := suite.app.RecoveryKeeper.GetParams(suite.ctx)
				params.EnableRecovery = false
				suite.app.RecoveryKeeper.SetParams(suite.ctx, params)
			},
			true,
			false,
		},
		{
			"continue - destination channel not authorized",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", ethsecpAddrEvmos, ethsecpAddrCosmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, sourceChannel, transfertypes.PortID, "channel-100", timeoutHeight, 0)
			},
			true,
			false,
		},
		{
			"continue - destination channel is EVM",
			func() {
				EVMChannels := suite.app.ClaimsKeeper.GetParams(suite.ctx).EVMChannels
				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", ethsecpAddrEvmos, ethsecpAddrCosmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, sourceChannel, transfertypes.PortID, EVMChannels[0], timeoutHeight, 0)
			},
			true,
			false,
		},
		{
			"fail - non ics20 packet",
			func() {
				packet = mockPacket
			},
			false,
			false,
		},
		{
			"fail - invalid sender - missing '1' ",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", "evmos", ethsecpAddrCosmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			false,
			false,
		},
		{
			"fail - invalid sender - invalid bech32",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", "badba1sv9m0g7ycejwr3s369km58h5qe7xj77hvcxrms", ethsecpAddrCosmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			false,
			false,
		},
		{
			"fail - invalid recipient",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", ethsecpAddrEvmos, "badbadhf0468jjpe6m6vx38s97z2qqe8ldu0njdyf625")
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			false,
			false,
		},
		{
			"fail - case: receiver address is in deny list",
			func() {
				blockedAddr := authtypes.NewModuleAddress(transfertypes.ModuleName)

				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", secpAddrCosmos, blockedAddr.String())
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			false,
			false,
		},
		{
			"continue - case 1: sender and receiver address are not the same",
			func() {
				pk1 := secp256k1.GenPrivKey()
				otherSecpAddrEvmos := sdk.AccAddress(pk1.PubKey().Address()).String()

				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", secpAddrCosmos, otherSecpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			true,
			false,
		},
		{
			"continue - receiver is a vesting account",
			func() {
				// Set vesting account
				bacc := authtypes.NewBaseAccount(ethsecpAddr, nil, 0, 0)
				acc := vestingtypes.NewClawbackVestingAccount(bacc, ethsecpAddr, nil, suite.ctx.BlockTime(), nil, nil)

				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", ethsecpAddrCosmos, ethsecpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			true,
			false,
		},
		{
			"continue - receiver is a module account",
			func() {
				incentivesAcc := suite.app.AccountKeeper.GetModuleAccount(suite.ctx, incentivestypes.ModuleName)
				suite.Require().NotNil(incentivesAcc)
				addr := incentivesAcc.GetAddress().String()
				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", addr, addr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			true,
			false,
		},
		{
			"continue - case 2: receiver pubkey is a supported key",
			func() {
				// Set account to generate a pubkey
				suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(ethsecpAddr, ethPk.PubKey(), 0, 0))

				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", ethsecpAddrCosmos, ethsecpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			true,
			false,
		},
		{
			"recovery - send uatom from cosmos to evmos",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", secpAddrCosmos, secpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			true,
			true,
		},
		{
			"recovery - send ibc/uosmo from cosmos to evmos",
			func() {
				denom = ibcOsmoDenom

				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", secpAddrCosmos, secpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			true,
			true,
		},
		{
			"recovery - send uosmo from osmosis to evmos",
			func() {
				// Setup Osmosis <=> Evmos IBC relayer
				denom = "uosmo"
				sourceChannel = "channel-204"
				evmosChannel = claimstypes.DefaultAuthorizedChannels[0]
				path = fmt.Sprintf("%s/%s", transfertypes.PortID, evmosChannel)

				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", secpAddrCosmos, secpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			true,
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			// Enable Recovery
			params := suite.app.RecoveryKeeper.GetParams(suite.ctx)
			params.EnableRecovery = true
			suite.app.RecoveryKeeper.SetParams(suite.ctx, params)

			tc.malleate()

			// Set Denom Trace
			denomTrace := transfertypes.DenomTrace{
				Path:      path,
				BaseDenom: denom,
			}
			suite.app.TransferKeeper.SetDenomTrace(suite.ctx, denomTrace)

			// Set Cosmos Channel
			channel := channeltypes.Channel{
				State:          channeltypes.INIT,
				Ordering:       channeltypes.UNORDERED,
				Counterparty:   channeltypes.NewCounterparty(transfertypes.PortID, sourceChannel),
				ConnectionHops: []string{sourceChannel},
			}
			suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, transfertypes.PortID, evmosChannel, channel)

			// Set Next Sequence Send
			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, transfertypes.PortID, evmosChannel, 1)

			// Mock the Transferkeeper to always return nil on SendTransfer(), as this
			// method requires a successfull handshake with the counterparty chain.
			// This, however, exceeds the requirements of the unit tests.
			mockTransferKeeper := &MockTransferKeeper{
				Keeper: suite.app.BankKeeper,
			}

			mockTransferKeeper.On("GetDenomTrace", mock.Anything, mock.Anything).Return(denomTrace, true)
			mockTransferKeeper.On("SendTransfer", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

			sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
			suite.Require().True(found)
			suite.app.RecoveryKeeper = keeper.NewKeeper(sp, suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.IBCKeeper.ChannelKeeper, mockTransferKeeper, suite.app.ClaimsKeeper)

			// Fund receiver account with EVMOS, ERC20 coins and IBC vouchers
			coins := sdk.NewCoins(
				sdk.NewCoin("aevmos", sdk.NewInt(1000)),
				sdk.NewCoin(ibcAtomDenom, sdk.NewInt(1000)),
				sdk.NewCoin(ibcOsmoDenom, sdk.NewInt(1000)),
				sdk.NewCoin(erc20Denom, sdk.NewInt(1000)),
			)
			testutil.FundAccount(suite.app.BankKeeper, suite.ctx, secpAddr, coins)

			// Perform IBC callback
			ack := suite.app.RecoveryKeeper.OnRecvPacket(suite.ctx, packet, expAck)

			// Check acknowledgement
			if tc.ackSuccess {
				suite.Require().True(ack.Success(), string(ack.Acknowledgement()))
				suite.Require().Equal(expAck, ack)
			} else {
				suite.Require().False(ack.Success(), string(ack.Acknowledgement()))
			}

			// Check recovery
			balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, secpAddr)
			if tc.expRecovery {
				suite.Require().True(balances.IsZero())
			} else {
				suite.Require().Equal(coins, balances)
			}
		})
	}
}
