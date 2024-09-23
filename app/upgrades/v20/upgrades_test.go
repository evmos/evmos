package v20_test

import (
	"testing"

	"cosmossdk.io/log"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	v20 "github.com/evmos/evmos/v20/app/upgrades/v20"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v20/x/evm/types"
)

func TestEnableGovPrecompile(t *testing.T) {
	var (
		nw  *network.UnitTestNetwork
		ctx sdk.Context
	)

	testCases := []struct {
		name        string
		setup       func()
		expFail     bool
		errContains string
	}{
		{
			name:        "fail - duplicated",
			setup:       func() {},
			expFail:     true,
			errContains: "duplicate precompile",
		},
		{
			name: "pass - enable gov precompile",
			setup: func() {
				params := nw.App.EvmKeeper.GetParams(ctx)
				params.ActiveStaticPrecompiles = []string{
					types.StakingPrecompileAddress,
					types.DistributionPrecompileAddress,
					types.ICS20PrecompileAddress,
					types.VestingPrecompileAddress,
				}
				require.NoError(t, nw.App.EvmKeeper.SetParams(ctx, params))
			},
			expFail: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()
			tc.setup()

			err := v20.EnableGovPrecompile(ctx, nw.App.EvmKeeper)
			if tc.expFail {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
				return
			}
			require.NoError(t, err)
			updatedParams := nw.App.EvmKeeper.GetParams(ctx)
			require.Contains(t, updatedParams.ActiveStaticPrecompiles, types.GovPrecompileAddress)
		})
	}
}

func TestUpdateExpeditedPropsParams(t *testing.T) {
	var (
		nw            *network.UnitTestNetwork
		ctx           sdk.Context
		initialParams govv1.Params
		err           error
	)

	testCases := []struct {
		name      string
		setup     func()
		postCheck func()
	}{
		{
			name:  "pass - default params, no-op",
			setup: func() {},
			postCheck: func() {
				params, err := nw.App.GovKeeper.Params.Get(ctx)
				require.NoError(t, err)
				require.Equal(t, initialParams, params)
			},
		},
		{
			name: "pass - expedited has 'stake' denom",
			setup: func() {
				// wrong exp min deposit denom
				initialParams.ExpeditedMinDeposit[0].Denom = "stake"
				err := nw.App.GovKeeper.Params.Set(ctx, initialParams)
				require.NoError(t, err)
			},
			postCheck: func() {
				params, err := nw.App.GovKeeper.Params.Get(ctx)
				require.NoError(t, err)
				require.Equal(t, "aevmos", params.ExpeditedMinDeposit[0].Denom)
				require.Equal(t, initialParams.ExpeditedMinDeposit[0].Amount, params.ExpeditedMinDeposit[0].Amount)
			},
		},
		{
			name: "pass - updates denom, amount and period",
			setup: func() {
				// wrong exp min deposit denom
				initialParams.ExpeditedMinDeposit[0].Denom = "stake"
				// wrong exp min deposit amount (< than min_deposit amt)
				initialParams.ExpeditedMinDeposit[0].Amount = initialParams.MinDeposit[0].Amount.SubRaw(1)
				// wrong exp voting period (> than voting period)
				expPeriod := *initialParams.VotingPeriod * 2
				initialParams.ExpeditedVotingPeriod = &expPeriod
				err := nw.App.GovKeeper.Params.Set(ctx, initialParams)
				require.NoError(t, err)
			},
			postCheck: func() {
				params, err := nw.App.GovKeeper.Params.Get(ctx)
				require.NoError(t, err)
				require.Equal(t, "aevmos", params.ExpeditedMinDeposit[0].Denom)
				require.Equal(t, initialParams.MinDeposit[0].Amount.MulRaw(govv1.DefaultMinExpeditedDepositTokensRatio), params.ExpeditedMinDeposit[0].Amount)
				require.Equal(t, *initialParams.VotingPeriod/2, *params.ExpeditedVotingPeriod)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()
			initialParams, err = nw.App.GovKeeper.Params.Get(ctx)
			require.NoError(t, err)
			// setup for testcase
			tc.setup()

			err = v20.UpdateExpeditedPropsParams(ctx, nw.App.GovKeeper)
			require.NoError(t, err)

			tc.postCheck()
		})
	}
}

