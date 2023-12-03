package evm_test

import (
	"errors"
	"math/big"

	"cosmossdk.io/math"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v15/app/ante/evm"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v15/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

const SENDER_KEYRING_KEY = 1

type TestTools struct {
	keyring     testkeyring.Keyring
	grpcHandler grpc.Handler
	txFactory   factory.TxFactory
	unitNetwork network.Network
}

func (suite *EvmAnteTestSuite) TestVerifyAccountBalance() {
	// Setup
	keyring := testkeyring.New(2)
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
		suite.Run(tc.name, func() {
			// Variable data
			statedb := unitNetwork.GetStateDB()
			statedbAccount := statedb.Keeper().GetAccount(unitNetwork.GetContext(), keyring.GetAddr(SENDER_KEYRING_KEY))
			txArgs, err := testTools.getTransactionArgs(suite.ethTxType)
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
				acc, err := testTools.grpcHandler.GetAccount(keyring.GetAccAddr(SENDER_KEYRING_KEY).String())
				suite.Require().NoError(err)
				suite.Require().NotEmpty(acc)
			}

			// Clean block for next test
			err = unitNetwork.NextBlock()
			suite.Require().NoError(err)
		})
	}
}

func (tt *TestTools) getTransactionArgs(txType uint8) (evmtypes.EvmTxArgs, error) {
	sender := tt.keyring.GetAddr(SENDER_KEYRING_KEY)
	defaultArgs := evmtypes.EvmTxArgs{
		Amount: big.NewInt(100),
	}
	switch txType {
	case gethtypes.DynamicFeeTxType:
		return tt.txFactory.PopulateEvmTxArgs(sender, defaultArgs)
	case gethtypes.AccessListTxType:
		defaultArgs.Accesses = &gethtypes.AccessList{{
			Address:     sender,
			StorageKeys: []common.Hash{{0}},
		}}
		defaultArgs.GasPrice = big.NewInt(1e9)
		return tt.txFactory.PopulateEvmTxArgs(sender, defaultArgs)
	case gethtypes.LegacyTxType:
		defaultArgs.GasPrice = big.NewInt(1e9)
		return tt.txFactory.PopulateEvmTxArgs(sender, defaultArgs)
	default:
		return evmtypes.EvmTxArgs{}, errors.New("tx type not supported")
	}
}

func getTxDataFromArgs(args *evmtypes.EvmTxArgs) (evmtypes.TxData, error) {
	ethTx := evmtypes.NewTx(args).AsTransaction()
	ethTx.Type()
	return evmtypes.NewTxDataFromTx(ethTx)
}
