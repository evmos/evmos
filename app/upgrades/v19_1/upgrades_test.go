package v191_test

import (
	"testing"

	"github.com/evmos/evmos/v19/app/upgrades/v19_1"
	testnetwork "github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	"github.com/stretchr/testify/require"
)

func TestEnableCustomEIPs(t *testing.T) {
	upgradeEIPs := []string{"evmos_0"}

	testCases := []struct {
		name       string
		activeEIPs []string
		expEIPsNum int
	}{
		{
			name:       "repeated EIP - skip",
			activeEIPs: []string{"evmos_0"},
			expEIPsNum: 1,
		},
		{
			name:       "all new EIP",
			activeEIPs: []string{"ethereum_3855"},
			expEIPsNum: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			network := testnetwork.NewUnitTestNetwork()

			require.NoError(t, network.NextBlock(), "failed to advance block")

			oldParams := network.App.EvmKeeper.GetParams(network.GetContext())
			oldParams.ExtraEIPs = tc.activeEIPs
			err := network.UpdateEvmParams(oldParams)
			require.NoError(t, err, "failed to update EVM params")

			logger := network.GetContext().Logger()
			err = v191.EnableCustomEIPs(network.GetContext(), logger, network.App.EvmKeeper)
			require.NoError(t, err)

			params := network.App.EvmKeeper.GetParams(network.GetContext())
			require.Equal(t, tc.expEIPsNum, len(params.ExtraEIPs))

			require.Subset(t, params.ExtraEIPs, upgradeEIPs, "expected all new EIPs to be present")
		})
	}
}
