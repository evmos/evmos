package gov_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/vm"

	"github.com/evmos/evmos/v18/precompiles/gov"
)

type proposalTestCases struct {
	name        string
	malleate    func() []interface{}
	postCheck   func(bz []byte)
	gas         uint64
	expErr      bool
	errContains string
}

func (s *PrecompileTestSuite) TestProposal() {
	method := s.precompile.Methods[gov.ProposalMethod]

	testCases := []proposalTestCases{}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), tc.gas)
			_ = contract
			_ = method
		})
	}
}
