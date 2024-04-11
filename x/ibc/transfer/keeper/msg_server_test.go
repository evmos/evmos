package keeper_test

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	"github.com/evmos/evmos/v16/contracts"
	erc20types "github.com/evmos/evmos/v16/x/erc20/types"
	"github.com/evmos/evmos/v16/x/ibc/transfer/keeper"
	"github.com/stretchr/testify/mock"
)

func (suite *KeeperTestSuite) TestTransfer() {
	mockChannelKeeper := &MockChannelKeeper{}
	mockICS4Wrapper := &MockICS4Wrapper{}
	mockChannelKeeper.On("GetNextSequenceSend", mock.Anything, mock.Anything, mock.Anything).Return(1, true)
	mockChannelKeeper.On("GetChannel", mock.Anything, mock.Anything, mock.Anything).Return(channeltypes.Channel{Counterparty: channeltypes.NewCounterparty("transfer", "channel-1")}, true)
	mockICS4Wrapper.On("SendPacket", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	testCases := []struct {
		name     string
		malleate func() *types.MsgTransfer
		expPass  bool
	}{
		{
			"pass - no token pair",
			func() *types.MsgTransfer {
				senderAcc := sdk.AccAddress(suite.address.Bytes())
				transferMsg := types.NewMsgTransfer("transfer", "channel-0", sdk.NewCoin("aevmos", math.NewInt(10)), senderAcc.String(), "", timeoutHeight, 0, "")

				coins := sdk.NewCoins(sdk.NewCoin("aevmos", math.NewInt(10)))
				err := suite.app.BankKeeper.MintCoins(suite.ctx, erc20types.ModuleName, coins)
				suite.Require().NoError(err)
				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, erc20types.ModuleName, senderAcc, coins)
				suite.Require().NoError(err)
				suite.Commit()
				return transferMsg
			},
			true,
		},
		{
			"error - invalid sender",
			func() *types.MsgTransfer {
				addr := ""
				contractAddr, err := suite.DeployContract("coin", "token", uint8(6))
				suite.Require().NoError(err)
				suite.Commit()

				// senderAcc := sdk.MustAccAddressFromBech32(addr)
				transferMsg := types.NewMsgTransfer("transfer", "channel-0", sdk.NewCoin("erc20/"+contractAddr.String(), math.NewInt(10)), addr, "", timeoutHeight, 0, "")
				return transferMsg
			},
			false,
		},
		{
			"no-op - disabled erc20 by params - sufficient sdk.Coins balance)",
			func() *types.MsgTransfer {
				contractAddr, err := suite.DeployContract("coin", "token", uint8(6))
				suite.Require().NoError(err)
				suite.Commit()

				pair, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
				suite.Require().NoError(err)
				suite.Commit()

				senderAcc := sdk.AccAddress(suite.address.Bytes())
				suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(10))
				suite.Commit()

				coin := sdk.NewCoin(pair.Denom, math.NewInt(10))
				coins := sdk.NewCoins(coin)

				err = suite.app.BankKeeper.MintCoins(suite.ctx, erc20types.ModuleName, coins)
				suite.Require().NoError(err)
				suite.Commit()

				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, erc20types.ModuleName, senderAcc, coins)
				suite.Require().NoError(err)
				suite.Commit()

				params := suite.app.Erc20Keeper.GetParams(suite.ctx)
				params.EnableErc20 = false
				err = suite.app.Erc20Keeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				suite.Commit()

				transferMsg := types.NewMsgTransfer("transfer", "channel-0", sdk.NewCoin(pair.Denom, math.NewInt(10)), senderAcc.String(), "", timeoutHeight, 0, "")

				return transferMsg
			},
			true,
		},
		{
			"error - disabled erc20 by params - insufficient sdk.Coins balance)",
			func() *types.MsgTransfer {
				contractAddr, err := suite.DeployContract("coin", "token", uint8(6))
				suite.Require().NoError(err)
				suite.Commit()

				pair, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
				suite.Require().NoError(err)
				suite.Commit()

				senderAcc := sdk.AccAddress(suite.address.Bytes())
				suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(10))
				suite.Commit()

				params := suite.app.Erc20Keeper.GetParams(suite.ctx)
				params.EnableErc20 = false
				err = suite.app.Erc20Keeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				suite.Commit()

				transferMsg := types.NewMsgTransfer("transfer", "channel-0", sdk.NewCoin(pair.Denom, math.NewInt(10)), senderAcc.String(), "", timeoutHeight, 0, "")

				return transferMsg
			},
			false,
		},
		{
			"no-op - pair not registered",
			func() *types.MsgTransfer {
				senderAcc := sdk.AccAddress(suite.address.Bytes())

				coin := sdk.NewCoin("test", math.NewInt(10))
				coins := sdk.NewCoins(coin)

				err := suite.app.BankKeeper.MintCoins(suite.ctx, erc20types.ModuleName, coins)
				suite.Require().NoError(err)

				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, erc20types.ModuleName, senderAcc, coins)
				suite.Require().NoError(err)
				suite.Commit()

				transferMsg := types.NewMsgTransfer("transfer", "channel-0", coin, senderAcc.String(), "", timeoutHeight, 0, "")

				return transferMsg
			},
			true,
		},
		{
			"no-op - pair is disabled",
			func() *types.MsgTransfer {
				contractAddr, err := suite.DeployContract("coin", "token", uint8(6))
				suite.Require().NoError(err)
				suite.Commit()

				pair, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
				suite.Require().NoError(err)
				pair.Enabled = false
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, *pair)

				coin := sdk.NewCoin(pair.Denom, math.NewInt(10))
				senderAcc := sdk.AccAddress(suite.address.Bytes())
				transferMsg := types.NewMsgTransfer("transfer", "channel-0", coin, senderAcc.String(), "", timeoutHeight, 0, "")

				// mint coins to perform the regular transfer without conversions
				err = suite.app.BankKeeper.MintCoins(suite.ctx, erc20types.ModuleName, sdk.NewCoins(coin))
				suite.Require().NoError(err)

				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, erc20types.ModuleName, senderAcc, sdk.NewCoins(coin))
				suite.Require().NoError(err)
				suite.Commit()

				return transferMsg
			},
			true,
		},
		{
			"pass - has enough balance in erc20 - need to convert",
			func() *types.MsgTransfer {
				contractAddr, err := suite.DeployContract("coin", "token", uint8(6))
				suite.Require().NoError(err)
				suite.Commit()

				pair, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
				suite.Require().NoError(err)
				suite.Commit()
				suite.Require().Equal("erc20/"+pair.Erc20Address, pair.Denom)

				senderAcc := sdk.AccAddress(suite.address.Bytes())
				transferMsg := types.NewMsgTransfer("transfer", "channel-0", sdk.NewCoin(pair.Denom, math.NewInt(10)), senderAcc.String(), "", timeoutHeight, 0, "")

				suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(10))
				suite.Commit()
				return transferMsg
			},
			true,
		},
		{
			"pass - has enough balance in coins",
			func() *types.MsgTransfer {
				contractAddr, err := suite.DeployContract("coin", "token", uint8(6))
				suite.Require().NoError(err)
				suite.Commit()

				pair, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
				suite.Require().NoError(err)
				suite.Commit()

				senderAcc := sdk.AccAddress(suite.address.Bytes())
				transferMsg := types.NewMsgTransfer("transfer", "channel-0", sdk.NewCoin(pair.Denom, math.NewInt(10)), senderAcc.String(), "", timeoutHeight, 0, "")

				coins := sdk.NewCoins(sdk.NewCoin(pair.Denom, math.NewInt(10)))
				err = suite.app.BankKeeper.MintCoins(suite.ctx, erc20types.ModuleName, coins)
				suite.Require().NoError(err)
				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, erc20types.ModuleName, senderAcc, coins)
				suite.Require().NoError(err)
				suite.Commit()

				return transferMsg
			},
			true,
		},
		{
			"error - fail conversion - no balance in erc20",
			func() *types.MsgTransfer {
				contractAddr, err := suite.DeployContract("coin", "token", uint8(6))
				suite.Require().NoError(err)
				suite.Commit()

				pair, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
				suite.Require().NoError(err)
				suite.Commit()

				senderAcc := sdk.AccAddress(suite.address.Bytes())
				transferMsg := types.NewMsgTransfer("transfer", "channel-0", sdk.NewCoin(pair.Denom, math.NewInt(10)), senderAcc.String(), "", timeoutHeight, 0, "")
				return transferMsg
			},
			false,
		},

		// STRV2
		{
			"no-op - pair not registered",
			func() *types.MsgTransfer {
				senderAcc := sdk.AccAddress(suite.address.Bytes())

				denom := "test"
				coinMetadata := banktypes.Metadata{
					Name:        "Generic IBC name",
					Symbol:      "IBC",
					Description: "Generic IBC token description",
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 0,
							Aliases:  []string{denom},
						},
						{
							Denom:    denom,
							Exponent: 18,
						},
					},
					Display: denom,
					Base:    denom,
				}

				coin := sdk.NewCoin(denom, math.NewInt(10))
				coins := sdk.NewCoins(coin)

				err := suite.app.BankKeeper.MintCoins(suite.ctx, erc20types.ModuleName, coins)
				suite.Require().NoError(err)

				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, erc20types.ModuleName, senderAcc, coins)
				suite.Require().NoError(err)
				suite.Commit()

				pair, err := suite.app.Erc20Keeper.RegisterCoin(suite.ctx, coinMetadata)
				suite.Require().Equal(pair.Denom, denom)
				suite.Require().NoError(err)

				transferMsg := types.NewMsgTransfer("transfer", "channel-0", coin, senderAcc.String(), "", timeoutHeight, 0, "")

				return transferMsg
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()

			_, err := suite.app.ScopedTransferKeeper.NewCapability(suite.ctx, host.ChannelCapabilityPath("transfer", "channel-0"))
			suite.Require().NoError(err)
			suite.app.TransferKeeper = keeper.NewKeeper(
				suite.app.AppCodec(), suite.app.GetKey(types.StoreKey), suite.app.GetSubspace(types.ModuleName),
				&MockICS4Wrapper{}, // ICS4 Wrapper: claims IBC middleware
				mockChannelKeeper, &suite.app.IBCKeeper.PortKeeper,
				suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.ScopedTransferKeeper,
				suite.app.Erc20Keeper, // Add ERC20 Keeper for ERC20 transfers
			)
			msg := tc.malleate()

			_, err = suite.app.TransferKeeper.Transfer(sdk.WrapSDKContext(suite.ctx), msg)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
	suite.mintFeeCollector = false
}

