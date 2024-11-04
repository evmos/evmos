package keeper_test

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	testutils "github.com/evmos/evmos/v20/testutil/integration/evmos/utils"
	evmostypes "github.com/evmos/evmos/v20/types"
	"github.com/evmos/evmos/v20/x/ibc/transfer/keeper"
	"github.com/stretchr/testify/mock"
)

func (suite *KeeperTestSuite) TestTransfer() {
	var (
		ctx    sdk.Context
		sender keyring.Key
	)
	mockChannelKeeper := &MockChannelKeeper{}
	mockICS4Wrapper := &MockICS4Wrapper{}
	mockChannelKeeper.On("GetNextSequenceSend", mock.Anything, mock.Anything, mock.Anything).Return(1, true)
	mockChannelKeeper.On("GetChannel", mock.Anything, mock.Anything, mock.Anything).Return(channeltypes.Channel{Counterparty: channeltypes.NewCounterparty("transfer", "channel-1")}, true)
	mockICS4Wrapper.On("SendPacket", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	authAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	testCases := []struct {
		name     string
		malleate func() *types.MsgTransfer
		expPass  bool
	}{
		{
			"pass - no token pair",
			func() *types.MsgTransfer {
				transferMsg := types.NewMsgTransfer("transfer", "channel-0", sdk.NewCoin(evmostypes.BaseDenom, math.NewInt(10)), sender.AccAddr.String(), "", timeoutHeight, 0, "")
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

				transferMsg := types.NewMsgTransfer("transfer", "channel-0", sdk.NewCoin("erc20/"+contractAddr.String(), math.NewInt(10)), addr, "", timeoutHeight, 0, "")
				return transferMsg
			},
			false,
		},
		{
			"no-op - disabled erc20 by params - sufficient sdk.Coins balance",
			func() *types.MsgTransfer {
				contractAddr, err := suite.DeployContract("coin", "token", uint8(6))
				suite.Require().NoError(err)

				pair, err := testutils.RegisterERC20(suite.factory, suite.network, testutils.ERC20RegistrationData{
					Addresses:    []string{contractAddr.Hex()},
					ProposerPriv: sender.Priv,
				})
				suite.Require().NoError(err)
				suite.Require().True(len(pair) == 1)

				amt := math.NewInt(10)
				_, err = suite.MintERC20Token(contractAddr, sender.Addr, amt.BigInt())
				suite.Require().NoError(err)

				// convert all ERC20 to IBC coin
				err = suite.ConvertERC20(sender, contractAddr, amt)
				suite.Require().NoError(err)

				params := suite.network.App.Erc20Keeper.GetParams(ctx)
				params.EnableErc20 = false

				err = testutils.UpdateERC20Params(testutils.UpdateParamsInput{
					Tf:      suite.factory,
					Network: suite.network,
					Pk:      sender.Priv,
					Params:  params,
				})
				suite.Require().NoError(err)

				coin := sdk.NewCoin(pair[0].Denom, amt)
				transferMsg := types.NewMsgTransfer("transfer", "channel-0", coin, sender.AccAddr.String(), "", timeoutHeight, 0, "")

				return transferMsg
			},
			true,
		},
		{
			"error - disabled erc20 by params - insufficient sdk.Coins balance",
			func() *types.MsgTransfer {
				contractAddr, err := suite.DeployContract("coin", "token", uint8(6))
				suite.Require().NoError(err)

				pair, err := testutils.RegisterERC20(suite.factory, suite.network, testutils.ERC20RegistrationData{
					Addresses:    []string{contractAddr.Hex()},
					ProposerPriv: sender.Priv,
				})
				suite.Require().NoError(err)
				suite.Require().True(len(pair) == 1)

				amt := math.NewInt(10)
				_, err = suite.MintERC20Token(contractAddr, sender.Addr, amt.BigInt())
				suite.Require().NoError(err)

				params := suite.network.App.Erc20Keeper.GetParams(ctx)
				params.EnableErc20 = false
				err = testutils.UpdateERC20Params(testutils.UpdateParamsInput{
					Tf:      suite.factory,
					Network: suite.network,
					Pk:      sender.Priv,
					Params:  params,
				})
				suite.Require().NoError(err)

				transferMsg := types.NewMsgTransfer("transfer", "channel-0", sdk.NewCoin(pair[0].Denom, amt), sender.AccAddr.String(), "", timeoutHeight, 0, "")

				return transferMsg
			},
			false,
		},
		{
			"no-op - pair not registered",
			func() *types.MsgTransfer {
				coin := sdk.NewCoin(suite.otherDenom, math.NewInt(10))
				transferMsg := types.NewMsgTransfer("transfer", "channel-0", coin, sender.AccAddr.String(), "", timeoutHeight, 0, "")
				return transferMsg
			},
			true,
		},
		{
			"no-op - pair is disabled",
			func() *types.MsgTransfer {
				contractAddr, err := suite.DeployContract("coin", "token", uint8(6))
				suite.Require().NoError(err)

				pair, err := testutils.RegisterERC20(suite.factory, suite.network, testutils.ERC20RegistrationData{
					Addresses:    []string{contractAddr.Hex()},
					ProposerPriv: sender.Priv,
				})
				suite.Require().NoError(err)
				suite.Require().True(len(pair) == 1)

				amt := math.NewInt(10)
				_, err = suite.MintERC20Token(contractAddr, sender.Addr, amt.BigInt())
				suite.Require().NoError(err)

				// convert all erc20 to coins to perform regular transfer without conversion
				err = suite.ConvertERC20(sender, contractAddr, amt)
				suite.Require().NoError(err)

				// disable token conversion
				err = testutils.ToggleTokenConversion(suite.factory, suite.network, sender.Priv, pair[0].Denom)
				suite.Require().NoError(err)

				coin := sdk.NewCoin(pair[0].Denom, math.NewInt(10))
				transferMsg := types.NewMsgTransfer("transfer", "channel-0", coin, sender.AccAddr.String(), "", timeoutHeight, 0, "")

				return transferMsg
			},
			true,
		},
		{
			"pass - has enough balance in erc20 - need to convert",
			func() *types.MsgTransfer {
				contractAddr, err := suite.DeployContract("coin", "token", uint8(6))
				suite.Require().NoError(err)

				res, err := testutils.RegisterERC20(suite.factory, suite.network, testutils.ERC20RegistrationData{
					Addresses:    []string{contractAddr.Hex()},
					ProposerPriv: sender.Priv,
				})
				suite.Require().NoError(err)
				suite.Require().True(len(res) == 1)
				pair := res[0]
				suite.Require().Equal("erc20/"+pair.Erc20Address, pair.Denom)

				amt := math.NewInt(10)
				_, err = suite.MintERC20Token(contractAddr, sender.Addr, amt.BigInt())
				suite.Require().NoError(err)

				transferMsg := types.NewMsgTransfer("transfer", "channel-0", sdk.NewCoin(pair.Denom, amt), sender.AccAddr.String(), "", timeoutHeight, 0, "")

				return transferMsg
			},
			true,
		},
		{
			"pass - has enough balance in coins",
			func() *types.MsgTransfer {
				contractAddr, err := suite.DeployContract("coin", "token", uint8(6))
				suite.Require().NoError(err)

				pair, err := testutils.RegisterERC20(suite.factory, suite.network, testutils.ERC20RegistrationData{
					Addresses:    []string{contractAddr.Hex()},
					ProposerPriv: sender.Priv,
				})
				suite.Require().NoError(err)
				suite.Require().True(len(pair) == 1)

				// mint some erc20 tokens
				amt := math.NewInt(10)
				_, err = suite.MintERC20Token(contractAddr, suite.keyring.GetAddr(0), amt.BigInt())
				suite.Require().NoError(err)

				// convert all to IBC coins
				sender := suite.keyring.GetKey(0)
				err = suite.ConvertERC20(sender, contractAddr, amt)
				suite.Require().NoError(err)

				transferMsg := types.NewMsgTransfer("transfer", "channel-0", sdk.NewCoin(pair[0].Denom, amt), sender.AccAddr.String(), "", timeoutHeight, 0, "")

				return transferMsg
			},
			true,
		},
		{
			"error - fail conversion - no balance in erc20",
			func() *types.MsgTransfer {
				contractAddr, err := suite.DeployContract("coin", "token", uint8(6))
				suite.Require().NoError(err)

				pair, err := testutils.RegisterERC20(suite.factory, suite.network, testutils.ERC20RegistrationData{
					Addresses:    []string{contractAddr.Hex()},
					ProposerPriv: sender.Priv,
				})
				suite.Require().NoError(err)
				suite.Require().True(len(pair) == 1)

				transferMsg := types.NewMsgTransfer("transfer", "channel-0", sdk.NewCoin(pair[0].Denom, math.NewInt(10)), sender.AccAddr.String(), "", timeoutHeight, 0, "")
				return transferMsg
			},
			false,
		},

		// STRV2
		// native coin - perform normal ibc transfer
		{
			"no-op - fail transfer",
			func() *types.MsgTransfer {
				senderAcc := suite.keyring.GetAccAddr(0)

				denom := "ibc/DF63978F803A2E27CA5CC9B7631654CCF0BBC788B3B7F0A10200508E37C70992"
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

				pair, err := suite.network.App.Erc20Keeper.RegisterERC20Extension(suite.network.GetContext(), coinMetadata.Base)
				suite.Require().Equal(pair.Denom, denom)
				suite.Require().NoError(err)

				transferMsg := types.NewMsgTransfer("transfer", "channel-0", coin, senderAcc.String(), "", timeoutHeight, 0, "")

				return transferMsg
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			sender = suite.keyring.GetKey(0)
			ctx = suite.network.GetContext()

			suite.network.App.TransferKeeper = keeper.NewKeeper(
				suite.network.App.AppCodec(), suite.network.App.GetKey(types.StoreKey), suite.network.App.GetSubspace(types.ModuleName),
				&MockICS4Wrapper{}, // ICS4 Wrapper
				mockChannelKeeper, suite.network.App.IBCKeeper.PortKeeper,
				suite.network.App.AccountKeeper, suite.network.App.BankKeeper, suite.network.App.ScopedTransferKeeper,
				suite.network.App.Erc20Keeper, // Add ERC20 Keeper for ERC20 transfers
				authAddr,
			)
			msg := tc.malleate()

			// get updated context with the latest changes
			ctx = suite.network.GetContext()

			_, err := suite.network.App.TransferKeeper.Transfer(ctx, msg)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
