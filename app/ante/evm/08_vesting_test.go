// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm_test

import (
	"math/big"
	"time"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/evmos/evmos/v19/app/ante/evm"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v19/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	vestingtypes "github.com/evmos/evmos/v19/x/vesting/types"
)

type AccountExpenses = map[string]*evm.EthVestingExpenseTracker

func (suite *EvmAnteTestSuite) TestCheckVesting() {
	keyring := testkeyring.New(1)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(unitNetwork)
	sender := keyring.GetAccAddr(0)
	addedExpense := math.NewInt(100)

	testCases := []struct {
		name                  string
		expectedError         error
		getAccountAndExpenses func() (authtypes.AccountI, AccountExpenses)
	}{
		{
			name:          "success: non clawback account should be successful",
			expectedError: nil,
			getAccountAndExpenses: func() (authtypes.AccountI, AccountExpenses) {
				account, err := grpcHandler.GetAccount(sender.String())
				suite.Require().NoError(err)
				return account, defaultAccountExpenses()
			},
		},
		{
			name:          "error: clawback account with balance 0 should fail",
			expectedError: errortypes.ErrInsufficientFunds,
			getAccountAndExpenses: func() (authtypes.AccountI, AccountExpenses) {
				newIndex := keyring.AddKey()
				unfundedAddr := keyring.GetAccAddr(newIndex)
				funder := keyring.GetAccAddr(0)
				vestingParams := defaultVestingParams(unitNetwork, funder, unfundedAddr)
				return generateNewVestingAccount(
					unitNetwork,
					vestingParams,
				), defaultAccountExpenses()
			},
		},
		{
			name:          "error: clawback account with not enough bank + not enough vested unlocked balance < total should fail",
			expectedError: vestingtypes.ErrInsufficientUnlockedCoins,
			getAccountAndExpenses: func() (authtypes.AccountI, AccountExpenses) {
				newIndex := keyring.AddKey()
				newAddr := keyring.GetAccAddr(newIndex)
				funder := keyring.GetAccAddr(0)

				// have insufficient bank balance but not zero
				insufficientAmount := addedExpense.Sub(math.NewInt(1))
				err := unitNetwork.FundAccount(
					newAddr,
					sdktypes.NewCoins(
						sdktypes.NewCoin(
							unitNetwork.GetDenom(),
							insufficientAmount,
						),
					),
				)
				suite.Require().NoError(err)

				vestingParams := defaultVestingParams(unitNetwork, funder, newAddr)
				vestingAccount := generateNewVestingAccount(
					unitNetwork,
					vestingParams,
				)
				return vestingAccount, defaultAccountExpenses()
			},
		},
		{
			name:          "error: clawback account with not enough bank + not enough vested unlocked balance < total + previousExpenses should fail",
			expectedError: vestingtypes.ErrInsufficientUnlockedCoins,
			getAccountAndExpenses: func() (authtypes.AccountI, AccountExpenses) {
				newIndex := keyring.AddKey()
				newAddr := keyring.GetAccAddr(newIndex)
				funder := keyring.GetAccAddr(0)

				// have insufficient bank balance but not zero
				enoughAmount := addedExpense
				err := unitNetwork.FundAccount(
					newAddr,
					sdktypes.NewCoins(
						sdktypes.NewCoin(
							unitNetwork.GetDenom(),
							enoughAmount,
						),
					),
				)
				suite.Require().NoError(err)

				vestingParams := defaultVestingParams(unitNetwork, funder, newAddr)
				vestingAccount := generateNewVestingAccount(
					unitNetwork,
					vestingParams,
				)

				accExpenses := defaultAccountExpenses()
				accExpenses[newAddr.String()] = &evm.EthVestingExpenseTracker{
					Total:     big.NewInt(1000),
					Spendable: big.NewInt(0),
				}

				return vestingAccount, accExpenses
			},
		},
		{
			name:          "success: clawback account with enough bank + not enough vested unlocked balance > total should be successful",
			expectedError: nil,
			getAccountAndExpenses: func() (authtypes.AccountI, AccountExpenses) {
				newIndex := keyring.AddKey()
				newAddr := keyring.GetAccAddr(newIndex)
				funder := keyring.GetAccAddr(0)

				// have more than enough bank balance
				enoughAmount := addedExpense.Add(math.NewInt(1e18))
				err := unitNetwork.FundAccount(
					newAddr,
					sdktypes.NewCoins(
						sdktypes.NewCoin(
							unitNetwork.GetDenom(),
							enoughAmount,
						),
					),
				)
				suite.Require().NoError(err)

				vestingParams := defaultVestingParams(unitNetwork, funder, newAddr)
				vestingAccount := generateNewVestingAccount(
					unitNetwork,
					vestingParams,
				)
				return vestingAccount, defaultAccountExpenses()
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			account, accountExpenses := tc.getAccountAndExpenses()

			// Function under test
			err := evm.CheckVesting(
				unitNetwork.GetContext(),
				unitNetwork.App.BankKeeper,
				account,
				accountExpenses,
				addedExpense.BigInt(),
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

type customVestingParams struct {
	FunderAddress    sdktypes.AccAddress
	BaseAccAddress   sdktypes.AccAddress
	StartVestingTime time.Time
	Period           sdkvesting.Period
	VestingAmount    math.Int
}

func defaultAccountExpenses() AccountExpenses {
	return make(map[string]*evm.EthVestingExpenseTracker)
}

func defaultVestingParams(network network.Network, funder, baseAddress sdktypes.AccAddress) customVestingParams {
	return customVestingParams{
		FunderAddress:    funder,
		BaseAccAddress:   baseAddress,
		StartVestingTime: time.Now(),
		Period: sdkvesting.Period{
			Length: 1000,
			Amount: sdktypes.NewCoins(sdktypes.NewInt64Coin(network.GetDenom(), 1000)),
		},
		VestingAmount: sdktypes.NewInt(1e18),
	}
}

func generateNewVestingAccount(
	unitNetwork *network.UnitTestNetwork,
	vestingParams customVestingParams,
) authtypes.AccountI {
	var (
		balances       = sdktypes.NewCoins(sdktypes.NewInt64Coin(unitNetwork.GetDenom(), 1000))
		lockupPeriods  = sdkvesting.Periods{{Length: 5000, Amount: balances}}
		vestingPeriods = sdkvesting.Periods{
			vestingParams.Period,
			vestingParams.Period,
			vestingParams.Period,
			vestingParams.Period,
			vestingParams.Period,
		}
		vestingCoins = sdktypes.NewCoins(
			sdktypes.NewCoin(unitNetwork.GetDenom(), vestingParams.VestingAmount),
		)
	)

	baseAcc := authtypes.NewBaseAccountWithAddress(vestingParams.BaseAccAddress)
	vestingAcc := vestingtypes.NewClawbackVestingAccount(
		baseAcc,
		vestingParams.FunderAddress,
		vestingCoins,
		vestingParams.StartVestingTime,
		lockupPeriods,
		vestingPeriods,
	)
	acc := unitNetwork.App.AccountKeeper.NewAccount(unitNetwork.GetContext(), vestingAcc)
	unitNetwork.App.AccountKeeper.SetAccount(unitNetwork.GetContext(), acc)
	return acc
}
