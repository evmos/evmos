// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v16_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/ethereum/go-ethereum/common"
	v16 "github.com/evmos/evmos/v16/app/upgrades/v16"
	"github.com/evmos/evmos/v16/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v16/testutil"
	testnetwork "github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	testutils "github.com/evmos/evmos/v16/testutil/integration/evmos/utils"
	utiltx "github.com/evmos/evmos/v16/testutil/tx"
	"github.com/evmos/evmos/v16/utils"
	erc20types "github.com/evmos/evmos/v16/x/erc20/types"
	incentives "github.com/evmos/evmos/v16/x/incentives/types"
	"github.com/stretchr/testify/require"
)

func (its *IntegrationTestSuite) TestMigrateFeeCollector() {
	its.SetupTest()

	feeCollectorModuleAccount := its.network.App.AccountKeeper.GetModuleAccount(its.network.GetContext(), authtypes.FeeCollectorName)
	modAcc, ok := feeCollectorModuleAccount.(*authtypes.ModuleAccount)
	its.Require().True(ok)

	oldFeeCollector := authtypes.NewModuleAccount(modAcc.BaseAccount, authtypes.FeeCollectorName)

	its.Require().NotNil(oldFeeCollector)
	its.Require().Len(oldFeeCollector.GetPermissions(), 0)

	// Create a new FeeCollector module account with the same address and the new permissions.
	newFeeCollectorModuleAccount := authtypes.NewModuleAccount(modAcc.BaseAccount, authtypes.FeeCollectorName, authtypes.Burner)
	its.network.App.AccountKeeper.SetModuleAccount(its.network.GetContext(), newFeeCollectorModuleAccount)

	newFeeCollector := its.network.App.AccountKeeper.GetModuleAccount(its.network.GetContext(), authtypes.FeeCollectorName)
	its.Require().True(ok)
	its.Require().Equal(feeCollectorModuleAccount.GetAccountNumber(), newFeeCollector.GetAccountNumber())
	its.Require().Equal(feeCollectorModuleAccount.GetAddress(), newFeeCollector.GetAddress())
	its.Require().Equal(feeCollectorModuleAccount.GetName(), newFeeCollector.GetName())
	its.Require().Equal(feeCollectorModuleAccount.GetPermissions(), newFeeCollector.GetPermissions())
}

func (its *IntegrationTestSuite) TestBurnUsageIncentivesPool() {
	its.SetupTest()
	// check initial balance of incentives mod account
	expIntialBalance := sdk.NewCoin(utils.BaseDenom, testnetwork.PrefundedAccountInitialBalance)
	res, err := its.handler.GetBalance(its.incentivesAcc, utils.BaseDenom)
	its.Require().NoError(err)
	its.Require().NotNil(res.Balance)
	its.Require().Equal(expIntialBalance, *res.Balance)

	err = v16.BurnUsageIncentivesPool(its.network.GetContext(), its.network.App.BankKeeper)
	its.Require().NoError(err)

	// Check incentives pool final balance - should be zero
	expFinalBalance := sdk.NewCoin(utils.BaseDenom, math.ZeroInt())
	res, err = its.handler.GetBalance(its.incentivesAcc, utils.BaseDenom)
	its.Require().NoError(err)
	its.Require().NotNil(res.Balance)
	its.Require().Equal(expFinalBalance, *res.Balance)
}

func (its *IntegrationTestSuite) TestUpdateInflationParams() {
	its.SetupTest()
	// check initial inflation params has incentive allocation > 0
	initialParams := its.network.App.InflationKeeper.GetParams(its.network.GetContext())
	its.Require().Equal(initialParams.InflationDistribution.UsageIncentives, initialInflDistr.UsageIncentives) //nolint:staticcheck

	err := v16.UpdateInflationParams(its.network.GetContext(), its.network.App.InflationKeeper)
	its.Require().NoError(err)

	// Check incentives allocation is zero
	finalParams := its.network.App.InflationKeeper.GetParams(its.network.GetContext())
	its.Require().Equal(math.LegacyZeroDec(), finalParams.InflationDistribution.UsageIncentives) //nolint:staticcheck
}

func (its *IntegrationTestSuite) TestDeleteDeprecatedProposals() {
	its.SetupTest()

	// Create 4 proposals. 3 will be deleted which correspond to the deprecated proposals
	expInitialProps := 4
	expFinalProps := 1
	prop1 := &incentives.RegisterIncentiveProposal{
		Title:       "Test",
		Description: "Test Register Incentive Proposal",
		Contract:    utiltx.GenerateAddress().String(),
		Allocations: sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
		Epochs:      100,
	}

	prop2 := &upgrade.SoftwareUpgradeProposal{ //nolint:staticcheck
		Title:       "Test",
		Description: "Test Software Upgrade Proposal",
		Plan:        upgrade.Plan{},
	}

	prop3 := &incentives.CancelIncentiveProposal{
		Title:       "Test",
		Description: "Test Cancel Incentive Proposal",
		Contract:    utiltx.GenerateAddress().String(),
	}

	prop4 := &erc20types.RegisterCoinProposal{ //nolint:staticcheck
		Title:       "Test",
		Description: "Test Register Coin Proposal",
		Metadata:    []banktypes.Metadata{},
	}

	privKey, _ := ethsecp256k1.GenerateKey()
	addrBz := privKey.PubKey().Address().Bytes()
	accAddr := sdk.AccAddress(addrBz)
	coins := sdk.NewCoins(sdk.NewCoin(its.network.GetDenom(), math.NewInt(5e18)))
	err := testutil.FundAccount(its.network.GetContext(), its.network.App.BankKeeper, accAddr, coins)
	its.Require().NoError(err)

	for _, prop := range []govtypesv1beta.Content{prop1, prop2, prop3, prop4} {
		its.createProposal(prop, accAddr)
	}

	// check the creation of the 3 proposals was successful
	allProposalsBefore := its.network.App.GovKeeper.GetProposals(its.network.GetContext())
	its.Require().Len(allProposalsBefore, expInitialProps)

	// Delete the corresponding proposals
	logger := its.network.GetContext().Logger()
	v16.DeleteDeprecatedProposals(its.network.GetContext(), its.network.App.GovKeeper, logger)

	allProposalsAfter := its.network.App.GovKeeper.GetProposals(its.network.GetContext())
	its.Require().Len(allProposalsAfter, expFinalProps)
}

