package osmosis_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/evmos/evmos/v15/precompiles/outposts/osmosis"
	testkeyring "github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"
)

func TestNewPrecompile(t *testing.T) {
	portID := "transfer"     //nolint:goconst
	channelID := "channel-0" //nolint:goconst
	keyring := testkeyring.New(2)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

	testCases := []struct {
		name            string
		contractAddress string
		expPass         bool
		errContains     string
	}{
		{
			name:            "fail - empty contract address",
			contractAddress: "",
			expPass:         false,
			errContains:     fmt.Sprintf(osmosis.ErrInvalidContractAddress),
		},
		{
			name:            "pass - not contract address",
			contractAddress: "osmo1qql8ag4cluz6r4dz28p3w00dnc9w8ueuhnecd2",
			expPass:         false,
			errContains:     fmt.Sprintf(osmosis.ErrInvalidContractAddress),
		},
		{
			name:            "fail - not osmosis smart contract",
			contractAddress: "evmos18rj46qcpr57m3qncrj9cuzm0gn3km08w5jxxlnw002c9y7xex5xsu74ytz",
			expPass:         false,
			errContains:     fmt.Sprintf(osmosis.ErrInvalidContractAddress),
		},
		{
			name:            "pass - valid contract address",
			contractAddress: osmosis.XCSContractTestnet,
			expPass:         true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			_, err := osmosis.NewPrecompile(
				unitNetwork.App.AuthzKeeper,
				portID,
				channelID,
				tc.contractAddress,
				unitNetwork.App.BankKeeper,
				unitNetwork.App.TransferKeeper,
				unitNetwork.App.StakingKeeper,
				unitNetwork.App.Erc20Keeper,
				unitNetwork.App.IBCKeeper.ChannelKeeper,
			)

			if tc.expPass {
				require.NoError(t, err, "expected no error while validating the contract address")
			} else {
				require.Error(t, err, "expected error while validating the contract address")
				require.Contains(t, err.Error(), tc.errContains)
			}
		})

	}
}
