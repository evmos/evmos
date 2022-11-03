package keeper_test

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v9/testutil"

	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v5/testing"
	ibcmock "github.com/cosmos/ibc-go/v5/testing/mock"
	erc20types "github.com/evmos/evmos/v9/x/erc20/types"

	"github.com/evmos/evmos/v9/contracts"
	claimstypes "github.com/evmos/evmos/v9/x/claims/types"
	"github.com/evmos/evmos/v9/x/erc20/keeper"
	"github.com/evmos/evmos/v9/x/erc20/types"
	vestingtypes "github.com/evmos/evmos/v9/x/vesting/types"
)

var erc20Denom = "erc20/0xdac17f958d2ee523a2206206994597c13d831ec7"

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

	coins := sdk.NewCoins(
		sdk.NewCoin("aevmos", sdk.NewInt(1000)),
		sdk.NewCoin(erc20Denom, sdk.NewInt(1000)),      // some ERC20 token
		sdk.NewCoin(cosmosTokenBase, sdk.NewInt(1000)), // some coin with a registered token pair
	)

	testCases := []struct {
		name          string
		malleate      func()
		ackSuccess    bool
		expConversion bool
		receiver      sdk.AccAddress
		expErc20s     *big.Int
		expCoins      sdk.Coins
		ibcConv       bool
		disableERC20  bool
	}{
		{
			name: "error - non ics-20 packet",
			malleate: func() {
				packet = mockPacket
			},
			ackSuccess:    false,
			expConversion: false,
			receiver:      secpAddr,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
			ibcConv:       false,
		},
		{
			name: "error - invalid sender (no '1')",		// TODO: sender address is not validated on the callback
			malleate: func() {
				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", "evmos", ethsecpAddrCosmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			ackSuccess:    false,
			expConversion: false,
			receiver:      secpAddr,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
			ibcConv:       false,
		},
		{
			name: "error - invalid sender (bad address)",
			malleate: func() {
				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", "badba1sv9m0g7ycejwr3s369km58h5qe7xj77hvcxrms", ethsecpAddrCosmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			ackSuccess:    false,
			expConversion: false,
			receiver:      secpAddr,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
			ibcConv:       false,
		},
		{
			name: "error - invalid recipient (bad address)",
			malleate: func() {
				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", ethsecpAddrEvmos, "badbadhf0468jjpe6m6vx38s97z2qqe8ldu0njdyf625")
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			ackSuccess:    false,
			expConversion: false,
			receiver:      secpAddr,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
			ibcConv:       false,
		},
		{
			name: "error - blocked sender",
			malleate: func() {
				blockedAddr := authtypes.NewModuleAddress(transfertypes.ModuleName)
				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", secpAddrCosmos, blockedAddr.String())
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			ackSuccess:    false,
			expConversion: false,
			receiver:      secpAddr,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
			ibcConv:       false,
		},
		{
			name: "error - blocked recipient",
			malleate: func() {
				blockedAddr := authtypes.NewModuleAddress(transfertypes.ModuleName)
				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", blockedAddr.String(), ethsecpAddrCosmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			ackSuccess:    false,
			expConversion: false,
			receiver:      secpAddr,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
			ibcConv:       false,
		},
		{
			name: "error - params disabled", // we disable params while running test
			malleate: func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", ethsecpAddrEvmos, ethsecpAddrCosmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			ackSuccess:    true,
			expConversion: false,
			receiver:      secpAddr,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
			ibcConv:       false,
			disableERC20:  true,
		},
		{
			name: "no-op - destination channel not authorized",
			malleate: func() {
				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", ethsecpAddrEvmos, ethsecpAddrCosmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, sourceChannel, transfertypes.PortID, "channel-100", timeoutHeight, 0)
			},
			ackSuccess:    true,
			expConversion: false,
			receiver:      secpAddr,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
			ibcConv:       false,
		},
		{
			name: "no-op - base denomination",
			malleate: func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", ethsecpAddrEvmos, ethsecpAddrCosmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			ackSuccess:    true,
			expConversion: false,
			receiver:      secpAddr,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
			ibcConv:       false,
		},
		{
			name: "no-op - erc20 denomination",
			malleate: func() {
				transfer := transfertypes.NewFungibleTokenPacketData(erc20Denom, "100", ethsecpAddrEvmos, ethsecpAddrCosmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			ackSuccess:    true,
			expConversion: false,
			receiver:      secpAddr,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
			ibcConv:       false,
		},
		{
			name: "error - invalid denomination",
			malleate: func() {
				transfer := transfertypes.NewFungibleTokenPacketData("b/d//s/ss/", "100", ethsecpAddrEvmos, ethsecpAddrCosmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			ackSuccess:    false,
			expConversion: false,
			receiver:      secpAddr,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
			ibcConv:       false,
		},
		{
			name: "ibc conversion - sender == receiver",
			malleate: func() {
				transfer := transfertypes.NewFungibleTokenPacketData(cosmosTokenBase, "100", secpAddrCosmos, secpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			ackSuccess:    true,
			expConversion: true,
			receiver:      secpAddr,
			expErc20s:     big.NewInt(100),
			expCoins: sdk.NewCoins(
				sdk.NewCoin("aevmos", sdk.NewInt(1000)),
				sdk.NewCoin(erc20Denom, sdk.NewInt(1000)),
				sdk.NewCoin(cosmosTokenBase, sdk.NewInt(900)),
			),
			ibcConv: false,
		},
		{
			name: "ibc conversion - sender != receiver",
			malleate: func() {
				pk1 := secp256k1.GenPrivKey()
				otherSecpAddrEvmos := sdk.AccAddress(pk1.PubKey().Address()).String()
				transfer := transfertypes.NewFungibleTokenPacketData(cosmosTokenBase, "500", otherSecpAddrEvmos, secpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			ackSuccess:    true,
			expConversion: true,
			receiver:      secpAddr,
			expErc20s:     big.NewInt(500),
			expCoins: sdk.NewCoins(
				sdk.NewCoin("aevmos", sdk.NewInt(1000)),
				sdk.NewCoin(erc20Denom, sdk.NewInt(1000)),
				sdk.NewCoin(cosmosTokenBase, sdk.NewInt(500)),
			),
			ibcConv: false,
		},
		{
			name: "conversion - receiver is a vesting account (eth address)",
			malleate: func() {
				// Set vesting account
				bacc := authtypes.NewBaseAccount(ethsecpAddr, nil, 0, 0)
				acc := vestingtypes.NewClawbackVestingAccount(bacc, ethsecpAddr, nil, suite.ctx.BlockTime(), nil, nil)

				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				// Fund receiver account with EVMOS, ERC20 coins and IBC vouchers
				// We do this since we are interested in the conversion portion w/ OnRecvPacket
				err := testutil.FundAccount(suite.ctx, suite.app.BankKeeper, ethsecpAddr, coins)
				suite.Require().NoError(err)

				transfer := transfertypes.NewFungibleTokenPacketData(cosmosTokenBase, "1000", ethsecpAddrCosmos, ethsecpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			ackSuccess:    true,
			expConversion: true,
			receiver:      ethsecpAddr,
			expErc20s:     big.NewInt(1000),
			expCoins: sdk.NewCoins(
				sdk.NewCoin("acoin", sdk.NewInt(1000)),
				sdk.NewCoin("aevmos", sdk.NewInt(1000)),
				sdk.NewCoin(erc20Denom, sdk.NewInt(1000)),
			),
			ibcConv: false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest() // reset

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

			sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
			suite.Require().True(found)
			suite.app.Erc20Keeper = keeper.NewKeeper(
				suite.app.GetKey(erc20types.StoreKey),
				suite.app.AppCodec(),
				sp,
				suite.app.AccountKeeper,
				suite.app.BankKeeper,
				suite.app.EvmKeeper,
				suite.app.StakingKeeper,
			)

			// Fund receiver account with EVMOS, ERC20 coins and IBC vouchers
			// We do this since we are interested in the conversion portion w/ OnRecvPacket
			err = testutil.FundAccount(suite.ctx, suite.app.BankKeeper, secpAddr, coins)
			suite.Require().NoError(err)

			// Enable ERC20
			params := suite.app.Erc20Keeper.GetParams(suite.ctx)
			params.EnableErc20 = true
			suite.app.Erc20Keeper.SetParams(suite.ctx, params)

			// Register Token Pair for testing
			pair := suite.setupRegisterCoin(metadataCoin)
			suite.Require().NotNil(pair)

			// For specific test, disable ERC20
			if tc.disableERC20 {
				params := suite.app.Erc20Keeper.GetParams(suite.ctx)
				params.EnableErc20 = false
				suite.app.Erc20Keeper.SetParams(suite.ctx, params)
			}

			// Perform IBC callback
			ack := suite.app.Erc20Keeper.OnRecvPacket(suite.ctx, packet, expAck)

			// Check acknowledgement
			if tc.ackSuccess {
				suite.Require().True(ack.Success(), string(ack.Acknowledgement()))
				suite.Require().Equal(expAck, ack)
			} else {
				suite.Require().False(ack.Success(), string(ack.Acknowledgement()))
			}

			// Check conversions
			if tc.ibcConv {
				// Check ERC20 balances
				balanceTokenAfter := suite.app.Erc20Keeper.BalanceOf(suite.ctx, contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(tc.receiver.Bytes()))
				suite.Require().Equal(tc.expErc20s, balanceTokenAfter)
				// Check Cosmos Coin Balances
				balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, secpAddr)
				suite.Require().Equal(tc.expCoins, balances)
			}
		})
	}
}
