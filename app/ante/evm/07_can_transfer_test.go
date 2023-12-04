// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm_test

import (
	"math/big"

	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v15/app/ante/evm"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

func (suite *EvmAnteTestSuite) TestCanTransfer() {
	keyring := testkeyring.New(1)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(unitNetwork)
	txFactory := factory.New(unitNetwork, grpcHandler)

	testTools := TestTools{
		keyring:     keyring,
		grpcHandler: grpcHandler,
		txFactory:   txFactory,
		unitNetwork: unitNetwork,
	}

	testCases := []struct {
		name          string
		expectedError error
		malleate      func()
	}{
		{
			name:          "fail: isLondon an insufficient fee",
			expectedError: errortypes.ErrInvalidSequence,
			malleate: func() {
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			senderKey := keyring.GetKey(1)
			account, err := grpcHandler.GetAccount(senderKey.AccAddr.String())
			suite.Require().NoError(err)
			preSequence := account.GetSequence()
			baseFeeResp, err := grpcHandler.GetBaseFee()
			suite.Require().NoError(err)
			ethCfg := unitNetwork.GetEVMChainConfig()
			evmParams, err := grpcHandler.GetEvmParams()
			suite.Require().NoError(err)
			ctx := unitNetwork.GetContext()
			signer := gethtypes.MakeSigner(ethCfg, big.NewInt(ctx.BlockHeight()))
			txArgs, err := testTools.getTransactionArgs(suite.ethTxType)
			msg, err := evmtypes.NewTx(&txArgs).AsMessage(signer, baseFeeResp.BaseFee.BigInt())
			suite.Require().NoError(err)

			tc.malleate()

			// Function under test
			err = evm.CanTransfer(
				unitNetwork.GetContext(),
				unitNetwork.App.EvmKeeper,
				msg,
				baseFeeResp.BaseFee.BigInt(),
				ethCfg,
				evmParams.Params,
				true,
			)

			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.expectedError.Error())
			} else {
				suite.Require().Equal(preSequence+1, account.GetSequence())
			}
		})
	}
}
