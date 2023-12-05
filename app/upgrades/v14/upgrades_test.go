package v14_test

import (
	v14 "github.com/evmos/evmos/v16/app/upgrades/v14"
	vestingprecompile "github.com/evmos/evmos/v16/precompiles/vesting"
)

func (s *UpgradesTestSuite) TestEnableVestingExtension() {
	s.SetupTest()

	vestingPrecompile := vestingprecompile.Precompile{}

	evmParams := s.app.EvmKeeper.GetParams(s.ctx)
	allPrecompiles := evmParams.ActivePrecompiles
	newPrecompiles := make([]string, 0, len(allPrecompiles)-1)
	for _, precompile := range allPrecompiles {
		if precompile == vestingPrecompile.Address().String() {
			continue
		}
		newPrecompiles = append(newPrecompiles, precompile)
	}
	s.Require().NotContains(newPrecompiles, vestingPrecompile.Address().String(),
		"expected vesting extension to be removed from active precompiles",
	)

	evmParams.ActivePrecompiles = newPrecompiles
	err := s.app.EvmKeeper.SetParams(s.ctx, evmParams)
	s.Require().NoError(err, "failed to set evm params")

	err = v14.EnableVestingExtension(s.ctx, s.app.EvmKeeper)
	s.Require().NoError(err, "failed to enable vesting extension")

	evmParams = s.app.EvmKeeper.GetParams(s.ctx)
	s.Require().Contains(evmParams.ActivePrecompiles, vestingPrecompile.Address().String(),
		"expected vesting extension to be contained in active precompiles",
	)
}
