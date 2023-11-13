// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package erc20_test

import (
	auth "github.com/evmos/evmos/v15/precompiles/authorization"
	"github.com/evmos/evmos/v15/precompiles/erc20"
)

func (s *PrecompileTestSuite) TestIsTransaction() {
	s.SetupTest()

	// Queries
	s.Require().False(s.precompile.IsTransaction(erc20.BalanceOfMethod))
	s.Require().False(s.precompile.IsTransaction(erc20.DecimalsMethod))
	s.Require().False(s.precompile.IsTransaction(erc20.NameMethod))
	s.Require().False(s.precompile.IsTransaction(erc20.SymbolMethod))
	s.Require().False(s.precompile.IsTransaction(erc20.TotalSupplyMethod))

	// Transactions
	s.Require().True(s.precompile.IsTransaction(auth.ApproveMethod))
	s.Require().True(s.precompile.IsTransaction(auth.IncreaseAllowanceMethod))
	s.Require().True(s.precompile.IsTransaction(auth.DecreaseAllowanceMethod))
	s.Require().True(s.precompile.IsTransaction(erc20.TransferMethod))
	s.Require().True(s.precompile.IsTransaction(erc20.TransferFromMethod))
}