func (its *IntegrationTestSuite) createProposal(content govtypesv1beta.Content, acc sdk.AccAddress) {
	allProposalsBefore := its.network.App.GovKeeper.GetProposals(its.network.GetContext())
	propID := len(allProposalsBefore) + 1

	legacyContent, err := govtypesv1.NewLegacyContent(
		content,
		sdk.MustBech32ifyAddressBytes(sdk.GetConfig().GetBech32AccountAddrPrefix(), acc),
	)
	its.Require().NoError(err)

	proposalMsgs := []sdk.Msg{legacyContent}
	newProposal, err := govtypesv1.NewProposal(proposalMsgs, uint64(propID), time.Now(), time.Now().Add(time.Hour*5), "", "Test", "Test", acc)
	its.Require().NoError(err)
	its.network.App.GovKeeper.SetProposal(its.network.GetContext(), newProposal)
}

func TestConvertToNativeCoinExtensions(t *testing.T) {
	ts, err := SetupConvertERC20CoinsTest(t)
	require.NoError(t, err, "failed to setup test")

	logger := ts.network.GetContext().Logger().With("upgrade")

	// Convert the coins back using the upgrade util
	err = v16.ConvertToNativeCoinExtensions(
		ts.network.GetContext(),
		logger,
		ts.network.App.AccountKeeper,
		ts.network.App.BankKeeper,
		ts.network.App.Erc20Keeper,
		ts.wevmosContract,
	)
	require.NoError(t, err, "failed to convert coins")

	err = ts.network.NextBlock()
	require.NoError(t, err, "failed to execute block")

	// NOTE: Here we check that the ERC20 converted coins have been added back to the bank balance.
	err = testutils.CheckBalances(ts.handler, testutils.ExpectedBalances{
		{Address: ts.keyring.GetAccAddr(testAccount), Coins: sdk.NewCoins(sdk.NewInt64Coin(XMPL, 300))},
		{Address: ts.keyring.GetAccAddr(erc20Deployer), Coins: sdk.NewCoins(sdk.NewInt64Coin(XMPL, 200))},
	})
	require.NoError(t, err, "failed to check balances")

	// NOTE: we check that the token pair was registered as an active precompile
	evmParams, err := ts.handler.GetEvmParams()
	require.NoError(t, err, "failed to get evm params")
	require.Contains(t, evmParams.Params.ActivePrecompiles, ts.tokenPair.GetERC20Contract().String(),
		"expected token pair precompile to be active",
	)
	require.NotContains(t, evmParams.Params.ActivePrecompiles, ts.nonNativeTokenPair.GetERC20Contract().String(),
		"expected non-native token pair not to be a precompile",
	)

	// NOTE: We check that the ERC20 contract for the token pair can still be called (now as an EVM extension)
	balance, err := GetERC20Balance(ts.factory, ts.keyring.GetPrivKey(testAccount), ts.tokenPair.GetERC20Contract())
	require.NoError(t, err, "failed to query ERC20 balance")
	require.Equal(t, int64(300), balance.Int64(), "expected different balance after converting ERC20")

	// NOTE: We check that the balance of the module address is empty after converting native ERC20s
	balancesRes, err := ts.handler.GetAllBalances(authtypes.NewModuleAddress(erc20types.ModuleName))
	require.NoError(t, err, "failed to get balances")
	require.True(t, balancesRes.Balances.IsZero(), "expected different balance for module account")

	// NOTE: We check that the erc20deployer account still has the minted balance after converting the native ERC20s only.
	balance, err = GetERC20Balance(ts.factory, ts.keyring.GetPrivKey(erc20Deployer), ts.nonNativeTokenPair.GetERC20Contract())
	require.NoError(t, err, "failed to query ERC20 balance")
	require.Equal(t, mintAmount, balance, "expected different balance after converting ERC20")

	// NOTE: We check that there all balance of the WEVMOS contract was withdrawn too.
	balance, err = GetERC20Balance(ts.factory, ts.keyring.GetPrivKey(testAccount), ts.wevmosContract)
	require.NoError(t, err, "failed to query ERC20 balance")
	require.Equal(t, common.Big0.Int64(), balance.Int64(), "expected no WEVMOS left after conversion")
}
