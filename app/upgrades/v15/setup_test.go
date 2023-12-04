// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v15_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmosapp "github.com/evmos/evmos/v16/app"
	"github.com/stretchr/testify/suite"
)

var s *UpgradesTestSuite

type UpgradesTestSuite struct {
	suite.Suite

	ctx        sdk.Context
	app        *evmosapp.Evmos
	validators []stakingtypes.Validator
	bondDenom  string
}

func TestUpgradeTestSuite(t *testing.T) {
	s = new(UpgradesTestSuite)
	suite.Run(t, s)
}

func (s *UpgradesTestSuite) SetupTest() {
	s.DoSetupTest()
}
