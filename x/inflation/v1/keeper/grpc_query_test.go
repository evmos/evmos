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
					EpochMintProvision: sdk.NewDecCoinFromDec(denomMint, expEpochMintProvision),
				}
			},
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			// reset
			nw = network.NewUnitTestNetwork(network.WithChainID(utils.MainnetChainID + "-1"))
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

	prefundedAccBalance := network.PrefundedAccountInitialBalance

	keyring := testkeyring.New(int(nAccs))

	// Foundation wallets are not considered in the circulating supply.
	foundationAcc := []sdk.AccAddress{
		utils.EthHexToCosmosAddr(types.FoundationWallets[0]),
		utils.EthHexToCosmosAddr(types.FoundationWallets[1]),
	}

	nw := network.NewUnitTestNetwork(
		network.WithAmountOfValidators(int(nVals)),
		network.WithPreFundedAccounts(append(keyring.GetAllAccAddrs(), foundationAcc...)...),
	)
	ctx := nw.GetContext()
	qc := nw.GetInflationClient()

	// Mint coins to increase the supply.
	mintDenom := nw.App.InflationKeeper.GetParams(ctx).MintDenom
	mintCoin := sdk.NewCoin(mintDenom, prefundedAccBalance.MulRaw(4))
	err := nw.App.InflationKeeper.MintCoins(ctx, mintCoin)
	require.NoError(t, err)

	// Expected circulating supply is composed only of the minted tokens plus
	// pre-funded accounts balances except foundation wallets.
	// Foundation wallets are removed in the computation, that's why we multiply
	// by 4 (minted coins) + number of EOA and we don't add the number of
	// foundation accounts.
	//
	// NOTE: wallets associated with nAccs have part of the balance delegated
	// but it is all considered in one place for simplicity.
	expCirculatingSupply := sdk.NewDecCoin(mintDenom, prefundedAccBalance.MulRaw(4+nAccs))

	// The total bonded tokens for the 4 accounts initialized on the setup (3
	// validators, 1 EOA).
	//
	// NOTE: the EOA delegate 1 token to every validator but it is already
	// accounted for in the expCirculatingSupply.
	bondedAmount := network.DefaultBondedAmount.MulRaw(nVals)
	bondedCoins := sdk.NewDecCoin(evmostypes.BaseDenom, bondedAmount)

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

	// Genesis available tokens are defined by the testing suite setup:
	//- Validators' self delegation.
	//- Tokens delegated by EOA.
	//- Free EOA tokens.
	valBondedAmt := network.DefaultBondedAmount.MulRaw(nVals)
	accsBondAmount := math.OneInt().MulRaw(nVals)
	accsFreeAmount := network.PrefundedAccountInitialBalance.MulRaw(nAccs).Sub(accsBondAmount)

	// Mint other coins to the inflation module to increase circulating supply.
	mintDenom := nw.App.InflationKeeper.GetParams(ctx).MintDenom
	mintAmount := network.PrefundedAccountInitialBalance.MulRaw(4)
	mintCoin := sdk.NewCoin(mintDenom, mintAmount)
	err := nw.App.InflationKeeper.MintCoins(ctx, mintCoin)
	require.NoError(t, err)

	circulatingSupply := valBondedAmt.Add(accsBondAmount).Add(accsFreeAmount).Add(mintAmount)

	epp := nw.App.InflationKeeper.GetEpochsPerPeriod(ctx)
	epochsPerPeriod := math.LegacyNewDec(epp)
	epochMintProvision := nw.App.InflationKeeper.GetEpochMintProvision(ctx)

	expInflationRate := epochMintProvision.Mul(epochsPerPeriod).Quo(math.LegacyNewDecFromInt(circulatingSupply)).Mul(math.LegacyNewDec(100))

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
