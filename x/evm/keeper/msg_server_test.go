// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package keeper_test

import (
	"math/big"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	gethtypes "github.com/ethereum/go-ethereum/core/types"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	gethparams "github.com/ethereum/go-ethereum/params"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/utils"
	utiltx "github.com/evmos/evmos/v16/testutil/tx"
	"github.com/evmos/evmos/v16/x/evm/types"
)

func (suite *EvmKeeperTestSuite) TestEthereumTx() {
	keyring := testkeyring.New(2)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(unitNetwork)
	txFactory := factory.New(unitNetwork, grpcHandler)

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
				tx, err := txFactory.GenerateSignedEthTx(keyring.GetPrivKey(0), args)
				suite.Require().NoError(err)
				return tx.GetMsgs()[0].(*types.MsgEthereumTx)
			},
			types.ErrInvalidGasCap,
		},
		{
			"success - transfer funds tx",
			func() *types.MsgEthereumTx {
				recipient := keyring.GetAddr(1)
				args := types.EvmTxArgs{
					To:     &recipient,
					Amount: big.NewInt(1e18),
				}
				tx, err := txFactory.GenerateSignedEthTx(keyring.GetPrivKey(0), args)
				suite.Require().NoError(err)
				return tx.GetMsgs()[0].(*types.MsgEthereumTx)
			},
			nil,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msg := tc.getMsg()

			// Function to be tested
			res, err := unitNetwork.App.EvmKeeper.EthereumTx(unitNetwork.GetContext(), msg)

			events := unitNetwork.GetContext().EventManager().Events()
			if tc.expectedErr != nil {
				suite.Require().Error(err)
				// no events should have been emitted
				suite.Require().Empty(events)
			} else {
				suite.Require().NoError(err)
				suite.Require().False(res.Failed())

				// check expected events were emitted
				suite.Require().NotEmpty(events)
				suite.Require().True(utils.ContainsEventType(events, types.EventTypeEthereumTx))
				suite.Require().True(utils.ContainsEventType(events, types.EventTypeTxLog))
				suite.Require().True(utils.ContainsEventType(events, sdktypes.EventTypeMessage))
			}

			err = unitNetwork.NextBlock()
			suite.Require().NoError(err)
		})
	}
}

func (suite *EvmKeeperTestSuite) TestUpdateParams() {
	keyring := testkeyring.New(1)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

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
			// Function to be tested
			msg := tc.getMsg()
			_, err := unitNetwork.App.EvmKeeper.UpdateParams(unitNetwork.GetContext(), msg)
			if tc.expectedErr != nil {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.expectedErr.Error())
			} else {
				suite.Require().NoError(err)
			}
		})

		err := unitNetwork.NextBlock()
		suite.Require().NoError(err)
	}
}

func (suite *KeeperTestSuite) createContractMsgTx(nonce uint64, signer gethtypes.Signer, gasPrice *big.Int) (*types.MsgEthereumTx, error) {
	contractCreateTx := &gethtypes.AccessListTx{
		GasPrice: gasPrice,
		Gas:      gethparams.TxGasContractCreation,
		To:       nil,
		Data:     []byte("contract_data"),
		Nonce:    nonce,
	}
	ethTx := gethtypes.NewTx(contractCreateTx)
	ethMsg := &types.MsgEthereumTx{}
	err := ethMsg.FromEthereumTx(ethTx)
	suite.Require().NoError(err)
	ethMsg.From = suite.keyring.GetAddr(0).Hex()
	krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
	return ethMsg, ethMsg.Sign(signer, krSigner)
}
