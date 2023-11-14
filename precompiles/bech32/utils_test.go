package bech32_test

import (
	"github.com/evmos/evmos/v15/precompiles/bech32"
)

// setupBech32Precompile is a helper function to set up an instance of the Bech32 precompile for
func (s *PrecompileTestSuite) setupBech32Precompile() *bech32.Precompile {
	precompile, err := bech32.NewPrecompile(6000)
	s.Require().NoError(err, "failed to create bech32 precompile")

	return precompile
}
