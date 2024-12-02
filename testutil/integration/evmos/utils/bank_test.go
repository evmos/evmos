package utils_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/evmos/v20/app"
	testkeyring "github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/utils"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestCheckBalances(t *testing.T) {
	testDenom := "atest"
	keyring := testkeyring.New(1)
	address := keyring.GetAccAddr(0).String()

	testcases := []struct {
		name         string
		decimals     uint8
		configurator app.ConfigFn
		expAmount    math.Int
		expPass      bool
		errContains  string
	}{
		{
			name:         "pass - eighteen decimals",
			decimals:     18,
			expAmount:    network.GetInitialAmount(evmtypes.EighteenDecimals),
			configurator: network.Test18DecimalsAppConfigurator,
			expPass:      true,
		},
		{
			name:         "pass - six decimals",
			decimals:     6,
			expAmount:    network.GetInitialAmount(evmtypes.SixDecimals),
			configurator: network.Test6DecimalsAppConfigurator,
			expPass:      true,
		},
		{
			name:        "fail - wrong amount",
			decimals:    18,
			expAmount:   math.NewInt(1),
			errContains: "expected balance",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			balances := []banktypes.Balance{{
				Address: address,
				Coins: sdk.NewCoins(
					sdk.NewCoin(testDenom, tc.expAmount),
				),
			}}

			nw := network.NewWithConfigurator(
				tc.configurator,
				network.WithBaseCoin(testDenom, tc.decimals),
				network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
			)

			err := utils.CheckBalances(nw.GetContext(), nw.GetBankClient(), balances)
			if tc.expPass {
				require.NoError(t, err, "unexpected error checking balances")
			} else {
				require.Error(t, err, "expected error checking balances")
				require.ErrorContains(t, err, tc.errContains, "expected different error checking balances")
			}
		})
	}
}