func (suite *KeeperTestSuite) TestStrV2() {
	mockChannelKeeper := &MockChannelKeeper{}
	mockICS4Wrapper := &MockICS4Wrapper{}
	mockChannelKeeper.On("GetNextSequenceSend", mock.Anything, mock.Anything, mock.Anything).Return(1, true)
	mockChannelKeeper.On("GetChannel", mock.Anything, mock.Anything, mock.Anything).Return(channeltypes.Channel{Counterparty: channeltypes.NewCounterparty("transfer", "channel-1")}, true)
	mockICS4Wrapper.On("SendPacket", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	testCases := []struct {
		name     string
		malleate func() *types.MsgTransfer
		expPass  bool
	}{
		// STRV2
		{
			"Registered coin with erc20 balance, but enough ibc balance - Do not add address to bookkeeping",
			func() *types.MsgTransfer {
				senderAcc := sdk.AccAddress(suite.address.Bytes())

				// mint native coins
				mintAmount := int64(1000)
				// amount to transfer
				transferAmount := int64(1000)
				// native coins in erc20 representation
				erc20Amount := int64(100)

				denom := "test"
				coinMetadata := banktypes.Metadata{
					Name:        "Generic IBC name",
					Symbol:      "IBC",
					Description: "Generic IBC token description",
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 0,
							Aliases:  []string{denom},
						},
						{
							Denom:    denom,
							Exponent: 18,
						},
					},
					Display: denom,
					Base:    denom,
				}

				coin := sdk.NewCoin(denom, math.NewInt(mintAmount))
				coins := sdk.NewCoins(coin)

				err := suite.app.BankKeeper.MintCoins(suite.ctx, erc20types.ModuleName, coins)
				suite.Require().NoError(err)

				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, erc20types.ModuleName, senderAcc, coins)
				suite.Require().NoError(err)
				suite.Commit()

				pair, err := suite.app.Erc20Keeper.RegisterCoin(suite.ctx, coinMetadata)
				suite.Require().Equal(pair.Denom, denom)
				suite.Require().NoError(err)

				erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
				contract := pair.GetERC20Contract()
				_, err = s.app.Erc20Keeper.CallEVM(s.ctx, erc20, erc20types.ModuleAddress, contract, true, "mint", common.BytesToAddress(senderAcc.Bytes()), big.NewInt(erc20Amount))
				s.Require().NoError(err)

				transferCoin := sdk.NewCoin(denom, math.NewInt(transferAmount))
				transferMsg := types.NewMsgTransfer("transfer", "channel-0", transferCoin, senderAcc.String(), "", timeoutHeight, 0, "")

				found := suite.app.Erc20Keeper.HasSTRv2Address(suite.ctx, s.address.Bytes())
				suite.Require().False(found)

				return transferMsg
			},
			false,
		},
		{
			"Registered coin with erc20 balance - Add address to bookkeeping",
			func() *types.MsgTransfer {
				senderAcc := sdk.AccAddress(suite.address.Bytes())

				// mint native coins
				mintAmount := int64(1000)
				// native coins in erc20 representation
				erc20Amount := int64(100)
				// native coins in native representation in account
				accountAmount := mintAmount - erc20Amount

				denom := "test"
				coinMetadata := banktypes.Metadata{
					Name:        "Generic IBC name",
					Symbol:      "IBC",
					Description: "Generic IBC token description",
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 0,
							Aliases:  []string{denom},
						},
						{
							Denom:    denom,
							Exponent: 18,
						},
					},
					Display: denom,
					Base:    denom,
				}

				mintCoin := sdk.NewCoin(denom, math.NewInt(mintAmount))
				mintCoins := sdk.NewCoins(mintCoin)

				err := suite.app.BankKeeper.MintCoins(suite.ctx, erc20types.ModuleName, mintCoins)
				suite.Require().NoError(err)

				accountCoin := sdk.NewCoin(denom, math.NewInt(accountAmount))
				accountCoins := sdk.NewCoins(accountCoin)

				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, erc20types.ModuleName, senderAcc, accountCoins)
				suite.Require().NoError(err)
				suite.Commit()

				pair, err := suite.app.Erc20Keeper.RegisterCoin(suite.ctx, coinMetadata)
				suite.Require().Equal(pair.Denom, denom)
				suite.Require().NoError(err)

				erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
				contract := pair.GetERC20Contract()
				_, err = s.app.Erc20Keeper.CallEVM(s.ctx, erc20, erc20types.ModuleAddress, contract, true, "mint", common.BytesToAddress(senderAcc.Bytes()), big.NewInt(erc20Amount))
				s.Require().NoError(err)

				// transfer the full minted coins (account amount + erc20 to convert)
				transferMsg := types.NewMsgTransfer("transfer", "channel-0", mintCoin, senderAcc.String(), "", timeoutHeight, 0, "")

				found := suite.app.Erc20Keeper.HasSTRv2Address(suite.ctx, s.address.Bytes())
				suite.Require().False(found)

				return transferMsg
			},
			true,
		},
		// STRV2
		{
			"Registered coin with erc20 balance - address already in bookkeeping",
			func() *types.MsgTransfer {
				senderAcc := sdk.AccAddress(suite.address.Bytes())

				denom := "test"
				coinMetadata := banktypes.Metadata{
					Name:        "Generic IBC name",
					Symbol:      "IBC",
					Description: "Generic IBC token description",
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 0,
							Aliases:  []string{denom},
						},
						{
							Denom:    denom,
							Exponent: 18,
						},
					},
					Display: denom,
					Base:    denom,
				}

				coin := sdk.NewCoin(denom, math.NewInt(1000))
				coins := sdk.NewCoins(coin)

				err := suite.app.BankKeeper.MintCoins(suite.ctx, erc20types.ModuleName, coins)
				suite.Require().NoError(err)

				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, erc20types.ModuleName, senderAcc, coins)
				suite.Require().NoError(err)
				suite.Commit()

				pair, err := suite.app.Erc20Keeper.RegisterCoin(suite.ctx, coinMetadata)
				suite.Require().Equal(pair.Denom, denom)
				suite.Require().NoError(err)

				erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
				contract := pair.GetERC20Contract()
				_, err = s.app.Erc20Keeper.CallEVM(s.ctx, erc20, erc20types.ModuleAddress, contract, true, "mint", common.BytesToAddress(senderAcc.Bytes()), big.NewInt(100))
				s.Require().NoError(err)

				s.app.Erc20Keeper.SetSTRv2Address(s.ctx, senderAcc)

				transferMsg := types.NewMsgTransfer("transfer", "channel-0", coin, senderAcc.String(), "", timeoutHeight, 0, "")

				return transferMsg
			},
			true,
		},
		// STRV2
		{
			"Not Registered coin - adress shouldnt be added to bookkeeping",
			func() *types.MsgTransfer {
				senderAcc := sdk.AccAddress(suite.address.Bytes())

				denom := "test"

				coin := sdk.NewCoin(denom, math.NewInt(1000))
				coins := sdk.NewCoins(coin)

				err := suite.app.BankKeeper.MintCoins(suite.ctx, erc20types.ModuleName, coins)
				suite.Require().NoError(err)

				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, erc20types.ModuleName, senderAcc, coins)
				suite.Require().NoError(err)
				suite.Commit()

				transferMsg := types.NewMsgTransfer("transfer", "channel-0", coin, senderAcc.String(), "", timeoutHeight, 0, "")

				found := suite.app.Erc20Keeper.HasSTRv2Address(suite.ctx, s.address.Bytes())
				suite.Require().False(found)
				return transferMsg
			},
			false,
		},
		// STRV2
		{
			"Registered coin without erc20 balance change - address shouldnt be added to bookkeeping",
			func() *types.MsgTransfer {
				senderAcc := sdk.AccAddress(suite.address.Bytes())

				denom := "test"
				coinMetadata := banktypes.Metadata{
					Name:        "Generic IBC name",
					Symbol:      "IBC",
					Description: "Generic IBC token description",
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 0,
							Aliases:  []string{denom},
						},
						{
							Denom:    denom,
							Exponent: 18,
						},
					},
					Display: denom,
					Base:    denom,
				}

				coin := sdk.NewCoin(denom, math.NewInt(1000))
				coins := sdk.NewCoins(coin)

				err := suite.app.BankKeeper.MintCoins(suite.ctx, erc20types.ModuleName, coins)
				suite.Require().NoError(err)

				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, erc20types.ModuleName, senderAcc, coins)
				suite.Require().NoError(err)
				suite.Commit()

				pair, err := suite.app.Erc20Keeper.RegisterCoin(suite.ctx, coinMetadata)
				suite.Require().Equal(pair.Denom, denom)
				suite.Require().NoError(err)

				transferMsg := types.NewMsgTransfer("transfer", "channel-0", coin, senderAcc.String(), "", timeoutHeight, 0, "")

				found := suite.app.Erc20Keeper.HasSTRv2Address(suite.ctx, s.address.Bytes())
				suite.Require().False(found)
				return transferMsg
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()

			_, err := suite.app.ScopedTransferKeeper.NewCapability(suite.ctx, host.ChannelCapabilityPath("transfer", "channel-0"))
			suite.Require().NoError(err)
			suite.app.TransferKeeper = keeper.NewKeeper(
				suite.app.AppCodec(), suite.app.GetKey(types.StoreKey), suite.app.GetSubspace(types.ModuleName),
				&MockICS4Wrapper{}, // ICS4 Wrapper: claims IBC middleware
				mockChannelKeeper, &suite.app.IBCKeeper.PortKeeper,
				suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.ScopedTransferKeeper,
				suite.app.Erc20Keeper, // Add ERC20 Keeper for ERC20 transfers
			)
			msg := tc.malleate()

			_, err = suite.app.TransferKeeper.Transfer(sdk.WrapSDKContext(suite.ctx), msg)
			suite.Require().NoError(err)

			found := suite.app.Erc20Keeper.HasSTRv2Address(suite.ctx, s.address.Bytes())
			suite.Require().Equal(tc.expPass, found)
		})
	}
	suite.mintFeeCollector = false
}
