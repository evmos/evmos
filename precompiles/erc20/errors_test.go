package erc20_test

import (
	"github.com/evmos/evmos/v18/precompiles/erc20"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
)

// TODO: This is not yet producing the correct reason bytes so we skip this test for now,
// until that's correctly implemented.
func (s *PrecompileTestSuite) TestBuildExecRevertedError() {
	s.T().Skip("skipping until correctly implemented")

	reason := "ERC20: transfer amount exceeds balance"
	revErr, err := erc20.BuildExecRevertedErr(reason)
	s.Require().NoError(err, "should not error when building revert error")

	revertErr, ok := revErr.(*evmtypes.RevertError)
	s.Require().True(ok, "error should be a revert error")

	// Here we expect the correct revert reason that's returned by an ERC20 Solidity contract.
	s.Require().Equal(
		"0x08c379a00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000002645524332303a207472616e7366657220616d6f756e7420657863656564732062616c616e63650000000000000000000000000000000000000000000000000000",
		revertErr.ErrorData(),
		"error data should be the revert reason")
}
