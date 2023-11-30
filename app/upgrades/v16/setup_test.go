// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v16_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"

	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"
	inflationtypes "github.com/evmos/evmos/v15/x/inflation/v1/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	network *network.UnitTestNetwork
	handler grpc.Handler
	keyring keyring.Keyring
	factory factory.TxFactory

	incentivesAcc sdk.AccAddress
}

// Initial inflation distribution is the previous default configuration
var initialInflDistr = inflationtypes.InflationDistribution{
	StakingRewards:  math.LegacyNewDecWithPrec(533333334, 9), // 0.53 = 40% / (1 - 25%)
	UsageIncentives: math.LegacyNewDecWithPrec(333333333, 9), // 0.33 = 25% / (1 - 25%)
	CommunityPool:   math.LegacyNewDecWithPrec(133333333, 9), // 0.13 = 10% / (1 - 25%)
}

func (its *IntegrationTestSuite) SetupTest() {
	keys := keyring.New(2)
	// Set some balance to the incentives module account
	its.incentivesAcc = authtypes.NewModuleAddress("incentives")
	accs := append(keys.GetAllAccAddrs(), its.incentivesAcc)
	nw := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(accs...),
	)
	gh := grpc.NewIntegrationHandler(nw)
	tf := factory.New(nw, gh)

	// Set inflation params to have UsageIncentives > 0
	updatedParams := inflationtypes.DefaultParams()
	updatedParams.InflationDistribution = initialInflDistr
	err := nw.UpdateInflationParams(updatedParams)
	its.Require().NoError(err)

	its.network = nw
	its.factory = tf
	its.handler = gh
	its.keyring = keys
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
