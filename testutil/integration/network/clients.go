// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package network

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/evmos/v14/app"
	"github.com/evmos/evmos/v14/encoding"
	evmtypes "github.com/evmos/evmos/v14/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v14/x/feemarket/types"
	infltypes "github.com/evmos/evmos/v14/x/inflation/types"
	revtypes "github.com/evmos/evmos/v14/x/revenue/v1/types"
)

func getQueryHelper(ctx sdktypes.Context) *baseapp.QueryServiceTestHelper {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	interfaceRegistry := encCfg.InterfaceRegistry
	return baseapp.NewQueryServerTestHelper(ctx, interfaceRegistry)
}

func (n *IntegrationNetwork) GetEvmClient() evmtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext())
	evmtypes.RegisterQueryServer(queryHelper, n.app.EvmKeeper)
	return evmtypes.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetRevenueClient() revtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext())
	revtypes.RegisterQueryServer(queryHelper, n.app.RevenueKeeper)
	return revtypes.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetBankClient() banktypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext())
	banktypes.RegisterQueryServer(queryHelper, n.app.BankKeeper)
	return banktypes.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetFeeMarketClient() feemarkettypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext())
	feemarkettypes.RegisterQueryServer(queryHelper, n.app.FeeMarketKeeper)
	return feemarkettypes.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetInflationClient() infltypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext())
	infltypes.RegisterQueryServer(queryHelper, n.app.InflationKeeper)
	return infltypes.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetAuthClient() authtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext())
	authtypes.RegisterQueryServer(queryHelper, n.app.AccountKeeper)
	return authtypes.NewQueryClient(queryHelper)
}
