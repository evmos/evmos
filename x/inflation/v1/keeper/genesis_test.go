package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	inflationkeeper "github.com/evmos/evmos/v16/x/inflation/v1/keeper"
	"github.com/stretchr/testify/require"
)

func TestInitGenesis(t *testing.T) {
	nw := network.NewUnitTestNetwork()
	ctx := nw.GetContext()
	// check calculated epochMintProvision at genesis
	epochMintProvision := nw.App.InflationKeeper.GetEpochMintProvision(ctx)
	expMintProvision := math.LegacyMustNewDecFromStr("847602739726027397260274.000000000000000000").Quo(math.LegacyNewDec(inflationkeeper.ReductionFactor))
	require.Equal(t, expMintProvision, epochMintProvision)
}
