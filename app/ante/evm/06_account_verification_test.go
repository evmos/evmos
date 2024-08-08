package evm_test

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/app/ante/evm"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v19/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v19/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
)

func (suite *EvmAnteTestSuite) TestVerifyAccountBalance() {
	// Setup
	keyring := testkeyring.New(2)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(unitNetwork)
	txFactory := factory.New(unitNetwork, grpcHandler)
	senderKey := keyring.GetKey(1)

	testCases := []struct {
		name                   string
		expectedError          error
		generateAccountAndArgs func() (*statedb.Account, evmtypes.EvmTxArgs)
	}{
		{
			name:          "fail: sender is not EOA",
			expectedError: errortypes.ErrInvalidType,
			generateAccountAndArgs: func() (*statedb.Account, evmtypes.EvmTxArgs) {
				statedbAccount := getDefaultStateDBAccount(unitNetwork, senderKey.Addr)
				txArgs, err := txFactory.GenerateDefaultTxTypeArgs(senderKey.Addr, suite.ethTxType)
				suite.Require().NoError(err)

				statedbAccount.CodeHash = []byte("test")
				suite.Require().NoError(err)
				return statedbAccount, txArgs
			},
		},
		{
			name:          "fail: sender balance is lower than the transaction cost",
			expectedError: errortypes.ErrInsufficientFunds,
			generateAccountAndArgs: func() (*statedb.Account, evmtypes.EvmTxArgs) {
				statedbAccount := getDefaultStateDBAccount(unitNetwork, senderKey.Addr)
				txArgs, err := txFactory.GenerateDefaultTxTypeArgs(senderKey.Addr, suite.ethTxType)
				suite.Require().NoError(err)

				// Make tx cost greater than balance
				balanceResp, err := grpcHandler.GetBalance(senderKey.AccAddr, unitNetwork.GetDenom())
				suite.Require().NoError(err)

				invalidAmount := balanceResp.Balance.Amount.Add(math.NewInt(100))
				txArgs.Amount = invalidAmount.BigInt()
				return statedbAccount, txArgs
			},
		},
		{
			name:          "fail: tx cost is negative",
			expectedError: errortypes.ErrInvalidCoins,
			generateAccountAndArgs: func() (*statedb.Account, evmtypes.EvmTxArgs) {
				statedbAccount := getDefaultStateDBAccount(unitNetwork, senderKey.Addr)
				txArgs, err := txFactory.GenerateDefaultTxTypeArgs(senderKey.Addr, suite.ethTxType)
				suite.Require().NoError(err)

				// Make tx cost negative. This has to be a big value because
				// it has to be bigger than the fee for the full cost to be negative
				invalidAmount := big.NewInt(-1e18)
				txArgs.Amount = invalidAmount
				return statedbAccount, txArgs
			},
		},
		{
			name:          "success: tx is successful and account is created if its nil",
			expectedError: errortypes.ErrInsufficientFunds,
			generateAccountAndArgs: func() (*statedb.Account, evmtypes.EvmTxArgs) {
				txArgs, err := txFactory.GenerateDefaultTxTypeArgs(senderKey.Addr, suite.ethTxType)
				suite.Require().NoError(err)
				return nil, txArgs
			},
		},
		{
			name:          "success: tx is successful if account is EOA and exists",
			expectedError: nil,
			generateAccountAndArgs: func() (*statedb.Account, evmtypes.EvmTxArgs) {
				statedbAccount := getDefaultStateDBAccount(unitNetwork, senderKey.Addr)
				txArgs, err := txFactory.GenerateDefaultTxTypeArgs(senderKey.Addr, suite.ethTxType)
				suite.Require().NoError(err)
				return statedbAccount, txArgs
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("%v_%v", evmtypes.GetTxTypeName(suite.ethTxType), tc.name), func() {
			// Perform test logic
			statedbAccount, txArgs := tc.generateAccountAndArgs()
			txData, err := txArgs.ToTxData()
			suite.Require().NoError(err)

			//  Function to be tested
			err = evm.VerifyAccountBalance(
				unitNetwork.GetContext(),
				unitNetwork.App.AccountKeeper,
				statedbAccount,
				senderKey.Addr,
				txData,
			)

			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.expectedError.Error())
			} else {
				suite.Require().NoError(err)
			}
			// Make sure the account is created either wa
			acc, err := grpcHandler.GetAccount(senderKey.AccAddr.String())
			suite.Require().NoError(err)
			suite.Require().NotEmpty(acc)

			// Clean block for next test
			err = unitNetwork.NextBlock()
			suite.Require().NoError(err)
		})
	}
}

func getDefaultStateDBAccount(unitNetwork *network.UnitTestNetwork, addr common.Address) *statedb.Account {
	statedb := unitNetwork.GetStateDB()
	return statedb.Keeper().GetAccount(unitNetwork.GetContext(), addr)
}
