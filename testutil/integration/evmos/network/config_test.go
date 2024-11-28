// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package network_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/evmos/evmos/v20/app"
	grpchandler "github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	evmostypes "github.com/evmos/evmos/v20/types"
	"github.com/evmos/evmos/v20/utils"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"

	"github.com/stretchr/testify/require"
)

func TestWithChainID(t *testing.T) {
	testCases := []struct {
		name             string
		chainID          string
		denom            string
		configurator     app.AppConfig
		expBalanceCosmos math.Int
	}{
		{
			name:         "18 decimals",
			chainID:      utils.MainnetChainID + "-1",
			denom:        "aevmos",
			configurator: network.Test18DecimalsAppConfigurator,
		},
		{
			name:         "6 decimals",
			chainID:      utils.SixDecChainID + "-1",
			denom:        "asevmos",
			configurator: network.Test6DecimalsAppConfigurator,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new network with 2 pre-funded accounts
			keyring := testkeyring.New(1)

			opts := []network.ConfigOption{
				network.WithChainID(tc.chainID),
				network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
			}

			nw := network.NewWithConfigurator(tc.configurator, opts...)

			handler := grpchandler.NewIntegrationHandler(nw)

			fmt.Println(network.PrefundedAccountInitialBalance.String())
			// Evm balance should always be in 18 decimals
			req, err := handler.GetBalanceFromEVM(keyring.GetAccAddr(0))
			require.NoError(t, err, "error getting balances")
			evmCoinDecimal := evmtypes.GetEVMCoinDecimals()
			require.Equal(t, network.PrefundedAccountInitialBalance.Mul(evmCoinDecimal.ConversionFactor()).String(), req.Balance, "expected amount to be in 18 decimals")

			// Bank balance should always be in the original amount
			cReq, err := handler.GetBalanceFromBank(keyring.GetAccAddr(0), tc.denom)
			require.NoError(t, err, "error getting balances")
			require.Equal(t, network.PrefundedAccountInitialBalance.String(), cReq.Balance.Amount.String(), "expected amount to be in original decimals")
		})
	}
}

func TestWithBalances(t *testing.T) {
	key1Balance := sdk.NewCoins(sdk.NewInt64Coin(evmostypes.BaseDenom, 1e18))
	key2Balance := sdk.NewCoins(
		sdk.NewInt64Coin(evmostypes.BaseDenom, 2e18),
		sdk.NewInt64Coin("other", 3e18),
	)

	// Create a new network with 2 pre-funded accounts
	keyring := testkeyring.New(2)
	balances := []banktypes.Balance{
		{
			Address: keyring.GetAccAddr(0).String(),
			Coins:   key1Balance,
		},
		{
			Address: keyring.GetAccAddr(1).String(),
			Coins:   key2Balance,
		},
	}
	nw := network.New(
		network.WithBalances(balances...),
	)
	handler := grpchandler.NewIntegrationHandler(nw)

	req, err := handler.GetAllBalances(keyring.GetAccAddr(0))
	require.NoError(t, err, "error getting balances")
	require.Len(t, req.Balances, 1, "wrong number of balances")
	require.Equal(t, balances[0].Coins, req.Balances, "wrong balances")

	req, err = handler.GetAllBalances(keyring.GetAccAddr(1))
	require.NoError(t, err, "error getting balances")
	require.Len(t, req.Balances, 2, "wrong number of balances")
	require.Equal(t, balances[1].Coins, req.Balances, "wrong balances")
}
