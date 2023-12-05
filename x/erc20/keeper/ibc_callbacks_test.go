package keeper_test

import (
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	inflationtypes "github.com/evmos/evmos/v16/x/inflation/v1/types"
)

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
				moduleAcc := suite.app.AccountKeeper.GetModuleAccount(suite.ctx, evmtypes.ModuleName)

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
				coinEvmos := sdk.NewCoin(pair.Denom, math.NewInt(10))
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
