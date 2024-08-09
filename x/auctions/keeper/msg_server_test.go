package keeper_test

import (
	"testing"

	testkeyring "github.com/evmos/evmos/v19/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	"github.com/stretchr/testify/require"

	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/auctions/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestBid(t *testing.T) {
	var validSenderKey testkeyring.Key
	var network *testnetwork.UnitTestNetwork

	testCases := []struct {
		name     string
		input    func() *types.MsgBid
		malleate func()
		expErr   bool
	}{
		{
			name: "pass",
			input: func() *types.MsgBid {
				return &types.MsgBid{
					Sender: validSenderKey.AccAddr.String(),
					Amount: sdk.NewCoin(utils.BaseDenom, sdk.NewInt(1)),
				}
			},
			malleate: func() {
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keyring := testkeyring.New(1)
			network = testnetwork.NewUnitTestNetwork(
				testnetwork.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
			)

			validSenderKey = keyring.GetKey(0)

			_, err := network.App.AuctionsKeeper.Bid(network.GetContext(), tc.input())

			if tc.expErr {
			} else {
				require.NoError(t, err)
			}
		})
	}
}
