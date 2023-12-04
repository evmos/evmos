package evm_test

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/evmos/evmos/v15/app/ante/evm"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v15/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

const SENDER_KEYRING_KEY = 1

func (suite *EvmAnteTestSuite) TestVerifyAccountBalance() {
	// Setup
	keyring := testkeyring.New(2)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(unitNetwork)
	txFactory := factory.New(unitNetwork, grpcHandler)

	testCases := []struct {
		name          string
		expectedError error
		malleate      func(statedbAccount *statedb.Account, args *evmtypes.EvmTxArgs)
	}{
		{
			name:          "fail: sender is not EOA",
			expectedError: errortypes.ErrInvalidType,
			malleate: func(statedbAccount *statedb.Account, _ *evmtypes.EvmTxArgs) {
				statedbAccount.CodeHash = []byte("test")
				err := unitNetwork.App.EvmKeeper.SetAccount(unitNetwork.GetContext(), keyring.GetAddr(SENDER_KEYRING_KEY), *statedbAccount)
				suite.Require().NoError(err)
			},
		}, {
			name:          "fail: sender balance is lower than the transaction cost",
			expectedError: errortypes.ErrInsufficientFunds,
			malleate: func(statedbAccount *statedb.Account, args *evmtypes.EvmTxArgs) {
				sender := keyring.GetKey(SENDER_KEYRING_KEY)
				// Make sure the account has no code hash
				statedbAccount.CodeHash = evmtypes.EmptyCodeHash
				err := unitNetwork.App.EvmKeeper.SetAccount(unitNetwork.GetContext(), sender.Addr, *statedbAccount)
				suite.Require().NoError(err)

				// Make tx cost greater than balance
				balanceResp, err := grpcHandler.GetBalance(sender.AccAddr, unitNetwork.GetDenom())
				suite.Require().NoError(err)

				invalidaAmount := balanceResp.Balance.Amount.Add(math.NewInt(100))
				args.Amount = invalidaAmount.BigInt()
			},
		}, {
			name:          "fail: tx cost is negative",
			expectedError: errortypes.ErrInvalidCoins,
			malleate: func(statedbAccount *statedb.Account, args *evmtypes.EvmTxArgs) {
				sender := keyring.GetKey(SENDER_KEYRING_KEY)
				// Make sure the account has no code hash
				statedbAccount.CodeHash = evmtypes.EmptyCodeHash
				err := unitNetwork.App.EvmKeeper.SetAccount(unitNetwork.GetContext(), sender.Addr, *statedbAccount)
				suite.Require().NoError(err)

				// Make tx cost is negative. This has to be a big value because the
				// it has to be bigger than the fee for the full cost to be negative
				invalidaAmount := big.NewInt(-1e18)
				args.Amount = invalidaAmount
			},
		}, {
			name:          "success: tx is succesfull and account is created if its nil",
			expectedError: nil,
			malleate: func(statedbAccount *statedb.Account, args *evmtypes.EvmTxArgs) {
				statedbAccount = nil
			},
		}, {
			name:          "success: tx is succesfull if account is EOA and exists",
			expectedError: nil,
			malleate: func(statedbAccount *statedb.Account, args *evmtypes.EvmTxArgs) {
				sender := keyring.GetKey(SENDER_KEYRING_KEY)
				// Make sure the account has no code hash
				statedbAccount.CodeHash = evmtypes.EmptyCodeHash
				err := unitNetwork.App.EvmKeeper.SetAccount(unitNetwork.GetContext(), sender.Addr, *statedbAccount)
				suite.Require().NoError(err)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("%v_%v", suite.getTxTypeTestName(), tc.name), func() {
			// Variable data
			sender := keyring.GetKey(SENDER_KEYRING_KEY)
			statedb := unitNetwork.GetStateDB()
			statedbAccount := statedb.Keeper().GetAccount(unitNetwork.GetContext(), sender.Addr)
			txArgs, err := txFactory.GenerateDefaultTxTypeArgs(sender.Addr, suite.ethTxType)
			suite.Require().NoError(err)

			// Perform test logic
			tc.malleate(statedbAccount, &txArgs)

			txData, err := getTxDataFromArgs(&txArgs)
			suite.Require().NoError(err)

			//  Function to be tested
			err = evm.VerifyAccountBalance(
				unitNetwork.GetContext(),
				unitNetwork.App.AccountKeeper,
				statedbAccount,
				keyring.GetAddr(SENDER_KEYRING_KEY),
				txData,
			)

			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.expectedError.Error())
			} else {
				suite.Require().NoError(err)

				// Make sure the account is created
				acc, err := grpcHandler.GetAccount(sender.AccAddr.String())
				suite.Require().NoError(err)
				suite.Require().NotEmpty(acc)
			}

			// Clean block for next test
			err = unitNetwork.NextBlock()
			suite.Require().NoError(err)
		})
	}
}

func getTxDataFromArgs(args *evmtypes.EvmTxArgs) (evmtypes.TxData, error) {
	ethTx := evmtypes.NewTx(args).AsTransaction()
	return evmtypes.NewTxDataFromTx(ethTx)
}
