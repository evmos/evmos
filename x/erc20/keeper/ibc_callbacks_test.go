package keeper_test

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/evmos/evmos/v11/x/erc20/keeper"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/ethereum/go-ethereum/common"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v11/testutil"

	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v6/testing"
	ibcmock "github.com/cosmos/ibc-go/v6/testing/mock"

	"github.com/evmos/evmos/v11/contracts"
	claimstypes "github.com/evmos/evmos/v11/x/claims/types"
	"github.com/evmos/evmos/v11/x/erc20/types"
	inflationtypes "github.com/evmos/evmos/v11/x/inflation/types"
	vestingtypes "github.com/evmos/evmos/v11/x/vesting/types"
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
	sourceChannel := "channel-292"
	evmosChannel := claimstypes.DefaultAuthorizedChannels[1]
	path := fmt.Sprintf("%s/%s", transfertypes.PortID, evmosChannel)

	timeoutHeight := clienttypes.NewHeight(0, 100)
	disabledTimeoutTimestamp := uint64(0)
	mockPacket := channeltypes.NewPacket(ibcgotesting.MockPacketData, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, disabledTimeoutTimestamp)
	packet := mockPacket
	expAck := ibcmock.MockAcknowledgement

	registeredDenom := cosmosTokenBase
	coins := sdk.NewCoins(
		sdk.NewCoin(claimstypes.DefaultClaimsDenom, sdk.NewInt(1000)),
		sdk.NewCoin(registeredDenom, sdk.NewInt(1000)), // some ERC20 token
		sdk.NewCoin(ibcBase, sdk.NewInt(1000)),         // some IBC coin with a registered token pair
	)

	testCases := []struct {
		name             string
		malleate         func()
		ackSuccess       bool
		receiver         sdk.AccAddress
		expErc20s        *big.Int
		expCoins         sdk.Coins
		checkBalances    bool
		disableERC20     bool
		disableTokenPair bool
	}{
		{
			name: "error - non ics-20 packet",
			malleate: func() {
				packet = mockPacket
			},
			receiver:      secpAddr,
			ackSuccess:    false,
			checkBalances: false,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
		},
		{
			name: "no-op - erc20 module param disabled",
			malleate: func() {
				transfer := transfertypes.NewFungibleTokenPacketData(registeredDenom, "100", ethsecpAddrEvmos, ethsecpAddrCosmos, "")
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			receiver:      secpAddr,
			disableERC20:  true,
			ackSuccess:    true,
			checkBalances: false,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
		},
		{
			name: "error - invalid sender (no '1')",
			malleate: func() {
				transfer := transfertypes.NewFungibleTokenPacketData(registeredDenom, "100", "evmos", ethsecpAddrCosmos, "")
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			receiver:      secpAddr,
			ackSuccess:    false,
			checkBalances: false,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
		},
		{
			name: "error - invalid sender (bad address)",
			malleate: func() {
				transfer := transfertypes.NewFungibleTokenPacketData(registeredDenom, "100", "badba1sv9m0g7ycejwr3s369km58h5qe7xj77hvcxrms", ethsecpAddrCosmos, "")
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			receiver:      secpAddr,
			ackSuccess:    false,
			checkBalances: false,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
		},
		{
			name: "error - invalid recipient (bad address)",
			malleate: func() {
				transfer := transfertypes.NewFungibleTokenPacketData(registeredDenom, "100", ethsecpAddrEvmos, "badbadhf0468jjpe6m6vx38s97z2qqe8ldu0njdyf625", "")
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			receiver:      secpAddr,
			ackSuccess:    false,
			checkBalances: false,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
		},
		{
			name: "no-op - sender == receiver, not from Evm channel",
			malleate: func() {
				transfer := transfertypes.NewFungibleTokenPacketData(registeredDenom, "100", ethsecpAddrEvmos, ethsecpAddrCosmos, "")
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, sourceChannel, transfertypes.PortID, "channel-100", timeoutHeight, 0)
			},
			ackSuccess:    true,
			receiver:      secpAddr,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
			checkBalances: true,
		},
		{
			name: "no-op - reciever is module account",
			malleate: func() {
				secpAddr = suite.app.AccountKeeper.GetModuleAccount(suite.ctx, "erc20").GetAddress()
				transfer := transfertypes.NewFungibleTokenPacketData(registeredDenom, "100", secpAddrCosmos, secpAddr.String(), "")
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			ackSuccess:    true,
			receiver:      secpAddr,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
			checkBalances: true,
		},
		{
			name: "no-op - base denomination",
			malleate: func() {
				// base denom should be prefixed
				sourcePrefix := transfertypes.GetDenomPrefix(transfertypes.PortID, sourceChannel)
				prefixedDenom := sourcePrefix + s.app.StakingKeeper.BondDenom(suite.ctx)
				transfer := transfertypes.NewFungibleTokenPacketData(prefixedDenom, "100", secpAddrCosmos, ethsecpAddrEvmos, "")
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			ackSuccess:    true,
			receiver:      ethsecpAddr,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
			checkBalances: true,
		},
		{
			name: "no-op - pair is not registered",
			malleate: func() {
				transfer := transfertypes.NewFungibleTokenPacketData(erc20Denom, "100", secpAddrCosmos, ethsecpAddrEvmos, "")
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			ackSuccess:    true,
			receiver:      ethsecpAddr,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
			checkBalances: true,
		},
		{
			name: "no-op - pair disabled",
			malleate: func() {
				pk1 := secp256k1.GenPrivKey()
				sourcePrefix := transfertypes.GetDenomPrefix(transfertypes.PortID, sourceChannel)
				prefixedDenom := sourcePrefix + registeredDenom
				otherSecpAddrEvmos := sdk.AccAddress(pk1.PubKey().Address()).String()
				transfer := transfertypes.NewFungibleTokenPacketData(prefixedDenom, "500", otherSecpAddrEvmos, ethsecpAddrEvmos, "")
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			ackSuccess: true,
			receiver:   ethsecpAddr,
			expErc20s:  big.NewInt(0),
			expCoins: sdk.NewCoins(
				sdk.NewCoin(claimstypes.DefaultClaimsDenom, sdk.NewInt(1000)),
				sdk.NewCoin(registeredDenom, sdk.NewInt(0)),
				sdk.NewCoin(ibcBase, sdk.NewInt(1000)),
			),
			checkBalances:    false,
			disableTokenPair: true,
		},
		{
			name: "no-op - sender == receiver and is not from evm chain", // getting failed to escrow coins - need to escrow coins
			malleate: func() {
				transfer := transfertypes.NewFungibleTokenPacketData(registeredDenom, "100", secpAddrCosmos, secpAddrEvmos, "")
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			receiver:      secpAddr,
			ackSuccess:    true,
			checkBalances: false,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
		},
		{
			name: "error - invalid denomination", // should fall as unregistered and not transfer any coins, but ack is Success
			malleate: func() {
				transfer := transfertypes.NewFungibleTokenPacketData("b/d//s/ss/", "100", ethsecpAddrEvmos, ethsecpAddrCosmos, "")
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			receiver:      secpAddr,
			ackSuccess:    true,
			checkBalances: true,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
		},
		{
			name: "ibc conversion - sender == receiver and from evm chain",
			malleate: func() {
				claimsParams := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				claimsParams.EVMChannels = []string{evmosChannel}
				suite.app.ClaimsKeeper.SetParams(suite.ctx, claimsParams)

				sourcePrefix := transfertypes.GetDenomPrefix(transfertypes.PortID, sourceChannel)
				prefixedDenom := sourcePrefix + registeredDenom
				transfer := transfertypes.NewFungibleTokenPacketData(prefixedDenom, "100", secpAddrCosmos, secpAddrEvmos, "")
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			receiver:      secpAddr,
			ackSuccess:    true,
			checkBalances: true,
			expErc20s:     big.NewInt(1000),
			expCoins: sdk.NewCoins(
				sdk.NewCoin(claimstypes.DefaultClaimsDenom, sdk.NewInt(1000)),
				sdk.NewCoin(registeredDenom, sdk.NewInt(0)),
				sdk.NewCoin(ibcBase, sdk.NewInt(1000)),
			),
		},
		{
			name: "ibc conversion - sender != receiver",
			malleate: func() {
				pk1 := secp256k1.GenPrivKey()
				sourcePrefix := transfertypes.GetDenomPrefix(transfertypes.PortID, sourceChannel)
				prefixedDenom := sourcePrefix + registeredDenom
				otherSecpAddrEvmos := sdk.AccAddress(pk1.PubKey().Address()).String()
				transfer := transfertypes.NewFungibleTokenPacketData(prefixedDenom, "500", otherSecpAddrEvmos, ethsecpAddrEvmos, "")
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			receiver:      ethsecpAddr,
			ackSuccess:    true,
			expErc20s:     big.NewInt(1000),
			checkBalances: true,
			expCoins: sdk.NewCoins(
				sdk.NewCoin(claimstypes.DefaultClaimsDenom, sdk.NewInt(1000)),
				sdk.NewCoin(registeredDenom, sdk.NewInt(0)),
				sdk.NewCoin(ibcBase, sdk.NewInt(1000)),
			),
		},
		{
			name: "ibc conversion - receiver is a vesting account (eth address)",
			malleate: func() {
				// Set vesting account
				bacc := authtypes.NewBaseAccount(ethsecpAddr, nil, 0, 0)
				acc := vestingtypes.NewClawbackVestingAccount(bacc, ethsecpAddr, nil, suite.ctx.BlockTime(), nil, nil)

				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
				sourcePrefix := transfertypes.GetDenomPrefix(transfertypes.PortID, sourceChannel)
				prefixedDenom := sourcePrefix + registeredDenom

				transfer := transfertypes.NewFungibleTokenPacketData(prefixedDenom, "1000", secpAddrCosmos, ethsecpAddrEvmos, "")
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, evmosChannel, timeoutHeight, 0)
			},
			receiver:      ethsecpAddr,
			ackSuccess:    true,
			checkBalances: true,
			expErc20s:     big.NewInt(1000),
			expCoins: sdk.NewCoins(
				sdk.NewCoin(ibcBase, sdk.NewInt(1000)),
				sdk.NewCoin(claimstypes.DefaultClaimsDenom, sdk.NewInt(1000)),
				sdk.NewCoin(registeredDenom, sdk.NewInt(0)),
			),
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
				BaseDenom: registeredDenom,
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

			suite.app.Erc20Keeper = keeper.NewKeeper(
				suite.app.GetKey(types.StoreKey),
				suite.app.AppCodec(),
				authtypes.NewModuleAddress(govtypes.ModuleName),
				suite.app.AccountKeeper,
				suite.app.BankKeeper,
				suite.app.EvmKeeper,
				suite.app.StakingKeeper,
				suite.app.ClaimsKeeper,
			)

			// Fund receiver account with EVMOS, ERC20 coins and IBC vouchers
			// We do this since we are interested in the conversion portion w/ OnRecvPacket
			err = testutil.FundAccount(suite.ctx, suite.app.BankKeeper, tc.receiver, coins)
			suite.Require().NoError(err)

			// Register Token Pair for testing
			pair := suite.setupRegisterCoin(metadataCoin)
			suite.Require().NotNil(pair)

			if tc.disableERC20 {
				params := suite.app.Erc20Keeper.GetParams(suite.ctx)
				params.EnableErc20 = false
				suite.app.Erc20Keeper.SetParams(suite.ctx, params)
			}

			if tc.disableTokenPair {
				_, err := suite.app.Erc20Keeper.ToggleConversion(suite.ctx, pair.Denom)
				suite.Require().NoError(err)
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

			if tc.checkBalances {
				// Check ERC20 balances
				balanceTokenAfter := suite.app.Erc20Keeper.BalanceOf(suite.ctx, contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(tc.receiver.Bytes()))
				suite.Require().Equal(tc.expErc20s.Int64(), balanceTokenAfter.Int64())
				// Check Cosmos Coin Balances
				balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, tc.receiver)
				suite.Require().Equal(tc.expCoins, balances)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestConvertCoinToERC20FromPacket() {
	senderAddr := "evmos1x2w87cvt5mqjncav4lxy8yfreynn273xn5335v"

	testCases := []struct {
		name     string
		malleate func() transfertypes.FungibleTokenPacketData
		transfer transfertypes.FungibleTokenPacketData
		expPass  bool
	}{
		{
			name: "error - invalid sender",
			malleate: func() transfertypes.FungibleTokenPacketData {
				return transfertypes.NewFungibleTokenPacketData("aevmos", "10", "", "", "")
			},
			expPass: false,
		},
		{
			name: "pass - is base denom",
			malleate: func() transfertypes.FungibleTokenPacketData {
				return transfertypes.NewFungibleTokenPacketData("aevmos", "10", senderAddr, "", "")
			},
			expPass: true,
		},
		{
			name: "pass - erc20 is disabled",
			malleate: func() transfertypes.FungibleTokenPacketData {
				pair := suite.setupRegisterCoin(metadataIbc)
				suite.Require().NotNil(pair)

				params := suite.app.Erc20Keeper.GetParams(suite.ctx)
				params.EnableErc20 = false
				suite.app.Erc20Keeper.SetParams(suite.ctx, params)
				return transfertypes.NewFungibleTokenPacketData(pair.Denom, "10", senderAddr, "", "")
			},
			expPass: true,
		},
		{
			name: "pass - denom is not registered",
			malleate: func() transfertypes.FungibleTokenPacketData {
				return transfertypes.NewFungibleTokenPacketData(metadataIbc.Base, "10", senderAddr, "", "")
			},
			expPass: true,
		},
		{
			name: "pass - denom is registered and has available balance",
			malleate: func() transfertypes.FungibleTokenPacketData {
				pair := suite.setupRegisterCoin(metadataIbc)
				suite.Require().NotNil(pair)

				sender := sdk.MustAccAddressFromBech32(senderAddr)

				// Mint coins on account to simulate receiving ibc transfer
				coinEvmos := sdk.NewCoin(pair.Denom, sdk.NewInt(10))
				coins := sdk.NewCoins(coinEvmos)
				err := suite.app.BankKeeper.MintCoins(suite.ctx, inflationtypes.ModuleName, coins)
				suite.Require().NoError(err)
				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, inflationtypes.ModuleName, sender, coins)
				suite.Require().NoError(err)

				return transfertypes.NewFungibleTokenPacketData(pair.Denom, "10", senderAddr, "", "")
			},
			expPass: true,
		},
		{
			name: "error - denom is registered but has no available balance",
			malleate: func() transfertypes.FungibleTokenPacketData {
				pair := suite.setupRegisterCoin(metadataIbc)
				suite.Require().NotNil(pair)

				return transfertypes.NewFungibleTokenPacketData(pair.Denom, "10", senderAddr, "", "")
			},
			expPass: false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest() // reset

			transfer := tc.malleate()

			err := suite.app.Erc20Keeper.ConvertCoinToERC20FromPacket(suite.ctx, transfer)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestOnAcknowledgementPacket() {
	var (
		data transfertypes.FungibleTokenPacketData
		ack  channeltypes.Acknowledgement
		pair *types.TokenPair
	)

	// secp256k1 account
	senderPk := secp256k1.GenPrivKey()
	sender := sdk.AccAddress(senderPk.PubKey().Address())

	receiverPk := secp256k1.GenPrivKey()
	receiver := sdk.AccAddress(receiverPk.PubKey().Address())
	fmt.Println(receiver)
	testCases := []struct {
		name     string
		malleate func()
		expERC20 *big.Int
		expPass  bool
	}{
		{
			name: "no-op - ack error sender is module account",
			malleate: func() {
				// Register Token Pair for testing
				pair = suite.setupRegisterCoin(metadataCoin)
				suite.Require().NotNil(pair)

				// for testing purposes we can only fund is not allowed to receive funds
				moduleAcc := suite.app.AccountKeeper.GetModuleAccount(suite.ctx, "erc20")
				sender = moduleAcc.GetAddress()
				err := testutil.FundModuleAccount(
					suite.ctx,
					suite.app.BankKeeper,
					moduleAcc.GetName(),
					sdk.NewCoins(
						sdk.NewCoin(pair.Denom, sdk.NewInt(100)),
					),
				)
				suite.Require().NoError(err)

				ack = channeltypes.NewErrorAcknowledgement(errors.New(""))
				data = transfertypes.NewFungibleTokenPacketData("", "", sender.String(), "", "")
			},
			expPass:  true,
			expERC20: big.NewInt(0),
		},
		{
			name: "conversion - convert ibc tokens to erc20 on ack error",
			malleate: func() {
				// Register Token Pair for testing
				pair = suite.setupRegisterCoin(metadataCoin)
				suite.Require().NotNil(pair)

				sender = sdk.AccAddress(senderPk.PubKey().Address())

				// Fund receiver account with EVMOS, ERC20 coins and IBC vouchers
				// We do this since we are interested in the conversion portion w/ OnRecvPacket
				err := testutil.FundAccount(
					suite.ctx,
					suite.app.BankKeeper,
					sender,
					sdk.NewCoins(
						sdk.NewCoin(pair.Denom, sdk.NewInt(100)),
					),
				)
				suite.Require().NoError(err)

				ack = channeltypes.NewErrorAcknowledgement(errors.New(""))
				data = transfertypes.NewFungibleTokenPacketData(pair.Denom, "100", sender.String(), receiver.String(), "")
			},
			expERC20: big.NewInt(100),
			expPass:  true,
		},
		{
			name: "no-op - positive ack",
			malleate: func() {
				// Register Token Pair for testing
				pair = suite.setupRegisterCoin(metadataCoin)
				suite.Require().NotNil(pair)

				sender = sdk.AccAddress(senderPk.PubKey().Address())

				// Fund receiver account with EVMOS, ERC20 coins and IBC vouchers
				// We do this since we are interested in the conversion portion w/ OnRecvPacket
				err := testutil.FundAccount(
					suite.ctx,
					suite.app.BankKeeper,
					sender,
					sdk.NewCoins(
						sdk.NewCoin(pair.Denom, sdk.NewInt(100)),
					),
				)
				suite.Require().NoError(err)

				ack = channeltypes.NewResultAcknowledgement([]byte{1})
			},
			expERC20: big.NewInt(0),
			expPass:  true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			err := suite.app.Erc20Keeper.OnAcknowledgementPacket(
				suite.ctx, channeltypes.Packet{}, data, ack,
			)
			suite.Require().NoError(err)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

			// check balance is the same as expected
			balance := suite.app.Erc20Keeper.BalanceOf(
				suite.ctx, contracts.ERC20MinterBurnerDecimalsContract.ABI,
				pair.GetERC20Contract(),
				common.BytesToAddress(sender.Bytes()),
			)
			suite.Require().Equal(tc.expERC20.Int64(), balance.Int64())
		})
	}
}

func (suite *KeeperTestSuite) TestOnTimeoutPacket() {
	senderAddr := "evmos1x2w87cvt5mqjncav4lxy8yfreynn273xn5335v"

	testCases := []struct {
		name     string
		malleate func() transfertypes.FungibleTokenPacketData
		transfer transfertypes.FungibleTokenPacketData
		expPass  bool
	}{
		{
			name: "no-op - sender is module account",
			malleate: func() transfertypes.FungibleTokenPacketData {
				// any module account can be passed here
				moduleAcc := suite.app.AccountKeeper.GetModuleAccount(suite.ctx, "claims")

				return transfertypes.NewFungibleTokenPacketData("", "10", moduleAcc.GetAddress().String(), "", "")
			},
			expPass: true,
		},
		{
			name: "pass - convert coin to erc20",
			malleate: func() transfertypes.FungibleTokenPacketData {
				pair := suite.setupRegisterCoin(metadataIbc)
				suite.Require().NotNil(pair)

				sender := sdk.MustAccAddressFromBech32(senderAddr)

				// Mint coins on account to simulate receiving ibc transfer
				coinEvmos := sdk.NewCoin(pair.Denom, sdk.NewInt(10))
				coins := sdk.NewCoins(coinEvmos)
				err := suite.app.BankKeeper.MintCoins(suite.ctx, inflationtypes.ModuleName, coins)
				suite.Require().NoError(err)
				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, inflationtypes.ModuleName, sender, coins)
				suite.Require().NoError(err)

				return transfertypes.NewFungibleTokenPacketData(pair.Denom, "10", senderAddr, "", "")
			},
			expPass: true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()

			data := tc.malleate()

			err := suite.app.Erc20Keeper.OnTimeoutPacket(suite.ctx, channeltypes.Packet{}, data)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
