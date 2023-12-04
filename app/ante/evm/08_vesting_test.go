// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm_test

import (
	"math/big"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/evmos/evmos/v15/app/ante/evm"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"
	vestingtypes "github.com/evmos/evmos/v15/x/vesting/types"
)

func (suite *EvmAnteTestSuite) TestCheckVesting() {
	keyring := testkeyring.New(1)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(unitNetwork)
	sender := keyring.GetAccAddr(0)

	testCases := []struct {
		name          string
		expectedError error
		malleate      func() authtypes.AccountI
	}{
		{
			name:          "success: non clawback account should be successful",
			expectedError: nil,
			malleate: func() authtypes.AccountI {
				account, err := grpcHandler.GetAccount(sender.String())
				suite.Require().NoError(err)
				return account
			},
		},
		{
			name:          "error: clawback account with balance 0 should fail",
			expectedError: errortypes.ErrInsufficientFunds,
			malleate: func() authtypes.AccountI {
				newIndex := keyring.AddKey()
				newAddr := keyring.GetAccAddr(newIndex)
				funder := keyring.GetAccAddr(0)
				return generateVestingAccount(unitNetwork, newAddr, funder)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			accountExpenses := make(map[string]*evm.EthVestingExpenseTracker)

			account := tc.malleate()
			addedExpense := big.NewInt(100)

			// Function under test
			err := evm.CheckVesting(
				unitNetwork.GetContext(),
				unitNetwork.App.BankKeeper,
				account,
				accountExpenses,
				addedExpense,
				unitNetwork.GetDenom(),
			)

			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.expectedError.Error())
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func generateVestingAccount(
	unitNetwork *network.UnitTestNetwork,
	newAddr sdktypes.AccAddress,
	funder sdktypes.AccAddress,
) authtypes.AccountI {
	var (
		balances       = sdktypes.NewCoins(sdktypes.NewInt64Coin(unitNetwork.GetDenom(), 1000))
		quarter        = sdktypes.NewCoins(sdktypes.NewInt64Coin(unitNetwork.GetDenom(), 250))
		lockupPeriods  = sdkvesting.Periods{{Length: 5000, Amount: balances}}
		vestingPeriods = sdkvesting.Periods{
			{Length: 2000, Amount: quarter},
			{Length: 2000, Amount: quarter},
			{Length: 2000, Amount: quarter},
			{Length: 2000, Amount: quarter},
		}
		vestingCoins = sdktypes.NewCoins(
			sdktypes.NewCoin(unitNetwork.GetDenom(), sdktypes.NewInt(1000000000)),
		)
		vestingTime = time.Now()
	)

	baseAcc := authtypes.NewBaseAccountWithAddress(newAddr)
	vestingAcc := vestingtypes.NewClawbackVestingAccount(
		baseAcc,
		funder,
		vestingCoins,
		vestingTime,
		lockupPeriods,
		vestingPeriods,
	)
	acc := unitNetwork.App.AccountKeeper.NewAccount(unitNetwork.GetContext(), vestingAcc)
	unitNetwork.App.AccountKeeper.SetAccount(unitNetwork.GetContext(), acc)
	return acc
}
