// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package keeper_test

import (
	"math/big"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/utils"
	"github.com/evmos/evmos/v20/x/evm/types"
)

func (suite *KeeperTestSuite) TestEthereumTx() {
	suite.enableFeemarket = true
	defer func() { suite.enableFeemarket = false }()
	suite.SetupTest()
	testCases := []struct {
		name        string
		getMsg      func() *types.MsgEthereumTx
		expectedErr error
	}{
		{
			"fail - insufficient gas",
			func() *types.MsgEthereumTx {
				args := types.EvmTxArgs{
					// Have insufficient gas
					GasLimit: 10,
				}
				tx, err := suite.factory.GenerateSignedEthTx(suite.keyring.GetPrivKey(0), args)
				suite.Require().NoError(err)
				return tx.GetMsgs()[0].(*types.MsgEthereumTx)
			},
			types.ErrInvalidGasCap,
		},
		{
			"success - transfer funds tx",
			func() *types.MsgEthereumTx {
				recipient := suite.keyring.GetAddr(1)
				args := types.EvmTxArgs{
					To:     &recipient,
					Amount: big.NewInt(1e18),
				}
				tx, err := suite.factory.GenerateSignedEthTx(suite.keyring.GetPrivKey(0), args)
				suite.Require().NoError(err)
				return tx.GetMsgs()[0].(*types.MsgEthereumTx)
			},
			nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			msg := tc.getMsg()

			// Function to be tested
			res, err := suite.network.App.EvmKeeper.EthereumTx(suite.network.GetContext(), msg)

			events := suite.network.GetContext().EventManager().Events()
			if tc.expectedErr != nil {
				suite.Require().Error(err)
				// no events should have been emitted
				suite.Require().Empty(events)
			} else {
				suite.Require().NoError(err)
				suite.Require().False(res.Failed())

				// check expected events were emitted
				suite.Require().NotEmpty(events)
				suite.Require().True(utils.ContainsEventType(events.ToABCIEvents(), types.EventTypeEthereumTx))
				suite.Require().True(utils.ContainsEventType(events.ToABCIEvents(), types.EventTypeTxLog))
				suite.Require().True(utils.ContainsEventType(events.ToABCIEvents(), sdktypes.EventTypeMessage))
			}

			err = suite.network.NextBlock()
			suite.Require().NoError(err)
		})
	}
	suite.enableFeemarket = false
}

func (suite *KeeperTestSuite) TestUpdateParams() {
	suite.SetupTest()
	testCases := []struct {
		name        string
		getMsg      func() *types.MsgUpdateParams
		expectedErr error
	}{
		{
			name: "fail - invalid authority",
			getMsg: func() *types.MsgUpdateParams {
				return &types.MsgUpdateParams{Authority: "foobar"}
			},
			expectedErr: govtypes.ErrInvalidSigner,
		},
		{
			name: "pass - valid Update msg",
			getMsg: func() *types.MsgUpdateParams {
				return &types.MsgUpdateParams{
					Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
					Params:    types.DefaultParams(),
				}
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run("MsgUpdateParams", func() {
			msg := tc.getMsg()
			_, err := suite.network.App.EvmKeeper.UpdateParams(suite.network.GetContext(), msg)
			if tc.expectedErr != nil {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.expectedErr.Error())
			} else {
				suite.Require().NoError(err)
			}
		})

		err := suite.network.NextBlock()
		suite.Require().NoError(err)
	}
}