var metaDatas = []banktypes.Metadata{
	{
		Description: "The native EVM, governance and staking token of the Evmos Hub",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:   "aevmos",
				Aliases: []string{"attoevmos"},
			},
			{
				Denom:    "evmos",
				Exponent: 18,
			},
		},
		Base:    "aevmos",
		Display: "evmos",
		Name:    "Evmos",
		Symbol:  "EVMOS",
	},
	{
		Description: "Cosmos coin token representation of 0x153A59d48AcEAbedbDCf7a13F67Ae52b434B810B",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom: "erc20/0x153A59d48AcEAbedbDCf7a13F67Ae52b434B810B",
			},
			{
				Denom:    "WrappedEtherCeler",
				Exponent: 18,
			},
		},
		Base:    "erc20/0x153A59d48AcEAbedbDCf7a13F67Ae52b434B810B",
		Display: "WrappedEtherCeler",
		Name:    "erc20/0x153A59d48AcEAbedbDCf7a13F67Ae52b434B810B",
		Symbol:  "ceWETH",
	},
	{
		Description: "Staking derivative stATOM for staked ATOM by Stride",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:   "ibc/0830AFFC2F4F7CD24F9CEC07024FEA64CE3C5ABBC520DBD803BFA97BC3DCCA85",
				Aliases: []string{"stuatom"},
			},
			{
				Denom:    "statom",
				Exponent: 6,
			},
		},
		Base:    "ibc/0830AFFC2F4F7CD24F9CEC07024FEA64CE3C5ABBC520DBD803BFA97BC3DCCA85",
		Display: "statom",
		Name:    "Stride Staked Atom",
		Symbol:  "stATOM",
	},
}

func TestFixDenomMetadata(t *testing.T) {
	var (
		nw  *network.UnitTestNetwork
		ctx sdk.Context
	)
	testCases := []struct {
		name  string
		setup func()
	}{
		{
			name: "all keys are OK - no change",
			setup: func() {
				for _, m := range metaDatas {
					nw.App.BankKeeper.SetDenomMetaData(ctx, m)
				}
			},
		},
		{
			name: "all corrupted keys",
			setup: func() {
				bk, ok := nw.App.BankKeeper.(bankkeeper.BaseKeeper)
				require.True(t, ok)
				for _, m := range metaDatas {
					corruptedKey := m.Base + "a"
					err := bk.BaseViewKeeper.DenomMetadata.Set(ctx, corruptedKey, m)
					require.NoError(t, err)
					// check that cannot retrieve metadata with correct key
					_, found := bk.GetDenomMetaData(ctx, m.Base)
					require.False(t, found)
				}
			},
		},
		{
			name: "some corrupted keys others OK",
			setup: func() {
				bk, ok := nw.App.BankKeeper.(bankkeeper.BaseKeeper)
				require.True(t, ok)
				corrupted := true
				for _, m := range metaDatas {
					key := m.Base
					if corrupted {
						key += "a"
					}
					err := bk.BaseViewKeeper.DenomMetadata.Set(ctx, key, m)
					require.NoError(t, err)
					// check that cannot retrieve metadata with correct key
					_, found := bk.GetDenomMetaData(ctx, m.Base)
					require.Equal(t, !corrupted, found)
					corrupted = !corrupted
				}
			},
		},
		{
			name: "duplicated entry with corrupted key (IBC coin)",
			setup: func() {
				for _, m := range metaDatas {
					nw.App.BankKeeper.SetDenomMetaData(ctx, m)
				}
				bk, ok := nw.App.BankKeeper.(bankkeeper.BaseKeeper)
				require.True(t, ok)
				ibcCoinMeta := metaDatas[len(metaDatas)-1]
				corruptedKey := ibcCoinMeta.Base + "i"
				err := bk.BaseViewKeeper.DenomMetadata.Set(ctx, corruptedKey, ibcCoinMeta)
				require.NoError(t, err)
				m, found := bk.GetDenomMetaData(ctx, ibcCoinMeta.Base)
				require.True(t, found)
				require.Equal(t, ibcCoinMeta, m)
				m, found = bk.GetDenomMetaData(ctx, corruptedKey)
				require.True(t, found)
				require.Equal(t, ibcCoinMeta, m)
				// there should be a duplicated entry
				// on the denom metas list
				res, err := nw.App.BankKeeper.DenomsMetadata(ctx, &banktypes.QueryDenomsMetadataRequest{})
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Len(t, res.Metadatas, len(metaDatas)+1)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()

			// setup for testcase
			tc.setup()

			err := v20.FixDenomMetadata(ctx, log.NewNopLogger(), nw.App.BankKeeper)
			require.NoError(t, err)

			for _, m := range metaDatas {
				res, err := nw.App.BankKeeper.DenomMetadata(ctx, &banktypes.QueryDenomMetadataRequest{Denom: m.Base})
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, m, res.Metadata)
			}

			// check the denom meta length
			res, err := nw.App.BankKeeper.DenomsMetadata(ctx, &banktypes.QueryDenomsMetadataRequest{})
			require.NoError(t, err)
			require.NotNil(t, res)
			require.Len(t, res.Metadatas, len(metaDatas))
		})
	}
}
