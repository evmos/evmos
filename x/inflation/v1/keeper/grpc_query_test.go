package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	testkeyring "github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	evmostypes "github.com/evmos/evmos/v20/types"
	"github.com/evmos/evmos/v20/utils"
	"github.com/evmos/evmos/v20/x/inflation/v1/types"
)

func TestPeriod(t *testing.T) { //nolint:dupl
	var (
		ctx    sdk.Context
		nw     *network.UnitTestNetwork
		req    *types.QueryPeriodRequest
		expRes *types.QueryPeriodResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"default period",
			func() {
				req = &types.QueryPeriodRequest{}
				expRes = &types.QueryPeriodResponse{}
			},
			true,
		},
		{
			"set period",
			func() {
				period := uint64(9)
				nw.App.InflationKeeper.SetPeriod(ctx, period)

				req = &types.QueryPeriodRequest{}
				expRes = &types.QueryPeriodResponse{Period: period}
			},
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			// reset
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()
			qc := nw.GetInflationClient()

			tc.malleate()

			res, err := qc.Period(ctx, req)
			if tc.expPass {
				require.NoError(t, err)
				require.Equal(t, expRes, res)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestEpochMintProvision(t *testing.T) {
	var (
		ctx         sdk.Context
		nw          *network.UnitTestNetwork
		req         *types.QueryEpochMintProvisionRequest
		expRes      *types.QueryEpochMintProvisionResponse
		bondedRatio math.LegacyDec
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"default epochMintProvision",
			func() {
				params := types.DefaultParams()
				defaultEpochMintProvision := types.CalculateEpochMintProvision(
					params,
					uint64(0),
					365,
					bondedRatio,
				)
				expEpochMintProvision := defaultEpochMintProvision.Quo(math.LegacyNewDec(types.ReductionFactor))
				req = &types.QueryEpochMintProvisionRequest{}
				expRes = &types.QueryEpochMintProvisionResponse{
					EpochMintProvision: sdk.NewDecCoinFromDec(types.DefaultInflationDenom, expEpochMintProvision),
				}
			},
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			// reset
			nw = network.NewUnitTestNetwork(network.WithChainID(utils.TestnetChainID + "-1"))
			ctx = nw.GetContext()
			qc := nw.GetInflationClient()

			// get bonded ratio
			var err error
			bondedRatio, err = nw.App.InflationKeeper.BondedRatio(ctx)
			require.NoError(t, err)

			tc.malleate()

			res, err := qc.EpochMintProvision(ctx, req)
			if tc.expPass {
				require.NoError(t, err)
				require.Equal(t, expRes, res)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestSkippedEpochs(t *testing.T) { //nolint:dupl
	var (
		ctx    sdk.Context
		nw     *network.UnitTestNetwork
		req    *types.QuerySkippedEpochsRequest
		expRes *types.QuerySkippedEpochsResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"default skipped epochs",
			func() {
				req = &types.QuerySkippedEpochsRequest{}
				expRes = &types.QuerySkippedEpochsResponse{}
			},
			true,
		},
		{
			"set skipped epochs",
			func() {
				skippedEpochs := uint64(9)
				nw.App.InflationKeeper.SetSkippedEpochs(ctx, skippedEpochs)

				req = &types.QuerySkippedEpochsRequest{}
				expRes = &types.QuerySkippedEpochsResponse{SkippedEpochs: skippedEpochs}
			},
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			// reset
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()
			qc := nw.GetInflationClient()

			tc.malleate()

			res, err := qc.SkippedEpochs(ctx, req)
			if tc.expPass {
				require.NoError(t, err)
				require.Equal(t, expRes, res)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestQueryCirculatingSupply(t *testing.T) {
	nAccs := int64(1)
	nVals := int64(3)

	keyring := testkeyring.New(int(nAccs))
	nw := network.NewUnitTestNetwork(
		network.WithAmountOfValidators(int(nVals)),
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	ctx := nw.GetContext()
	qc := nw.GetInflationClient()

	// Mint coins to increase supply
	mintDenom := nw.App.InflationKeeper.GetParams(ctx).MintDenom
	mintCoin := sdk.NewCoin(mintDenom, sdk.TokensFromConsensusPower(int64(400_000_000), evmostypes.PowerReduction))
	err := nw.App.InflationKeeper.MintCoins(ctx, mintCoin)
	require.NoError(t, err)

	// team allocation is zero if not on mainnet
	expCirculatingSupply := sdk.NewDecCoin(mintDenom, sdk.TokensFromConsensusPower(200_000_000, evmostypes.PowerReduction))

	// the total bonded tokens for the 4 accounts initialized on the setup (3 validators, 1 EOA)
	bondedAmount := network.DefaultBondedAmount.MulRaw(nVals)
	bondedAmount = bondedAmount.Add(network.PrefundedAccountInitialBalance.MulRaw(nAccs))
	bondedCoins := sdk.NewDecCoin(evmostypes.AttoEvmos, bondedAmount)

	res, err := qc.CirculatingSupply(ctx, &types.QueryCirculatingSupplyRequest{})
	require.NoError(t, err)
	require.Equal(t, expCirculatingSupply.Add(bondedCoins), res.CirculatingSupply)
}

func TestQueryInflationRate(t *testing.T) {
	nAccs := int64(1)
	nVals := int64(3)

	keyring := testkeyring.New(int(nAccs))
	nw := network.NewUnitTestNetwork(
		network.WithAmountOfValidators(int(nVals)),
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	ctx := nw.GetContext()
	qc := nw.GetInflationClient()

	// the total bonded tokens for the 4 accounts initialized on the setup (3 validators, 1 EOA)
	bondedAmt := network.DefaultBondedAmount.MulRaw(nVals)                          // Add the allocation for the validators
	bondedAmt = bondedAmt.Add(network.PrefundedAccountInitialBalance.MulRaw(nAccs)) // Add the allocation for the EOA

	// Mint coins to increase supply
	mintDenom := nw.App.InflationKeeper.GetParams(ctx).MintDenom
	mintCoin := sdk.NewCoin(mintDenom, sdk.TokensFromConsensusPower(int64(400_000_000), evmostypes.PowerReduction).Sub(bondedAmt))
	err := nw.App.InflationKeeper.MintCoins(ctx, mintCoin)
	require.NoError(t, err)

	expInflationRate := math.LegacyMustNewDecFromStr("51.5625").Quo(math.LegacyNewDec(types.ReductionFactor))
	res, err := qc.InflationRate(ctx, &types.QueryInflationRateRequest{})
	require.NoError(t, err)
	require.Equal(t, expInflationRate, res.InflationRate)
}

func TestQueryParams(t *testing.T) {
	nw := network.NewUnitTestNetwork()
	ctx := nw.GetContext()
	qc := nw.GetInflationClient()

	expParams := types.DefaultParams()

	res, err := qc.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, expParams, res.Params)
}
