// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package network

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	epochstypes "github.com/evmos/evmos/v20/x/epochs/types"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v20/x/feemarket/types"
	infltypes "github.com/evmos/evmos/v20/x/inflation/v1/types"
	vestingtypes "github.com/evmos/evmos/v20/x/vesting/types"
)

func getQueryHelper(ctx sdktypes.Context, encCfg testutil.TestEncodingConfig) *baseapp.QueryServiceTestHelper {
	interfaceRegistry := encCfg.InterfaceRegistry
	// This is needed so that state changes are not committed in precompiles
	// simulations.
	cacheCtx, _ := ctx.CacheContext()
	return baseapp.NewQueryServerTestHelper(cacheCtx, interfaceRegistry)
}

func (n *IntegrationNetwork) GetERC20Client() erc20types.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	erc20types.RegisterQueryServer(queryHelper, n.app.Erc20Keeper)
	return erc20types.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetEvmClient() evmtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	evmtypes.RegisterQueryServer(queryHelper, n.app.EvmKeeper)
	return evmtypes.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetGovClient() govtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	govtypes.RegisterQueryServer(queryHelper, govkeeper.NewQueryServer(&n.app.GovKeeper))
	return govtypes.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetBankClient() banktypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	banktypes.RegisterQueryServer(queryHelper, n.app.BankKeeper)
	return banktypes.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetFeeMarketClient() feemarkettypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	feemarkettypes.RegisterQueryServer(queryHelper, n.app.FeeMarketKeeper)
	return feemarkettypes.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetInflationClient() infltypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	infltypes.RegisterQueryServer(queryHelper, n.app.InflationKeeper)
	return infltypes.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetAuthClient() authtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	authtypes.RegisterQueryServer(queryHelper, authkeeper.NewQueryServer(n.app.AccountKeeper))
	return authtypes.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetAuthzClient() authz.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	authz.RegisterQueryServer(queryHelper, n.app.AuthzKeeper)
	return authz.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetStakingClient() stakingtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	stakingtypes.RegisterQueryServer(queryHelper, stakingkeeper.Querier{Keeper: n.app.StakingKeeper.Keeper})
	return stakingtypes.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetDistrClient() distrtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	distrtypes.RegisterQueryServer(queryHelper, distrkeeper.Querier{Keeper: n.app.DistrKeeper})
	return distrtypes.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetEpochsClient() epochstypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	epochstypes.RegisterQueryServer(queryHelper, n.app.EpochsKeeper)
	return epochstypes.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetVestingClient() vestingtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext(), n.GetEncodingConfig())
	vestingtypes.RegisterQueryServer(queryHelper, n.app.VestingKeeper)
	return vestingtypes.NewQueryClient(queryHelper)
}
