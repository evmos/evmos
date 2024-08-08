package keeper_test

import (
	"errors"
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/erc20/keeper"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/ethereum/go-ethereum/common"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/evmos/evmos/v19/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v19/testutil"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v8/testing"
	ibcmock "github.com/cosmos/ibc-go/v8/testing/mock"

	"github.com/evmos/evmos/v19/contracts"
	"github.com/evmos/evmos/v19/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
)

var erc20Denom = "erc20/0xdac17f958d2ee523a2206206994597c13d831ec7"

func (suite *KeeperTestSuite) TestOnRecvPacket() {
	var ctx sdk.Context
	// secp256k1 account
	secpPk := secp256k1.GenPrivKey()
	secpAddr := sdk.AccAddress(secpPk.PubKey().Address())
	secpAddrCosmos := sdk.MustBech32ifyAddressBytes(sdk.Bech32MainPrefix, secpAddr)

	// ethsecp256k1 account
	ethPk, err := ethsecp256k1.GenerateKey()
	suite.Require().Nil(err)
	ethsecpAddr := sdk.AccAddress(ethPk.PubKey().Address())
	ethsecpAddrEvmos := sdk.AccAddress(ethPk.PubKey().Address()).String()
	ethsecpAddrCosmos := sdk.MustBech32ifyAddressBytes(sdk.Bech32MainPrefix, ethsecpAddr)

	// Setup Cosmos <=> Evmos IBC relayer
	sourceChannel := "channel-292"
	evmosChannel := "channel-3"
	path := fmt.Sprintf("%s/%s", transfertypes.PortID, evmosChannel)

	timeoutHeight := clienttypes.NewHeight(0, 100)
	disabledTimeoutTimestamp := uint64(0)
	mockPacket := channeltypes.NewPacket(ibcgotesting.MockPacketData, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, disabledTimeoutTimestamp)
	packet := mockPacket
	expAck := ibcmock.MockAcknowledgement

	registeredDenom := cosmosTokenBase
	coins := sdk.NewCoins(
		sdk.NewCoin(utils.BaseDenom, math.NewInt(1000)),
		sdk.NewCoin(registeredDenom, math.NewInt(1000)), // some ERC20 token
		sdk.NewCoin(ibcBase, math.NewInt(1000)),         // some IBC coin with a registered token pair
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
			name: "error - sender == receiver, not from Evm channel",
			malleate: func() {
				transfer := transfertypes.NewFungibleTokenPacketData(registeredDenom, "100", ethsecpAddrEvmos, ethsecpAddrCosmos, "")
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, sourceChannel, transfertypes.PortID, "channel-100", timeoutHeight, 0)
			},
			ackSuccess:    false,
			receiver:      secpAddr,
			expErc20s:     big.NewInt(0),
			expCoins:      coins,
			checkBalances: false,
		},
		{
			name: "no-op - receiver is module account",
			malleate: func() {
				secpAddr = suite.network.App.AccountKeeper.GetModuleAccount(ctx, "erc20").GetAddress()
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
				bondDenom, err := suite.network.App.StakingKeeper.BondDenom(ctx)
				suite.Require().NoError(err)
				prefixedDenom := sourcePrefix + bondDenom
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
				sdk.NewCoin(utils.BaseDenom, math.NewInt(1000)),
				sdk.NewCoin(registeredDenom, math.NewInt(0)),
				sdk.NewCoin(ibcBase, math.NewInt(1000)),
			),
			checkBalances:    false,
			disableTokenPair: true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest() // reset
			ctx = suite.network.GetContext()

			tc.malleate()

			// Register Token Pair for testing
			contractAddr, err := suite.setupRegisterERC20Pair(contractMinterBurner)
			suite.Require().NoError(err, "failed to register pair")
			// get updated context after registering ERC20 pair
			ctx = suite.network.GetContext()

			// Set Denom Trace
			denomTrace := transfertypes.DenomTrace{
				Path:      path,
				BaseDenom: registeredDenom,
			}
			suite.network.App.TransferKeeper.SetDenomTrace(ctx, denomTrace)

			// Set Cosmos Channel
			channel := channeltypes.Channel{
				State:          channeltypes.INIT,
				Ordering:       channeltypes.UNORDERED,
				Counterparty:   channeltypes.NewCounterparty(transfertypes.PortID, sourceChannel),
				ConnectionHops: []string{sourceChannel},
			}
			suite.network.App.IBCKeeper.ChannelKeeper.SetChannel(ctx, transfertypes.PortID, evmosChannel, channel)

			// Set Next Sequence Send
			suite.network.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(ctx, transfertypes.PortID, evmosChannel, 1)

			suite.network.App.Erc20Keeper = keeper.NewKeeper(
				suite.network.App.GetKey(types.StoreKey),
				suite.network.App.AppCodec(),
				authtypes.NewModuleAddress(govtypes.ModuleName),
				suite.network.App.AccountKeeper,
				suite.network.App.BankKeeper,
				suite.network.App.EvmKeeper,
				suite.network.App.StakingKeeper,
				suite.network.App.AuthzKeeper,
				&suite.network.App.TransferKeeper,
			)

			// Fund receiver account with EVMOS, ERC20 coins and IBC vouchers
			// We do this since we are interested in the conversion portion w/ OnRecvPacket
			err = testutil.FundAccount(ctx, suite.network.App.BankKeeper, tc.receiver, coins)
			suite.Require().NoError(err)

			id := suite.network.App.Erc20Keeper.GetTokenPairID(ctx, contractAddr.String())
			pair, _ := suite.network.App.Erc20Keeper.GetTokenPair(ctx, id)
			suite.Require().NotNil(pair)

			if tc.disableERC20 {
				params := suite.network.App.Erc20Keeper.GetParams(ctx)
				params.EnableErc20 = false
				suite.network.App.Erc20Keeper.SetParams(ctx, params) //nolint:errcheck
			}

			if tc.disableTokenPair {
				_, err := suite.network.App.Erc20Keeper.ToggleConversion(ctx, &types.MsgToggleConversion{
					Authority: authtypes.NewModuleAddress("gov").String(),
					Token:     pair.Denom,
				})
				suite.Require().NoError(err)
			}

			// Perform IBC callback
			ack := suite.network.App.Erc20Keeper.OnRecvPacket(ctx, packet, expAck)

			// Check acknowledgement
			if tc.ackSuccess {
				suite.Require().True(ack.Success(), string(ack.Acknowledgement()))
				suite.Require().Equal(expAck, ack)
			} else {
				suite.Require().False(ack.Success(), string(ack.Acknowledgement()))
			}

			if tc.checkBalances {
				// Check ERC20 balances
				balanceTokenAfter := suite.network.App.Erc20Keeper.BalanceOf(ctx, contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(tc.receiver.Bytes()))
				suite.Require().Equal(tc.expErc20s.Int64(), balanceTokenAfter.Int64())
				// Check Cosmos Coin Balances
				balances := suite.network.App.BankKeeper.GetAllBalances(ctx, tc.receiver)
				suite.Require().Equal(tc.expCoins, balances)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestConvertCoinToERC20FromPacket() {
	var ctx sdk.Context
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
				// Register Token Pair for testing
				contractAddr, err := suite.setupRegisterERC20Pair(contractMinterBurner)
				suite.Require().NoError(err, "failed to register pair")
				ctx = suite.network.GetContext()
				id := suite.network.App.Erc20Keeper.GetTokenPairID(ctx, contractAddr.String())
				pair, _ := suite.network.App.Erc20Keeper.GetTokenPair(ctx, id)
				suite.Require().NotNil(pair)

				params := suite.network.App.Erc20Keeper.GetParams(ctx)
				params.EnableErc20 = false
				_ = suite.network.App.Erc20Keeper.SetParams(ctx, params)
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
			name: "pass - erc20 is disabled",
			malleate: func() transfertypes.FungibleTokenPacketData {
				// Register Token Pair for testing
				contractAddr, err := suite.setupRegisterERC20Pair(contractMinterBurner)
				suite.Require().NoError(err, "failed to register pair")
				ctx = suite.network.GetContext()
				id := suite.network.App.Erc20Keeper.GetTokenPairID(ctx, contractAddr.String())
				pair, _ := suite.network.App.Erc20Keeper.GetTokenPair(ctx, id)
				suite.Require().NotNil(pair)

				err = testutil.FundAccount(
					ctx,
					suite.network.App.BankKeeper,
					sdk.MustAccAddressFromBech32(senderAddr),
					sdk.NewCoins(
						sdk.NewCoin(pair.Denom, math.NewInt(100)),
					),
				)
				suite.Require().NoError(err)

				_, err = suite.network.App.EvmKeeper.CallEVM(ctx, contracts.ERC20MinterBurnerDecimalsContract.ABI, suite.keyring.GetAddr(0), contractAddr, true, "mint", types.ModuleAddress, big.NewInt(10))
				suite.Require().NoError(err)

				return transfertypes.NewFungibleTokenPacketData(pair.Denom, "10", senderAddr, "", "")
			},
			expPass: true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			defer func() { suite.mintFeeCollector = false }()

			suite.SetupTest() // reset
			ctx = suite.network.GetContext()

			transfer := tc.malleate()

			err := suite.network.App.Erc20Keeper.ConvertCoinToERC20FromPacket(ctx, transfer)
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
		ctx  sdk.Context
		data transfertypes.FungibleTokenPacketData
		ack  channeltypes.Acknowledgement
		pair types.TokenPair
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
				contractAddr, err := suite.setupRegisterERC20Pair(contractMinterBurner)
				suite.Require().NoError(err, "failed to register pair")
				ctx = suite.network.GetContext()
				id := suite.network.App.Erc20Keeper.GetTokenPairID(ctx, contractAddr.String())
				pair, _ = suite.network.App.Erc20Keeper.GetTokenPair(ctx, id)
				suite.Require().NotNil(pair)

				// for testing purposes we can only fund is not allowed to receive funds
				moduleAcc := suite.network.App.AccountKeeper.GetModuleAccount(ctx, "erc20")
				sender = moduleAcc.GetAddress()
				err = testutil.FundModuleAccount(
					ctx,
					suite.network.App.BankKeeper,
					moduleAcc.GetName(),
					sdk.NewCoins(
						sdk.NewCoin(pair.Denom, math.NewInt(100)),
					),
				)
				suite.Require().NoError(err)

				ack = channeltypes.NewErrorAcknowledgement(errors.New(""))
				data = transfertypes.NewFungibleTokenPacketData(pair.Denom, "100", sender.String(), receiver.String(), "")
			},
			expPass:  true,
			expERC20: big.NewInt(0),
		},
		{
			name: "no-op - positive ack",
			malleate: func() {
				// Register Token Pair for testing
				contractAddr, err := suite.setupRegisterERC20Pair(contractMinterBurner)
				suite.Require().NoError(err, "failed to register pair")
				ctx = suite.network.GetContext()
				id := suite.network.App.Erc20Keeper.GetTokenPairID(ctx, contractAddr.String())
				pair, _ = suite.network.App.Erc20Keeper.GetTokenPair(ctx, id)
				suite.Require().NotNil(pair)

				sender = sdk.AccAddress(senderPk.PubKey().Address())

				// Fund receiver account with EVMOS, ERC20 coins and IBC vouchers
				// We do this since we are interested in the conversion portion w/ OnRecvPacket
				err = testutil.FundAccount(
					ctx,
					suite.network.App.BankKeeper,
					sender,
					sdk.NewCoins(
						sdk.NewCoin(pair.Denom, math.NewInt(100)),
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
			ctx = suite.network.GetContext()

			tc.malleate()

			err := suite.network.App.Erc20Keeper.OnAcknowledgementPacket(
				ctx, channeltypes.Packet{}, data, ack,
			)
			suite.Require().NoError(err)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

			// check balance is the same as expected
			balance := suite.network.App.Erc20Keeper.BalanceOf(
				ctx, contracts.ERC20MinterBurnerDecimalsContract.ABI,
				pair.GetERC20Contract(),
				common.BytesToAddress(sender.Bytes()),
			)
			suite.Require().Equal(tc.expERC20.Int64(), balance.Int64())
		})
	}
}

func (suite *KeeperTestSuite) TestOnTimeoutPacket() {
	var ctx sdk.Context
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
				moduleAcc := suite.network.App.AccountKeeper.GetModuleAccount(ctx, evmtypes.ModuleName)

				return transfertypes.NewFungibleTokenPacketData("", "10", moduleAcc.GetAddress().String(), "", "")
			},
			expPass: true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			ctx = suite.network.GetContext()

			data := tc.malleate()

			err := suite.network.App.Erc20Keeper.OnTimeoutPacket(ctx, channeltypes.Packet{}, data)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
